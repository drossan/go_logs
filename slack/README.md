# go_logs/slack

Notificador de Slack para [go_logs](../README.md). Es un módulo aparte (`github.com/drossan/go_logs/slack/v3`) para que el core no arrastre `slack-go/slack` ni `gorilla/websocket` a quien no usa Slack.

```bash
go get github.com/drossan/go_logs/slack/v3
```

## Uso

```go
import (
    "log"

    "github.com/drossan/go_logs/v3"
    "github.com/drossan/go_logs/v3/hooks"
    "github.com/drossan/go_logs/slack/v3"
)

func main() {
    // Desde SLACK_TOKEN y SLACK_CHANNEL_ID (o el legado SLACK_CHANEL_ID)…
    n, err := slack.NewNotifierFromEnv()
    // …o explícito: n, err := slack.NewNotifier("xoxb-...", "C1234567890")
    if err != nil {
        log.Printf("Slack desactivado: %v", err) // n queda deshabilitado: no envía nada
    }

    // API v2: ErrorLog, FatalLog… se notifican con NOTIFICATIONS_SLACK_ENABLED=1
    // y NOTIFICATION_<NIVEL>_LOG=1.
    go_logs.SetNotifier(n)
    go_logs.ErrorLog("fallo")

    // API v3: hook por nivel.
    logger, _ := go_logs.New(go_logs.WithHooks(hooks.NewSlackHook(n, go_logs.ErrorLevel)))
    logger.Error("fallo")
}
```

- Si falta el token o el canal, los constructores devuelven un `*Notifier` **deshabilitado** (no nil) y un error comparable con `errors.Is(err, slack.ErrTokenMissing)` / `slack.ErrChannelMissing`. El valor del token nunca aparece en errores ni avisos.
- `SendNotificationWithAttachments` envía `[]slack.Attachment` de `github.com/slack-go/slack`.

## Desarrollo

`go.mod` lleva `require github.com/drossan/go_logs/v3 v3.1.0` y `replace github.com/drossan/go_logs/v3 => ../`. El `replace` solo se aplica al trabajar dentro del repo; quien consume el módulo lo ignora y resuelve el `require`. Los tests se lanzan desde este directorio:

```bash
cd slack && go test -race ./...
```

La versión se publica con el tag `slack/v3.1.0`.
