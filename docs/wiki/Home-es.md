# go_logs v3

**[English](Home)** | **[Espanol](#espanol)**

---

<a name="espanol"></a>
## Espanol

Una biblioteca de logging moderna y estructurada para Go, sin dependencias externas (excepto integracion opcional con Slack).

[![Go Reference](https://pkg.go.dev/badge/github.com/drossan/go_logs.svg)](https://pkg.go.dev/github.com/drossan/go_logs)
[![Go Report Card](https://goreportcard.com/badge/github.com/drossan/go_logs)](https://goreportcard.com/report/github.com/drossan/go_logs)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Descripcion General

**go_logs** es una biblioteca de logging lista para produccion que proporciona logging estructurado con campos tipados, multiples formateadores de salida, child loggers, propagacion de contexto y un sistema de hooks extensible. Sigue los principios de Clean Architecture con una clara separacion entre interfaces e implementaciones.

La biblioteca mantiene **100% de compatibilidad hacia atras** entre v2 (funciones globales legacy) y v3 (interfaz Logger moderna), permitiendo una migracion gradual sin cambios disruptivos.

## Caracteristicas Principales

- **Logging Estructurado** - Campos con tipo seguro (String, Int, Int64, Float64, Bool, Err, Any)
- **Multiples Formateadores** - TextFormatter para desarrollo (colores), JSONFormatter para produccion (ELK, Loki, Datadog)
- **Child Loggers** - Crea loggers contextuales con campos heredados usando `With()`
- **Propagacion de Contexto** - Extraccion automatica de trace_id, span_id, request_id del contexto
- **Sistema de Hooks** - Arquitectura extensible para procesamiento personalizado (Slack, metricas, etc.)
- **Rotacion de Archivos** - RotatingFileWriter integrado con rotacion por tamano (sin dependencias)
- **Zero Allocations** - Creacion de campos y filtrado de nivel sin asignaciones
- **Thread-Safe** - Todas las operaciones son seguras para uso concurrente
- **Niveles estilo Syslog** - Trace=10, Debug=20, Info=30, Warn=40, Error=50, Fatal=60
- **Redaccion de Datos Sensibles** - Enmascaramiento automatico de passwords, tokens, API keys
- **Informacion del Caller** - Opcional archivo:linea y nombre de funcion en las entradas de log
- **Stack Traces** - Captura automatica para nivel Error y superiores

## Comparacion Rapida: v2 vs v3

| Caracteristica | v2 (Legacy) | v3 (Moderna) |
|----------------|-------------|--------------|
| Estilo de API | Funciones globales | Interfaz Logger |
| Campos Estructurados | No | Si (campos tipados) |
| Child Loggers | No | Si (`With()`) |
| Soporte de Contexto | No | Si (`LogCtx()`) |
| Multiples Instancias | No | Si |
| Hooks | No | Si |
| Inyeccion de Dependencias | No | Si (basada en interfaces) |
| Compatible hacia atras | - | Si (100%) |

### Ejemplo v2 (Legacy - Todavia Soportado)

```go
import "github.com/drossan/go_logs"

func main() {
    go_logs.Init() // Opcional, se auto-inicializa
    go_logs.InfoLog("Aplicacion iniciada")
    go_logs.ErrorLog("Algo salio mal")
}
```

### Ejemplo v3 (Moderna - Recomendado)

```go
import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Crear un logger configurado
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Logging estructurado con campos tipados
    logger.Info("Servidor iniciado",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )

    // Child logger con contexto heredado
    reqLogger := logger.With(
        go_logs.String("request_id", "abc-123"),
        go_logs.String("user_id", "user-456"),
    )
    reqLogger.Info("Peticion recibida")
}
```

## Instalacion

```bash
go get github.com/drossan/go_logs@latest
```

Ver [Instalacion](Installation-es.md) para instrucciones detalladas.

## Indice de Documentacion

### Primeros Pasos

- [Inicio Rapido](Getting-Started-es.md) - Guia de inicio rapido con tu primer logger
- [Instalacion](Installation-es.md) - Instalacion y dependencias

### Conceptos Core

- [Referencia API](API-Reference-es.md) - Documentacion completa de la API
- [Logging Estructurado](Structured-Logging-es.md) - Tipos de campos y mejores practicas
- [Formateadores](Formatters-es.md) - Configuracion de TextFormatter y JSONFormatter

### Caracteristicas Avanzadas

- [Contexto y Tracing](Context-and-Tracing-es.md) - Soporte para trazabilidad distribuida
- [Hooks](Hooks-es.md) - Procesamiento personalizado de logs e integracion con Slack
- [Rotacion de Archivos](File-Rotation-es.md) - Configuracion de RotatingFileWriter

### Configuracion

- [Configuracion](Configuration-es.md) - Variables de entorno y opciones programaticas

### Migracion

- [Migracion v2 a v3](Migration-v2-to-v3-es.md) - Guia de migracion paso a paso

### Ejemplos y Rendimiento

- [Ejemplos](Examples-es.md) - Ejemplos de codigo practicos para escenarios comunes
- [Rendimiento](Performance-es.md) - Benchmarks y tips de optimizacion

## Destacados de Rendimiento

| Operacion | Rendimiento |
|-----------|-------------|
| Filtrado fast-path | 0.32 ns/op |
| Creacion de campo | 0.34 ns/op, 0 allocs |
| TextFormatter | 220.6 ns/op |
| JSONFormatter | 249.3 ns/op |
| RotatingFileWriter | 16M msg/seg |

Ver [Rendimiento](Performance-es.md) para benchmarks detallados.

## Ejemplos Rapidos

### Logger Basico

```go
logger, _ := go_logs.New()
logger.Info("Hola, Mundo!")
```

### Logger con Rotacion de Archivos

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### Logger con Tracing de Contexto

```go
func handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    ctx = go_logs.WithTraceID(ctx, r.Header.Get("X-Trace-ID"))
    logger.LogCtx(ctx, go_logs.InfoLevel, "Procesando peticion")
}
```

### Ejemplo de Middleware HTTP

```go
func loggingMiddleware(logger go_logs.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            reqID := generateRequestID()

            reqLogger := logger.With(
                go_logs.String("request_id", reqID),
                go_logs.String("method", r.Method),
                go_logs.String("path", r.URL.Path),
            )

            ctx := go_logs.WithRequestID(r.Context(), reqID)
            next.ServeHTTP(w, r.WithContext(ctx))

            reqLogger.Info("Peticion completada",
                go_logs.Int64("duration_ms", time.Since(start).Milliseconds()),
                go_logs.Int("status", 200),
            )
        })
    }
}
```

## Contribuir

Las contribuciones son bienvenidas! Por favor lee las guias de contribucion antes de enviar PRs.

## Licencia

Licencia MIT - ver [LICENSE](../LICENSE) para mas detalles.

## Soporte

- **Issues**: [GitHub Issues](https://github.com/drossan/go_logs/issues)
- **Documentacion**: Esta wiki
- **Referencia Go**: [pkg.go.dev](https://pkg.go.dev/github.com/drossan/go_logs)
