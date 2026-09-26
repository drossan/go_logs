# Session log — go_logs-instalable-bugs-prod-04

> Append-only.

## 2026-09-26 23:01 — Arranque

- Rama `plan/go_logs/instalable-bugs-prod` (ya existente). `depends_on: [01]` → 01 en `completed/`. `active/` vacío.
- Tarea movida a `.claude/tasks/active/go_logs/` (`status: active`).
- Plan: Red (`spyWriter` en `logger_test.go` + tests 1:1 con el Gherkin en `flusher_test.go`) → Green (`Flusher`, `Flush()` en los 4 writers, `writeEntry` solo con `Flush`, `Logger.Sync()` real con `isIgnorableSyncErr`) → gofmt/vet/`GOOS=windows` → docs → benchmark → verificación → fact-checker → cierre.
- Follow-up de la tarea 03 (carrera de escritura entre hermanos en `writeEntry`): **no entra**. Su arreglo es un mutex de escritura compartido por el árbol, lo que el task file excluye de forma explícita ("Separar el mutex de escritura del de configuración", C3). Esta tarea solo cambia la llamada posterior a `Write`; la carrera sigue en el plan aparte.

## 2026-09-26 ~23:04 — Red

- `spyWriter` en `logger_test.go` (cuenta `Write`/`Flush`/`Sync`, errores inyectables, thread-safe). Tests 1:1 con el Gherkin en `flusher_test.go` (las comprobaciones de `Flush` se hacen por aserción de interfaz en runtime, para que el Red fuera de comportamiento y no un fallo de compilación; las aserciones de compilación `var _ Flusher = ...` se añadieron justo después).
- Rojo: 10 tests fallan. Sync por entrada 100 (se esperaba 0) y Flush 0 (se esperaba 100); `SamplingWriter` no llega al fichero; `MultiWriter`/`SamplingWriter`/`RotatingFileWriter` sin `Flush`; `bufio.Writer` vacío tras `Sync`; `Sync` del output llamado 10 veces (se esperaba 1); `ErrClosedPipe`/"disk full" devolvían nil; `MultiWriter` con "disk full" devolvía nil.
- Pasan ya con el código original (quedan como regresión): lectura inmediata del fichero, `MultiWriter`(fichero+buffer), output sin `Sync`, `EINVAL`/`ENOTTY`/`EBADF`→nil, `os.Stdout`, `MultiWriter` con EBADF, concurrencia Log+Sync, fichero real "durable".

## 2026-09-26 ~23:07 — Green / Refactor

- `writer.go`: `Flusher` e `isIgnorableSyncErr`; lista de errno en `writer_errno.go` (`!plan9`: EINVAL, ENOTTY, EBADF) y `writer_errno_plan9.go` (EINVAL). Plan 9 no define `syscall.ENOTTY`/`EBADF` y HEAD sí compilaba con `GOOS=plan9 GOARCH=amd64`: el build tag evita esa regresión (lo previsto en la Spec).
- `logger_impl.go`: `writeEntry` → `Write` + `Flush` si es `Flusher`, sin `Sync`; `Sync()` real mediante `getOutput()` (bajo `RLock`). Godoc de `Sync` en `logger.go` y `logger_impl.go`.
- `Flush()` en `RotatingFileWriter` y `EnhancedRotatingFileWriter` (bajo `w.mu`, no-op tras `Close`); `MultiWriter.Flush()`; `SamplingWriter.Flush()`/`Sync()` delegan.
- Test extra `TestIsIgnorableSyncErr` (nil, errno solo, `%w` como en `RotatingFileWriter.Sync`, join todo ignorable/mixto) y `TestRotatingWriters_FlushWritesToFileWithoutClose`.
- `gofmt -w` sobre `rotating_writer.go` y `sampling.go`: ya estaban sin formatear en HEAD (comentario godoc y alineación de un struct); se formatean porque esta tarea los toca.
- Bajo `go test`, `os.Stdout.Sync()` devuelve aquí `sync /dev/stdout: bad file descriptor`: `TestLoggerSync_Stdout` recorre de verdad la rama del error ignorado.

## 2026-09-26 ~23:10 — Verificación (go1.27.0, darwin/arm64)

| Comando | Resultado |
|---|---|
| `go test -race -count=5 -run '<tests de la tarea + TestWithMultiOutput\|TestRotatingFile_\|TestLogger_Sync>' .` | ok |
| `go test -count=1 ./...` | 8 paquetes ok (domain sin tests) |
| `go test -race -count=1 ./...` | ok salvo `async/TestAsyncLogger_WithFields` (`async_test.go:220`, fallo conocido → tarea 05) |
| `go vet ./...` | limpio |
| `GOOS=windows go build ./...` / `GOOS=windows go vet .` | ok / limpio |
| `GOOS=plan9 GOARCH=amd64 go build ./...`, `GOOS=linux GOARCH=amd64 go build ./...` | ok |
| `gofmt -l` sobre los 12 ficheros tocados | sin salida |
| `BenchmarkLoggerToRotatingFile -benchmem` (3 corridas) | **3236-3367 ns/op**, 1016 B/op, 16 allocs/op (objetivo < 10 µs) |
| Mismo benchmark sobre HEAD (worktree, 2 corridas, `-benchtime=2s`) | 3822951-3898601 ns/op (≈260 msg/s, el B6 del informe) |

## 2026-09-26 23:13 — Cierre

**Resumen**: escribir una entrada ya no hace fsync. `writeEntry` llama a `Flush()` (interfaz nueva `Flusher`) y nunca a `Sync()`. Los cuatro writers del paquete implementan `Flush()`. `Logger.Sync()` deja de ser un no-op: llama una vez al `Sync()` del output (flush + fsync) e ignora por errno `EINVAL`/`ENOTTY`/`EBADF`. El logging a fichero pasa de ~3,9 ms/op a ~3,3 µs/op.

**Decisiones + porqué**
- `MultiWriter.Sync()`/`Flush()` combinan errores con `errors.Join` en vez de devolver el último, e `isIgnorableSyncErr` solo descarta un error combinado si **todos** sus componentes son ignorables. Con "último error", `MultiWriter(spy "disk full", stdout EBADF)` devolvía EBADF → ignorado → se perdía el fallo real. Eso contradice el escenario "propaga errores reales de **cualquier** escritor" (cubierto con 3 órdenes en `TestLoggerSync_MultiWriterPropagatesRealErrors`). La Spec pedía "mismo patrón que su `Sync()`" para `Flush`, y se mantiene: los dos usan `Join`.
- Errno por plataforma en `writer_errno*.go`: plan 9 no tiene `ENOTTY`/`EBADF` y HEAD compilaba en plan 9.
- `Logger.Sync()` sigue la Spec al pie de la letra: no llama a `Flush` en outputs que solo son `Flusher` (p. ej. `bufio.Writer`), porque ya se flushean tras cada entrada.
- Follow-up de la tarea 03 (carrera de escritura entre hermanos): fuera, ver Arranque.

**Tests corridos**: ver la tabla de Verificación (suite completa ok; `-race` ok salvo el fallo conocido de async → tarea 05).

**Fact-checker**: 20/20 VERIFICADO (subagente Sonnet), 0 INCORRECTO, 0 NO VERIFICABLE.

**Docs actualizadas**: godoc de `Flusher`, `isIgnorableSyncErr`, `ignorableSyncErrnos`, cada `Flush()`, `Syncer`, `MultiWriter.Sync`, `SamplingWriter.Sync`, `Logger.Sync` (interfaz + impl) y `writeEntry`/`getOutput`. `README.md` § "Flush por entrada y `Sync()`" (nueva, dentro de Rotating File Writer). `CLAUDE.md` § RotatingFileWriter (flush por entrada, `Sync`, errno, `Join`, cifra del benchmark) y § Organización del Código (`writer.go`). Lo pide la DoD. La tabla de rendimiento de README/CLAUDE.md (16M msg/s) no se toca: la regenera la tarea 10 con esta cifra.

**Ficheros**: `writer.go`, `writer_errno.go`, `writer_errno_plan9.go` (nuevos); `logger.go`, `logger_impl.go`, `multiwriter.go`, `rotating_writer.go`, `rotating_writer_enhanced.go`, `sampling.go`; tests `flusher_test.go` (nuevo), `logger_test.go`, `benchmark_v3_test.go`; `README.md`, `CLAUDE.md`, task file, plan, este log.

**Tiempo real**: ~0,25 h (23:01–23:15, más la espera del fact-checker; horas intermedias aproximadas) frente a 3 h estimadas.

**Follow-ups**
- Preexistentes y fuera de alcance (ya fallaban en HEAD): `GOOS=js GOARCH=wasm go build ./...` falla en `signal/signal.go` (`syscall.SIGHUP`); `GOOS=windows go vet ./...` falla en `signal/signal_test.go:37` (`syscall.Kill`). Sugerencia: build tags en `signal/` (plan aparte o tarea 07 si el CI llega a cubrir Windows).
- `flusher_test.go` usa `syscall.ENOTTY`/`EBADF` y no compila en plan 9 (el código de producción sí). No hay CI en plan 9: sin acción.
- `file_writer_test.go:151` todavía dice "due to auto-sync" en un comentario; el test sigue siendo válido (flush por entrada). Se deja intacto porque la DoD pide no tocar el fichero.
- Carrera de escritura entre hermanos en `writeEntry` (heredada de la tarea 03) → plan «separar mutex de configuración y de escritura».
