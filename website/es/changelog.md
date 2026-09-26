---
outline: [2, 3]
---

# Changelog

Notas de versión de `github.com/drossan/go_logs`. Las versiones siguen [Semantic Versioning](https://semver.org/lang/es/); el módulo de Slack `github.com/drossan/go_logs/slack/v3` se etiqueta aparte como `slack/vX.Y.Z`.

## v3.1.0 (sin publicar)

Primera versión que se puede instalar de verdad como v3: `go get github.com/drossan/go_logs/v3@v3.1.0`.

::: warning v3.0.x nunca fue instalable bajo /v3
Los tags `v3.0.0`–`v3.0.4` declaran `module github.com/drossan/go_logs` (sin sufijo `/v3`) en su `go.mod`, así que el proxy de Go no puede servirlos como `github.com/drossan/go_logs/v3`. Siguen publicados pero no son válidos para `/v3`; usa `v3.1.0` o posterior.
:::

### 🚨 Cambios incompatibles

- La ruta del módulo pasa a ser `github.com/drossan/go_logs/v3`. `async`, `hooks`, `http`, `otel` y `signal` son paquetes de ese único módulo (`github.com/drossan/go_logs/v3/async`, …) en lugar de módulos separados.
- Slack sale del core: el notificador vive en el módulo aparte `github.com/drossan/go_logs/slack/v3` (`slack.NewNotifier`, `slack.NewNotifierFromEnv`) y el paquete `adapters` desaparece. La API v2 solo envía notificaciones tras `go_logs.SetNotifier(n)`; ver la [guía de migración](/es/migration-v2-to-v3) y [Configuración](/es/configuration).

### 🚀 Mejoras

- `go_logs.SetNotifier(domain.Notifier)` registra el notificador que usa la API v2 (`nil` lo desactiva).
- Interfaz `go_logs.Flusher`: los writers se vacían tras cada entrada. La implementan `RotatingFileWriter`, `EnhancedRotatingFileWriter`, `MultiWriter` y `SamplingWriter`.
- `Logger.Sync()` ahora vacía y sincroniza la salida, e ignora los errores `EINVAL`/`ENOTTY`/`EBADF` que devuelven terminales y tuberías.
- El nivel de log lo comparten un logger y todos sus hijos: `SetLevel` en cualquiera de ellos afecta a todo el árbol.

### 🩹 Correcciones

- `Init()` (y la auto-inicialización v2) nunca termina el proceso: las variables vacías usan su valor por defecto, los valores inválidos usan el defecto y avisan por stderr, y un fichero de log que no se puede abrir desactiva la salida a fichero con un aviso. Sin entorno el nivel por defecto es `info`, así que `InfoLog("x")` escribe en consola.
- `With()` ya no comparte memoria de campos entre loggers hermanos (una data race que podía loguear los campos de otra petición).
- Sin `fsync` por entrada al escribir a fichero (limitaba el logging a fichero a unos cientos de mensajes por segundo).
- `async`: `Sync()` en un hijo espera a sus entradas, `Close()` es idempotente en la raíz y no-op en los hijos, y `Fatal` en un hijo incluye sus campos.
- `GetCaller` informa del nombre de paquete sin el sufijo `/vN`.
- El paquete `hooks` compila en Windows (allí se excluye el hook de syslog).

### 📖 Documentación

- Este sitio de documentación (inglés y español), con los ejemplos de código corregidos.
- Añadido `LICENSE` (MIT).

### 🤖 CI

- Workflow que ejecuta `gofmt`, `go vet`, `go test -race` (módulo raíz y `slack/`) y una compilación cruzada a Windows en cada push y pull request.

## v3.0.0 – v3.0.4 (2026-03-02)

`v3.0.1`–`v3.0.4` solo cambiaron el workflow de release. `v3.0.0` introdujo la API v3; las funcionalidades de abajo se listaban antes como "v3.1"–"v3.5", pero todas salieron en `v3.0.0`.

### 🚀 Mejoras

- Interfaz `Logger` con campos estructurados tipados, child loggers (`With()`) y propagación de contexto (`trace_id`, `span_id`).
- `TextFormatter` y `JSONFormatter`, hooks extensibles, redacción de datos sensibles y un `RotatingFileWriter` sin dependencias.
- Información del caller y stack traces (`WithCaller`, `WithStackTrace`, `WithCallerLevel`).
- `MultiWriter`, rotación por tiempo (diaria/horaria), compresión gzip y retención `MaxAge` (`EnhancedRotatingFileWriter`).
- Sampling (`SamplingWriter`), logger global (`SetDefault`), utilidades de test (`CaptureBuffer`, `MockLogger`).
- Exportador OpenTelemetry (`otel`) y hooks de syslog, local y remoto.
- Métricas, logging asíncrono (`async`), nivel dinámico por HTTP (`http`) y rotación por SIGHUP (`signal`).
- La API global v2 sigue funcionando.

## v1.0.0 – v1.2.5 (2023-12-30 – 2024-05-14)

La API global legacy (`InfoLog`, `ErrorLog`, …), a la que esta documentación llama "v2".

- `v1.0.0`: versión inicial con logging básico.
- `v1.2.0`: logging centralizado y notificaciones Slack.
- `v1.2.1`–`v1.2.5`: correcciones en la configuración de notificaciones y en la gestión del fichero de log.
