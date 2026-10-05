---
id: go_logs-instalable-bugs-prod-03
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: [go_logs-instalable-bugs-prod-01]
estimate: 2h
actual: 0.3h
created: 2026-09-26
updated: 2026-09-26
---

# `With()` sin memoria compartida entre hijos, copia bajo lock y nivel compartido por el árbol

## Description

`LoggerImpl.With()` (`logger_impl.go:254-273`) hace `childFields := append(l.fields, fields...)`. Cuando `l.fields` tiene capacidad sobrante, dos hijos creados del mismo padre escriben en el mismo array subyacente: el logger del request A acaba emitiendo el `request_id` del request B (reproducido en el informe, B5) y `-race` lo detecta. `With()` además lee `level`, `output`, `formatter`, `hooks` y `redactor` sin tomar `l.mu` (carrera con `SetLevel`, A3 del informe), copia `level` por valor (así que `SetLevel` del padre, que es lo que hace `http.DynamicLevelHandler`, no llega a hijos ya creados) y guarda un campo `parent` que nunca se lee. Decisión del owner: el nivel se comparte por todo el árbol de loggers. Ver plan, Objetivo 4.

## Spec

- `With()`: `childFields := make([]Field, 0, len(l.fields)+len(fields)); append(parent...); append(new...)`. Nunca `append` sobre `l.fields`.
- Copiar `output`, `formatter`, `hooks`, `redactor`, `flags`, `enableCaller`, `callerSkip`, `callerLevel`, `enableStackTrace`, `stackTraceLevel`, `metrics` bajo `l.mu.RLock()`.
- Nivel compartido: sustituir `level Level` por `level *atomic.Int32` (o `*atomic.Int64`) creado en `NewLogger` y **compartido por puntero** en `With()`. `SetLevel` hace `Store`, `GetLevel`/`getLevel` hacen `Load` sin mutex. Documentar en godoc de `With` y `SetLevel` que el nivel es compartido por padre e hijos.
- Eliminar el campo `parent` de `LoggerImpl` (no se lee en ningún sitio; comprobar con `grep -n 'parent' *.go`).
- `l.mu` sigue protegiendo `output`/`formatter`/`hooks`/`redactor` y la escritura (la separación del mutex de escritura es plan aparte).
- Actualizar `testing.go` `MockLogger.With()` **no** entra (plan aparte); solo `LoggerImpl`.
- `gofmt -w logger_impl.go` (hoy no está formateado).
- Tests en `logger_test.go`: padre con 3 campos (capacidad 4 tras append) y dos hijos; N goroutines haciendo `With` + `SetLevel` concurrente bajo `-race`; nivel propagado en ambas direcciones.

## Fuera de alcance

- Que el logger mute el slice de campos del llamante en `combineFields` (A10 del informe): plan aparte.
- Separar el mutex de configuración del de escritura y el deadlock reentrante del writer (A3/C3 del informe): plan aparte.
- `async.With()`: tarea 05.
- `MockLogger.With()` que ignora campos (`testing.go`): plan aparte.

## Scenarios (Gherkin)

```gherkin
Feature: Los loggers hijos no comparten memoria de campos

  Scenario: Dos hijos del mismo padre emiten solo sus propios campos
    Given un logger padre con los campos a=1, b=2 y c=3 (slice con capacidad sobrante)
    And un hijo A creado con el campo req="AAA"
    And un hijo B creado con el campo req="BBB"
    When el hijo A registra "handling A"
    And el hijo B registra "handling B"
    Then la línea de A contiene req="AAA" y no contiene "BBB"
    And la línea de B contiene req="BBB" y no contiene "AAA"

  Scenario: Crear un hijo de un logger sin campos previos
    Given un logger padre sin campos
    When se crea un hijo con el campo x=1
    Then el hijo emite x=1
    And el padre sigue sin campos

  Scenario: With sin campos nuevos devuelve un hijo equivalente al padre
    Given un logger padre con el campo a=1
    When se llama a With sin argumentos
    Then el hijo emite a=1 igual que el padre

  Scenario: Un hijo puede redefinir el valor de un campo heredado sin sobrescribirlo
    Given un logger padre con el campo env="prod"
    When se crea un hijo con el campo env="staging"
    And el hijo registra un mensaje
    Then la línea contiene ambas apariciones de env: "prod" y "staging"

  Scenario: Crear un hijo no altera los campos del padre
    Given un logger padre con el campo a=1
    When se crea un hijo con el campo b=2
    And el padre registra "solo padre"
    Then la línea del padre contiene a=1 y no contiene b=2

  Scenario: Creación concurrente de hijos junto a cambios de nivel
    Given un logger padre con tres campos
    When 50 goroutines crean cada una un hijo con un campo propio y registran un mensaje
    And otra goroutine cambia el nivel del padre repetidamente mientras tanto
    Then cada línea contiene únicamente el campo de la goroutine que la emitió
    And el race detector no reporta ninguna carrera

Feature: El nivel de log es compartido por todo el árbol de loggers

  Scenario: Cambiar el nivel del padre afecta a un hijo creado antes
    Given un logger padre en nivel Info
    And un hijo creado a partir de él
    When el nivel del padre se cambia a Debug
    And el hijo registra un mensaje de nivel Debug
    Then el mensaje del hijo se emite

  Scenario: Cambiar el nivel desde un hijo afecta al padre
    Given un logger padre en nivel Info
    And un hijo creado a partir de él
    When el nivel del hijo se cambia a Error
    And el padre registra un mensaje de nivel Info
    Then el mensaje del padre no se emite

  Scenario: Un nieto también comparte el nivel con el abuelo
    Given un logger raíz, un hijo creado con With y un nieto creado con With del hijo
    When el nieto cambia el nivel a Error
    Then el abuelo no emite mensajes de nivel Warn

  Scenario: Consultar el nivel desde el hijo refleja el del árbol
    Given un logger padre en nivel Warn
    When se crea un hijo y se consulta su nivel
    Then el nivel es Warn

  Scenario: El filtrado por nivel no requiere el mutex de configuración
    Given un logger en nivel Error
    When se registran 1000 mensajes de nivel Debug desde 8 goroutines
    Then no se escribe nada en la salida
    And el benchmark de filtrado se mantiene por debajo de 5 ns por operación (cifra informativa)
```

## Nota de diseño

`combineFields`/`With` **no deduplican** claves de campo repetidas entre padre e hijo: si ambos declaran la misma clave, la línea de log lleva las dos apariciones (comportamiento actual conservado, fijado por test). Deduplicar es una mejora de UX distinta, fuera de esta tarea.

## Provides

- `LoggerImpl.level` como atómico compartido: la tarea 05 (`async`) y el `http.DynamicLevelHandler` pueden confiar en que `SetLevel` sobre el logger raíz afecta a todos los hijos.
- `With()` seguro para concurrencia: la tarea 04 puede asumir que `output`/`formatter` se copian bajo lock.

## Definition of Done

- [x] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor
- [x] Cada escenario Gherkin tiene al menos un test (camino feliz + bordes/errores)
- [x] Todos los tests en verde: `go test -race .` (paquete raíz) y `go test ./...`
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `gofmt -l logger_impl.go logger_test.go` sin salida; `go vet ./...` limpio
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **godoc en el código** — `With`, `SetLevel`, `GetLevel` documentan el nivel compartido
  - [x] **Doc técnica (contexto)** — nota en README ("Child loggers") y CLAUDE.md sobre el nivel compartido
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-03.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-03: <conventional commit>`
