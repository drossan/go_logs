# Session log — go_logs-instalable-bugs-prod-01

> Append-only.

## 2026-09-26 22:13 — Arranque

- Rama `plan/go_logs/instalable-bugs-prod` creada desde `develop` (local 1 commit por delante de `origin/develop`; `git pull` sin cambios).
- Plan movido a `.claude/plans/active/go_logs/` (`status: active`); tarea movida a `.claude/tasks/active/go_logs/` (`status: active`). `depends_on: []`.
- Plan de pasos: Red (`packageFromFuncName` + race del mock de `signal`) → Green (helper, mutex, `/v3`, imports, borrar submódulos/`go.work`, `vendor/`, build tag, `LICENSE`, docs) → verificación → fact-checker → cierre.

## 2026-09-26 22:14 — Red

- `TestPackageFromFuncName` (tabla con los 9 casos del Scenario Outline) + `TestPackageFromFuncName_Generic` en `caller_test.go` → rojo por compilación (`undefined: packageFromFuncName`).
- `signal/signal_test.go`: `TestSIGHUPHandler_MultipleRotations` pasa a leer `rotator.Calls()` y exige `== 3` → rojo por compilación (`rotator.Calls undefined`).
- Race original reproducida con el test de `HEAD`: `go test -race -run MultipleRotations` → `WARNING: DATA RACE` en `signal_test.go:122` (el `count++` del closure frente a la lectura del test).
- `GOOS=windows go build ./...` antes del build tag → falla con `undefined: syslog.Dial`/`syslog.New`/`syslog.LOG_*` en `hooks/syslog_hook.go`.

## 2026-09-26 22:15-22:21 — Green / Refactor

- `caller.go`: `packageFromFuncName` + `isMajorVersionSuffix` (no exportados, con godoc). `GetCaller` los usa. Los genéricos se tratan cortando la ruta en el primer `[`.
- `signal/signal_test.go`: `mockRotator` cuenta `Rotate` bajo `sync.Mutex` y expone `Calls()`.
- `go.mod` → `module github.com/drossan/go_logs/v3`. Imports reescritos en 13 `.go` (`config.go`, `example_formatters_test.go`, `async/`, `hooks/`, `http/`, `otel/`, `signal/`).
- Borrados `async|hooks|http|otel|signal/go.mod` (`git rm`) y `go.work` (no trackeado). `git rm -r --cached vendor` (480 ficheros) y además se borró `vendor/` del disco: con un `vendor/` presente Go lo usa automáticamente y no se habría verificado el build sin él (se puede regenerar con `go mod vendor`).
- `hooks/syslog_hook.go`: `//go:build !windows && !plan9` + comentario del motivo. **No existe** `hooks/syslog_hook_test.go` (la Spec lo menciona): no hay nada que etiquetar.
- `LICENSE` MIT, `Copyright (c) 2023-2026 Daniel Rosselló`.
- Docs: import path y `go get` a `/v3` en `README.md`, `MIGRATION.md`, `CLAUDE.md` y `docs/wiki/*.md`. Criterio: se cambian los import paths entre comillas, `go get` (a `/v3@latest`, porque `v3.0.x` nunca será válida bajo `/v3`), `require` (a `/v3 v3.1.0`) y los enlaces de pkg.go.dev (que son por módulo). Las URLs de GitHub y goreportcard (que son por repo) se quedan.
- **Decisión del owner a mitad de sesión**: directiva `go 1.21` → `go 1.27.0`. Se avisó de que la directiva es un mínimo que se impone a todo consumidor; el owner lo aceptó ("el principal consumidor soy yo"). Se comprobó que `go1.27.0` existe (`GOTOOLCHAIN=go1.27.0 go version` → `go1.27.0 darwin/arm64`); el toolchain local es 1.26.0 y con `GOTOOLCHAIN=auto` usa 1.27.0. Requisitos actualizados en la wiki (`Getting-Started*`, `Installation*`: `Go 1.27+`, `go-version: '1.27'`, `golang:1.27-alpine`), el HOW-TO (bloque Stack) y el plan (Recursos + Registro de cambios). Spec de la tarea anotada.
- gofmt: `async/async_test.go`, `config.go`, `otel/exporter.go` y `signal/signal.go` ya estaban sin formatear en `HEAD` y los toqué por el import; se pasaron por `gofmt -w` (solo espaciado y comentarios) para cumplir la regla de gofmt limpio en los ficheros tocados.

## 2026-09-26 22:22 — Verificación (todo con go1.27.0)

| Comando | Resultado |
|---|---|
| `go list -m` | `github.com/drossan/go_logs/v3` |
| `go list ./...` | 8 paquetes: raíz, adapters, async, domain, hooks, http, otel, signal |
| `find . -name go.mod` | solo `./go.mod` |
| `git ls-files vendor \| wc -l` | 0 |
| `go build ./...`, `go vet ./...` (sin vendor/) | OK |
| `GOOS=windows go build ./...` / `GOOS=plan9 GOARCH=amd64 go build ./...` | OK / OK |
| `go test -count=1 ./...` | 7 paquetes con tests `ok` (domain no tiene tests) |
| `go test -race -count=1 ./...` | todo `ok` salvo `async`: `TestAsyncLogger_WithFields` (DATA RACE en `async_test.go:220` + timeout de 10 s) — fallo conocido, fuera de alcance, lo resuelve la tarea 05 |
| `go test -race -count=3 ./signal/` | ok |
| e2e: módulo en `/tmp/e2e01` con `replace` que importa `v3` y `v3/async` | `go build` OK; ejecución imprime `INFO hola e2e`, exit 0 |
| `git diff --name-only \| grep .go$ \| xargs gofmt -l` | sin salida |

Cobertura de escenarios: módulo /v3, paquetes, e2e, Windows, vendor y LICENSE se verifican con los comandos de la tabla; Scenario Outline y genéricos → `TestPackageFromFuncName*`; "El test existente de GetCaller sigue reportando go_logs" → `TestGetCaller` en verde con el módulo en `/v3`; signal → `TestSIGHUPHandler_MultipleRotations` (`== 3`) con `-race`.

## 2026-09-26 22:26 — Cierre

**Resumen**: el repo es un único módulo `github.com/drossan/go_logs/v3` (8 paquetes), sin submódulos, sin `go.work` ni `vendor/` trackeado, compila en Windows/plan9, con `LICENSE` MIT, `GetCaller` correcto bajo `/vN` y el mock de `signal` sin race. Directiva `go 1.27.0` por decisión del owner.

**Decisiones + porqué**
- `packageFromFuncName` corta en el primer `[` antes de buscar el último `/`: los tipos genéricos instanciados pueden contener rutas con `/`.
- `vN` solo cuenta como sufijo de versión mayor si N ≥ 2 y no empieza por cero (`v1`, `v01` y `v2tools` son nombres de paquete).
- `go get … /v3@latest` en la documentación: `v3.0.x` nunca resolverá bajo `/v3`, y `v3.1.0` aún no está tagueado.
- `go 1.27.0`: decisión explícita del owner; el riesgo (mínimo que se impone a los consumidores) se avisó y quedó registrado en el plan.
- gofmt en 4 ficheros heredados sin formato: lo exige la regla de gofmt limpio en los ficheros tocados; son cambios solo de espaciado y comentarios.

**Tests**: ver la tabla de verificación. `go test ./...` en verde; `go test -race ./...` en verde salvo `async` (fallo conocido → tarea 05).

**Fact-checker**: 16/16 VERIFICADO (subagente Sonnet), 0 INCORRECTO, 0 NO VERIFICABLE.

**Docs actualizadas**: godoc de `packageFromFuncName`/`isMajorVersionSuffix`, comentario del build tag en `syslog_hook.go`; import path `/v3` en `README.md`, `MIGRATION.md`, `CLAUDE.md`, `docs/wiki/*.md` (lo pide la Spec); requisito Go 1.27 en la wiki y el HOW-TO (decisión del owner); plan (Recursos + Registro de cambios).

**Ficheros**: `go.mod`, `caller.go`, `caller_test.go`, `signal/signal_test.go`, `hooks/syslog_hook.go`, `LICENSE`, 13 `.go` con import reescrito, 5 `go.mod` borrados, `vendor/` fuera del índice, docs. Commit `bf977ac`: `go_logs-instalable-bugs-prod-01: refactor!: single v3 module, LICENSE, /vN-aware GetCaller, windows build`.

**Tiempo real**: ~0.3 h (22:13–22:28) frente a 3 h estimadas.

**Follow-ups**
- `CLAUDE.md` todavía describe la "Arquitectura Híbrida" de submódulos y la tabla de rendimiento → tarea 10 (ya está en su alcance).
- `.claude/agents|skills` (de otro proyecto) mencionan Go 1.21 → no se tocan (el plan deja fuera la limpieza de `.claude/` ajeno).
- La tarea 07 (CI) debe usar un Go ≥ 1.27 ("stable" hoy lo cumple).
- Race de `async` → tarea 05.
