# CLAUDE.md

@.claude/honesty-rules.md

Este archivo proporciona guía a Claude Code (claude.ai/code) para trabajar con el código en este repositorio.

## Resumen del Proyecto

Esta es una biblioteca de registro moderna para Go (`go_logs` v3) que proporciona logging estructurado con campos tipados, múltiples formatters (Text/JSON), child loggers, context propagation, y sistema de hooks extensible. El paquete sigue Clean Architecture con separación clara entre interfaces y implementaciones.

**Versiones**: v2 (legacy, funciones globales) y v3 (moderna, interfaz Logger). Ambas coexisten con 100% backward compatibility.

## Arquitectura v3

### Arquitectura Híbrida

El proyecto sigue una **arquitectura híbrida** con el core en el paquete principal y features opcionales como submódulos:

```
go_logs/
├── Core (paquete principal - siempre incluido)
│   ├── logger.go              # Interfaz Logger principal
│   ├── logger_impl.go         # Implementación thread-safe de Logger
│   ├── level.go               # Syslog-style Level
│   ├── field.go               # Structured Fields
│   ├── entry.go               # Log entry
│   ├── formatter.go           # Interfaz Formatter
│   ├── text_formatter.go      # TextFormatter (desarrollo)
│   ├── json_formatter.go      # JSONFormatter (producción)
│   ├── context.go             # Context propagation
│   ├── hook.go                # Hook interface
│   ├── rotating_writer.go     # RotatingFileWriter
│   ├── options.go             # Option pattern
│   ├── config.go              # Configuración
│   ├── metrics.go             # Metrics/Stats (zero overhead)
│   └── api.go                 # API v2 (backward compatible)
│
├── async/                 # Submódulo: Async logging (opt-in)
│   └── async.go           # Non-blocking logging con goroutine
│
├── http/                  # Submódulo: HTTP endpoints (opt-in)
│   └── dynamic_level.go   # Log level dinámico via HTTP
│
├── signal/                # Submódulo: Signal handlers (opt-in)
│   └── signal.go          # SIGHUP handler para rotación
│
├── domain/                # Interfaces del dominio (v2 legacy)
│   └── notification.go    # Interfaz Notifier
│
├── hooks/                 # Hooks v3
│   └── slack_hook.go      # SlackHook (v3 moderno)
│
└── slack/                 # ÚNICO submódulo: github.com/drossan/go_logs/slack/v3 (go.mod propio)
    └── notifier.go        # slack.Notifier (NewNotifier, NewNotifierFromEnv)
```

### Capas Principales

```
go_logs/
├── logger.go              # Interfaz Logger principal
├── logger_impl.go         # Implementación thread-safe de Logger
├── level.go               # Syslog-style Level (Trace=10, Debug=20, Info=30, etc.)
├── field.go               # Structured Fields (String, Int, Err, etc.)
├── entry.go               # Log entry con timestamp, level, message, fields
├── formatter.go           # Interfaz Formatter
├── text_formatter.go      # TextFormatter con colores (desarrollo)
├── json_formatter.go      # JSONFormatter (producción)
├── context.go             # Context management (WithTraceID, WithFields)
├── hook.go                # Interfaz Hook para extensibilidad
├── rotating_writer.go     # RotatingFileWriter sin dependencias
├── options.go             # Option pattern para configuración
├── config.go              # Configuración y variables de entorno
├── metrics.go             # Metrics/Stats (zero overhead)
├── api.go                 # API v2 (backward compatible)
├── domain/                # Interfaces del dominio
│   └── notification.go    # Interfaz Notifier (v2 legacy)
├── hooks/                 # Hooks v3
│   └── slack_hook.go      # SlackHook (v3 moderno)
└── slack/                 # Submódulo github.com/drossan/go_logs/slack/v3
    └── notifier.go        # slack.Notifier
```

### Patrones Arquitectónicos

1. **Dependency Inversion**: `LoggerImpl` depende de interfaces (`Formatter`, `Hook`), no de implementaciones
2. **Option Pattern**: Configuración flexible con `New(WithLevel(...), WithFormatter(...))`
3. **Interface Segregation**: Interfaces pequeñas y enfocadas (`Logger`, `Formatter`, `Hook`)
4. **Zero Dependencies**: el core solo usa `fatih/color`. `slack-go/slack` vive en el submódulo `slack/` (`go list -deps ./ | grep slack-go` vacío en la raíz)

## API v3 vs v2

### v3 API (Moderna - Recomendada)

```go
import "github.com/drossan/go_logs/v3"

// Crear logger con configuración
logger := go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
)

// Logging estructurado
logger.Info("connection established",
    go_logs.String("host", "db.example.com"),
    go_logs.Int("port", 5432),
    go_logs.Err(err),
)

// Child logger con campos heredados
reqLogger := logger.With(
    go_logs.String("request_id", "abc-123"),
    go_logs.String("user_id", "user-456"),
)
reqLogger.Info("request started") // Incluye request_id y user_id

// Context propagation
ctx = go_logs.WithTraceID(ctx, "trace-789")
logger.LogCtx(ctx, go_logs.InfoLevel, "processing request")
```

### v2 API (Legacy - Backward Compatible)

```go
import "github.com/drossan/go_logs/v3"

go_logs.Init() // Opcional, auto-inicializa
go_logs.InfoLog("message")
go_logs.ErrorLog("error message")
go_logs.Infof("formatted %s", "message")
```

## Configuración

### Variables de Entorno

```bash
# Nivel de log (syslog-style)
LOG_LEVEL=info              # trace, debug, info, warn, error, fatal, silent (default: info)

# Formato de salida
LOG_FORMAT=text             # text (dev) o json (prod)

# Archivo de log
SAVE_LOG_FILE=1
LOG_FILE_NAME=app.log
LOG_FILE_PATH=/var/log

# Rotación de archivos (v3)
LOG_MAX_SIZE=100            # MB antes de rotar
LOG_MAX_BACKUPS=5           # Archivos backup a mantener

# Slack
SLACK_TOKEN=xoxb-xxx
SLACK_CHANNEL_ID=C123456
```

Todas son opcionales. `Init()` (y la auto-inicialización de la API v2) **nunca termina el proceso**:

- Variable vacía o ausente → se usa el valor por defecto sin avisar (`LOG_LEVEL=info`, booleanos a `false`).
- Valor no válido (p. ej. `SAVE_LOG_FILE=quizas` o `LOG_LEVEL=loud`) → valor por defecto + aviso por stderr con el nombre de la variable y el valor recibido.
- Fichero de log que no se puede abrir → aviso por stderr y guardado en fichero desactivado.

### Configuración Programática (v3)

```go
logger := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithHook(go_logs.NewFuncHook(func(e *go_logs.Entry) error {
        // Custom hook logic
        return nil
    })),
    go_logs.WithCommonRedaction(), // Enmascarar password, token, etc.
)
```

## Comandos Comunes de Desarrollo

```bash
# Dependencias
GO111MODULE=on go mod tidy

# Tests
GO111MODULE=on go test ./...
GO111MODULE=on go test -v ./...
GO111MODULE=on go test -race ./...  # Race detector

# Benchmarks
GO111MODULE=on go test -bench=. -benchmem

# Cobertura
GO111MODULE=on go test -coverprofile=coverage.out ./...
GO111MODULE=on go tool cover -html=coverage.out

# Build
GO111MODULE=on go build ./...
```

### CI (`.github/workflows/ci.yml`)

En cada push y PR, un job en `ubuntu-latest` con Go `stable` ejecuta, en este orden y parando en el primer fallo, la misma secuencia que conviene lanzar en local antes de hacer push:

```bash
test -z "$(gofmt -l $(git ls-files '*.go'))"   # solo .go trackeados: excluye website/ y vendor/
go vet ./... && (cd slack && go vet ./...)
go test -race -count=1 ./...
(cd slack && go test -race -count=1 ./...)       # el ./... de la raíz no incluye slack/
GOOS=windows go build ./...                      # compilación cruzada, sin tests
```

`permissions: contents: read` y `concurrency` con `cancel-in-progress` por ref. Sin matriz de versiones, sin lint ni cobertura (fuera del plan `instalable-bugs-prod`).

## Niveles de Log (Syslog-style)

| Nivel | Valor | Uso |
|-------|-------|-----|
| Trace | 10 | Traza detallada de ejecución |
| Debug | 20 | Información de debugging |
| Info | 30 | Eventos informativos |
| Warn | 40 | Advertencias |
| Error | 50 | Errores |
| Fatal | 60 | Errores fatales (termina programa) |
| Silent | 0 | Deshabilita todos los logs |

## Componentes v3

### Logger Interface

```go
type Logger interface {
    Log(level Level, msg string, fields ...Field)
    LogCtx(ctx context.Context, level Level, msg string, fields ...Field)

    Trace(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)

    With(fields ...Field) Logger
    SetLevel(level Level)
    GetLevel() Level
    Sync() error
}
```

### Structured Fields

```go
go_logs.String("key", "value")
go_logs.Int("count", 42)
go_logs.Int64("id", 123456789)
go_logs.Float64("rate", 3.14)
go_logs.Bool("enabled", true)
go_logs.Err(err)
go_logs.Any("data", struct{...}{...})
```

### Formatters

- **TextFormatter**: Salida legible con colores ANSI para desarrollo
- **JSONFormatter**: JSON estructurado para ELK, Loki, Datadog

### Hooks

```go
// Hook personalizado
type Hook interface {
    Run(entry *Entry) error
}

// Usar hook
logger := go_logs.New(
    go_logs.WithHook(myCustomHook),
)
```

### RotatingFileWriter

Rotación por tamaño sin dependencias externas:
- Rota cuando alcanza LOG_MAX_SIZE MB
- Mantiene hasta LOG_MAX_BACKUPS archivos
- Buffered writes para performance
- **Flush por entrada, sin fsync**: `writeEntry` llama a `Flush()` si el output implementa `Flusher` (`writer.go`) y nunca a `Sync()`. Lo implementan `RotatingFileWriter`, `EnhancedRotatingFileWriter`, `MultiWriter` y `SamplingWriter` (estos dos lo propagan a sus writers)
- **`Logger.Sync()`** llama una vez al `Sync()` del output (flush + fsync en los writers de fichero) e ignora `EINVAL`/`ENOTTY`/`EBADF` por errno (`isIgnorableSyncErr`: stdout, tuberías, stdout dentro de `MultiWriter`). `MultiWriter.Sync/Flush` combinan errores con `errors.Join`, así que un error real no queda oculto por el de un terminal
- Benchmark end-to-end `BenchmarkLoggerToRotatingFile` (Logger → JSON → fichero): ~3,3 µs/op (antes, con fsync por entrada, ~3,9 ms/op)

### Redactor

Enmascara automáticamente campos sensibles:
- password, passwd, pwd
- token, secret, api_key, apikey
- authorization, auth

```go
logger := go_logs.New(
    go_logs.WithCommonRedaction(),
)
// password=***, token=*** en todos los logs
```

## Performance

| Métrica | Resultado | Target |
|---------|-----------|--------|
| Fast-path filtering | 0.32 ns/op | < 5 ns |
| Field creation | 0.34 ns/op, 0 allocs | < 10 ns |
| TextFormatter | 220.6 ns/op | < 500 ns |
| JSONFormatter | 249.3 ns/op | < 1 µs |
| RotatingFileWriter | 16M msg/sec | - |

## Migración v2 → v3

Ver **MIGRATION.md** para guía completa de migración.

Opciones:
1. **Drop-in**: Actualizar dependencia, sin cambios de código
2. **Gradual**: Mezclar v2 y v3 en misma aplicación
3. **Full**: Adoptar todas las features de v3

## Notas Importantes

- **Thread-safe**: Toda la implementación es thread-safe con mutex
- **Child loggers**: `With()` copia los campos del padre en un slice nuevo (los hijos nunca comparten memoria de campos) y copia la configuración bajo `RLock`. El **nivel es compartido por todo el árbol** (`*atomic.Int32` por puntero): `SetLevel` en padre, hijo o nieto afecta a todos, y el filtrado por nivel no toma el mutex
- **Zero allocations**: Fields y level filtering no hacen allocations
- **Backward compatible**: v2 API funciona sin cambios
- **Context support**: Extrae trace_id, span_id automáticamente
- **Clean Architecture**: Interfaces separadas de implementaciones

## Organización del Código

### Archivos Principales v3
- `logger.go`: Interfaz Logger
- `logger_impl.go`: Implementación thread-safe
- `level.go`: Syslog-style Level type
- `field.go`: Structured Fields
- `entry.go`: Log entry representation
- `formatter.go`: Interfaz Formatter
- `text_formatter.go`: TextFormatter con colores
- `json_formatter.go`: JSONFormatter
- `context.go`: Context propagation
- `hook.go`: Hook interface
- `rotating_writer.go`: RotatingFileWriter
- `writer.go`: interfaz `Flusher` e `isIgnorableSyncErr` (errno por plataforma en `writer_errno*.go`)
- `options.go`: Option pattern

### Archivos v2 (Backward Compatible)
- `api.go`: Funciones globales v2 (InfoLog, ErrorLog, etc.)
- `logs.go`: Implementación v2 legacy
- `save.go`: Guardar a archivo v2
- `config.go`: Configuración compartida
- `domain/notification.go`: Interfaz Notifier v2 (la implementa `slack.Notifier`; se registra con `SetNotifier`)

### Archivos v3
- `hooks/slack_hook.go`: SlackHook v3

### Slack (submódulo `slack/`)

- Módulo `github.com/drossan/go_logs/slack/v3` (el sufijo `/vN` va al **final** de la ruta: `.../v3/slack` no admitiría versiones v3). `go.mod` con `require github.com/drossan/go_logs/v3 v3.1.0` + `replace => ../`; tag de release `slack/v3.1.0`. Tests: `cd slack && go test -race ./...` (el `./...` de la raíz no lo incluye).
- API v2: `go_logs.SetNotifier(n domain.Notifier)` (nil desactiva, protegido por `notifierMu`). Con `NOTIFICATIONS_SLACK_ENABLED=1` y sin notificador, aviso por `warnOutput` una sola vez por proceso (`sync.Once` que `SetNotifier` no re-arma). Los errores del notificador se ignoran. `Init()` ya no lee `SLACK_TOKEN`/`SLACK_CHANNEL_ID`: lo hace `slack.NewNotifierFromEnv()`.

## Testing

- 60+ test functions
- 30+ benchmarks
- Race detector clean
- ~90% coverage en código crítico

## Features Opcionales (Submódulos)

### Metrics (Core - Siempre habilitado)

```go
logger, _ := go_logs.New()
metrics := logger.GetMetrics()

// Métricas disponibles
fmt.Printf("Total logs: %d\n", metrics.Total())
fmt.Printf("Errors: %d\n", metrics.Count(go_logs.ErrorLevel))
fmt.Printf("Dropped: %d\n", metrics.Dropped())

// Snapshot para monitoring
snapshot := metrics.Snapshot()
```

### Async Logging (Submódulo - Opt-in)

```go
import "github.com/drossan/go_logs/v3/async"

syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
asyncLogger := async.Wrap(syncLogger, 1000) // buffer size 1000
defer asyncLogger.Close() // drena y detiene el worker; Sync() no lo detiene

// Non-blocking logging
asyncLogger.Info("Server started", go_logs.Int("port", 8080))
```

Los hijos (`With()`) comparten por puntero un `core` no exportado (buffer, worker, contador `pending`, `closeOnce`), así que `Sync()` en cualquier nodo espera a todo el árbol. `Close()` solo actúa en la raíz (idempotente con `sync.Once`); en un hijo es no-op. `Fatal` drena primero y añade los campos del hijo.

### Dynamic Level via HTTP (Submódulo - Opt-in)

```go
import httplogs "github.com/drossan/go_logs/v3/http"

logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
handler := httplogs.NewDynamicLevelHandler(logger, httplogs.Config{
    Endpoint:  "/debug/level",
    AuthToken: "secret-token",
})

http.Handle("/debug/", handler)
// GET /debug/level → {"level": "INFO"}
// PUT /debug/level {"level": "debug"} → 200 OK
// GET /debug/level/metrics → métricas de logging
```

### SIGHUP Handler (Submódulo - Opt-in)

```go
import "github.com/drossan/go_logs/v3/signal"

writer, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
logger, _ := go_logs.New(go_logs.WithOutput(writer))

handler := signal.NewSIGHUPHandler(signal.WrapRotator(writer))
handler.Register()
defer handler.Stop()

// Logs rotarán al recibir SIGHUP del sistema
```
