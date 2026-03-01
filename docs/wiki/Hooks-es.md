# Hooks

Los hooks proporcionan un punto de extension para procesar entradas de log antes de ser escritas. Usa hooks para enviar logs a sistemas externos, recopilar metricas, disparar alertas o implementar logica de procesamiento personalizada.

**[English](Hooks.md)** | **Espanol**

## Interfaz Hook

Todos los hooks implementan la interfaz `Hook`:

```go
type Hook interface {
    Run(entry *Entry) error
}
```

El metodo `Run` se llama para cada entrada de log que pasa el filtro de nivel. Si el hook retorna un error, se loguea pero no impide que otros hooks o la salida procedan.

## Tipo Entry

Los hooks reciben una `Entry` con toda la informacion del log:

```go
type Entry struct {
    Level      Level
    Message    string
    Fields     []Field
    Timestamp  time.Time
    Caller     *CallerInfo
    StackTrace []byte
}
```

### Metodos de Entry

```go
// Verificar si tiene un campo
entry.HasField("user_id")

// Obtener valor de un campo
entry.GetFieldValue("user_id")

// Clonar la entrada para modificacion
cloned := entry.Clone()

// Agregar mas campos
modified := entry.WithFields(go_logs.String("extra", "valor"))

// Verificar nivel
if entry.Level >= go_logs.ErrorLevel {
    // Esto es un error o peor
}
```

## Crear Hooks

### Usando HookFunc

La forma mas simple de crear un hook es usando `HookFunc` o `NewFuncHook`:

```go
hook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
    fmt.Printf("Log: [%s] %s\n", entry.Level, entry.Message)
    return nil
})

logger, _ := go_logs.New(
    go_logs.WithHooks(hook),
)
```

### Implementando la Interfaz Hook

Para hooks con estado, implementa la interfaz directamente:

```go
type MetricsHook struct {
    counter *prometheus.CounterVec
}

func NewMetricsHook() *MetricsHook {
    return &MetricsHook{
        counter: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "log_entries_total",
                Help: "Numero total de entradas de log",
            },
            []string{"level"},
        ),
    }
}

func (h *MetricsHook) Run(entry *go_logs.Entry) error {
    h.counter.WithLabelValues(entry.Level.String()).Inc()
    return nil
}

// Uso
logger, _ := go_logs.New(
    go_logs.WithHooks(NewMetricsHook()),
)
```

## Patrones de Hook Integrados

### Hook de Recopilacion de Metricas

Rastrear conteos de logs por nivel:

```go
type MetricsHook struct {
    counts map[go_logs.Level]int64
    mu     sync.RWMutex
}

func NewMetricsHook() *MetricsHook {
    return &MetricsHook{
        counts: make(map[go_logs.Level]int64),
    }
}

func (h *MetricsHook) Run(entry *go_logs.Entry) error {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.counts[entry.Level]++
    return nil
}

func (h *MetricsHook) GetCount(level go_logs.Level) int64 {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.counts[level]
}
```

### Hook de Alertas de Error

Enviar alertas para errores:

```go
type AlertHook struct {
    notifier Notifier
    levels   map[go_logs.Level]bool
}

func NewAlertHook(notifier Notifier) *AlertHook {
    return &AlertHook{
        notifier: notifier,
        levels: map[go_logs.Level]bool{
            go_logs.ErrorLevel: true,
            go_logs.FatalLevel: true,
        },
    }
}

func (h *AlertHook) Run(entry *go_logs.Entry) error {
    if !h.levels[entry.Level] {
        return nil // Saltar niveles que no son error
    }

    // Construir mensaje de alerta
    msg := fmt.Sprintf("[%s] %s", entry.Level, entry.Message)

    // Agregar campos
    for _, field := range entry.Fields {
        msg += fmt.Sprintf("\n  %s: %v", field.Key(), field.Value())
    }

    // Enviar alerta asincronamente
    go h.notifier.Send(msg)
    return nil
}
```

### Hook de Filtrado

Filtrar logs basado en contenido:

```go
type FilterHook struct {
    blockedKeys map[string]bool
}

func NewFilterHook(blockedKeys ...string) *FilterHook {
    h := &FilterHook{
        blockedKeys: make(map[string]bool),
    }
    for _, key := range blockedKeys {
        h.blockedKeys[key] = true
    }
    return h
}

func (h *FilterHook) Run(entry *go_logs.Entry) error {
    // Remover campos bloqueados
    for i := len(entry.Fields) - 1; i >= 0; i-- {
        if h.blockedKeys[entry.Fields[i].Key()] {
            entry.Fields = append(entry.Fields[:i], entry.Fields[i+1:]...)
        }
    }
    return nil
}
```

## Hook de Slack

Enviar logs de error a Slack:

```go
type SlackHook struct {
    webhookURL string
    channel    string
    username   string
    levels     map[go_logs.Level]bool
}

func NewSlackHook(webhookURL, channel, username string) *SlackHook {
    return &SlackHook{
        webhookURL: webhookURL,
        channel:    channel,
        username:   username,
        levels: map[go_logs.Level]bool{
            go_logs.ErrorLevel: true,
            go_logs.FatalLevel: true,
        },
    }
}

func (h *SlackHook) Run(entry *go_logs.Entry) error {
    if !h.levels[entry.Level] {
        return nil
    }

    // Construir mensaje de Slack
    message := map[string]interface{}{
        "channel": h.channel,
        "username": h.username,
        "text": fmt.Sprintf("[%s] %s", entry.Level, entry.Message),
    }

    // Enviar a Slack
    return h.send(message)
}

// Uso
func main() {
    slackHook := NewSlackHook(
        "https://hooks.slack.com/services/XXX/YYY/ZZZ",
        "#alertas",
        "Logger",
    )

    logger, _ := go_logs.New(
        go_logs.WithHooks(slackHook),
    )

    logger.Error("Fallo conexion a base de datos", go_logs.Err(err))
}
```

## Multiples Hooks

Agregar multiples hooks a un logger:

```go
logger, _ := go_logs.New(
    go_logs.WithHooks(
        NewMetricsHook(),
        NewAlertHook(notifier),
        NewEnrichmentHook(),
    ),
)
```

Los hooks se ejecutan en el orden en que se agregan.

## Manejo de Errores de Hook

Si un hook retorna un error, se loguea pero no detiene otros hooks o la salida:

```go
func (h *MiHook) Run(entry *go_logs.Entry) error {
    if err := h.process(entry); err != nil {
        // Loguear el error pero no fallar
        log.Printf("error de hook: %v", err)
        return err
    }
    return nil
}
```

## Mejores Practicas

### 1. Mantener Hooks Rapidos

Los hooks se ejecutan sincronamente (a menos que los hagas async). Evita operaciones lentas:

```go
// Mal: Llamada HTTP bloqueante
func (h *Hook) Run(entry *go_logs.Entry) error {
    http.Post(url, "application/json", body) // Bloquea!
    return nil
}

// Bien: Procesamiento async
func (h *Hook) Run(entry *go_logs.Entry) error {
    go h.sendAsync(entry.Clone())
    return nil
}
```

### 2. Manejar Errores Con Elegancia

No dejes que los errores de hook crashee tu aplicacion:

```go
func (h *Hook) Run(entry *go_logs.Entry) error {
    if err := h.process(entry); err != nil {
        // Loguear pero no propagar
        fmt.Fprintf(os.Stderr, "error de hook: %v\n", err)
        return nil // o return err para loguearlo
    }
    return nil
}
```

### 3. Clonar Entradas para Modificacion

Si modificas entradas o las usas asincronamente, clonalas:

```go
func (h *Hook) Run(entry *go_logs.Entry) error {
    // Clonar para uso async
    cloned := entry.Clone()

    go func() {
        // Seguro usar cloned
        h.process(cloned)
    }()

    return nil
}
```

### 4. Filtrar Temprano

Verifica condiciones antes de operaciones costosas:

```go
func (h *SlackHook) Run(entry *go_logs.Entry) error {
    // Saltar no-errores inmediatamente
    if entry.Level < go_logs.ErrorLevel {
        return nil
    }

    // Solo ahora hacer trabajo costoso
    return h.sendToSlack(entry)
}
```

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Documentacion de interfaz Hook
- [Ejemplos](Examples-es.md) - Mas ejemplos de hooks
- [Configuracion](Configuration-es.md) - Configurar hooks via opciones
