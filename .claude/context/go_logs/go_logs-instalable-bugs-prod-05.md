# Session log — go_logs-instalable-bugs-prod-05

> Append-only.

## 2026-09-26 23:15 — Arranque

- Rama `plan/go_logs/instalable-bugs-prod` (ya existente). `depends_on: [01]` → 01 en `completed/`. `active/` vacío.
- Tarea movida a `.claude/tasks/active/go_logs/` (`status: active`).
- Plan: Red (tests 1:1 con los 12 escenarios en `async/async_test.go`, rojo con `-race`) → Green (`core` compartido, `With()` copia solo campos, `Close()` con `sync.Once` y no-op en hijos, `Fatal` drena y lleva campos, `time.Ticker`) → gofmt/vet → docs (godoc, README, wiki Optional-Modules) → suite completa con `-race` → fact-checker → cierre.
- `MockLogger` de `testing.go` no sirve para el escenario de `Fatal`: su `Fatal` llama a `Log`, pero `With()` devuelve el propio mock y no registra orden de llamadas `Fatal` vs `Log` de forma distinguible más allá del nivel. Se usa un logger de prueba local en el test (captura `Log`/`Fatal` en orden, sin terminar el proceso).

## 2026-09-26 ~23:17 — Red

- 13 tests nuevos en `async/async_test.go` (12 escenarios + `TestAsyncClose_ChildRepeatedCloseIsSafe`), con `safeBuffer` (buffer con mutex) y `recordingLogger` (registra `Log`/`Fatal` en orden; `gate` opcional para simular un output lento; `GetMetrics` propio).
- `go test -race ./async/`: 2 `DATA RACE`; fallan `TestAsyncLogger_WithFields` (el conocido), `ParentSyncDoesNotWaitTimeout` (0 de 10 mensajes: `parent.Sync` retorna antes de tiempo), `FatalDrainsAndIncludesFields` (Fatal antes de "previous", sin `component`), `Close_Idempotent` (`close of closed channel`), `Close_ChildIsNoOp` (pánico). Relanzando sin los que hacen pánico: `ChildCreatedAfterCloseIsDiscarded` (Sync 5,0 s y la entrada se escribía) y `ConcurrentWithChildLogging` (Close 500,1 ms > 500 ms).
- Ya pasaban con el código original (quedan como regresión): `SyncWaitsForChildEntry` (pasó por timing: `child.Sync` no esperaba, pero el worker escribió antes de la lectura; `safeBuffer` evita la carrera que sí tiene `TestAsyncLogger_WithFields` sobre `bytes.Buffer`), `FieldsDoNotLeakBetweenSiblings`, `NonPositiveShutdownTimeoutDefaults`, `LogAfterCloseIsDiscarded` y los dos de buffer lleno.

## 2026-09-26 ~23:20 — Green / Refactor

- `core` compartido por puntero; `Logger{core, fields, isRoot}`; `With()` → `combineFields` (make + append); `worker`/`processEntry`/`enqueue` son métodos de `core`; `processEntry` decrementa `pending` con `defer` tras escribir.
- `Close()`: no-op en hijos; en la raíz `closeOnce.Do(shutdown + close(done))` y luego espera. `Fatal`: `Sync()` y después `syncLogger.Fatal(l.fields + fields)`. Espera común `core.waitFor()` con `time.Ticker`; al agotar el timeout no hay error.
- Se borra el campo `mu sync.RWMutex` (no se usaba).
- **Diferencia con la Spec**: se añade `workerDone chan struct{}` a `core` (lo cierra el worker al salir) y `waitFor` también retorna cuando se cierra. Motivo: una entrada que pase el chequeo de `shutdown` justo antes de `Close()` puede llegar al canal después de que el worker haya vaciado el buffer y salido; `pending` se queda en 1 para siempre y `Close()`/`Sync()` esperarían el timeout entero (el escenario de `Close` concurrente lo dejaba al límite). Con `workerDone`, esa entrada se descarta, que es la semántica documentada para lo registrado tras `Close`.

## 2026-09-26 ~23:22 — Verificación (go1.27.0, darwin/arm64)

| Comando | Resultado |
|---|---|
| `go test -race -count=3 ./async/` | ok |
| `go test -race -count=30 -run '<tests de la tarea + WithFields + ShutdownTimeout>' ./async/` | ok (estabilidad de los tests con tiempos) |
| `go test -count=1 ./...` | 8 paquetes ok (domain sin tests) |
| `go test -race -count=1 ./...` | 8 paquetes ok, **sin excepciones** (se cierra el último rojo conocido) |
| `go vet ./...` | limpio |
| `gofmt -l async/` | sin salida |
| `GOOS=windows go build ./...` | ok |

## 2026-09-26 23:22 — Cierre

**Resumen**: `async.Logger` y sus hijos comparten ahora un único `core` (buffer, worker, `pending`, `shutdown`, `closeOnce`). `child.Sync()` espera a sus entradas y `parent.Sync()` ya no espera el timeout entero tras logs de un hijo. `Close()` es idempotente en la raíz y no-op en hijos. `Fatal` drena primero y lleva los campos del hijo. `go test -race ./...` queda en verde en todo el repo.

**Decisiones + porqué**
- `workerDone` añadido a `core` (desviación de la Spec, ver Green): evita que una entrada en carrera con `Close()` bloquee `Close()`/`Sync()` durante todo el timeout.
- `Close()` en la raíz espera por `pending == 0` **o** por la salida del worker, con timeout. Una segunda llamada solo repite esa espera (retorna enseguida porque el worker ya salió).
- Test propio `recordingLogger` en vez del `MockLogger` de `testing.go` (motivo en Arranque).
- Fuga de goroutine con `Wrap()` sin `Close()`: solo documentada en el godoc de `Wrap`, `WrapWithConfig` y `Sync`, como decidió el owner.

**Tests corridos**: ver la tabla de Verificación. Todo en verde.

**Fact-checker**: 10/10 VERIFICADO (subagente Sonnet), 0 INCORRECTO, 0 NO VERIFICABLE. Comprobó en un worktree de HEAD que `TestAsyncLogger_WithFields` fallaba con `-race` antes del cambio.

**Docs actualizadas**: godoc del paquete (el ejemplo usa `defer Close()`), `core`, `Logger`, `Wrap`, `WrapWithConfig`, `With`, `Sync`, `Close`, `Fatal` y los helpers no exportados. `README.md`, `docs/wiki/Optional-Modules.md` y `-es.md`: ejemplo con `defer asyncLogger.Close()` y nota sobre hijos, `Close()` y `Sync()` (lo pide la DoD). `CLAUDE.md` § Async: mismo ejemplo y descripción del `core` compartido (el patrón cambia, así que se documenta en el mismo cambio, como pide el HOW-TO).

**Ficheros**: `async/async.go`, `async/async_test.go`, `README.md`, `CLAUDE.md`, `docs/wiki/Optional-Modules.md`, `docs/wiki/Optional-Modules-es.md`, task file, plan, este log.

**Tiempo real**: ~0,4 h (23:15–23:22 más la espera del fact-checker; horas intermedias aproximadas) frente a 2 h estimadas.

**Follow-ups**
- El snippet async del README declara `asyncLogger :=` dos veces en el mismo bloque (Wrap y WrapWithConfig), así que no compila si se copia tal cual. Ya estaba en HEAD y no es uno de los cuatro patrones de la tarea 10, pero encaja en su pasada de snippets.
- `TestAsyncLogger_NonBlocking` exige < 10 ms para 100 logs; con `-race` en una máquina cargada podría dar un falso rojo en CI (tarea 07). Hoy pasa.
- Las cabeceras "Submódulo - Opt-in" de README y CLAUDE.md siguen describiendo `async` como submódulo: lo reescribe la tarea 10.
