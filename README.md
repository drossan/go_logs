# Go_Logs - Biblioteca de Registro para Go

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/drossan/go_logs)

## Descripción

Este paquete está diseñado para manejar el registro de eventos y errores de una manera centralizada en aplicaciones Go. Proporciona funcionalidades para registrar mensajes de diferentes niveles de severidad con colores en terminal, archivo persistente, y notificaciones a través de Slack.

### Características Principales

- ✅ **Múltiples niveles de registro**: Fatal, Error, Warning, Info, Success
- ✅ **Salida con colores**: Mensajes coloreados en terminal para mejor legibilidad
- ✅ **Archivo persistente**: Logging eficiente con buffering (98% más rápido que versiones anteriores)
- ✅ **Notificaciones Slack**: Configurable para cada nivel de log
- ✅ **Soporte para formateo**: Funciones tipo `fmt.Sprintf` (Infof, Errorf, etc.)
- ✅ **Soporte para contexto**: Funciones con `context.Context` para distributed tracing
- ✅ **Thread-safe**: Protección contra race conditions con mutex
- ✅ **Cero dependencias externas**: Solo usa bibliotecas estándar de Go

## Niveles de Registro

El paquete ofrece los siguientes niveles de registro:

| Función | Nivel | Color | Uso |
|---------|-------|-------|-----|
| `FatalLog()` | FATAL | Rojo 💣 | Errores fatales que terminan el programa |
| `ErrorLog()` | ERROR | Rojo | Errores que no requieren terminar el programa |
| `WarningLog()` | WARNING | Amarillo | Advertencias potenciales |
| `InfoLog()` | INFO | Amarillo | Eventos informativos |
| `SuccessLog()` | SUCCESS | Verde | Operaciones exitosas |

## Configuración

La configuración se maneja a través de variables de entorno:

```dotenv
# Archivo de Log
SAVE_LOG_FILE=1                    # Habilitar logging a archivo (0 o 1)
LOG_FILE_NAME=app.log              # Nombre del archivo (default: log.txt)
LOG_FILE_PATH=/var/log/app         # Directorio para logs (default: actual)

# Notificaciones por Nivel
NOTIFICATION_FATAL_LOG=1           # Enviar fatal logs a Slack (0 o 1)
NOTIFICATION_ERROR_LOG=1           # Enviar error logs a Slack (0 o 1)
NOTIFICATION_WARNING_LOG=1         # Enviar warning logs a Slack (0 o 1)
NOTIFICATION_INFO_LOG=1            # Enviar info logs a Slack (0 o 1)
NOTIFICATION_SUCCESS_LOG=1         # Enviar success logs a Slack (0 o 1)

# Configuración de Slack
NOTIFICATIONS_SLACK_ENABLED=1      # Habilitar notificaciones (0 o 1)
SLACK_TOKEN=xoxb-your-token        # Token de bot de Slack
SLACK_CHANNEL_ID=C1234567890       # ID del canal de Slack
```

### Variables de Entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `SAVE_LOG_FILE` | Habilitar logging a archivo | 0 |
| `LOG_FILE_NAME` | Nombre del archivo de log | log.txt |
| `LOG_FILE_PATH` | Directorio para logs | Directorio actual |
| `NOTIFICATION_FATAL_LOG` | Notificar fatal a Slack | 0 |
| `NOTIFICATION_ERROR_LOG` | Notificar errors a Slack | 0 |
| `NOTIFICATION_WARNING_LOG` | Notificar warnings a Slack | 0 |
| `NOTIFICATION_INFO_LOG` | Notificar info a Slack | 0 |
| `NOTIFICATION_SUCCESS_LOG` | Notificar success a Slack | 0 |
| `NOTIFICATIONS_SLACK_ENABLED` | Habilitar Slack | 0 |
| `SLACK_TOKEN` | Token de Slack | - |
| `SLACK_CHANNEL_ID` | ID del canal | - |

## Uso Básico

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    // Inicializar el logger
    go_logs.Init()
    defer go_logs.Close() // Cerrar archivo al finalizar

    // Uso básico
    go_logs.InfoLog("Servidor iniciado en el puerto 8080")
    go_logs.SuccessLog("Conexión a base de datos establecida")
    go_logs.WarningLog("Conexión a base de datos cerca del límite")
    go_logs.ErrorLog("Fallo al conectar a cache")
}
```

## Uso Avanzado

### Funciones con Formato

Similar a `fmt.Sprintf`, puedes usar las funciones con `f` al final:

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    go_logs.Init()

    // Formato con argumentos
    go_logs.Infof("Servidor iniciado en el puerto %d", 8080)
    go_logs.Successf("Conectado a %s en %dms", host, latency)
    go_logs.Errorf("Error al procesar solicitud de %s: %v", user, err)
    go_logs.Warningf("Usando deprecated API versión %s", apiVersion)
}
```

### Funciones con Contexto

Para distributed tracing y cancelación:

```go
package main

import (
    "context"
    "github.com/drossan/go_logs"
    "time"
)

func main() {
    go_logs.Init()

    ctx := context.Background()

    // Con contexto
    go_logs.InfoLogCtx(ctx, "Procesando solicitud")

    // Combinado: contexto + formato
    go_logs.InfoLogCtxf(ctx, "Usuario %s inició sesión", username)
    go_logs.ErrorLogCtxf(ctx, "Fallo en operación: %v", err)
}
```

### Ejemplo Completo con Cleanup

```go
package main

import (
    "github.com/drossan/go_logs"
    "os"
)

func main() {
    // Configurar variables de entorno
    os.Setenv("SAVE_LOG_FILE", "1")
    os.Setenv("LOG_FILE_NAME", "app.log")
    os.Setenv("NOTIFICATION_INFO_LOG", "1")

    // Inicializar
    go_logs.Init()
    defer go_logs.Close() // Importante: cerrar al finalizar

    // Verificar si Slack está configurado
    if go_logs.IsNotifierEnabled() {
        go_logs.InfoLog("Notificaciones de Slack activas")
    } else {
        go_logs.WarningLog("Slack no está configurado")
    }

    // Logging en diferentes niveles
    go_logs.InfoLog("Aplicación iniciada")
    go_logs.SuccessLog("Configuración cargada correctamente")

    // Simular operaciones
    if err := runApplication(); err != nil {
        go_logs.Errorf("Error fatal: %v", err)
        go_logs.FatalLog("Terminando aplicación")
    }
}

func runApplication() error {
    go_logs.Infof("Procesando %d registros", 1000)
    // ... lógica de la aplicación
    return nil
}
```

## API Completa

### Funciones Básicas

- `FatalLog(message string)` - Log fatal y termina el programa
- `ErrorLog(message string)` - Log de error
- `WarningLog(message string)` - Log de advertencia
- `InfoLog(message string)` - Log informativo
- `SuccessLog(message string)` - Log de éxito

### Funciones con Formato (fmt.Sprintf-like)

- `Fatalf(format string, args ...interface{})`
- `Errorf(format string, args ...interface{})`
- `Warningf(format string, args ...interface{})`
- `Infof(format string, args ...interface{})`
- `Successf(format string, args ...interface{})`

### Funciones con Contexto

- `ErrorLogCtx(ctx context.Context, message string)`
- `WarningLogCtx(ctx context.Context, message string)`
- `InfoLogCtx(ctx context.Context, message string)`
- `SuccessLogCtx(ctx context.Context, message string)`

### Funciones Combinadas (Contexto + Formato)

- `ErrorLogCtxf(ctx context.Context, format string, args ...interface{})`
- `WarningLogCtxf(ctx context.Context, format string, args ...interface{})`
- `InfoLogCtxf(ctx context.Context, format string, args ...interface{})`
- `SuccessLogCtxf(ctx context.Context, format string, args ...interface{})`

### Funciones de Control

- `Init()` - Inicializa el logger con configuración de entorno
- `Close()` - Cierra el archivo de log (flush de buffers)
- `IsNotifierEnabled() bool` - Verifica si Slack notifications están activas

## Mejoras de Rendimiento

v2.0+ incluye mejoras significativas de rendimiento:

- **Archivo persistente**: 98% más rápido (23,099ns → 400ns por operación)
- **Buffering con bufio**: 63% menos allocations de memoria
- **Thread-safe**: Protección con mutex para concurrencia
- **Caché de env vars**: Evita llamadas repetidas a os.Getenv()

## Testing

El paquete incluye más de 30 tests con 83.5% de cobertura:

```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar con cobertura
go test -cover ./...

# Ejecutar benchmarks
go test -bench=. -benchmem ./...
```

## Contribución

Las contribuciones son bienvenidas. Por favor:

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## Licencia

Este paquete se ofrece bajo los términos de la Licencia MIT. Lee el archivo LICENSE para más detalles.

## Changelog

### v2.0 (Última)

- ✅ **API mejorada**: Funciones con formato y soporte de contexto
- ✅ **Performance**: 98% más rápido con archivo persistente y buffering
- ✅ **WarningLog**: Función de advertencia agregada
- ✅ **Tests**: 83.5% de cobertura con 30+ tests
- ✅ **Thread-safe**: Protección contra race conditions
- ✅ **Documentación**: Godoc completa en todas las funciones exportadas

### v1.0

- Versión inicial con logging básico y notificaciones Slack
