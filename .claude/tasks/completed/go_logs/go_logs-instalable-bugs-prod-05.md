---
id: go_logs-instalable-bugs-prod-05
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: [go_logs-instalable-bugs-prod-01]
estimate: 2h
actual: 0.4h
created: 2026-09-26
updated: 2026-09-26
---

# `async`: núcleo compartido entre padre e hijos, `Close()` idempotente y no-op en hijos

## Description

`async.Logger.With()` (`async/async.go:230-243`) devuelve un `&Logger{}` nuevo que comparte `buffer` y `done` con el padre pero tiene sus propios `pending atomic.Int64` y `shutdown atomic.Bool` a cero. El worker (goroutine del padre) decrementa `pending` del padre, así que el del hijo crece sin límite y el del padre se vuelve negativo: `child.Sync()` retorna sin esperar y `parent.Sync()`/`Close()` esperan el timeout completo de 5 s. `TestAsyncLogger_WithFields` falla con `-race` por eso. `Close()` hace `close(l.done)` sin `sync.Once` (pánico al repetir) y `Fatal` del hijo ignora `l.fields`. Decisión del owner: `Close()` desde un hijo es no-op. Ver plan, Objetivo 6.

## Spec

- Nuevo tipo no exportado `core` en `async/async.go` con: `syncLogger go_logs.Logger`, `buffer chan *logEntry`, `pending atomic.Int64`, `shutdown atomic.Bool`, `config Config`, `done chan struct{}`, `closeOnce sync.Once`.
- `Logger` pasa a ser `struct{ core *core; fields []go_logs.Field; isRoot bool }`. `WrapWithConfig` crea el `core`, arranca el worker y devuelve `&Logger{core: c, isRoot: true}`.
- `With()`: copia `fields` (make + append, nunca `append` sobre `l.fields`) y devuelve `&Logger{core: l.core, fields: combined, isRoot: false}`.
- `enqueue`, `worker`, `processEntry`, `Sync`: operan sobre `l.core`. `Sync()` sigue esperando a `pending == 0` con timeout; con el contador compartido ya es correcto para padre e hijos.
- `Close()`: si `!l.isRoot` devuelve `nil` sin efecto (godoc: "Close solo tiene efecto en el logger raíz devuelto por Wrap; en un hijo es no-op"). En la raíz: `closeOnce.Do(func(){ shutdown.Store(true); close(done) })` y luego la espera con timeout. Segunda llamada: repite la espera y devuelve `nil`, sin pánico.
- `Fatal(msg, fields...)`: `l.Sync()` primero para drenar; luego `l.core.syncLogger.Fatal(msg, combinados...)` con `l.fields` + `fields`.
- Sustituir el bucle de espera con `time.After(1ms)` por un `time.Ticker` (evita crear un timer por iteración). Mantener el comportamiento de no devolver error al agotar el timeout.
- **Fuga de goroutine conocida, no corregida en esta tarea** (decisión del owner): si un consumidor hace `Wrap()` y solo llama a `Sync()` (nunca a `Close()`), el worker queda vivo indefinidamente. Documentar explícitamente en el godoc de `Wrap` y `Sync` que `Close()` es la única forma de detener el worker, y que `Sync()` no lo hace. No se introduce un finalizer ni un contexto cancelable: queda para otro plan si se decide corregir.
- Tests en `async/async_test.go`: `MockLogger` de `go_logs/testing.go` o un logger sobre `bytes.Buffer` protegido; medir que `parent.Sync()` tras logs del hijo retorna en menos de 1 s con `ShutdownTimeout: 5s`.
- `gofmt -w async/async.go async/async_test.go`.

## Fuera de alcance

- Los `Close()` no idempotentes de `signal` y `otel`: planes aparte.
- Cambiar la política de descarte cuando el buffer está lleno (se conserva: drop + `IncrementDropped`).
- Hacer que `Wrap()` no arranque la goroutine hasta el primer log.
- El `SamplingWriter`/métricas de dropped del core.

## Scenarios (Gherkin)

```gherkin
Feature: Los loggers hijos asíncronos comparten el pipeline del padre

  Scenario: Sync desde un hijo espera a que su entrada se escriba
    Given un logger asíncrono envolviendo un logger síncrono sobre un buffer
    And un hijo creado con el campo request_id="req-123"
    When el hijo registra "child message" y llama a Sync
    Then el buffer contiene "request_id" y "req-123"
    And el race detector no reporta ninguna carrera

  Scenario: Sync desde el padre no espera el timeout tras logs de un hijo
    Given un logger asíncrono con ShutdownTimeout de 5 segundos
    And un hijo que ha registrado 10 mensajes
    When el padre llama a Sync
    Then Sync retorna en menos de 1 segundo
    And los 10 mensajes están en el buffer

  Scenario: Los campos del hijo no contaminan a otro hijo
    Given un logger asíncrono con un campo común app="x"
    And un hijo A con req="A" y un hijo B con req="B"
    When cada hijo registra un mensaje y se sincroniza
    Then la línea de A contiene req="A" y no "B"
    And la línea de B contiene req="B" y no "A"

  Scenario: Fatal desde un hijo incluye los campos heredados y drena antes
    Given un logger asíncrono envolviendo un logger de prueba que captura Fatal sin terminar el proceso
    And un hijo con el campo component="db"
    And un mensaje previo del hijo aún en el buffer
    When el hijo llama a Fatal con "boom"
    Then el logger de prueba recibió el mensaje previo antes de Fatal
    And la llamada a Fatal incluye el campo component="db"

Feature: Cierre del logger asíncrono

  Scenario: Close en la raíz es idempotente
    Given un logger asíncrono raíz con mensajes pendientes
    When se llama a Close dos veces
    Then ambas llamadas devuelven nil
    And no se produce ningún pánico
    And todos los mensajes pendientes se han escrito

  Scenario: Close desde un hijo no apaga el pipeline
    Given un logger asíncrono raíz y un hijo
    When el hijo llama a Close
    Then devuelve nil
    And el padre puede seguir registrando mensajes que se escriben tras Sync

  Scenario: Un hijo creado tras Close no puede volver a loguear
    Given un logger asíncrono raíz cerrado
    When se crea un hijo a partir de él y registra un mensaje
    Then no se produce ningún pánico
    And el mensaje no aparece en el buffer

  Scenario: Close en la raíz concurrente con logs de un hijo no produce pánico ni carreras
    Given un logger asíncrono raíz y un hijo escribiendo continuamente
    When se llama a Close en la raíz mientras el hijo sigue registrando
    Then no hay pánico
    And el race detector no reporta ninguna carrera
    And Close termina dentro del ShutdownTimeout configurado

  Scenario: ShutdownTimeout no positivo cae al valor por defecto
    Given WrapWithConfig con ShutdownTimeout de 0
    When se consulta el timeout efectivo
    Then es de 5 segundos

  Scenario: Registrar tras Close en la raíz se descarta sin pánico
    Given un logger asíncrono raíz cerrado
    When se registra "late"
    Then no se produce ningún pánico
    And "late" no aparece en el buffer

  Scenario: Buffer lleno descarta y contabiliza
    Given un logger asíncrono con buffer de tamaño 1 cuyo logger síncrono bloquea la escritura
    When se registran 10 mensajes
    Then el contador de dropped de las métricas es mayor que 0
    And el registro no bloquea al llamante

  Scenario: Buffer lleno descarta y contabiliza también cuando loguea un hijo
    Given un logger asíncrono con buffer de tamaño 1 cuyo logger síncrono bloquea la escritura
    And un hijo creado a partir de él
    When el hijo registra 10 mensajes
    Then el contador de dropped de las métricas es mayor que 0
    And el registro no bloquea al llamante
```

## Provides

- `async.Logger` seguro con hijos: `Sync()`/`Close()` fiables. `go test -race ./...` desde la raíz pasa a ser viable (lo exige la tarea 07).

## Definition of Done

- [x] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor
- [x] Cada escenario Gherkin tiene al menos un test (camino feliz + bordes/errores)
- [x] Todos los tests en verde: `go test -race ./...` desde la raíz (todos los paquetes)
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `gofmt -l async/` sin salida; `go vet ./...` limpio
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **godoc en el código** — `Logger`, `With`, `Close` (no-op en hijos), `Fatal`, `core`
  - [x] **Doc técnica (contexto)** — README y wiki "Optional-Modules": `defer asyncLogger.Close()` en el ejemplo (no `Sync`), nota sobre hijos
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-05.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-05: <conventional commit>`
