---
id: go_logs-instalable-bugs-prod-01
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: []
estimate: 3h
actual: 0.3h
created: 2026-09-26
updated: 2026-09-26
---

# Módulo único `github.com/drossan/go_logs/v3`, LICENSE, `GetCaller` con `/vN`, build Windows, sin `vendor/`

## Description

Hoy `go get github.com/drossan/go_logs@v3` no resuelve: el `go.mod` no lleva el sufijo `/v3` que Go exige para módulos v2+, y los cinco subdirectorios `async/`, `hooks/`, `http/`, `otel/`, `signal/` tienen `go.mod` propios con `replace => ../` que ningún consumidor puede usar. Esta tarea convierte el repo en un único módulo `github.com/drossan/go_logs/v3` y deja el árbol compilando en Linux, macOS y Windows, con `LICENSE` MIT. Contexto completo en el plan (`.claude/plans/active/go_logs/instalable-bugs-prod.md`, "Contexto y problema") y en `.claude/reports/analisis-completo-20260926.md` (B1, B2, B3, A2 del design-review).

## Spec

- `go.mod` raíz: `module github.com/drossan/go_logs/v3`. ~~Mantener `go 1.21`~~ → `go 1.27.0` (decisión del owner en sesión, 2026-09-26; ver Registro de cambios del plan) y las mismas dependencias.
- Reescribir el import en los 13 `.go` que hoy importan `github.com/drossan/go_logs` o `github.com/drossan/go_logs/adapters` (lista: `config.go`, `example_formatters_test.go`, `async/*.go`, `hooks/*.go`, `http/*.go`, `otel/*.go`, `signal/*.go`). También en `README.md`, `MIGRATION.md`, `CLAUDE.md` y `docs/wiki/*.md` (solo la ruta de import / `go get`; los snippets rotos los corrige la tarea 10).
- Borrar `async/go.mod`, `hooks/go.mod`, `http/go.mod`, `otel/go.mod`, `signal/go.mod` y el `go.work` local (gitignorado). Ejecutar `go mod tidy`.
- `git rm -r --cached vendor` (ya está en `.gitignore:21`). Sin `vendor/`, `go build` resuelve desde el module cache; verificar que `go build ./...` y `go vet ./...` funcionan sin `-mod=vendor`.
- `caller.go` (`GetCaller`, líneas 30-37): al derivar `pkgName` del nombre de función, si el segmento de ruta anterior al último `/` acaba en `/vN` (N entero ≥ 2), el paquete es el segmento previo. Implementar como helper no exportado `packageFromFuncName(fnName string) (pkg, fn string)` para poder testearlo con tablas.
- `hooks/syslog_hook.go` y `hooks/syslog_hook_test.go`: añadir `//go:build !windows && !plan9` en la primera línea. `GOOS=windows go build ./...` debe compilar.
- `signal/signal_test.go`: el `mockRotator` protege su contador con `sync.Mutex` (o `atomic.Int32`); `go test -race ./signal/` en verde.
- `LICENSE` en la raíz: texto MIT estándar, `Copyright (c) 2023-2026 Daniel Rosselló`.
- No tocar `adapters/` (lo mueve la tarea 06) ni `logger_impl.go`/`config.go` más allá del import.

## Fuera de alcance

- Sacar `adapters/` (slack-go) del módulo raíz: tarea 06.
- Corregir los snippets rotos de README y wiki (`New` con dos valores, `WithHook`, etc.): tarea 10.
- Los `Close()` no idempotentes de `signal`/`otel` y el resto de hallazgos de `http`/`otel`/`hooks`: planes aparte.
- El bug de `async.With()` (`go test -race ./async/` seguirá rojo hasta la tarea 05).
- Ejecutar `git tag`: acción del owner tras el merge.

## Scenarios (Gherkin)

```gherkin
Feature: El repositorio es un único módulo Go v3 instalable

  Scenario: El módulo se identifica con el sufijo de versión mayor
    Given el repositorio en la raíz
    When se consulta la identidad del módulo con `go list -m`
    Then el resultado es "github.com/drossan/go_logs/v3"
    And no existe ningún fichero go.mod fuera de la raíz

  Scenario: Todos los paquetes forman parte del módulo raíz
    Given el módulo único
    When se listan los paquetes con `go list ./...`
    Then aparecen exactamente los ocho paquetes raíz, adapters, domain, async, hooks, http, otel y signal
    And `go build ./...` y `go vet ./...` terminan sin error sin usar vendor/

  Scenario: Un consumidor externo importa el core y un subpaquete
    Given un módulo temporal fuera del repo con `replace github.com/drossan/go_logs/v3 => <ruta-repo>`
    And un main.go que importa "github.com/drossan/go_logs/v3" y "github.com/drossan/go_logs/v3/async"
    When se compila con `go build`
    Then compila sin error

  Scenario: El módulo compila en Windows
    Given el módulo único
    When se compila con GOOS=windows `go build ./...`
    Then compila sin error
    And hooks/syslog_hook.go queda excluido por build tag en windows y plan9

  Scenario: vendor/ no está trackeado
    Given el repositorio
    When se listan los ficheros trackeados con `git ls-files vendor`
    Then la salida está vacía

  Scenario: Existe la licencia MIT
    Given el repositorio
    When se lee LICENSE en la raíz
    Then contiene "MIT License" y "Copyright (c) 2023-2026 Daniel Rosselló"

Feature: GetCaller reporta el nombre de paquete correcto con módulos /vN

  Scenario Outline: Extracción de paquete y función desde el nombre calificado
    Given el nombre de función calificado <fn_name>
    When se extraen paquete y función
    Then el paquete es <pkg> y la función es <fn>

    Examples:
      | fn_name                                                | pkg      | fn          |
      | github.com/drossan/go_logs/v3.TestGetCaller            | go_logs  | TestGetCaller |
      | github.com/drossan/go_logs/v3.(*LoggerImpl).Log        | go_logs  | (*LoggerImpl).Log |
      | github.com/drossan/go_logs/v3/async.(*Logger).With     | async    | (*Logger).With |
      | example.com/x/v2.Func                                  | x        | Func        |
      | example.com/v2tools.Func                               | v2tools  | Func        |
      | github.com/u/pkg.Func                                  | pkg      | Func        |
      | main.main                                              | main     | main        |
      | github.com/drossan/go_logs/v1.Func                     | v1       | Func        |
      | github.com/example/pkg/v10.Func                        | pkg      | Func        |

  Scenario: Un nombre calificado con tipo genérico no rompe la extracción
    Given el nombre de función calificado "github.com/drossan/go_logs/v3.List[go.shape.int].Get"
    When se extraen paquete y función
    Then el paquete es "go_logs"
    And no se produce ningún pánico

  Scenario: El test existente de GetCaller sigue reportando go_logs
    Given un logger con captura de caller activada
    When se registra un mensaje desde un test del paquete raíz
    Then el campo Package del caller es "go_logs"

Feature: La suite de signal es limpia bajo el race detector

  Scenario: Múltiples rotaciones por SIGHUP no producen carreras en el test
    Given el handler SIGHUP registrado con un rotador de prueba
    When se envían tres señales SIGHUP y se espera a que se procesen
    Then el contador de rotaciones vale 3
    And `go test -race ./signal/` termina sin "DATA RACE"
```

## Provides

- Import path `github.com/drossan/go_logs/v3` y subpaquetes `.../v3/{async,hooks,http,otel,signal,adapters,domain}` en un solo módulo. Todas las tareas siguientes editan ficheros bajo este path.
- Helper `packageFromFuncName` en `caller.go` (no exportado).
- Repo sin `vendor/` ni `go.work`: `go build ./...`, `go vet ./...`, `go test ./...` funcionan desde la raíz sobre todos los paquetes.
- `LICENSE` presente.

## Definition of Done

- [x] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor (aplica a `packageFromFuncName` y al mock de signal)
- [x] Cada escenario Gherkin tiene al menos un test o una verificación registrada en el session log (los de módulo/licencia/Windows se verifican con los comandos del escenario)
- [x] Todos los tests en verde: `go test ./...` (con `-race`: todo excepto `async`, cuyo fallo conocido queda anotado)
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `gofmt -l caller.go signal/signal_test.go hooks/syslog_hook.go` sin salida; `go vet ./...` limpio
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **godoc en el código** — `packageFromFuncName` documentado; comentario del build tag en `syslog_hook.go`
  - [x] **Doc técnica (contexto)** — import path `/v3` en README, MIGRATION.md, CLAUDE.md y wiki
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-01.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-01: <conventional commit>`
