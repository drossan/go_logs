# CLAUDE.md

Este archivo proporciona guía a Claude Code (claude.ai/code) para trabajar con el código en este repositorio.

## Resumen del Proyecto

Esta es una biblioteca de registro para Go (`go_logs`) que proporciona registro centralizado de eventos y errores con niveles de severidad configurables y notificaciones opcionales de Slack. El paquete sigue un patrón de arquitectura limpia con separación entre la lógica del dominio y los adaptadores externos.

## Arquitectura

La base de código está organizada en tres capas principales:

- **Paquete raíz** (`/`): Funcionalidad principal de registro con funciones de registro exportadas (FatalLog, ErrorLog, InfoLog, SuccessLog) y gestión de configuración mediante `Init()`
- **Capa de dominio** (`domain/`): Contiene la interfaz `Notifier` - un puerto para implementaciones de notificaciones
- **Capa de adaptadores** (`adapters/`): Implementa la interfaz `Notifier` - actualmente proporciona `SlackNotifier` para la integración con Slack

Patrón arquitectónico clave: El paquete usa inyección de dependencias donde `config.go` contiene una variable `notifier *adapters.SlackNotifier` que implementa la interfaz del dominio `Notifier`. Esto permite fácil extensión con otros proveedores de notificaciones (ej. Email, Discord) agregando nuevos adaptadores.

## Configuración

La configuración está basada en variables de entorno (ver `.env.example`). Llama a `Init()` antes de cualquier función de registro para:
- Cargar variables de entorno
- Configurar la salida al archivo de registro (opcional)
- Configurar qué niveles de registro activan notificaciones
- Inicializar el notificador de Slack (si está habilitado)

La función `Init()` debe ser llamada antes de usar cualquier función de registro, aunque `saveLog()` se auto-inicializará si es necesario.

## Comandos Comunes de Desarrollo

Compilación y dependencias:
```bash
go mod tidy        # Limpiar dependencias
go generate ./...  # Generar código (se ejecuta antes de las compilaciones vía goreleaser)
```

Pruebas:
```bash
go test ./...      # Ejecutar todas las pruebas (actualmente no existen archivos de prueba)
go test -v ./...   # Ejecutar con salida detallada
```

Ejecutar paquete específico:
```bash
go run .           # Ejecutar desde la raíz (si existe main.go)
```

## Proceso de Lanzamiento

El proyecto usa GoReleaser para los lanzamientos. Configuración en `.goreleaser.yaml`:
- Omite la compilación (paquete solo biblioteca)
- Ejecuta `go mod tidy` y `go generate ./...` como hooks previos al lanzamiento
- Crea lanzamientos en GitHub con changelog auto-generado
- El changelog excluye commits que empiezan con `docs:` o `test:`

## Organización del Código

- `logs.go`: API pública - funciones de registro exportadas con salida a terminal codificada por color
- `config.go`: Configuración e inicialización - carga variables de entorno, gestiona el estado global
- `save.go`: Enrutamiento interno - determina si los registros deben guardarse en archivo y/o enviarse como notificaciones
- `adapters/slack_notifier.go`: Integración con Slack - implementa la interfaz `Notifier`
- `domain/notification.go`: Definición del puerto - interfaz `Notifier` para extensibilidad

## Notas Importantes

- El paquete usa estado global a nivel de paquete (variables en `config.go`) - esto es intencional para una biblioteca de registro
- Los archivos de registro se abren/cierran para cada operación de escritura (ver `registerMessage()` en `save.go`)
- Las notificaciones de Slack usan la variable de entorno `SLACK_CHANEL_ID` (nota el error ortográfico - "CHANEL" vs "CHANNEL")
- FatalLog llama a `log.Fatal()` que termina el programa después de registrar
