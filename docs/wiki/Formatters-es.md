# Formateadores

Los formateadores controlan como las entradas de log se convierten en bytes para salida. go_logs proporciona dos formateadores integrados: TextFormatter para desarrollo y JSONFormatter para produccion.

**[English](Formatters.md)** | **Espanol**

## Interfaz Formatter

Todos los formateadores implementan la interfaz `Formatter`:

```go
type Formatter interface {
    Format(entry *Entry) ([]byte, error)
}
```

El metodo `Format` recibe una `Entry` y retorna bytes formateados. Las implementaciones deben ser thread-safe.

## TextFormatter

TextFormatter produce salida legible por humanos con colores ANSI opcionales. Este es el formateador por defecto y es ideal para desarrollo.

### Formato de Salida

```
[TIMESTAMP] LEVEL mensaje clave=valor clave2="valor con espacios"
```

### Ejemplo de Salida

```
[2026/02/28 17:30:00] INFO Servidor iniciado host=localhost port=8080
[2026/02/28 17:30:01] WARN Uso de memoria alto usage_percent=85.5
[2026/02/28 17:30:02] ERROR Fallo conexion a base de datos error=conexion rechazada host=db.ejemplo.com
```

### Crear un TextFormatter

```go
// Configuracion por defecto
formatter := go_logs.NewTextFormatter()

// Configuracion personalizada
formatter := go_logs.NewTextFormatterWithConfig(go_logs.FormatterConfig{
    EnableColors:    true,
    EnableTimestamp: true,
    EnableLevel:     true,
    TimestampFormat: "2006/01/02 15:04:05",
})
```

### Metodos de Configuracion

```go
// Habilitar/deshabilitar colores ANSI
formatter.SetEnableColors(true)

// Habilitar/deshabilitar timestamp
formatter.SetEnableTimestamp(true)

// Habilitar/deshabilitar nivel
formatter.SetEnableLevel(true)

// Establecer formato de timestamp (formato de tiempo de referencia Go)
formatter.SetTimestampFormat("2006-01-02 15:04:05.000")
```

### Mapeo de Colores

| Nivel | Color |
|-------|-------|
| Trace | Cyan |
| Debug | HiBlue (azul brillante) |
| Info | Yellow |
| Warn | HiYellow (amarillo brillante) |
| Error | Red |
| Fatal | HiRed (rojo brillante) |
| Success | Green |

### Ejemplo de Uso

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Crear TextFormatter con configuracion por defecto
    formatter := go_logs.NewTextFormatter()

    logger, _ := go_logs.New(
        go_logs.WithFormatter(formatter),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Debug("Mensaje de debug")
    logger.Info("Mensaje de info")
    logger.Warn("Mensaje de advertencia")
    logger.Error("Mensaje de error")
}
```

### Deshabilitar Colores

Para salida no-TTY (archivos, pipes), deshabilita colores:

```go
formatter := go_logs.NewTextFormatter()
formatter.SetEnableColors(false)

logger, _ := go_logs.New(
    go_logs.WithFormatter(formatter),
)
```

## JSONFormatter

JSONFormatter produce salida JSON estructurada ideal para sistemas de agregacion de logs como ELK, Loki, Datadog y Splunk.

### Formato de Salida

```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"mensaje","fields":{"clave":"valor"}}
```

### Estructura de Salida

```go
type JSONLogEntry struct {
    Timestamp  string                 `json:"timestamp,omitempty"`
    Level      string                 `json:"level,omitempty"`
    Message    string                 `json:"message"`
    Fields     map[string]interface{} `json:"fields,omitempty"`
    Caller     string                 `json:"caller,omitempty"`
    CallerFunc string                 `json:"caller_func,omitempty"`
    StackTrace string                 `json:"stack_trace,omitempty"`
}
```

### Crear un JSONFormatter

```go
// Configuracion por defecto
formatter := go_logs.NewJSONFormatter()

// Configuracion personalizada
formatter := go_logs.NewJSONFormatterWithConfig(go_logs.FormatterConfig{
    EnableTimestamp: true,
    EnableLevel:     true,
    TimestampFormat: time.RFC3339,
})
```

### Ejemplo de Uso

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Info("Servidor iniciado",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )
}
```

**Salida:**
```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"Servidor iniciado","fields":{"host":"localhost","port":8080}}
```

## Integracion con Agregadores de Logs

### ELK Stack (Elasticsearch, Logstash, Kibana)

Configura Filebeat o Logstash para leer logs JSON:

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  paths:
    - /var/log/app/*.log
  json.keys_under_root: true
  json.add_error_key: true
```

### Grafana Loki

Usa Promtail o el driver Docker de Loki:

```yaml
# promtail.yml
scrape_configs:
- job_name: app
  static_configs:
  - targets:
      - localhost
    labels:
      job: app
      __path__: /var/log/app/*.log
  pipeline_stages:
  - json:
      expressions:
        level: level
        message: message
```

### Datadog

Usa el Datadog Agent con parsing JSON:

```yaml
# datadog.yaml
logs_enabled: true

# conf.d/app.d/conf.yaml
logs:
  - type: file
    path: /var/log/app/app.log
    service: miapp
    source: go
```

## Elegir un Formateador

| Caso de Uso | Formateador Recomendado |
|-------------|------------------------|
| Desarrollo | TextFormatter con colores |
| Produccion | JSONFormatter |
| Agregacion de logs (ELK, Loki) | JSONFormatter |
| Lectura directa de archivo | TextFormatter (sin colores) |
| Salida personalizada | Formateador Custom |

## Consideraciones de Rendimiento

| Formateador | Rendimiento Tipico | Notas |
|-------------|-------------------|-------|
| TextFormatter | ~220 ns/op | Rapido, asignaciones minimas |
| JSONFormatter | ~250 ns/op | Usa encoding/json, escaping apropiado |
| Custom | Varia | Perfilar para uso en produccion |

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Documentacion de interfaz Formatter
- [Configuracion](Configuration-es.md) - Variable de entorno LOG_FORMAT
- [Rotacion de Archivos](File-Rotation-es.md) - Combinar formateadores con salida a archivo
