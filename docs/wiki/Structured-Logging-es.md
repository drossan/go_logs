# Logging Estructurado

El logging estructurado agrega contexto a tus mensajes de log usando pares clave-valor tipados llamados campos. Esto hace que los logs sean buscables, filtrables y mas faciles de analizar en sistemas de agregacion de logs.

**[English](Structured-Logging.md)** | **Espanol**

## Que es el Logging Estructurado?

El logging tradicional usa mensajes de texto plano:

```
[2026/02/28 10:30:00] INFO Usuario juan conectado desde 192.168.1.1
```

El logging estructurado separa el mensaje de los datos:

```
[2026/02/28 10:30:00] INFO Usuario conectado username=juan ip=192.168.1.1
```

En formato JSON, esto es aun mas potente:

```json
{
  "timestamp": "2026-02-28T10:30:00Z",
  "level": "INFO",
  "message": "Usuario conectado",
  "fields": {
    "username": "juan",
    "ip": "192.168.1.1"
  }
}
```

Esta estructura te permite:
- Buscar por valores de campos (`username:juan`)
- Filtrar por presencia de campo (`exists:ip`)
- Agregar y analizar datos (contar logins por username)

## Tipos de Campos

go_logs proporciona constructores de campos tipados para diferentes tipos de datos:

### String

Para valores de texto como usernames, hostnames, IDs, etc.

```go
go_logs.String(key, value string) Field
```

**Ejemplo:**

```go
logger.Info("Usuario conectado",
    go_logs.String("username", "juan"),
    go_logs.String("email", "juan@ejemplo.com"),
    go_logs.String("ip", "192.168.1.1"),
)
```

**Salida (Texto):**
```
[2026/02/28 10:30:00] INFO Usuario conectado username=juan email=juan@ejemplo.com ip=192.168.1.1
```

### Int

Para valores enteros como puertos, contadores, codigos de estado, etc.

```go
go_logs.Int(key string, value int) Field
```

**Ejemplo:**

```go
logger.Info("Peticion HTTP",
    go_logs.Int("status_code", 200),
    go_logs.Int("response_size", 1024),
    go_logs.Int("port", 8080),
)
```

### Int64

Para enteros grandes como tamanos de archivo, timestamps, IDs de base de datos, etc.

```go
go_logs.Int64(key string, value int64) Field
```

**Ejemplo:**

```go
logger.Info("Archivo procesado",
    go_logs.Int64("bytes", 1536299520),
    go_logs.Int64("file_id", 9876543210),
    go_logs.Int64("timestamp_ns", time.Now().UnixNano()),
)
```

### Float64

Para valores de punto flotante como porcentajes, mediciones, tasas, etc.

```go
go_logs.Float64(key string, value float64) Field
```

**Ejemplo:**

```go
logger.Info("Metricas del sistema",
    go_logs.Float64("cpu_percent", 45.7),
    go_logs.Float64("memory_gb", 1.23),
    go_logs.Float64("request_rate", 1250.5),
)
```

### Bool

Para flags y estados booleanos.

```go
go_logs.Bool(key string, value bool) Field
```

**Ejemplo:**

```go
logger.Info("Feature flags",
    go_logs.Bool("debug_mode", true),
    go_logs.Bool("cache_enabled", false),
    go_logs.Bool("maintenance_mode", false),
)
```

### Err

Para valores de error. La clave se establece automaticamente como `"error"`.

```go
go_logs.Err(err error) Field
```

**Ejemplo:**

```go
err := errors.New("conexion rechazada")
logger.Error("Fallo conexion a base de datos",
    go_logs.Err(err),
    go_logs.String("host", "db.ejemplo.com"),
)
```

**Salida:**
```
[2026/02/28 10:30:00] ERROR Fallo conexion a base de datos error=conexion rechazada host=db.ejemplo.com
```

### Any

Para valores arbitrarios que no encajan en otros tipos. Usa `fmt.Sprintf("%v")` para formateo.

```go
go_logs.Any(key string, value interface{}) Field
```

**Ejemplo:**

```go
type Usuario struct {
    ID   int
    Nombre string
}

usuario := Usuario{ID: 123, Nombre: "Juan"}
logger.Info("Usuario creado",
    go_logs.Any("usuario", usuario),
)
```

**Salida:**
```
[2026/02/28 10:30:00] INFO Usuario creado usuario={123 Juan}
```

## Referencia de Tipos de Campo

| Funcion | Tipo Go | Tipo JSON | Caso de Uso |
|---------|---------|-----------|-------------|
| `String` | string | string | Valores de texto (nombres, IDs, URLs) |
| `Int` | int | number | Enteros pequenos (contadores, puertos) |
| `Int64` | int64 | number | Enteros grandes (tamanos de archivo, timestamps) |
| `Float64` | float64 | number | Decimales (porcentajes, mediciones) |
| `Bool` | bool | boolean | Flags y estados |
| `Err` | error | string | Valores de error (clave es "error") |
| `Any` | interface{} | varies | Valores arbitrarios |

## Mejores Practicas

### Usar Nombres de Campo Consistentes

Usa el mismo nombre de campo para el mismo concepto en toda tu aplicacion:

```go
// Bien: Nomenclatura consistente
logger.Info("Usuario conectado", go_logs.String("user_id", "123"))
logger.Info("Pedido creado", go_logs.String("user_id", "123"))
logger.Info("Pago procesado", go_logs.String("user_id", "123"))

// Mal: Nomenclatura inconsistente
logger.Info("Usuario conectado", go_logs.String("user_id", "123"))
logger.Info("Pedido creado", go_logs.String("uid", "123"))
logger.Info("Pago procesado", go_logs.String("customer_id", "123"))
```

### Usar Claves Descriptivas

Elige nombres de campo claros y descriptivos:

```go
// Bien
logger.Info("Peticion completada",
    go_logs.Int("response_time_ms", 45),
    go_logs.Int("status_code", 200),
)

// Mal
logger.Info("Peticion completada",
    go_logs.Int("time", 45),
    go_logs.Int("code", 200),
)
```

### Usar Tipos Apropiados

Coincide los tipos de campo con los datos:

```go
// Bien: Usar Int64 para numeros grandes
logger.Info("Archivo procesado",
    go_logs.Int64("bytes", 1536299520),
)

// Mal: Usar String para numeros (pierde indexacion numerica)
logger.Info("Archivo procesado",
    go_logs.String("bytes", "1536299520"),
)
```

### Usar Child Loggers para Contexto

Crea child loggers con campos comunes para evitar repeticion:

```go
// En lugar de repetir campos:
logger.Info("Peticion iniciada", go_logs.String("request_id", "abc-123"))
logger.Info("Procesando", go_logs.String("request_id", "abc-123"))
logger.Info("Peticion completada", go_logs.String("request_id", "abc-123"))

// Usar un child logger:
reqLogger := logger.With(go_logs.String("request_id", "abc-123"))
reqLogger.Info("Peticion iniciada")
reqLogger.Info("Procesando")
reqLogger.Info("Peticion completada")
```

### Incluir Contexto de Error

Al loguear errores, incluye contexto relevante:

```go
err := db.Connect()
if err != nil {
    logger.Error("Fallo conexion a base de datos",
        go_logs.Err(err),
        go_logs.String("host", cfg.DB.Host),
        go_logs.Int("port", cfg.DB.Port),
        go_logs.String("database", cfg.DB.Name),
    )
}
```

## Patrones Comunes

### Logging de Peticiones HTTP

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := r.Header.Get("X-Request-ID")

    // Crear logger con alcance de peticion
    reqLogger := h.logger.With(
        go_logs.String("request_id", requestID),
        go_logs.String("method", r.Method),
        go_logs.String("path", r.URL.Path),
        go_logs.String("remote_addr", r.RemoteAddr),
    )

    reqLogger.Info("Peticion iniciada")

    // ... manejar peticion ...

    reqLogger.Info("Peticion completada",
        go_logs.Int("status", statusCode),
        go_logs.Int64("duration_ms", time.Since(start).Milliseconds()),
        go_logs.Int("response_bytes", responseSize),
    )
}
```

### Logging de Operaciones de Base de Datos

```go
func (db *Database) Query(ctx context.Context, query string, args ...interface{}) {
    start := time.Now()

    logger.Debug("Query de base de datos iniciada",
        go_logs.String("query", query),
        go_logs.Any("args", args),
    )

    result, err := db.exec(ctx, query, args...)

    duration := time.Since(start)
    logger.Debug("Query de base de datos completada",
        go_logs.String("query", query),
        go_logs.Int64("duration_ms", duration.Milliseconds()),
        go_logs.Int("rows_affected", result.RowsAffected),
    )

    if err != nil {
        logger.Error("Query de base de datos fallo",
            go_logs.Err(err),
            go_logs.String("query", query),
        )
    }

    return result, err
}
```

## Convenciones de Nomenclatura de Campos

Sigue estas convenciones para nombres de campo consistentes:

| Concepto | Nombre de Campo Recomendado | Tipo |
|----------|----------------------------|------|
| Identificador de usuario | `user_id` | String |
| Identificador de peticion | `request_id` | String |
| Identificador de traza | `trace_id` | String |
| Identificador de span | `span_id` | String |
| Metodo HTTP | `method` | String |
| Ruta HTTP | `path` | String |
| Estado HTTP | `status_code` | Int |
| Tiempo de respuesta | `duration_ms` | Int64 |
| Error | `error` | Err |
| Host de base de datos | `db_host` | String |
| Nombre de base de datos | `database` | String |
| Ruta de archivo | `file_path` | String |
| Tamano de archivo | `bytes` | Int64 |

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Referencia completa de funciones de campo
- [Contexto y Tracing](Context-and-Tracing-es.md) - Campos de contexto automaticos
- [Ejemplos](Examples-es.md) - Mas ejemplos practicos
