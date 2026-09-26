---
id: go_logs-instalable-bugs-prod-06
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: [go_logs-instalable-bugs-prod-01, go_logs-instalable-bugs-prod-02]
estimate: 3h
actual: 0.2h
created: 2026-09-26
updated: 2026-09-26
---

# Slack fuera del core: submódulo `slack/`, `SetNotifier`, `config.go` sin `adapters`

## Description

`adapters/slack_notifier.go` importa `slack-go/slack`, que arrastra `gorilla/websocket`, y `config.go:5` importa `adapters`: cualquier consumidor que solo quiera loguear por consola se lleva un cliente de Slack en su árbol de dependencias y en su superficie de CVE, contradiciendo el "zero dependencies" del README. Hoy no existe ningún consumidor instalable de v3, así que es la única ventana en la que sacarlo no rompe a nadie (B2 del design-review). El core conservará la capacidad de notificar vía una interfaz inyectada. Ver plan, Objetivo 2.

## Spec

> **Desviación aprobada por el owner (2026-09-26, durante la ejecución):** la ruta del módulo es `github.com/drossan/go_logs/slack/v3` (tag `slack/v3.1.0`), no `github.com/drossan/go_logs/v3/slack`: con esa ruta Go rechaza `v3.1.0` (`should be v0 or v1, not v3`). La directiva `go` de `slack/go.mod` es `1.27.0`, no `1.21`: la impone el `replace ../`. Ver el session log y el Registro de cambios del plan.

- Crear `slack/` como submódulo:
  - `slack/go.mod`: `module github.com/drossan/go_logs/v3/slack`, `go 1.21`, `require github.com/drossan/go_logs/v3 v3.1.0` **y** `replace github.com/drossan/go_logs/v3 => ../` (patrón OpenTelemetry: el `replace` sirve al desarrollo local y se ignora en consumidores; el `require` real se resuelve tras el tag `v3.1.0`). `require github.com/slack-go/slack v0.12.5`. `go mod tidy` dentro de `slack/`.
  - Mover `adapters/slack_notifier.go` → `slack/notifier.go` (package `slack`) y su test. Renombrar `NewSlackNotifier()` → `NewNotifierFromEnv() (*Notifier, error)` conservando la lectura de `SLACK_TOKEN` y `SLACK_CHANNEL_ID`/`SLACK_CHANEL_ID`. Añadir `NewNotifier(token, channelID string) (*Notifier, error)` sin env.
  - `slack/README.md` breve con `go get github.com/drossan/go_logs/v3/slack` y ejemplo de uso con `go_logs.SetNotifier`.
- Borrar `adapters/` por completo. `domain/notification.go` se queda (interfaz `Notifier`).
- Raíz (`config.go`):
  - `var notifier domain.Notifier` (antes `*adapters.SlackNotifier`), protegido por el `RWMutex` que ya protege `notificationSettings` o uno propio.
  - Función exportada:
    ```go
    // SetNotifier registra el notificador usado por la API v2 (ErrorLog, FatalLog…)
    // cuando NOTIFICATIONS_SLACK_ENABLED está activo. nil lo desactiva. Los
    // notificadores concretos viven fuera del core (p. ej. go_logs/v3/slack).
    func SetNotifier(n domain.Notifier)
    ```
  - Eliminar `loadSlackConfig()` y el import de `adapters`. `isNotifierEnabled()` (o el equivalente en `config.go:387-390`) devuelve `notifier != nil && notificationsEnabled`.
  - Si `NOTIFICATIONS_SLACK_ENABLED=1` y no hay notificador registrado en el momento de enviar, emitir **una sola vez por proceso** (guardado con `sync.Once`, no se re-arma aunque luego se llame a `SetNotifier(nil)`) por `warnOutput` (tarea 02) el aviso `go_logs: NOTIFICATIONS_SLACK_ENABLED activo pero no hay notificador; llama a go_logs.SetNotifier (ver go_logs/v3/slack)` y continuar.
- `hooks/slack_hook.go`: actualizar su godoc (referencia hoy a `adapters.NewSlackNotifier()`, paquete que esta tarea borra) para que cite `github.com/drossan/go_logs/v3/slack` en su lugar. Sin cambios funcionales en ese fichero.
- `hooks/slack_hook.go` no cambia (depende solo de su interfaz `SlackNotifier`); `*slack.Notifier` la satisface. Añadir en `slack/notifier_test.go` la aserción `var _ hooks.SlackNotifier = (*Notifier)(nil)`.
- `go list -deps ./ | grep -c slack-go` en la raíz debe ser 0. `go.mod` raíz: quitar `slack-go/slack` y `gorilla/websocket` con `go mod tidy`.
- `MIGRATION.md`: sección "Slack en v3.1" con el antes (`NOTIFICATIONS_SLACK_ENABLED=1` bastaba) y el después (`import ".../v3/slack"`; `n, _ := slack.NewNotifierFromEnv(); go_logs.SetNotifier(n)`).
- Tests en raíz con un `fakeNotifier` que registra mensajes; en `slack/` los tests existentes del adaptador.

## Fuera de alcance

- Hacer que `SlackHook.Run` no bloquee el camino de log ni sincronizar su `SetLevel` (hallazgos de `hooks`): plan aparte.
- Reescribir el resto de documentación de Slack más allá de `MIGRATION.md`, `slack/README.md` y el ajuste de la sección de README raíz (tarea 10 pasa el repaso completo).
- Ejecutar `git tag slack/v3.1.0`: checklist de release del owner.
- Mover `hooks/` o `otel/` a submódulos.

## Scenarios (Gherkin)

```gherkin
Feature: El core no depende de Slack

  Scenario: El grafo de dependencias del paquete raíz no incluye slack-go
    Given el módulo raíz
    When se listan sus dependencias con `go list -deps ./`
    Then ninguna contiene "slack-go" ni "gorilla/websocket"

  Scenario: adapters ya no existe y todo compila
    Given el módulo raíz
    When se compila `go build ./...` y se ejecutan los tests
    Then no existe el directorio adapters
    And hooks/slack_hook.go compila sin cambios

Feature: Notificaciones v2 mediante un notificador inyectado

  Scenario: Con notificador registrado y notificaciones activas se envía el mensaje
    Given NOTIFICATIONS_SLACK_ENABLED=1 y NOTIFICATION_ERROR_LOG=1
    And un notificador de prueba registrado con SetNotifier
    When se registra "fallo" con ErrorLog
    Then el notificador de prueba recibió un mensaje que contiene "fallo"

  Scenario: Sin notificador registrado no se envía nada y se avisa una sola vez
    Given NOTIFICATIONS_SLACK_ENABLED=1 y NOTIFICATION_ERROR_LOG=1
    And ningún notificador registrado
    When se registran dos errores con ErrorLog
    Then el proceso sigue vivo
    And el aviso en stderr sobre SetNotifier aparece exactamente una vez
    And ambos errores se escriben en la salida estándar

  Scenario: El aviso no se re-arma tras alternar el notificador
    Given NOTIFICATIONS_SLACK_ENABLED=1 y NOTIFICATION_ERROR_LOG=1
    And ya se emitió el aviso de "no hay notificador" una vez en este proceso
    When se registra un notificador con SetNotifier, se desregistra con SetNotifier(nil) y se registra otro error
    Then el aviso no se vuelve a emitir

  Scenario: Un notificador que falla no interrumpe el registro del error
    Given un notificador de prueba cuyo SendNotification siempre devuelve error
    And NOTIFICATIONS_SLACK_ENABLED=1 con notificaciones de error activas
    When se registra un error con ErrorLog
    Then el error se escribe en la salida estándar igualmente
    And el proceso sigue vivo

  Scenario: SetNotifier con nil desactiva las notificaciones
    Given un notificador de prueba registrado
    When se registra nil con SetNotifier y luego un error con ErrorLog
    Then el notificador de prueba no recibe ningún mensaje

  Scenario: Con notificaciones desactivadas el notificador registrado no se usa
    Given NOTIFICATIONS_SLACK_ENABLED=0 y un notificador de prueba registrado
    When se registra un error con ErrorLog
    Then el notificador de prueba no recibe ningún mensaje

  Scenario: Registro concurrente de notificador y envío de errores
    Given 20 goroutines alternando SetNotifier y ErrorLog
    When terminan
    Then el race detector no reporta ninguna carrera

Feature: Submódulo slack

  Scenario: El submódulo se identifica y compila contra el core local
    Given el directorio slack/
    When se ejecuta `go list -m` y `go test ./...` dentro de slack/
    Then el módulo es "github.com/drossan/go_logs/v3/slack"
    And los tests pasan

  Scenario Outline: Construcción del notificador desde entorno
    Given SLACK_TOKEN=<token> y SLACK_CHANNEL_ID=<canal>
    When se construye con NewNotifierFromEnv
    Then el resultado es <resultado>

    Examples:
      | token     | canal   | resultado                       |
      | ""        | ""      | error que menciona SLACK_TOKEN  |
      | "xoxb-1"  | ""      | error que menciona el canal     |
      | "xoxb-1"  | "C123"  | notificador no nulo             |

  Scenario: El canal legado SLACK_CHANEL_ID se usa cuando falta el nombre correcto
    Given SLACK_TOKEN="xoxb-1", SLACK_CHANNEL_ID="" y SLACK_CHANEL_ID="C999"
    When se construye con NewNotifierFromEnv
    Then el resultado es un notificador no nulo con canal "C999"

  Scenario Outline: Construcción directa del notificador sin pasar por el entorno
    Given token=<token> y canal=<canal> pasados directamente a NewNotifier
    When se construye el notificador
    Then el resultado es <resultado>

    Examples:
      | token    | canal | resultado                   |
      | ""       | ""    | error que menciona el token |
      | "xoxb-1" | ""    | error que menciona el canal |
      | "xoxb-1" | "C1"  | notificador no nulo         |

  Scenario: El notificador satisface la interfaz del hook v3
    Given el tipo slack.Notifier
    When se comprueba en compilación que implementa hooks.SlackNotifier
    Then compila

  Scenario: Un consumidor externo importa el submódulo slack
    Given un módulo temporal con replace de go_logs/v3 y de go_logs/v3/slack a las rutas locales
    And un main.go que importa ambos y llama a go_logs.SetNotifier(slack.NewNotifier(...))
    When se compila
    Then compila sin error
```

## Provides

- `go_logs.SetNotifier(domain.Notifier)` en el core; `github.com/drossan/go_logs/v3/slack` con `NewNotifier` y `NewNotifierFromEnv`.
- Módulo raíz sin `slack-go` ni `gorilla/websocket`: la tarea 07 ejecuta CI en raíz y en `slack/`; la tarea 10 documenta y verifica `go list -deps`.

## Definition of Done

- [x] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor
- [x] Cada escenario Gherkin tiene al menos un test (camino feliz + bordes/errores)
- [x] Todos los tests en verde: `go test -race ./...` en la raíz y `(cd slack && go test -race ./...)`
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `gofmt -l` sin salida en raíz y `slack/`; `go vet ./...` en ambos
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **godoc en el código** — `SetNotifier`, `slack.Notifier`, `NewNotifier`, `NewNotifierFromEnv`, package doc de `slack`
  - [x] **Doc técnica (contexto)** — `MIGRATION.md` ("Slack en v3.1"), `slack/README.md`, sección Slack del README raíz y CLAUDE.md (estructura: un módulo + `slack/`)
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-06.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-06: <conventional commit>`
