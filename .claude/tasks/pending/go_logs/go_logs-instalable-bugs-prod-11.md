---
id: go_logs-instalable-bugs-prod-11
package: go_logs
plan: instalable-bugs-prod
status: pending
priority: 2
depends_on: [go_logs-instalable-bugs-prod-01]
estimate: 1h
actual:
created: 2026-09-26
updated: 2026-09-26
---

# `LogCtx(nil, ...)` no panica; `SuccessLevel` es alcanzable con el nivel por defecto

## Description

La pasada de `scenario-coverage` sobre el set completo encontró dos bugs reales que ni el plan original ni el informe de análisis excluían explícitamente: `LogCtx(nil, level, msg)` hace panic porque los extractores de `context.go` llaman a `ctx.Value(...)` sin comprobar `ctx == nil`; y `SuccessLevel` (25) queda por debajo de `InfoLevel` (30), el nivel por defecto tras la tarea 02, así que `logger.Log(SuccessLevel, ...)` nunca se emite con la configuración por defecto — inconsistente con la API v2, donde `success` sí se trata como visible por defecto (tarea 02). Decisión del owner: tarea breve y acotada, con test de regresión para cada caso. Ver plan, `## Registro de cambios del plan`.

## Spec

- `context.go`: en cada función extractora que recibe `ctx context.Context` (`GetTraceID`, `GetSpanID`, `ExtractFieldsFromContext` y cualquier otra que llame a `ctx.Value(...)`), añadir al inicio `if ctx == nil { ctx = context.Background() }` o el guard equivalente más simple que evite el nil-deref sin cambiar la firma pública.
- `LoggerImpl.LogCtx(ctx, level, msg, fields...)`: si `ctx == nil`, tratarlo como `context.Background()` antes de extraer campos (coherente con `context.Context` de la stdlib, que documenta ese mismo patrón en varios sitios).
- `level.go`: subir `SuccessLevel` a un valor `> InfoLevel` y `< WarnLevel` (p. ej. `35`, entre 30 y 40) de modo que sea visible con el nivel por defecto `InfoLevel`. Actualizar la tabla de niveles de `CLAUDE.md`/README si ya existe (si no, lo hace la tarea 10; aquí basta con el valor correcto y su test). Revisar `metrics.go` (`levelToIndex`): hoy devuelve -1 para `SuccessLevel`; ajustar el mapeo para que `Count(SuccessLevel)` y el `Snapshot()` lo incluyan correctamente tras el cambio de valor.
- No tocar la semántica de `Fatal`/`FatalLog` ni ningún otro nivel.

## Fuera de alcance

- El resto de bugs del informe no mencionados aquí (deadlock reentrante, redactor, NaN en JSON, etc.): planes aparte, según el `Fuera de alcance` del plan.
- Renombrar o eliminar `SuccessLevel` de la API pública.
- Cambiar el valor de cualquier otro nivel (`TraceLevel`, `DebugLevel`, `InfoLevel`, `WarnLevel`, `ErrorLevel`, `FatalLevel`).

## Scenarios (Gherkin)

```gherkin
Feature: LogCtx tolera un contexto nulo

  Scenario: LogCtx con ctx=nil no panica
    Given un logger v3 configurado con salida a un buffer
    When se llama a LogCtx con ctx=nil, nivel Info y mensaje "hola"
    Then no se produce ningún pánico
    And el buffer contiene "hola"

  Scenario: LogCtx con un contexto real sigue extrayendo sus campos
    Given un logger v3 configurado con salida a un buffer
    And un contexto con trace_id "abc-123"
    When se llama a LogCtx con ese contexto, nivel Info y mensaje "con contexto"
    Then el buffer contiene "trace_id" y "abc-123"

Feature: SuccessLevel es visible con el nivel por defecto

  Scenario: Un mensaje de nivel Success se emite con la configuración por defecto
    Given un logger v3 con el nivel por defecto (Info)
    When se registra un mensaje de nivel Success
    Then el mensaje se emite

  Scenario Outline: Orden relativo de niveles tras el ajuste
    Given los niveles Info, Success y Warn
    When se comparan entre sí
    Then <relacion>

    Examples:
      | relacion                          |
      | Info < Success                    |
      | Success < Warn                    |

  Scenario: Las métricas contabilizan correctamente los mensajes de nivel Success
    Given un logger v3 con métricas activadas
    When se registran 3 mensajes de nivel Success
    Then metrics.Count(SuccessLevel) es 3
    And el snapshot de métricas incluye la clave de Success con valor 3
```

## Provides

- `context.go` y `LogCtx` seguros ante `ctx == nil`. `SuccessLevel` visible por defecto y correctamente contabilizado en métricas.

## Definition of Done

- [ ] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor
- [ ] Cada escenario Gherkin tiene al menos un test (camino feliz + bordes/errores)
- [ ] Todos los tests en verde: `go test -race ./...`
- [ ] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [ ] Lint / format / typecheck OK: `gofmt -l context.go logger_impl.go level.go metrics.go` sin salida; `go vet ./...` limpio
- [ ] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [ ] Documentación actualizada — tres capas:
  - [ ] **godoc en el código** — `LogCtx` documenta que `ctx == nil` se trata como `context.Background()`; comentario en `level.go` sobre el nuevo valor de `SuccessLevel`
  - [ ] **Doc técnica (contexto)** — tabla de niveles de README/CLAUDE.md actualizada con el valor correcto de Success
  - [ ] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-11.md`
- [ ] Commit en la rama del plan: `go_logs-instalable-bugs-prod-11: <conventional commit>`
