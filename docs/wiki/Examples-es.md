# Ejemplos

Esta pagina contiene ejemplos practicos para casos de uso comunes con go_logs v3.

**[English](Examples.md)** | **Espanol**

## Tabla de Contenidos

- [Logger Basico](#logger-basico)
- [Logger con Rotacion](#logger-con-rotacion)
- [Logger con Notificaciones Slack](#logger-con-notificaciones-slack)
- [Middleware HTTP](#middleware-http)
- [Interceptor gRPC](#interceptor-grpc)
- [Logging de Cola de Trabajos](#logging-de-cola-de-trabajos)
- [Microservicios](#microservicios)
- [Testing con Logs](#testing-con-logs)

---

## Logger Basico

### Logger Simple de Consola

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithOutput(os.Stdout),
    )
    if err != nil {
        panic(err)
    }

    logger.Info("Aplicacion iniciada")
    logger.Debug("Esto no se logueara (el nivel es Info)")
    logger.Error("Algo salio mal", go_logs.String("component", "main"))
}
```

### Logger con Campos Estructurados

```go
package main

import (
    "errors"
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New()

    // Campos String
    logger.Info("Accion de usuario",
        go_logs.String("user_id", "123"),
        go_logs.String("action", "login"),
    )

    // Campos numericos
    logger.Info("Peticion completada",
        go_logs.Int("status_code", 200),
        go_logs.Int64("duration_ms", 45),
    )

    // Campo de error
    err := errors.New("conexion rechazada")
    logger.Error("Error de base de datos",
        go_logs.Err(err),
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 5432),
    )

    // Campo booleano
    logger.Info("Verificacion de feature",
        go_logs.Bool("feature_enabled", true),
    )

    // Campo flotante
    logger.Info("Metricas de rendimiento",
        go_logs.Float64("cpu_percent", 45.7),
        go_logs.Float64("memory_gb", 1.23),
    )
}
```

---

## Logger con Rotacion

### Rotacion de Archivo con Formato JSON

```go
package main

import (
    "os"
    "time"
    "github.com/drossan/go_logs"
)

func main() {
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/miapp/app.log", 100, 5),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("Aplicacion iniciada",
        go_logs.String("version", "1.0.0"),
        go_logs.Int("pid", os.Getpid()),
    )

    // Simular trabajo de aplicacion
    for i := 0; i < 1000; i++ {
        logger.Info("Procesando peticion",
            go_logs.Int("request_id", i),
        )
        time.Sleep(100 * time.Millisecond)
    }

    logger.Info("Aplicacion finalizandose")
}
```

### Salida a Consola y Archivo

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Crear writer de archivo con rotacion
    fileWriter, err := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
    if err != nil {
        panic(err)
    }

    // Crear logger con multi-salida
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithMultiOutput(fileWriter, os.Stdout),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Los logs van tanto a archivo como a consola
    logger.Info("Aplicacion iniciada")
}
```

---

## Middleware HTTP

### Middleware de Logging de Peticiones

```go
package main

import (
    "context"
    "net/http"
    "os"
    "time"

    "github.com/drossan/go_logs"
    "github.com/google/uuid"
)

var logger go_logs.Logger

func main() {
    var err error
    logger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    mux := http.NewServeMux()
    mux.HandleFunc("/api/usuarios", usuariosHandler)
    mux.HandleFunc("/api/pedidos", pedidosHandler)

    // Envolver con middleware de logging
    wrappedMux := loggingMiddleware(mux)

    logger.Info("Iniciando servidor", go_logs.Int("port", 8080))
    http.ListenAndServe(":8080", wrappedMux)
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Crear logger con alcance de peticion
        reqLogger := logger.With(
            go_logs.String("request_id", requestID),
            go_logs.String("method", r.Method),
            go_logs.String("path", r.URL.Path),
            go_logs.String("remote_addr", r.RemoteAddr),
        )

        // Agregar trace ID del header o generar
        traceID := r.Header.Get("X-Trace-ID")
        if traceID != "" {
            ctx := go_logs.WithTraceID(r.Context(), traceID)
            r = r.WithContext(ctx)
        }

        // Log inicio de peticion
        reqLogger.Info("Peticion iniciada")

        // Envolver response writer para capturar status
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Llamar siguiente handler
        next.ServeHTTP(wrapped, r)

        // Log completacion de peticion
        duration := time.Since(start)
        reqLogger.Info("Peticion completada",
            go_logs.Int("status", wrapped.statusCode),
            go_logs.Int64("duration_ms", duration.Milliseconds()),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

func usuariosHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Handler usuarios llamado")
    w.Write([]byte(`{"usuarios": []}`))
}

func pedidosHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Handler pedidos llamado")
    w.Write([]byte(`{"pedidos": []}`))
}
```

---

## Microservicios

### Fabrica de Logger de Servicio

```go
package logging

import (
    "os"
    "github.com/drossan/go_logs"
)

var baseLogger go_logs.Logger

func Init(serviceName string) {
    var err error
    baseLogger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
        go_logs.WithCommonRedaction(),
    )
    if err != nil {
        panic(err)
    }

    baseLogger = baseLogger.With(
        go_logs.String("service", serviceName),
    )
}

func Logger() go_logs.Logger {
    return baseLogger
}
```

### Flujo de Peticion Cross-Service

**API Gateway:**

```go
func manejarPeticion(w http.ResponseWriter, r *http.Request) {
    traceID := uuid.New().String()
    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion recibida",
        go_logs.String("path", r.URL.Path),
    )

    // Llamar servicio downstream con trace ID
    req, _ := http.NewRequest("GET", "http://servicio-usuarios/api", nil)
    req.Header.Set("X-Trace-ID", traceID)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        logger.LogCtx(ctx, go_logs.ErrorLevel, "Llamada downstream fallo",
            go_logs.Err(err),
        )
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion completada")
}
```

**Servicio de Usuarios:**

```go
func manejarPeticion(w http.ResponseWriter, r *http.Request) {
    traceID := r.Header.Get("X-Trace-ID")
    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion recibida en servicio-usuarios")

    // Procesar peticion...

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion completada en servicio-usuarios")
}
```

---

## Testing con Logs

### Capturar Logs en Tests

```go
package main

import (
    "bytes"
    "testing"

    "github.com/drossan/go_logs"
)

func TestAlgo(t *testing.T) {
    // Crear buffer para capturar logs
    var buf bytes.Buffer

    logger, _ := go_logs.New(
        go_logs.WithOutput(&buf),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )

    // Ejecutar codigo que loguea
    procesarPedido(logger, "pedido-123")

    // Verificar que los logs contienen datos esperados
    output := buf.String()
    if !bytes.Contains(output, []byte("pedido-123")) {
        t.Error("Se esperaba ID de pedido en los logs")
    }
}
```

### Helper de Test

```go
package testing

import (
    "bytes"
    "testing"

    "github.com/drossan/go_logs"
)

// TestLogger crea un logger que captura salida para testing
func TestLogger(t *testing.T) (go_logs.Logger, *bytes.Buffer) {
    var buf bytes.Buffer
    logger, err := go_logs.New(
        go_logs.WithOutput(&buf),
        go_logs.WithLevel(go_logs.DebugLevel),
    )
    if err != nil {
        t.Fatalf("Fallo al crear logger de test: %v", err)
    }
    return logger, &buf
}

// Uso
func TestMiFuncion(t *testing.T) {
    logger, buf := TestLogger(t)

    result := miFuncion(logger)

    if !bytes.Contains(buf.Bytes(), []byte("mensaje esperado")) {
        t.Error("Mensaje de log esperado no encontrado")
    }
}
```

---

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Documentacion completa de API
- [Configuracion](Configuration-es.md) - Opciones de configuracion
- [Contexto y Tracing](Context-and-Tracing-es.md) - Tracing distribuido
- [Hooks](Hooks-es.md) - Procesamiento personalizado de logs
