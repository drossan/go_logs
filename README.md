# Go_Logs v3 - Biblioteca de Logging Moderna para Go

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/drossan/go_logs)
[![Go Report Card](https://goreportcard.com/badge/github.com/drossan/go_logs)](https://goreportcard.com/report/github.com/drossan/go_logs)

Biblioteca de logging estructurado para Go con campos tipados, múltiples formatters (Text/JSON), child loggers, context propagation, sistema de hooks extensible, y rotación de archivos. **100% backward compatible con v2**.

## Características Principales

### v3 Features (Modernas)

- Structured logging con campos tipados (`String`, `Int`, `Err`, etc.)
- Child loggers con propagación de campos (`logger.With(...)`)
- Context propagation real (`WithTraceID`, `WithSpanID`)
- Dual formatters: TextFormatter (dev) + JSONFormatter (prod)
- Sistema de hooks extensible (Slack, Sentry, métricas)
- RotatingFileWriter sin dependencias externas
- Redactor automático de datos sensibles (password, token, etc.)
- Interfaz Logger inyectable para testing

### v2 Features (Legacy - Backward Compatible)

- Múltiples niveles de registro: Trace, Debug, Info, Warn, Error, Fatal, Success
- Salida con colores ANSI para terminal
- Archivo persistente con buffering
- Notificaciones Slack configurables
- Thread-safe con mutex
- Cero dependencias externas (solo color y slack)

## Instalación

```bash
go get github.com/drossan/go_logs@v3
```

## Inicio Rápido

### v3 API (Recomendada)

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Crear logger con configuración
    logger := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewTextFormatter()),
        go_logs.WithOutput(os.Stdout),
    )

    // Logging estructurado con campos tipados
    logger.Info("connection established",
        go_logs.String("host", "db.example.com"),
        go_logs.Int("port", 5432),
    )

    // Child logger con campos heredados
    reqLogger := logger.With(
        go_logs.String("request_id", "abc-123"),
        go_logs.String("user_id", "user-456"),
    )
    reqLogger.Info("request started") // Incluye request_id y user_id
}
```

### v2 API (Legacy - Drop-in Compatible)

```go
package main

import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()         // Opcional, auto-inicializa
    defer go_logs.Close()

    go_logs.InfoLog("Servidor iniciado")
    go_logs.ErrorLog("Error de conexión")
    go_logs.Infof("Puerto: %d", 8080)
}
```

## Niveles de Log (Syslog-style)

| Nivel | Valor | Color | Uso |
|-------|-------|-------|-----|
| Trace | 10 | Cyan | Traza detallada de ejecución |
| Debug | 20 | HiBlue | Información de debugging |
| Info | 30 | Yellow | Eventos informativos |
| Warn | 40 | HiYellow | Advertencias |
| Error | 50 | Red | Errores |
| Fatal | 60 | HiRed | Errores fatales (termina programa) |
| Success | - | Green | Operaciones exitosas (v2 legacy) |
| Silent | 0 | - | Deshabilita todos los logs |

## Formatters

### TextFormatter (Desarrollo)

Salida legible con colores ANSI:

```
[2026/02/28 17:30:00] INFO connection established host=db.example.com port=5432
```

### JSONFormatter (Producción)

JSON estructurado para ELK, Loki, Datadog:

```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"connection established","fields":{"host":"db.example.com","port":5432}}
```

```go
// Usar JSONFormatter
logger := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

## Structured Fields

```go
logger.Info("user action",
    go_logs.String("action", "login"),
    go_logs.String("username", "john"),
    go_logs.Int("duration_ms", 150),
    go_logs.Bool("success", true),
    go_logs.Float64("rate", 3.14),
    go_logs.Err(err),
    go_logs.Any("metadata", map[string]string{"ip": "192.168.1.1"}),
)
```

## Child Loggers

Crea loggers con campos pre-inyectados que se heredan:

```go
// Logger base
logger := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

// Child logger con campos de request
reqLogger := logger.With(
    go_logs.String("request_id", "abc-123"),
    go_logs.String("user_id", "user-456"),
)

// Todos estos logs incluyen request_id y user_id
reqLogger.Info("request started")
reqLogger.Info("processing data")
reqLogger.Error("validation failed")
```

## Caller Info

Muestra automáticamente archivo, línea y función en cada log:

```go
logger, _ := go_logs.New(
    go_logs.WithCaller(true),
)

logger.Info("user logged in")
// Output: [2026/02/28 17:30:00] INFO main.go:42 handleLogin user logged in
```

Output en JSON:
```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","caller":"main.go:42","caller_func":"handleLogin","message":"user logged in"}
```

## Stack Traces

Captura automática de stack trace en errores:

```go
logger, _ := go_logs.New(
    go_logs.WithStackTrace(true),
    go_logs.WithStackTraceLevel(go_logs.ErrorLevel), // Solo en Error+
)

logger.Error("database connection failed")
// Output incluye stack trace completo
```

Configuración de niveles:
- `WarnLevel`: Stack trace para Warn, Error, Fatal
- `ErrorLevel`: Stack trace para Error, Fatal (default)
- `FatalLevel`: Stack trace solo para Fatal

## Context Propagation

Extrae automáticamente trace_id y span_id del contexto:

```go
import (
    "context"
    "github.com/drossan/go_logs"
)

func handler(ctx context.Context) {
    // Inyectar trace ID en contexto
    ctx = go_logs.WithTraceID(ctx, "trace-789")
    ctx = go_logs.WithSpanID(ctx, "span-101")

    // El logger extrae automáticamente del contexto
    logger := go_logs.New()
    logger.LogCtx(ctx, go_logs.InfoLevel, "processing request")
    // Output incluye: trace_id=trace-789 span_id=span-101
}
```

## Hooks System

Sistema extensible para enviar logs a múltiples destinos:

```go
// Hook personalizado
func myHook(entry *go_logs.Entry) error {
    // Enviar a Sentry, métricas, etc.
    return nil
}

logger := go_logs.New(
    go_logs.WithHook(go_logs.NewFuncHook(myHook)),
)
```

### Slack Hook

```go
import "github.com/drossan/go_logs/hooks"

slackHook := hooks.NewSlackHook("xoxb-token", "C123456")
logger := go_logs.New(
    go_logs.WithHook(slackHook),
)
```

### OpenTelemetry / OTLP Hook

Envía logs a un collector OpenTelemetry:

```go
import "github.com/drossan/go_logs/otel"

// Hook simple
otelHook := otel.NewOTLPHook("http://localhost:4318/v1/logs")
defer otelHook.Close()

logger := go_logs.New(
    go_logs.WithHook(otelHook),
)

// Con configuración avanzada
exporter := otel.NewOTLPExporterWithConfig(otel.OTLPConfig{
    Endpoint:      "http://otel-collector:4318/v1/logs",
    Headers:       map[string]string{"Authorization": "Bearer token"},
    MaxPending:    100,
    FlushInterval: 5 * time.Second,
    Timeout:       30 * time.Second,
})
hook := otel.NewOTLPHookWithExporter(exporter, go_logs.InfoLevel)
```

### Syslog Hook

Envía logs al syslog local o remoto:

```go
import "github.com/drossan/go_logs/hooks"

// Syslog local
syslogHook, _ := hooks.NewSyslogHook("myapp")
defer syslogHook.Close()

// Syslog remoto (RFC5424)
remoteHook, _ := hooks.NewNetworkSyslogHook("tcp", "logs.example.com:514", "myapp")

// Con formateador RFC5424
syslogHook.SetFormatter(hooks.RFC5424Formatter("myapp"))

logger := go_logs.New(
    go_logs.WithHook(syslogHook),
)
```

## Rotating File Writer

Rotación por tamaño sin dependencias externas:

```go
logger := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    // 100MB max, 5 backups
)
```

O via variables de entorno:

```bash
LOG_MAX_SIZE=100      # MB antes de rotar
LOG_MAX_BACKUPS=5     # Archivos backup a mantener
```

### Rotación por Tiempo (Daily/Hourly)

```go
// Rotación diaria con compresión
logger, _ := go_logs.New(
    go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
        Filename:     "/var/log/app.log",
        MaxSizeMB:    100,
        MaxBackups:   30,           // Mantener 30 días
        RotationType: go_logs.RotateDaily,
        Compress:     true,         // Gzip archivos antiguos
        MaxAge:       30,           // Eliminar después de 30 días
    }),
)
```

Tipos de rotación:
- `RotateSize`: Por tamaño (default)
- `RotateDaily`: Cada día a medianoche
- `RotateHourly`: Cada hora

### Múltiples Outputs (File + Console)

```go
file, _ := go_logs.NewRotatingFileWriter("app.log", 100, 5)
multi := go_logs.NewMultiWriter(file, os.Stdout)

logger, _ := go_logs.New(
    go_logs.WithOutput(multi),
)
// Logs van a archivo Y consola simultáneamente
```

## Redactor de Datos Sensibles

Enmascara automáticamente campos sensibles:

```go
logger := go_logs.New(
    go_logs.WithCommonRedaction(),
)

logger.Info("user login",
    go_logs.String("username", "john"),
    go_logs.String("password", "secret123"), // Output: password=***
)
```

Campos enmascarados por defecto: `password`, `passwd`, `pwd`, `token`, `secret`, `api_key`, `apikey`, `authorization`, `auth`.

## Submódulos Opcionales (Arquitectura Híbrida)

### Metrics (Core - Siempre habilitado)

Estadísticas de logging con zero overhead:

```go
logger, _ := go_logs.New()
metrics := logger.GetMetrics()

// Métricas disponibles
fmt.Printf("Total logs: %d\n", metrics.Total())
fmt.Printf("Errors: %d\n", metrics.Count(go_logs.ErrorLevel))
fmt.Printf("Dropped: %d\n", metrics.Dropped())

// Snapshot para monitoring
snapshot := metrics.Snapshot()
// snapshot.Total, snapshot.ByLevel[InfoLevel], etc.

// Reset para testing
metrics.Reset()
```

### Async Logging (Submódulo - Opt-in)

Logging non-blocking para alta carga:

```go
import "github.com/drossan/go_logs/async"

// Crear logger síncrono base
syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

// Envolver con async (buffer size 1000)
asyncLogger := async.Wrap(syncLogger, 1000)
defer asyncLogger.Sync()

// Non-blocking logging
asyncLogger.Info("Server started", go_logs.Int("port", 8080))

// Configuración avanzada
asyncLogger := async.WrapWithConfig(syncLogger, async.Config{
    BufferSize:      10000,
    ShutdownTimeout: 10 * time.Second,
})

// Child logger con campos
childLogger := asyncLogger.With(go_logs.String("request_id", "abc-123"))
```

### Dynamic Level via HTTP (Submódulo - Opt-in)

Cambiar nivel de log en runtime via HTTP:

```go
import httplogs "github.com/drossan/go_logs/http"

logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

handler := httplogs.NewDynamicLevelHandler(logger, httplogs.Config{
    Endpoint:  "/debug/level",
    AuthToken: "secret-token",      // Bearer token auth
    RateLimit: 10,                  // 10 req/seg
    AllowedIPs: []string{"10.0.0.0/8", "192.168.1.100"},
})

http.Handle("/debug/", handler)

// Endpoints disponibles:
// GET /debug/level → {"level": "INFO", "timestamp": "..."}
// PUT /debug/level {"level": "debug"} → 200 OK
// GET /debug/level/metrics → {"total": 1234, "by_level": {...}, "dropped": 0}
```

### SIGHUP Handler (Submódulo - Opt-in)

Rotación de logs al recibir señal del sistema:

```go
import "github.com/drossan/go_logs/signal"

writer, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
logger, _ := go_logs.New(go_logs.WithOutput(writer))

// Crear handler para SIGHUP
handler := signal.NewSIGHUPHandler(signal.WrapRotator(writer))
handler.Register()
defer handler.Stop()

// Logs rotarán al recibir SIGHUP del sistema (ej: logrotate)
// También funciona con señales personalizadas:
handler := signal.NewSIGHUPHandler(rotator, syscall.SIGUSR1, syscall.SIGUSR2)
```

## Configuración

### Variables de Entorno

```bash
# Nivel de log (syslog-style)
LOG_LEVEL=info              # trace, debug, info, warn, error, fatal, silent

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

# Notificaciones por nivel (v2 legacy)
NOTIFICATION_FATAL_LOG=1
NOTIFICATION_ERROR_LOG=1
NOTIFICATION_WARNING_LOG=1
NOTIFICATION_INFO_LOG=1
NOTIFICATION_SUCCESS_LOG=1
```

### Configuración Programática (v3)

```go
logger := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithHook(myCustomHook),
    go_logs.WithCommonRedaction(),
)
```

## API v3 Completa

### Logger Interface

```go
type Logger interface {
    // Logging básico
    Log(level Level, msg string, fields ...Field)
    LogCtx(ctx context.Context, level Level, msg string, fields ...Field)

    // Convenience methods
    Trace(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)

    // Child logger
    With(fields ...Field) Logger

    // Control
    SetLevel(level Level)
    GetLevel() Level
    Sync() error
}
```

### Field Helpers

```go
go_logs.String("key", "value")
go_logs.Int("count", 42)
go_logs.Int64("id", 123456789)
go_logs.Float64("rate", 3.14)
go_logs.Bool("enabled", true)
go_logs.Err(err)
go_logs.Any("data", struct{...}{...})
```

### Options

```go
// Nivel de log
go_logs.WithLevel(go_logs.InfoLevel)

// Salida
go_logs.WithOutput(os.Stdout)
go_logs.WithFormatter(go_logs.NewJSONFormatter())

// Caller info (archivo:línea:función)
go_logs.WithCaller(true)           // Habilitar caller info
go_logs.WithCallerSkip(2)          // Ajustar frames a saltar

// Stack traces
go_logs.WithStackTrace(true)                    // Habilitar stack traces
go_logs.WithStackTraceLevel(go_logs.ErrorLevel) // Nivel mínimo

// Hooks y redacción
go_logs.WithHooks(myHook)
go_logs.WithCommonRedaction()

// Rotación de archivos
go_logs.WithRotatingFile("/var/log/app.log", 100, 5)
```

## API v2 (Legacy)

### Funciones Básicas

```go
go_logs.FatalLog(message string)
go_logs.ErrorLog(message string)
go_logs.WarningLog(message string)
go_logs.InfoLog(message string)
go_logs.SuccessLog(message string)
```

### Funciones con Formato

```go
go_logs.Fatalf(format string, args ...interface{})
go_logs.Errorf(format string, args ...interface{})
go_logs.Warningf(format string, args ...interface{})
go_logs.Infof(format string, args ...interface{})
go_logs.Successf(format string, args ...interface{})
```

### Funciones con Contexto

```go
go_logs.ErrorLogCtx(ctx context.Context, message string)
go_logs.WarningLogCtx(ctx context.Context, message string)
go_logs.InfoLogCtx(ctx context.Context, message string)
go_logs.SuccessLogCtx(ctx context.Context, message string)
```

## Performance

| Métrica | Resultado | Target |
|---------|-----------|--------|
| Fast-path filtering | 0.32 ns/op | < 5 ns |
| Field creation | 0.34 ns/op, 0 allocs | < 10 ns |
| TextFormatter | 220.6 ns/op | < 500 ns |
| JSONFormatter | 249.3 ns/op | < 1 µs |
| RotatingFileWriter | 16M msg/sec | - |

## Sampling / Rate Limiting

Controla el volumen de logs en sistemas de alta carga:

```go
// Permitir máximo 100 logs por minuto
sampler := go_logs.NewSampler(100, time.Minute)

if sampler.Allow() {
    logger.Info("high frequency event")
}

// SamplingWriter para envolver cualquier writer
buf := go_logs.NewCaptureBuffer()
sw := go_logs.NewSamplingWriter(buf, 100, time.Minute)

// Con callback para monitorear drops
sw := go_logs.NewSamplingWriterWithCallback(w, 100, time.Minute, func(dropped int) {
    metrics.LogsDropped.Add(dropped)
})
```

## Global Logger

Acceso global sin pasar logger por parámetros:

```go
// Configurar logger global
logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
go_logs.SetDefault(logger)

// Usar desde cualquier parte del código
go_logs.GlobalInfo("application started")
go_logs.GlobalError("something failed", go_logs.Err(err))
go_logs.GlobalWarn("deprecated feature used")

// También con nivel dinámico
go_logs.SetDefaultLevel(go_logs.DebugLevel) // Cambiar nivel globalmente
```

## Testing Utilities

Herramientas para tests con logs:

### CaptureBuffer

Captura output de logs en tests:

```go
func TestMyFeature(t *testing.T) {
    buf := go_logs.NewCaptureBuffer()
    logger, _ := go_logs.New(go_logs.WithOutput(buf))

    logger.Info("test message")

    // Verificar contenido
    if !buf.Contains("test message") {
        t.Error("expected log message")
    }

    // Verificar múltiples substrings
    if !buf.ContainsAll("test", "message") {
        t.Error("expected both substrings")
    }

    // Obtener líneas
    lines := buf.Lines()
    lastLine := buf.LastLine()
}
```

### MockLogger

Mock completo para tests unitarios:

```go
func TestBusinessLogic(t *testing.T) {
    mock := go_logs.NewMockLogger()
    mock.SetLevel(go_logs.DebugLevel) // Capturar todos los niveles

    // Usar mock en lugar del logger real
    processOrder(mock, order)

    // Verificar logs generados
    if mock.Count() != 2 {
        t.Errorf("expected 2 logs, got %d", mock.Count())
    }

    if !mock.HasMessage("order processed") {
        t.Error("expected 'order processed' log")
    }

    if !mock.HasLevel(go_logs.ErrorLevel) {
        t.Error("expected an error log")
    }

    // Obtener última entrada
    last := mock.LastEntry()
    // Verificar campos
}
```

## Migración v2 → v3

Ver [MIGRATION.md](MIGRATION.md) para guía completa.

### Opciones de Migración

1. **Drop-in**: Actualizar dependencia, sin cambios de código
2. **Gradual**: Mezclar v2 y v3 en misma aplicación
3. **Full**: Adoptar todas las features de v3

### Ejemplo de Migración Gradual

```go
// v2 (sigue funcionando)
go_logs.InfoLog("legacy code")

// v3 (nuevo código)
logger := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
logger.Info("new code", go_logs.String("feature", "v3"))
```

## Testing

```bash
# Ejecutar tests
GO111MODULE=on go test ./...

# Con race detector
GO111MODULE=on go test -race ./...

# Benchmarks
GO111MODULE=on go test -bench=. -benchmem

# Cobertura
GO111MODULE=on go test -coverprofile=coverage.out ./...
GO111MODULE=on go tool cover -html=coverage.out
```

- 60+ test functions
- 30+ benchmarks
- ~90% coverage en código crítico
- Race detector clean

## Documentación

- [MIGRATION.md](MIGRATION.md) - Guía de migración v2 → v3
- [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) - Resumen técnico
- [CLAUDE.md](CLAUDE.md) - Guía para Claude Code
- [GoDoc](https://pkg.go.dev/github.com/drossan/go_logs) - Referencia de API

## Changelog

### v3.5 (Actual)

**Arquitectura Híbrida** - Core + submódulos opcionales

- **Metrics (Core)**: Estadísticas de logging con zero overhead
  - Contadores por nivel (Total, Info, Error, etc.)
  - Contador de logs dropeados (async)
  - Snapshots para monitoring
  - Thread-safe con atomic operations

- **async/** (Submódulo): Logging asíncrono non-blocking
  - Buffered channel para alto throughput
  - Drop-on-overflow cuando buffer lleno
  - Graceful shutdown con timeout
  - Métricas compartidas con logger síncrono

- **http/** (Submódulo): Log level dinámico via HTTP
  - GET /log-level - Obtener nivel actual
  - PUT /log-level - Cambiar nivel
  - GET /log-level/metrics - Métricas de logging
  - Bearer token authentication
  - Rate limiting
  - IP whitelisting (exacto y CIDR)

- **signal/** (Submódulo): SIGHUP handler para rotación
  - Rotación de logs al recibir señal del sistema
  - Compatible con logrotate de Linux
  - Señales personalizables
  - Thread-safe

### v3.4

- OpenTelemetry/OTLP: Envío de logs a collectors OpenTelemetry
- Syslog Hook: Integración con syslog local y remoto (RFC5424)
- OTLPExporter con buffering y flush automático
- NetworkSyslogHook para servidores remotos

### v3.3

- Sampling/Rate Limiting: Control de volumen de logs
- Global Logger: Acceso global con SetDefault() y Global*()
- Testing Utils: CaptureBuffer y MockLogger para tests
- SamplingWriter con callbacks de drop

### v3.2

- MultiWriter: Salida simultánea a múltiples destinos
- Rotación por tiempo: Daily (diario) y Hourly (horario)
- Compresión gzip: Archivos rotados comprimidos automáticamente
- MaxAge: Limpieza automática de archivos antiguos
- EnhancedRotatingFileWriter con configuración completa

### v3.1

- Caller Info: archivo, línea y función en cada log
- Stack Traces: captura automática en Error+
- WithCaller(), WithStackTrace(), WithStackTraceLevel() options

### v3.0

- Structured logging con campos tipados
- Child loggers con propagación de campos
- Context propagation real (trace_id, span_id)
- Dual formatters: TextFormatter + JSONFormatter
- Sistema de hooks extensible
- RotatingFileWriter sin dependencias
- Redactor de datos sensibles
- Interfaz Logger inyectable
- 100% backward compatible con v2

### v2.0

- API mejorada con funciones con formato
- Performance 98% más rápido con buffering
- Soporte de contexto para distributed tracing
- WarningLog agregado
- 83.5% de cobertura con tests

### v1.0

- Versión inicial con logging básico y Slack

## Licencia

MIT License - ver [LICENSE](LICENSE) para más detalles.
