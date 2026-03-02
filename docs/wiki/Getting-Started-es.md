# Inicio Rapido

Esta guia te ayudara a comenzar con go_logs en solo unos minutos.

**[English](Getting-Started.md)** | **Espanol**

## Requisitos Previos

- **Go 1.21+** - La biblioteca usa caracteristicas modernas de Go
- Un proyecto con Go modules habilitado

## Inicio Rapido

### 1. Instalar el Paquete

```bash
go get github.com/drossan/go_logs@latest
```

### 2. Crear Tu Primer Logger

Crea un nuevo archivo `main.go`:

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Crear un logger con configuracion por defecto
    logger, err := go_logs.New()
    if err != nil {
        panic(err)
    }

    // Logging basico
    logger.Info("Hola, go_logs!")
}
```

### 3. Ejecutar Tu Programa

```bash
go run main.go
```

**Salida esperada:**
```
[2026/02/28 10:30:00] INFO Hola, go_logs!
```

## Tu Primer Log Estructurado

El logging estructurado agrega contexto a tus mensajes con campos tipados:

```go
package main

import (
    "os"
    "errors"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
    )

    // Log con campos estructurados
    logger.Info("Servidor iniciado",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
        go_logs.String("environment", "development"),
    )

    // Loguear un error
    err := errors.New("conexion rechazada")
    logger.Error("Fallo al conectar a la base de datos",
        go_logs.Err(err),
        go_logs.String("host", "db.example.com"),
        go_logs.Int("port", 5432),
    )
}
```

**Salida esperada:**
```
[2026/02/28 10:30:00] INFO Servidor iniciado host=localhost port=8080 environment=development
[2026/02/28 10:30:01] ERROR Fallo al conectar a la base de datos error=conexion rechazada host=db.example.com port=5432
```

## Usando Child Loggers

Los child loggers heredan campos de su padre, reduciendo la repeticion:

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    // Logger base con informacion del servicio
    baseLogger, _ := go_logs.New()

    // Crear un child logger con contexto de peticion
    requestLogger := baseLogger.With(
        go_logs.String("request_id", "req-abc-123"),
        go_logs.String("user_id", "user-456"),
    )

    // Todos los logs de requestLogger incluyen request_id y user_id
    requestLogger.Info("Peticion recibida")
    requestLogger.Info("Procesando pago")
    requestLogger.Info("Peticion completada")
}
```

**Salida esperada:**
```
[2026/02/28 10:30:00] INFO Peticion recibida request_id=req-abc-123 user_id=user-456
[2026/02/28 10:30:01] INFO Procesando pago request_id=req-abc-123 user_id=user-456
[2026/02/28 10:30:02] Peticion completada request_id=req-abc-123 user_id=user-456
```

## Cambiando a Formato JSON

Para entornos de produccion, usa formato JSON para compatibilidad con agregadores de logs:

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Info("Servidor iniciado",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )
}
```

**Salida esperada:**
```json
{"timestamp":"2026-02-28T10:30:00Z","level":"INFO","message":"Servidor iniciado","fields":{"host":"localhost","port":8080}}
```

## Escribir a Archivo con Rotacion

Para logging persistente con rotacion automatica:

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithRotatingFile("/var/log/myapp/app.log", 100, 5),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )
    defer logger.Sync() // Flush antes de salir

    logger.Info("Aplicacion iniciada")
}
```

Esto crea:
- `/var/log/myapp/app.log` - Archivo de log actual
- `/var/log/myapp/app.log.1` - Backup mas reciente
- Hasta 5 archivos de backup, rotando cuando el archivo alcanza 100MB

## Usando Contexto para Tracing Distribuido

Pasa trace IDs a traves de tu aplicacion para correlacion de peticiones:

```go
package main

import (
    "context"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New()

    // Agregar informacion de traza al contexto
    ctx := context.Background()
    ctx = go_logs.WithTraceID(ctx, "trace-abc-123")
    ctx = go_logs.WithSpanID(ctx, "span-xyz-789")

    // Log con contexto - trace_id y span_id se incluyen automaticamente
    logger.LogCtx(ctx, go_logs.InfoLevel, "Procesando peticion",
        go_logs.String("operation", "create_user"),
    )
}
```

**Salida esperada:**
```
[2026/02/28 10:30:00] INFO Procesando peticion trace_id=trace-abc-123 span_id=span-xyz-789 operation=create_user
```

## Usando Variables de Entorno

Configura el logging sin cambiar codigo:

```bash
# Establecer nivel de log
export LOG_LEVEL=debug

# Establecer formato (text o json)
export LOG_FORMAT=json

# Habilitar logging a archivo
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_FILE_PATH=/var/log/myapp

# Configurar rotacion
export LOG_MAX_SIZE=100
export LOG_MAX_BACKUPS=5
```

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    // El logger automaticamente recoge las variables de entorno
    logger, _ := go_logs.New()

    logger.Debug("Esto se logueara si LOG_LEVEL=debug")
    logger.Info("Aplicacion iniciada")
}
```

## Ejemplo Completo: Servidor HTTP

Aqui hay un ejemplo completo mostrando las mejores practicas:

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

    // Configurar logger para produccion
    logger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
        go_logs.WithCaller(true), // Incluir archivo:linea
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Crear servidor HTTP con middleware de logging
    mux := http.NewServeMux()
    mux.HandleFunc("/api/hello", helloHandler)

    wrappedMux := loggingMiddleware(mux)

    logger.Info("Iniciando servidor",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )

    if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
        logger.Fatal("Servidor fallo", go_logs.Err(err))
    }
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Handler hello llamado")
    w.Write([]byte("Hola, Mundo!"))
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := uuid.New().String()

        // Crear logger con alcance de peticion
        reqLogger := logger.With(
            go_logs.String("request_id", requestID),
            go_logs.String("method", r.Method),
            go_logs.String("path", r.URL.Path),
            go_logs.String("remote_addr", r.RemoteAddr),
        )

        // Agregar request ID al contexto
        ctx := go_logs.WithRequestID(r.Context(), requestID)

        // Log inicio de peticion
        reqLogger.Info("Peticion iniciada")

        // Llamar siguiente handler
        next.ServeHTTP(w, r.WithContext(ctx))

        // Log completacion de peticion
        duration := time.Since(start)
        reqLogger.Info("Peticion completada",
            go_logs.Int64("duration_ms", duration.Milliseconds()),
        )
    })
}
```

## Siguientes Pasos

Ahora que tienes go_logs funcionando, explora estos temas:

- [Referencia API](API-Reference-es.md) - Documentacion completa de la API
- [Logging Estructurado](Structured-Logging-es.md) - Todos los tipos de campos y mejores practicas
- [Formateadores](Formatters-es.md) - Personalizar formato de salida
- [Contexto y Tracing](Context-and-Tracing-es.md) - Tracing distribuido
- [Hooks](Hooks-es.md) - Procesamiento personalizado de logs
- [Rotacion de Archivos](File-Rotation-es.md) - Logging a archivo en produccion
- [Configuracion](Configuration-es.md) - Todas las opciones de configuracion
- [Ejemplos](Examples-es.md) - Mas ejemplos practicos

## Problemas Comunes

### No Aparece Salida

Asegurate de que el nivel de log esta configurado apropiadamente:

```go
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel), // Habilitar debug y superiores
)
```

### Colores No Funcionan en Terminal

Los colores estan habilitados por defecto. Si ves codigos de escape, tu terminal puede no soportar colores ANSI. Deshabilitalos:

```go
formatter := go_logs.NewTextFormatter()
formatter.SetEnableColors(false)
logger, _ := go_logs.New(
    go_logs.WithFormatter(formatter),
)
```

### Logs No Aparecen en Archivo

Asegurate de llamar `Sync()` antes de salir:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("app.log", 100, 5),
)
defer logger.Sync() // Importante!
```
