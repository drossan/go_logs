# Paquete Go_Logs
## Descripción

Este paquete está diseñado para manejar el registro de eventos y errores de una manera centralizada en aplicaciones Go. Proporciona funcionalidades para registrar mensajes de diferentes niveles de severidad e incluso permite el envío de notificaciones a través de Slack.

El paquete ofrece los siguientes niveles de registro:
- FatalLog: Registros de errores fatales que requieren que el programa se detenga.
- ErrorLog: Registros de errores que no requieren que el programa se detenga.
- InfoLog: Registros de eventos informativos de la aplicación.
- SuccessLog: Registro de operaciones exitosas en la aplicación.
 
### Configuración
La configuración se maneja a través de variables de entorno, como se muestra en el archivo .env.example:

```dotenv
# ENABLED LOG FILE
SAVE_LOG_FILE=1
LOG_FILE_NAME=log.txt
LOG_FILE_PATH=

# NOTIFICATION LVL
NOTIFICATION_FATAL_LOG=1
NOTIFICATION_ERROR_LOG=1
NOTIFICATION_WARNING_LOG=1
NOTIFICATION_INFO_LOG=1
NOTIFICATION_SUCCESS_LOG=1

# SLACK
NOTIFICATIONS_SLACK_ENABLED=1
SLACK_TOKEN=YOUR_TOKEN
SLACK_CHANNEL_ID=CHANNEL_ID
```

Aquí está el desglose de las variables de configuración por sección:

- ENABLED LOG FILE: Controla si se creará, o no, el archivo de registro.
    - SAVE_LOG_FILE: Un valor de 1 habilita la creación del archivo de registro. Cualquier otro valor lo deshabilitará.
    - LOG_FILE_NAME: Nombre del archivo de registro.
    - LOG_FILE_PATH: Ruta donde se guardará el archivo de registro.
- NOTIFICATION LVL: Controla el nivel de eventos que generarán una notificación.
    - NOTIFICATION_FATAL_LOG: Un valor de 1 habilita las notificaciones para eventos fatales. Cualquier otro valor las desactivará.
    - NOTIFICATION_ERROR_LOG: Un valor de 1 habilita las notificaciones para eventos de error. Cualquier otro valor las desactivará.
    - NOTIFICATION_WARNING_LOG: Un valor de 1 habilita las notificaciones para eventos de advertencia. Cualquier otro valor las desactivará.
    - NOTIFICATION_INFO_LOG: Un valor de 1 habilita las notificaciones para eventos informativos. Cualquier otro valor las desactivará.
    - NOTIFICATION_SUCCESS_LOG: Un valor de 1 habilita las notificaciones para eventos de éxito. Cualquier otro valor las desactivará.
- SLACK: Controla la configuración para las notificaciones de Slack.
    - NOTIFICATIONS_SLACK_ENABLED: Un valor de 1 habilita las notificaciones de Slack. Cualquier otro valor las deshabilitará.
    - SLACK_TOKEN: El token de tu aplicación Slack.
    - SLACK_CHANNEL_ID: ID del canal donde se enviarán las notificaciones.

### Uso
Para utilizar este paquete, en primer lugar, necesitarás importarlo y luego llamar a la función Init() para inicializar el registrador con tu configuración.

```Go
import (
    "github.com/your_username/go_logs"
)

func main() {
    go_logs.Init()

    /* ... */

    go_logs.InfoLog("This is an info message")
    go_logs.SuccessLog("This is a success message")
    go_logs.ErrorLog("This is an error message")
    go_logs.FatalLog("This is a fatal message")
}
```

### Contribución
Cualquier contribución a este paquete es bienvenida, ya sea corrigiendo errores, añadiendo nuevas funcionalidades, o mejorando la documentación. Por favor, abra una nueva solicitud de Pull Requests para proponer cambios.

### Licencia
Este paquete se ofrece bajo los términos de la Licencia MIT.
Por favor, lee el archivo LICENSE para más detalles.