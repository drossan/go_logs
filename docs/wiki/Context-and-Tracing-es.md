# Contexto y Tracing

go_logs proporciona soporte integrado para tracing distribuido a traves de propagacion de contexto. Esto te permite correlacionar logs a traves de multiples servicios y operaciones usando trace IDs, span IDs y request IDs.

**[English](Context-and-Tracing.md)** | **Espanol**

## Descripcion General

El tracing distribuido es esencial para debugging y monitoreo de microservicios. Cuando una peticion fluye a traves de multiples servicios, cada servicio loguea con el mismo trace ID, permitiendote:

1. **Correlacionar logs** entre servicios
2. **Trazar el flujo de peticiones** de inicio a fin
3. **Identificar cuellos de botella** y fallos
4. **Depurar problemas** en sistemas distribuidos

## Claves de Contexto

go_logs usa las siguientes claves de contexto para tracing:

| Clave | Descripcion | Valor de Ejemplo |
|-------|-------------|------------------|
| `trace_id` | Identificador unico para todo el flujo de la peticion | `abc-123-def-456` |
| `span_id` | Identificador para una operacion especifica | `span-xyz-789` |
| `request_id` | Identificador con alcance a un solo servicio | `req-456` |
| `user_id` | Usuario asociado con la peticion | `user-123` |

## Agregar Valores de Contexto

### WithTraceID

Agrega un trace ID al contexto.

```go
func WithTraceID(ctx context.Context, traceID string) context.Context
```

**Ejemplo:**

```go
ctx := go_logs.WithTraceID(context.Background(), "trace-abc-123")
```

### WithSpanID

Agrega un span ID al contexto.

```go
func WithSpanID(ctx context.Context, spanID string) context.Context
```

**Ejemplo:**

```go
ctx := go_logs.WithSpanID(ctx, "span-xyz-789")
```

### WithRequestID

Agrega un request ID al contexto.

```go
func WithRequestID(ctx context.Context, requestID string) context.Context
```

**Ejemplo:**

```go
ctx := go_logs.WithRequestID(ctx, "req-456")
```

### WithUserID

Agrega un user ID al contexto.

```go
func WithUserID(ctx context.Context, userID string) context.Context
```

**Ejemplo:**

```go
ctx := go_logs.WithUserID(ctx, "user-123")
```

## Extraer Valores de Contexto

### GetTraceID

Extrae el trace ID del contexto.

```go
func GetTraceID(ctx context.Context) string
```

Retorna un string vacio si no esta establecido.

### GetSpanID

Extrae el span ID del contexto.

```go
func GetSpanID(ctx context.Context) string
```

### GetRequestID

Extrae el request ID del contexto.

```go
func GetRequestID(ctx context.Context) string
```

### GetUserID

Extrae el user ID del contexto.

```go
func GetUserID(ctx context.Context) string
```

## Logging con Contexto

### LogCtx

Loguea un mensaje con soporte de contexto. Extrae automaticamente informacion de traza.

```go
func (l Logger) LogCtx(ctx context.Context, level Level, msg string, fields ...Field)
```

**Ejemplo:**

```go
ctx := go_logs.WithTraceID(context.Background(), "trace-123")
ctx = go_logs.WithSpanID(ctx, "span-456")

logger.LogCtx(ctx, go_logs.InfoLevel, "Procesando peticion")
```

**Salida:**
```
[2026/02/28 10:30:00] INFO Procesando peticion trace_id=trace-123 span_id=span-456
```

## Ejemplo de Handler HTTP

Aqui hay un ejemplo completo mostrando propagacion de contexto en handlers HTTP:

```go
package main

import (
    "context"
    "net/http"
    "os"

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

    // Envolver con middleware de tracing
    wrappedMux := tracingMiddleware(mux)

    logger.Info("Iniciando servidor", go_logs.Int("port", 8080))
    http.ListenAndServe(":8080", wrappedMux)
}

// tracingMiddleware extrae o genera trace IDs
func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extraer trace ID del header de la peticion o generar uno nuevo
        traceID := r.Header.Get("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.New().String()
        }

        // Agregar trace ID al contexto
        ctx := go_logs.WithTraceID(r.Context(), traceID)

        // Agregar request ID (unico por peticion)
        requestID := uuid.New().String()
        ctx = go_logs.WithRequestID(ctx, requestID)

        // Agregar a headers de respuesta para servicios downstream
        w.Header().Set("X-Trace-ID", traceID)
        w.Header().Set("X-Request-ID", requestID)

        // Lamar siguiente handler con contexto enriquecido
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Ejemplo de Microservicios

Cuando una peticion fluye a traves de multiples servicios, propaga el trace ID:

### Servicio A (API Gateway)

```go
func manejarPeticion(w http.ResponseWriter, r *http.Request) {
    // Generar o extraer trace ID
    traceID := r.Header.Get("X-Trace-ID")
    if traceID == "" {
        traceID = uuid.New().String()
    }

    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion recibida")

    // Llamar Servicio B
    resp, err := llamarServicioB(ctx)
    if err != nil {
        logger.LogCtx(ctx, go_logs.ErrorLevel, "Servicio B fallo",
            go_logs.Err(err),
        )
        http.Error(w, "Error", 500)
        return
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion completada")
}

func llamarServicioB(ctx context.Context) (*http.Response, error) {
    traceID := go_logs.GetTraceID(ctx)

    req, _ := http.NewRequest("GET", "http://servicio-b/api", nil)
    req.Header.Set("X-Trace-ID", traceID)

    logger.LogCtx(ctx, go_logs.DebugLevel, "Llamando Servicio B")
    return http.DefaultClient.Do(req)
}
```

### Servicio B

```go
func manejarPeticion(w http.ResponseWriter, r *http.Request) {
    // Extraer trace ID de la peticion entrante
    traceID := r.Header.Get("X-Trace-ID")
    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion recibida en Servicio B")

    // Procesar peticion...

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion completada en Servicio B")
}
```

### Logs Correlacionados

Servicio A:
```json
{"timestamp":"2026-02-28T10:30:00Z","level":"INFO","message":"Peticion recibida","fields":{"trace_id":"abc-123"}}
{"timestamp":"2026-02-28T10:30:01Z","level":"DEBUG","message":"Llamando Servicio B","fields":{"trace_id":"abc-123"}}
{"timestamp":"2026-02-28T10:30:02Z","level":"INFO","message":"Peticion completada","fields":{"trace_id":"abc-123"}}
```

Servicio B:
```json
{"timestamp":"2026-02-28T10:30:01Z","level":"INFO","message":"Peticion recibida en Servicio B","fields":{"trace_id":"abc-123"}}
{"timestamp":"2026-02-28T10:30:02Z","level":"INFO","message":"Peticion completada en Servicio B","fields":{"trace_id":"abc-123"}}
```

Todos los logs con `trace_id=abc-123` pueden ser correlacionados para trazar el flujo completo de la peticion.

## Mejores Practicas

### 1. Siempre Propagar Contexto

Pasa contexto a traves de todas las llamadas de funcion:

```go
// Bien
func procesarPedido(ctx context.Context, pedido *Pedido) error {
    logger.LogCtx(ctx, go_logs.InfoLevel, "Procesando pedido")
    return validarPedido(ctx, pedido)
}

// Mal - pierde contexto
func procesarPedido(pedido *Pedido) error {
    logger.Info("Procesando pedido") // Sin info de traza!
    return nil
}
```

### 2. Usar Middleware para HTTP

Siempre usa middleware para configurar el contexto de tracing:

```go
func main() {
    mux := http.NewServeMux()
    // ... registrar handlers ...

    // Envolver con middleware de tracing
    http.ListenAndServe(":8080", tracingMiddleware(mux))
}
```

### 3. Incluir Trace ID en Respuestas

Retorna trace IDs en headers de respuesta para debugging:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    traceID := go_logs.GetTraceID(ctx)

    w.Header().Set("X-Trace-ID", traceID)
    // ... manejar peticion ...
}
```

### 4. Generar Trace IDs en Puntos de Entrada

Genera trace IDs en los limites del servicio:

```go
func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := r.Header.Get("X-Trace-ID")
        if traceID == "" {
            // Generar nuevo trace ID si no se proporciona
            traceID = uuid.New().String()
        }
        ctx := go_logs.WithTraceID(r.Context(), traceID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Documentacion de funciones de contexto
- [Ejemplos](Examples-es.md) - Mas ejemplos de contexto
- [Hooks](Hooks-es.md) - Crear hooks que usan contexto
