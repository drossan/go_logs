# Session log — go_logs-instalable-bugs-prod-03

> Append-only.

## 2026-09-26 22:49 — Arranque

- Rama `plan/go_logs/instalable-bugs-prod` (ya existente). `depends_on: [01]` → 01 en `completed/`. `active/` vacío.
- Tarea movida a `.claude/tasks/active/go_logs/` (`status: active`).
- Plan: Red (`logger_test.go`, 1:1 con el Gherkin; carreras vistas con `-race`) → Green (`With()` con `make`+`append` bajo `RLock`, `level *atomic.Int32` compartido, sin `parent`) → gofmt/vet → docs → verificación → fact-checker → cierre.

## 2026-09-26 22:50 — Red

- 11 tests en `logger_test.go` (1:1 con el Gherkin) + helpers `lockedBuffer`, `parentWithSpareCapacity` (fuerza `fields` len 3 / cap 8 para no depender del crecimiento de `append`) y `lineWith`.
- Rojo: fallan 6 — `SiblingsDoNotShareFields` (línea de A: `a=1 b=2 c=3 req=BBB`, el bug B5 reproducido), `ConcurrentChildrenAndSetLevel` (goroutines emitiendo el `g` de otra), los 3 `SharedLevel_*` de propagación y `LevelFilteringDoesNotTakeMutex` (bloqueado en `l.mu`). `go test -race -run TestLogger_With_` → 4 `WARNING: DATA RACE` (`logger_impl.go:256` append, `:259` lectura de `level` frente a `:280` `SetLevel`).
- Pasan ya con el código original (quedan como regresión): `FromParentWithoutFields`, `NoArgsIsEquivalentToParent`, `DuplicateKeyKeepsBoth`, `DoesNotAlterParent`, `SharedLevel_ChildGetLevel`.

## 2026-09-26 22:52 — Green / Refactor

- `logger_impl.go`: `level *atomic.Int32` (creado en `NewLogger`, `Store` en `LevelOption`/`SetLevel`, `Load` en `getLevel`/`GetLevel`, sin mutex); `With()` con `make`+2×`append` y copia del resto de estado bajo `l.mu.RLock()`, compartiendo `level` y `metrics` por puntero; campo `parent` eliminado; `gofmt -w`.
- Godoc de `With`/`SetLevel`/`GetLevel` en `logger_impl.go` y en la interfaz de `logger.go`.

## 2026-09-26 23:00 — Verificación (go1.27.0)

| Comando | Resultado |
|---|---|
| `go test -race -count=1 .` | ok |
| `go test -race -count=5 -run 'TestLogger_With_\|TestLogger_SharedLevel\|TestLogger_LevelFiltering' .` | ok |
| `go test -count=1 ./...` | 8 paquetes ok (domain sin tests) |
| `go test -race -count=1 ./...` | ok salvo `async/TestAsyncLogger_WithFields` (`async_test.go:220`, fallo conocido → tarea 05) |
| `go vet ./...` | limpio |
| `gofmt -l logger.go logger_impl.go logger_test.go` | sin salida |
| `BenchmarkLogger_FastPathFiltering` (3 corridas) | antes 14,5-18,3 ns/op → después 2,2-3,7 ns/op, 0 allocs (objetivo < 5 ns) |

## 2026-09-26 23:05 — Cierre

**Resumen**: `With()` ya no comparte el array de campos entre hermanos, copia la configuración bajo `RLock` y el nivel es un `*atomic.Int32` compartido por todo el árbol (padre ↔ hijo ↔ nieto). El filtrado por nivel ya no toma el mutex. Campo `parent` eliminado.

**Decisiones + porqué**
- La precondición de «capacidad sobrante» se fuerza en el test asignando `fields` con cap 8, en vez de confiar en cómo crece `append`: con el arreglo, `With()` crea slices de capacidad exacta y el test seguiría ejercitando el caso peligroso.
- «El filtrado no requiere el mutex» se comprueba con un assert real (los `Debug` filtrados terminan con `l.mu` bloqueado por el test), no solo con el benchmark, que es informativo.
- En el test concurrente los hijos escriben en un `lockedBuffer`: cada hijo tiene su propio `l.mu`, así que un `bytes.Buffer` compartido tendría una carrera ajena a esta tarea (ver follow-ups).
- El `enableCaller`/`callerLevel`/… que lee `Log()` sin lock no se toca: se fija al crear el logger y no hay setters.

**Fact-checker**: 11/11 VERIFICADO (subagente Sonnet), 0 INCORRECTO, 0 NO VERIFICABLE. Matiz en la cifra base del benchmark (una corrida a 18,34 ns/op): corregido el rango en este log.

**Docs actualizadas**: godoc (`logger.go`, `logger_impl.go`: `With`, `SetLevel`, `GetLevel`, `getLevel`, campo `level`); `README.md` § Child Loggers (copia de campos, claves duplicadas, nivel compartido + ejemplo); `CLAUDE.md` § Notas Importantes. Lo pide la DoD. El snippet `logger := go_logs.New(...)` del README sigue roto a propósito: lo corrige la tarea 10.

**Ficheros**: `logger_impl.go`, `logger_test.go`, `logger.go`, `README.md`, `CLAUDE.md`, task file, plan, este log. Commit `464c835`: `go_logs-instalable-bugs-prod-03: fix(logger): With() copies fields, config under RLock, level shared by logger tree`.

**Tiempo real**: ~0,3 h (22:49–23:05) frente a 2 h estimadas.

**Follow-ups**
- **Carrera preexistente (reproducida con `-race`, confirmada también en HEAD por el fact-checker)**: hermanos que loguean a la vez en un output que no es thread-safe (`bytes.Buffer`, `bufio.Writer`) chocan en `writeEntry` (`l.output.Write`), porque cada hijo tiene su propio `l.mu`. Arreglo natural: un mutex de escritura compartido por puntero en el árbol (como `level`). Encaja con el plan aparte «separar mutex de configuración y de escritura» (A3/C3); a considerar en la tarea 04 si toca `writeEntry`.
- La tabla de rendimiento de `CLAUDE.md` (0,32 ns fast-path) no casa con lo medido (~2,3 ns) → tarea 10.
