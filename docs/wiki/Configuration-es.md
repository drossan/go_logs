# Configuracion

go_logs puede configurarse a traves de variables de entorno, opciones programaticas, o una combinacion de ambas.

**[English](Configuration.md)** | **Espanol**

## Tabla de Contenidos

- [Variables de Entorno](#variables-de-entorno)
- [Configuracion Programatica](#configuracion-programatica)
- [Configuracion por Entorno](#configuracion-por-entorno)
- [Redaccion de Datos Sensibles](#redaccion-de-datos-sensibles)

---

## Variables de Entorno

### Configuracion Core

| Variable | Descripcion | Por Defecto | Ejemplo |
|----------|-------------|-------------|---------|
| `LOG_LEVEL` | Nivel minimo de log | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | Formato de salida | `text` | `text`, `json` |

### Logging a Archivo (Compatibilidad v2)

| Variable | Descripcion | Por Defecto | Ejemplo |
|----------|-------------|-------------|---------|
| `SAVE_LOG_FILE` | Habilitar logging a archivo | `0` | `1` |
| `LOG_FILE_NAME` | Nombre del archivo de log | `log.txt` | `app.log` |
| `LOG_FILE_PATH` | Directorio del archivo de log | dir actual | `/var/log/miapp` |

### Rotacion de Archivos

| Variable | Descripcion | Por Defecto | Ejemplo |
|----------|-------------|-------------|---------|
| `LOG_MAX_SIZE` | Tamano max de archivo en MB | `100` | `100` |
| `LOG_MAX_BACKUPS` | Max archivos de backup | `5` | `10` |

### Notificaciones Slack (v2)

| Variable | Descripcion | Por Defecto |
|----------|-------------|-------------|
| `NOTIFICATIONS_SLACK_ENABLED` | Habilitar notificaciones Slack | `0` |
| `SLACK_TOKEN` | Token del bot de Slack | - |
| `SLACK_CHANNEL_ID` | ID del canal de Slack | - |

### Referencia de Niveles de Log

| Nivel | Valor | Descripcion |
|-------|-------|-------------|
| `trace` | 10 | Extremadamente detallado |
| `debug` | 20 | Informacion de diagnostico |
| `info` | 30 | Operacional general |
| `warn`, `warning` | 40 | Problemas potenciales |
| `error` | 50 | Errores |
| `fatal` | 60 | Errores criticos |
| `silent`, `none`, `disable` | 0 | Deshabilitar todo logging |

### Establecer Variables de Entorno

**Linux/macOS:**
```bash
export LOG_LEVEL=debug
export LOG_FORMAT=json
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_FILE_PATH=/var/log/miapp
```

**Docker Compose:**
```yaml
services:
  app:
    image: miapp
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      - SAVE_LOG_FILE=1
      - LOG_FILE_NAME=app.log
      - LOG_FILE_PATH=/var/log
```

**Kubernetes:**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: miapp
spec:
  containers:
  - name: app
    image: miapp
    env:
    - name: LOG_LEVEL
      value: "info"
    - name: LOG_FORMAT
      value: "json"
```

---

## Configuracion Programatica

### Patron de Opciones

Configurar el logger usando opciones funcionales:

```go
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
    go_logs.WithCaller(true),
    go_logs.WithCommonRedaction(),
)
```

### Opciones Disponibles

#### WithLevel

Establece el nivel minimo de log.

```go
go_logs.WithLevel(go_logs.InfoLevel)
```

#### WithOutput

Establece el destino de salida.

```go
go_logs.WithOutput(os.Stdout)
go_logs.WithOutput(archivo) // Cualquier io.Writer
```

#### WithFormatter

Establece el formateador de log.

```go
go_logs.WithFormatter(go_logs.NewTextFormatter())
go_logs.WithFormatter(go_logs.NewJSONFormatter())
```

#### WithHooks

Agrega hooks para procesamiento personalizado.

```go
go_logs.WithHooks(miHook1, miHook2)
```

#### WithCaller

Habilita informacion del caller (archivo:linea funcion).

```go
go_logs.WithCaller(true)
```

**Salida:**
```
[2026/02/28 10:30:00] INFO main.go:42 main.procesarPeticion Procesando peticion
```

#### WithCommonRedaction

Habilita redaccion de campos sensibles comunes.

```go
go_logs.WithCommonRedaction()
```

Redacta: password, passwd, pwd, token, api_key, apikey, api-key, secret, authorization, auth, cookie, session, credit_card, ssn, social_security

#### WithRotatingFile

Crea un writer de archivo con rotacion.

```go
go_logs.WithRotatingFile("/var/log/app.log", 100, 5)
```

---

## Configuracion por Entorno

### Desarrollo

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func NewLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.DebugLevel),
        go_logs.WithFormatter(go_logs.NewTextFormatter()), // Colores
        go_logs.WithOutput(os.Stdout),
        go_logs.WithCaller(true), // Mostrar archivo:linea
    )
    return logger
}
```

### Staging

```go
func NewLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/app.log", 50, 10),
    )
    return logger
}
```

### Produccion

```go
func NewLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/app.log", 100, 30),
        go_logs.WithCommonRedaction(),
        go_logs.WithStackTraceLevel(go_logs.ErrorLevel),
    )
    return logger
}
```

### Fabrica Basada en Entorno

```go
func NewLogger() go_logs.Logger {
    env := os.Getenv("ENV")

    switch env {
    case "production":
        return newProductionLogger()
    case "staging":
        return newStagingLogger()
    default:
        return newDevelopmentLogger()
    }
}
```

---

## Redaccion de Datos Sensibles

### Redaccion Integrada

Usa `WithCommonRedaction()` para campos sensibles comunes:

```go
logger, _ := go_logs.New(
    go_logs.WithCommonRedaction(),
)

logger.Info("Login usuario",
    go_logs.String("username", "juan"),
    go_logs.String("password", "secreto123"), // Sera redactado
    go_logs.String("token", "abc-123"),       // Sera redactado
)
```

**Salida:**
```
[2026/02/28 10:30:00] INFO Login usuario username=juan password=*** token=***
```

### Redaccion Personalizada

Especifica campos personalizados a redactar:

```go
logger, _ := go_logs.New(
    go_logs.WithRedactor("password", "token", "credit_card", "ssn"),
)
```

### Lista de Campos Redactados

Campos comunes redactados por `WithCommonRedaction()`:

| Categoria | Campos |
|-----------|--------|
| Passwords | password, passwd, pwd |
| Tokens | token, api_key, apikey, api-key |
| Secretos | secret, authorization, auth |
| Sesion | cookie, session |
| Financieros | credit_card, ssn, social_security |

---

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Todas las opciones de configuracion
- [Formateadores](Formatters-es.md) - Configuracion de formateadores
- [Rotacion de Archivos](File-Rotation-es.md) - Configuracion de rotacion
- [Hooks](Hooks-es.md) - Configuracion de hooks
