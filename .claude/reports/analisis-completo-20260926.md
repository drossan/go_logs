# Análisis completo de go_logs — 2026-09-26

Revisión en profundidad del paquete raíz, los submódulos (`async`, `http`, `signal`, `hooks`, `otel`, `adapters`) y el empaquetado. Todos los hallazgos marcados como verificados se reprodujeron ejecutando código (programas de prueba en módulo temporal, `go test -race`, benchmarks, compilación real de los snippets del README).

## Veredicto

Atractivo actual: **3/10**. La arquitectura v3 es sensata (interfaces `Logger`/`Formatter`/`Hook` pequeñas, inyección limpia, campos tipados, wiki bilingüe) pero hoy fallan todas las primeras impresiones: no se puede instalar, no tiene licencia, el quickstart mata el proceso, el primer snippet del README no compila, `With()` corrompe campos entre requests y el logging a fichero va a ~260 msg/s. Todo es reparable en días.

## Bloqueantes (impiden usar el paquete)

| # | Hallazgo | Dónde | Evidencia |
|---|----------|-------|-----------|
| B1 | **Módulo no instalable en v3.** `module github.com/drossan/go_logs` sin sufijo `/v3` con tags v3.0.x | `go.mod:1` | `go list -m -versions` vía proxy.golang.org solo devuelve v1.0.0…v1.2.5 |
| B2 | **Submódulos no publicables.** `require go_logs v0.0.0` + `replace => ../` (los `replace` de una dependencia se ignoran) | `async/go.mod`, `hooks/`, `http/`, `otel/`, `signal/` | `go get .../async` fallaría siempre |
| B3 | **Sin `LICENSE`.** README y wiki anuncian MIT | raíz | `git ls-files` no lo contiene |
| B4 | **`Init()` mata el proceso del consumidor.** `strconv.ParseBool("")` falla → `log.Fatalf` → `os.Exit(1)`. `save.go:10` auto-llama a `Init()` | `config.go:97-100, 115-118, 198-221, 299` | Binario con solo `go_logs.InfoLog("hola")` sin env: `exit status 1` |
| B5 | **`With()` comparte el array subyacente.** `append(l.fields, fields...)` con capacidad sobrante → dos hijos se pisan. Data race confirmada | `logger_impl.go:256` | Request A loga el `request_id` de B |
| B6 | **fsync por cada línea.** `writeEntry` llama a `output.Sync()` en cada log; `RotatingFileWriter.Sync()` hace `file.Sync()` | `logger_impl.go:355-358` | 3,8 ms/op ≈ 260 msg/s frente a "16M msg/s" del README (medido: writer crudo 9M msg/s) |
| B7 | **README no compila** en 4 patrones: `logger := go_logs.New(...)` (devuelve 2 valores, 12 sitios), `WithHook` no existe (es `WithHooks`), `hooks.NewSlackHook("tok","chan")` firma incorrecta, `logger.GetMetrics()` no está en la interfaz `Logger` | `README.md`, `CLAUDE.md`, wiki Optional-Modules | Salida real del compilador |

## Bugs de corrección en el core

### Altos

- **A1 `Sync()` es no-op** (`logger_impl.go:291-295`). La interfaz promete flush. Con `bufio.Writer` como output se pierden los últimos bytes al salir. Cuando se arregle B6, esto pasa a perder logs al cerrar.
- **A2 `With()` lee estado compartido sin lock** (`logger_impl.go:258-273`) frente a `SetLevel`. Race confirmada.
- **A3 Deadlock reentrante**: se escribe al `io.Writer` del usuario con `l.mu.Lock()` tomado (`logger_impl.go:351-358`). Un writer que loguee su propio fallo bloquea para siempre. Afecta al `errorHandler` de `MultiWriter`.
- **A4 Rotador enhanced borra ficheros ajenos**: `HasPrefix(name, prefix)` considera backup cualquier fichero que empiece por `app` (`rotating_writer_enhanced.go:326-341`). Verificado: `appointments.db` eliminado por la purga.
- **A5 Error de formato inyecta texto plano en stream JSON** (`logger_impl.go:343-348`). `Float64("ratio", NaN)` o `Any(chan)` rompen la línea para Loki/Fluentd. Claves duplicadas se colapsan por usar `map[string]interface{}`.
- **A6 Estado global v2 sin sincronizar**: `isInit`, `saveLogFile`, `logLevel`, `notifier`… solo se protegió `notificationSettings`. Race confirmada `save.go:10` vs `config.go:95`.
- **A7 `compressFile` borra el original aunque falle** (`rotating_writer_enhanced.go:294-315`). Errores de `Read`, `gzWriter.Close` y `dstFile.Close` descartados; `os.Remove(src)` incondicional.
- **A8 Caller incorrecto vía `Log()`**: skip fijo de 2 → reporta `proc.go` del runtime. `GetStackTrace(skip)` ignora `skip` (`caller.go:68`).
- **A9 Redactor sensible a mayúsculas, no anida, no toca el mensaje** (`options.go:103-113`). `Authorization`, `Password` y `Any("user", {password})` se fugan. Además muta el slice del llamante y deja `Int` con valor `"***"`.
- **A10 El logger muta el slice del llamante** (`combineFields`, `logger_impl.go:314-317`): un `[]Field` reutilizado queda enmascarado para siempre.
- **A11 Writer de `WithRotatingFile` inalcanzable**: no hay `Close()` en `Logger`; el fd vive todo el proceso.

### Medios

- ANSI escrito a ficheros: `fatih/color` decide por `os.Stdout`, no por el writer (`text_formatter.go:48-56`). Sin detección TTY/`NO_COLOR` por writer.
- Timestamps JSON con `RFC3339` sin sub-segundo (`json_formatter.go:36`).
- `Fatal()` hace `os.Exit(1)` sin `Sync()` ni hook configurable (`logger_impl.go:247-251`).
- `Default()` crea un logger nuevo por llamada cuando no hay default (`global.go:17-28`): métricas siempre 0, caller `global.go:73`.
- `maxBackups=0` conserva un backup (`rotating_writer.go:153-167`); comentario describe otro algoritmo.
- Rotación fallida deja fd cerrado en `w.file` (`rotating_writer.go:142-181`).
- `type Option interface{}` (`options.go:11`): `New(42)` compila y se ignora; structs de opción exportados sin necesidad.
- Setters de formatters sin lock (`text_formatter.go:162-179`, `json_formatter.go:121-134`).
- `LogCtx(nil, ...)` → panic (`context.go:123`).
- `SuccessLevel=25 < InfoLevel=30`: no se imprime con config por defecto y `Metrics` lo ignora (`level.go:30`, `metrics.go:103-119`).
- `slack-go/slack` + `gorilla/websocket` en el módulo raíz vía `config.go:5 → adapters` (contradice "zero deps").
- `MultiWriter.Write` solo `RLock` → writes concurrentes al mismo writer; `Close()` no idempotente.
- `WithCallerLevel(SilentLevel)` captura caller en TODOS los niveles; `WithCaller(false)` no desactiva Error+ (`options.go:205`, `logger_impl.go:177`).
- `save.go:15-17`: guardar a fichero v2 depende del ajuste de notificaciones Slack.
- `Entry.RemoveField` reordena campos (`entry.go:215`).

### Deuda

- Tres parsers de nivel (`level.go:96`, `config.go:129`, `config.go:150`) y dos juegos de constantes.
- `rotating_writer.go` y `rotating_writer_enhanced.go` duplican `openFile/Write/Sync/Close` con comportamiento divergente; `WithRotatingFile` usa el simple pese a documentar lo contrario (`options.go:262`).
- Comentarios "Phase N" obsoletos por todo el código; `api.go` `*LogCtx` con `TODO` ignorando `ctx` mientras `context.go` ya lo implementa.
- `Sampler` no conectado al `Logger`; `SamplingWriter` muestrea bytes ya formateados; `Metrics.IncrementDropped()` nunca se llama desde el core.
- `testing.go:141` `lastErr` sin usar; `MockLogger.With()` ignora campos; `testing.go` compila en producción (moverlo a `logtest/`).
- Tests: 364 pasan con `-race` pero no cubren ninguno de B4-B6 ni A1-A5. `TestLogger_Sync` solo comprueba `err == nil`.

## Submódulos

- **async** (`async.go:236-243`): `With()` crea `pending` nuevo; el worker decrementa el del padre → `child.Sync()` retorna sin esperar, `parent.Sync()/Close()` esperan el timeout completo. Race confirmada en `async_test.go:218-220`. `Close()` sin `sync.Once` → panic al repetir. `Wrap()` fuga una goroutine si solo se hace `Sync()`. `Fatal` del hijo pierde campos y no drena el buffer.
- **signal** (`signal.go:86-114`): `Stop()` doble → panic; `Register()` doble duplica goroutines. Race del test en el mock (`signal_test.go:103/122`).
- **http** (`dynamic_level.go`): auth abierto si `AuthToken==""` (fail-open en endpoint de control); comparación no constant-time (`:268`); rate limit después del auth → fuerza bruta sin freno; rate limit inoperante porque la clave incluye el puerto efímero (`:278`) y la ventana es de 1 min aunque se documenta req/s; IPv6 bloqueado por `LastIndex(":")` (`:233`); sin `MaxBytesReader` en el PUT.
- **otel** (`exporter.go:142-174`): no emite OTLP válido (falta `scopeLogs/logRecords`, `timeUnixNano`, `body` como `AnyValue`, `attributes` como array, `severityNumber`); no importa nada de `go.opentelemetry.io`; `Entry` no tiene `TraceID/SpanID`. `Close()` cierra `flushCh` muerto y panica al repetir; `Export` lanza `go e.Flush()` sin límite.
- **hooks**: `SlackHook.Run` bloqueante en el path de log; `SetLevel` sin mutex. `syslog_hook.go` sin reconexión, `writer` sin lock frente a `Close()`, y `log/syslog` hace que todo el paquete `hooks` no compile en Windows. Cobertura 24%.
- **adapters/domain**: v2 legacy duplicando `hooks/slack_hook.go`; único motivo de la dependencia Slack en el core.

## Empaquetado y DX

- **Sin CI de tests**: solo `release.yaml` y `sync-wiki.yml`. README afirma "race clean" y "~90% cobertura": medido raíz 79%, async y signal **fallan** `-race`, hooks 24%. `go test ./...` desde la raíz no toca ningún submódulo.
- **`gofmt -l`**: 16 ficheros sin formatear, incluidos `logger_impl.go`, `config.go`, `level.go`.
- **Basura trackeada**: 480 ficheros de `vendor/` (76% del repo, y `.gitignore` lo excluye), `test.log`, `IMPLEMENTATION_SUMMARY.md`, `PHASE2_SUMMARY.md`, `.claude/agents|commands|skills` de otro proyecto ("Reverence Hotels API").
- **Sin** `doc.go`, `examples/`, `CHANGELOG.md`, `CONTRIBUTING.md`. Portada vacía en pkg.go.dev.
- **README solo en español**; el wiki ya es bilingüe.
- **Versionado**: tags v1.x y v3.0.x, README dice "v3.5 (Actual)". Sin tags de submódulos. `.goreleaser.yaml` con `builds: skip` y `changelog.skip` deprecado para una librería sin binario.
- **Tabla de performance incorrecta**: fast-path real 17 ns (RWMutex) frente a 0,32 anunciado; JSON 390 ns frente a 249; fichero 260 msg/s frente a 16M. Ningún `sync.Pool`; `Error()` cuesta ~1,5 µs y 15 allocs por `runtime.Caller` implícito.
- **243 símbolos exportados** en el paquete raíz (slog ~40, zerolog ~150).

## Lo que falta para ser atractivo frente a slog / zap / zerolog

1. **`slog.Handler` adapter** (`NewSlogHandler(logger)` y `FromSlog`). Cero referencias a `log/slog` en el código. Es la puerta de entrada estándar desde Go 1.21; sin ella el paquete es una isla.
2. Zero-alloc real en el camino caliente: `Field` sin boxing de escalares, `sync.Pool` de `Entry` y buffers, encoder JSON append-based.
3. Nivel con `atomic` en lugar de `RWMutex`.
4. Sampling integrado en el `Logger`, no como writer aparte.
5. Bridge `*log.Logger` (`StdLogger()`), output por nivel, niveles custom.
6. OTel bridge real sobre `go.opentelemetry.io/otel/log`, o renombrar `otel/` y retirar la promesa.
7. Color por writer con detección TTY y `NO_COLOR`.

Nicho defendible: "baterías incluidas sobre slog" (rotación nativa, redactor, nivel dinámico por HTTP, métricas). Hoy está enterrado bajo los bloqueantes.

## Plan priorizado

1. **Instalable y legal (1 día)**: `module …/go_logs/v3`, actualizar imports, tag `v3.1.0`; submódulos con `require …/v3 v3.1.0` sin `replace` y tags `async/v3.1.0` etc.; `LICENSE` MIT.
2. **Tres bugs de producción (1-2 días)**: quitar `Sync()` por entrada y hacer que `Logger.Sync()` flushee; clonar slice en `With()`; `Init()` tolerante sin `log.Fatalf`. Test de regresión por cada uno.
3. **CI real (medio día)**: matriz Go 1.21→stable × linux/macos/windows, por módulo: `go vet`, `gofmt -l`, `go test -race -coverprofile`, `golangci-lint`. Arreglar races de async y signal. Badges.
4. **README en inglés que compile (1 día)**: corregir los 4 patrones, tabla de performance real, `examples/` compilados en CI, `README.es.md`.
5. **`slog.Handler` (2-3 días)**: mayor retorno por esfuerzo.
6. **Limpiar y adelgazar (medio día)**: `git rm --cached vendor .claude/{agents,commands,skills} test.log *_SUMMARY.md`; `adapters/` a submódulo `slack/`; `testing.go` a `logtest/`; `doc.go`, `CHANGELOG.md`, `CONTRIBUTING.md`; `gofmt -w`.
7. **Endurecer submódulos (2 días)**: `sync.Once` en todos los `Close/Stop`; `pending` compartido en `async.With`; http con token obligatorio, `ConstantTimeCompare`, `SplitHostPort`, rate limit antes del auth, `MaxBytesReader`; decidir el destino de `otel/`.

Después: locking separado config/escritura, `Close()` en `Logger`, `Fatal` con hook, redactor case-insensitive y recursivo, encoder JSON manual, `Option` tipado, un solo rotador, un solo sistema de niveles.
