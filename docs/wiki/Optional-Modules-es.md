# Módulos Opcionales (Arquitectura Híbrida)

go_logs v3.5 introduce una **arquitectura híbrida** con el core en el paquete principal y features opcionales como submódulos.

## Resumen de Arquitectura

```
go_logs/
├── Core (siempre incluido)
│   ├── logger.go, logger_impl.go
│   ├── metrics.go              ← NUEVO: Métricas zero-overhead
│   └── ... (otros archivos core)
│
├── async/                       ← Opt-in
│   └── async.go                ← Logging non-blocking
│
├── http/                        ← Opt-in
│   └── dynamic_level.go        ← Control de nivel via HTTP
│
└── signal/                      ← Opt-in
    └── signal.go               ← Handler SIGHUP
```

## Metrics (Core - Siempre Habilitado)

Estadísticas de logging con zero overhead disponibles en cada logger:

```go
logger, _ := go_logs.New()
metrics := logger.GetMetrics()

// Métricas disponibles
fmt.Printf("Total logs: %d\n", metrics.Total())
fmt.Printf("Logs info: %d\n", metrics.Count(go_logs.InfoLevel))
fmt.Printf("Errores: %d\n", metrics.Count(go_logs.ErrorLevel))
fmt.Printf("Dropeados: %d\n", metrics.Dropped())

// Obtener snapshot para monitoring
snapshot := metrics.Snapshot()
// snapshot.Total
// snapshot.ByLevel[go_logs.InfoLevel]

// Reset para testing
metrics.Reset()
```

### Thread-Safety

Todas las operaciones de métricas usan operaciones atómicas, haciéndolas seguras para uso concurrente con zero overhead de locks.

### Integración con Monitoring

```go
// Exponer métricas a Prometheus
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    metrics := logger.GetMetrics()
    snapshot := metrics.Snapshot()

    fmt.Fprintf(w, "# HELP go_logs_total Total de entradas de log\n")
    fmt.Fprintf(w, "# TYPE go_logs_total counter\n")
    fmt.Fprintf(w, "go_logs_total %d\n", snapshot.Total)

    for level, count := range snapshot.ByLevel {
        fmt.Fprintf(w, "go_logs_by_level{level=\"%s\"} %d\n", level.String(), count)
    }

    fmt.Fprintf(w, "go_logs_dropped %d\n", metrics.Dropped())
})
```

## Async Logging (Submódulo - Opt-in)

Logging non-blocking para aplicaciones de alto throughput:

```go
import "github.com/drossan/go_logs/async"

// Crear logger síncrono base
syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

// Envolver con async (tamaño buffer 1000)
asyncLogger := async.Wrap(syncLogger, 1000)
defer asyncLogger.Sync()

// Logging non-blocking
asyncLogger.Info("Servidor iniciado", go_logs.Int("puerto", 8080))
```

### Configuración

```go
asyncLogger := async.WrapWithConfig(syncLogger, async.Config{
    BufferSize:      10000,              // Capacidad del buffer
    ShutdownTimeout: 10 * time.Second,   // Max espera en Sync()
})
```

### Comportamiento

- **Non-blocking**: Las llamadas de log retornan inmediatamente
- **Drop-on-overflow**: Si el buffer está lleno, se dropean logs (contados en métricas)
- **Graceful shutdown**: `Sync()` espera todos los logs pendientes
- **Métricas compartidas**: Los logs dropeados incrementan el contador compartido

### Cuándo Usar

- Aplicaciones de alto throughput (>10k logs/seg)
- Logging a outputs lentos (red, servicios remotos)
- Cuando puedes tolerar pérdida ocasional de logs por performance

### Cuándo NO Usar

- Logging financiero/audit donde cada log debe persistirse
- Aplicaciones de bajo throughput (<1k logs/seg)
- Cuando debuggeas issues donde logs pueden perderse

## Nivel Dinámico via HTTP (Submódulo - Opt-in)

Cambiar nivel de log en runtime via endpoints HTTP:

```go
import httplogs "github.com/drossan/go_logs/http"

logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

handler := httplogs.NewDynamicLevelHandler(logger, httplogs.Config{
    Endpoint:  "/debug/level",
    AuthToken: "token-secreto",
    RateLimit: 10,  // peticiones por segundo
    AllowedIPs: []string{"10.0.0.0/8", "192.168.1.100"},
})

http.Handle("/debug/", handler)
```

### Endpoints

| Método | Path | Descripción |
|--------|------|-------------|
| GET | `/debug/level` | Obtener nivel actual |
| PUT | `/debug/level` | Establecer nivel |
| GET | `/debug/level/metrics` | Obtener métricas de logging |

### Ejemplos de Uso

```bash
# Obtener nivel actual
curl -H "Authorization: Bearer token-secreto" http://localhost:8080/debug/level
# Response: {"level": "INFO", "timestamp": "2026-03-01T10:00:00Z"}

# Cambiar a debug
curl -X PUT -H "Authorization: Bearer token-secreto" \
  -H "Content-Type: application/json" \
  -d '{"level": "debug"}' \
  http://localhost:8080/debug/level
# Response: {"level": "DEBUG", "timestamp": "2026-03-01T10:01:00Z"}

# Obtener métricas
curl -H "Authorization: Bearer token-secreto" http://localhost:8080/debug/level/metrics
# Response: {"total": 1234, "by_level": {"INFO": 1000, "ERROR": 234}, "dropped": 0}
```

### Features de Seguridad

- **Autenticación Bearer token**: Requerido si `AuthToken` está configurado
- **Rate limiting**: Previene abuso (peticiones por segundo)
- **IP whitelisting**: IPs exactas o rangos CIDR

### Casos de Uso

- Debuggear issues en producción sin redeploy
- Aumentar logging temporalmente durante incidentes
- Reducir logging durante horas pico

## Handler SIGHUP (Submódulo - Opt-in)

Rotar logs al recibir señales del sistema:

```go
import "github.com/drossan/go_logs/signal"

writer, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
logger, _ := go_logs.New(go_logs.WithOutput(writer))

// Crear handler para SIGHUP
handler := signal.NewSIGHUPHandler(signal.WrapRotator(writer))
handler.Register()
defer handler.Stop()

// Los logs rotarán al recibir SIGHUP
```

### Integración con logrotate

Crear `/etc/logrotate.d/miapp`:

```
/var/log/app.log {
    daily
    rotate 30
    compress
    missingok
    notifempty
    postrotate
        kill -HUP $(cat /var/run/miapp.pid)
    endscript
}
```

### Señales Personalizadas

```go
// Manejar múltiples señales
handler := signal.NewSIGHUPHandler(rotator, syscall.SIGUSR1, syscall.SIGUSR2)
```

## Resumen

| Feature | Paquete | Overhead | Caso de Uso |
|---------|---------|----------|-------------|
| **Metrics** | Core | Zero (atomic) | Monitoring & observabilidad |
| **Async** | `async/` | Goroutine + channel | Alto throughput |
| **HTTP** | `http/` | HTTP endpoint | Debugging en producción |
| **Signal** | `signal/` | Signal handler | Integración con logrotate |
