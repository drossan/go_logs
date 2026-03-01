# Referencia API

Documentacion completa de referencia para la API de go_logs v3.

**[English](API-Reference.md)** | **Espanol**

## Tabla de Contenidos

- [Interfaz Logger](#interfaz-logger)
- [Crear Loggers](#crear-loggers)
- [Niveles de Log](#niveles-de-log)
- [Campos Estructurados](#campos-estructurados)
- [Opciones de Configuracion](#opciones-de-configuracion)
- [Formateadores](#formateadores)
- [Funciones de Contexto](#funciones-de-contexto)
- [Hooks](#hooks)
- [Rotacion de Archivos](#rotacion-de-archivos)
- [Tipo Entry](#tipo-entry)
- [Logger Global](#logger-global)
- [API Legacy v2](#api-legacy-v2)

---

## Interfaz Logger

La interfaz `Logger` es la abstraccion principal para todas las operaciones de logging.

```go
type Logger interface {
    // Metodos principales de logging
    Log(level Level, msg string, fields ...Field)
    LogCtx(ctx context.Context, level Level, msg string, fields ...Field)

    // Metodos de conveniencia para cada nivel
    Trace(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)

    // Creacion de child logger
    With(fields ...Field) Logger

    // Configuracion
    SetLevel(level Level)
    GetLevel() Level

    // Ciclo de vida
    Sync() error
}
```

### Log

Registra un mensaje en el nivel especificado con campos estructurados.

```go
func (l Logger) Log(level Level, msg string, fields ...Field)
```

**Parametros:**
- `level` - El nivel de severidad (TraceLevel, DebugLevel, etc.)
- `msg` - El mensaje de log
- `fields` - Cero o mas campos estructurados

**Ejemplo:**

```go
logger.Log(go_logs.InfoLevel, "Usuario conectado",
    go_logs.String("username", "juan"),
    go_logs.String("ip", "192.168.1.1"),
)
```

### LogCtx

Registra un mensaje con soporte de contexto para tracing distribuido.

```go
func (l Logger) LogCtx(ctx context.Context, level Level, msg string, fields ...Field)
```

**Parametros:**
- `ctx` - Contexto que contiene trace_id, span_id, etc.
- `level` - El nivel de severidad
- `msg` - El mensaje de log
- `fields` - Cero o mas campos estructurados

**Ejemplo:**

```go
ctx := go_logs.WithTraceID(context.Background(), "trace-123")
logger.LogCtx(ctx, go_logs.InfoLevel, "Procesando peticion")
```

### Trace

Registra en TraceLevel. Usar para logging extremadamente detallado y de alto volumen.

```go
func (l Logger) Trace(msg string, fields ...Field)
```

**Ejemplo:**

```go
logger.Trace("Funcion entrada",
    go_logs.String("function", "procesarDatos"),
    go_logs.Any("args", args),
)
```

### Debug

Registra en DebugLevel. Usar para informacion de diagnostico detallada.

```go
func (l Logger) Debug(msg string, fields ...Field)
```

**Ejemplo:**

```go
logger.Debug("Procesando peticion",
    go_logs.String("endpoint", "/api/usuarios"),
    go_logs.String("method", "GET"),
)
```

### Info

Registra en InfoLevel. Usar para mensajes informativos generales.

```go
func (l Logger) Info(msg string, fields ...Field)
```

**Ejemplo:**

```go
logger.Info("Servidor iniciado",
    go_logs.String("host", "localhost"),
    go_logs.Int("port", 8080),
)
```

### Warn

Registra en WarnLevel. Usar para situaciones potencialmente dainosas.

```go
func (l Logger) Warn(msg string, fields ...Field)
```

**Ejemplo:**

```go
logger.Warn("Uso de memoria alto",
    go_logs.Float64("usage_percent", 85.5),
    go_logs.Int64("bytes_used", 17179869184),
)
```

### Error

Registra en ErrorLevel. Usar para eventos de error.

```go
func (l Logger) Error(msg string, fields ...Field)
```

**Ejemplo:**

```go
err := errors.New("conexion rechazada")
logger.Error("Fallo conexion a base de datos",
    go_logs.Err(err),
    go_logs.String("host", "db.example.com"),
)
```

### Fatal

Registra en FatalLevel y termina la aplicacion con `os.Exit(1)`.

```go
func (l Logger) Fatal(msg string, fields ...Field)
```

**Ejemplo:**

```go
if config == nil {
    logger.Fatal("No se puede cargar configuracion",
        go_logs.String("file", "/etc/app/config.yaml"),
    )
}
```

**Advertencia**: `Fatal` termina la aplicacion. Usar solo para errores irrecuperables.

### With

Crea un child logger con campos pre-agregados.

```go
func (l Logger) With(fields ...Field) Logger
```

**Retorna:** Una nueva instancia de Logger con configuracion y campos heredados.

**Ejemplo:**

```go
baseLogger, _ := go_logs.New()

// Crear child logger con contexto de peticion
requestLogger := baseLogger.With(
    go_logs.String("request_id", "req-123"),
    go_logs.String("user_id", "user-456"),
)

// Todos los logs de requestLogger incluyen request_id y user_id
requestLogger.Info("Procesando peticion")
requestLogger.Info("Peticion completada")
```

### SetLevel

Cambia el umbral minimo de nivel de log en tiempo de ejecucion.

```go
func (l Logger) SetLevel(level Level)
```

**Ejemplo:**

```go
// Habilitar debug logging dinamicamente
if debugMode {
    logger.SetLevel(go_logs.DebugLevel)
}
```

### GetLevel

Retorna el umbral minimo de nivel de log actual.

```go
func (l Logger) GetLevel() Level
```

**Ejemplo:**

```go
if logger.GetLevel() >= go_logs.DebugLevel {
    // Debug logging esta habilitado, hacer logging costoso
    logger.Debug("Estado detallado", go_logs.Any("state", volcarEstado()))
}
```

### Sync

Vacía cualquier entrada de log en buffer. Llamar antes de salir de la aplicacion.

```go
func (l Logger) Sync() error
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("app.log", 100, 5),
)
defer logger.Sync() // Asegurar que todos los logs se escriban
```

---

## Crear Loggers

### New

Crea un nuevo Logger con las opciones dadas.

```go
func New(opts ...Option) (Logger, error)
```

**Parametros:**
- `opts` - Cero o mas funciones Option

**Retorna:**
- `Logger` - Instancia de logger configurada
- `error` - Error de configuracion (ej: ruta de archivo invalida)

**Configuracion por Defecto:**
- Nivel: InfoLevel
- Salida: os.Stdout
- Formateador: TextFormatter
- Sin hooks
- Sin redaccion

**Ejemplo:**

```go
// Logger basico con valores por defecto
logger, err := go_logs.New()
if err != nil {
    panic(err)
}

// Logger configurado
logger, err := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
    go_logs.WithCaller(true),
)
```

---

## Niveles de Log

Los niveles de log siguen el ordenamiento numerico estilo syslog para filtrado de umbral rapido.

### Tipo Level

```go
type Level int
```

### Constantes de Level

```go
const (
    TraceLevel   Level = 10  // Extremadamente detallado
    DebugLevel   Level = 20  // Informacion de diagnostico
    InfoLevel    Level = 30  // Mensajes operacionales generales
    SuccessLevel Level = 25  // Operaciones exitosas (compatibilidad v2)
    WarnLevel    Level = 40  // Situaciones potencialmente dainosas
    ErrorLevel   Level = 50  // Eventos de error
    FatalLevel   Level = 60  // Errores criticos (termina app)
    SilentLevel  Level = 0   // Deshabilita todo logging
)
```

### ParseLevel

Convierte un string a un Level.

```go
func ParseLevel(levelStr string) Level
```

**Valores soportados:** trace, debug, info, warn, warning, error, fatal, silent, none, disable

**Ejemplo:**

```go
level := go_logs.ParseLevel("debug") // Retorna DebugLevel
level := go_logs.ParseLevel("WARNING") // Retorna WarnLevel (case-insensitive)
level := go_logs.ParseLevel("unknown") // Retorna InfoLevel (default)
```

### Level.String

Retorna la representacion en string.

```go
func (l Level) String() string
```

**Ejemplo:**

```go
level := go_logs.InfoLevel
fmt.Println(level.String()) // Output: INFO
```

---

## Campos Estructurados

Los campos son pares clave-valor tipados para logging estructurado. Ver [Logging Estructurado](Structured-Logging-es.md) para documentacion detallada.

### Funciones de Field

```go
func String(key, val string) Field
func Int(key string, val int) Field
func Int64(key string, val int64) Field
func Float64(key string, val float64) Field
func Bool(key string, val bool) Field
func Err(err error) Field
func Any(key string, val interface{}) Field
```

---

## Opciones de Configuracion

Las opciones configuran el logger usando el patron de opciones funcionales.

### WithLevel

Establece el umbral minimo de nivel de log.

```go
func WithLevel(level Level) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
)
```

### WithOutput

Establece el destino de salida.

```go
func WithOutput(w io.Writer) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithOutput(os.Stdout),
)
```

### WithFormatter

Establece el formateador de log.

```go
func WithFormatter(f Formatter) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### WithHooks

Agrega hooks para procesamiento personalizado de logs.

```go
func WithHooks(hooks ...Hook) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithHooks(miHookPersonalizado),
)
```

### WithCommonRedaction

Habilita la redaccion de campos sensibles comunes.

```go
func WithCommonRedaction() Option
```

Redacta: password, passwd, pwd, token, api_key, apikey, api-key, secret, authorization, auth, cookie, session, credit_card, ssn, social_security

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithCommonRedaction(),
)
logger.Info("Login usuario", go_logs.String("password", "secreto123"))
// Output: password=***
```

### WithCaller

Habilita informacion del caller (archivo:linea funcion).

```go
func WithCaller(enabled bool) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithCaller(true),
)
logger.Info("Hola")
// Output: [2026/02/28 10:30:00] INFO main.go:42 main.Hola Hola
```

### WithRotatingFile

Crea un writer de archivo con rotacion.

```go
func WithRotatingFile(filename string, maxSizeMB int, maxBackups int) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
)
```

---

## Funciones de Contexto

Funciones para gestionar valores de contexto para tracing distribuido.

### WithTraceID

Agrega un trace ID al contexto.

```go
func WithTraceID(ctx context.Context, traceID string) context.Context
```

### WithSpanID

Agrega un span ID al contexto.

```go
func WithSpanID(ctx context.Context, spanID string) context.Context
```

### WithRequestID

Agrega un request ID al contexto.

```go
func WithRequestID(ctx context.Context, requestID string) context.Context
```

### WithUserID

Agrega un user ID al contexto.

```go
func WithUserID(ctx context.Context, userID string) context.Context
```

---

## Hooks

### Interfaz Hook

```go
type Hook interface {
    Run(entry *Entry) error
}
```

### NewFuncHook

Crea un Hook desde una funcion.

```go
func NewFuncHook(fn func(entry *Entry) error) Hook
```

**Ejemplo:**

```go
hook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
    // Enviar errores a monitoreo
    if entry.Level >= go_logs.ErrorLevel {
        monitoreo.TrackError(entry.Message)
    }
    return nil
})

logger, _ := go_logs.New(
    go_logs.WithHooks(hook),
)
```

---

## Rotacion de Archivos

### NewRotatingFileWriter

Crea un writer de archivo con rotacion.

```go
func NewRotatingFileWriter(filename string, maxSizeMB int, maxBackups int) (*RotatingFileWriter, error)
```

**Parametros:**
- `filename` - Ruta al archivo de log
- `maxSizeMB` - Tamano maximo en MB antes de rotar
- `maxBackups` - Numero maximo de archivos de backup a mantener

**Ejemplo:**

```go
writer, err := go_logs.NewRotatingFileWriter("app.log", 100, 5)
if err != nil {
    panic(err)
}
defer writer.Close()

logger, _ := go_logs.New(
    go_logs.WithOutput(writer),
)
```

---

## Tipo Entry

Entry representa una entrada de log completa.

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

---

## Logger Global

### SetGlobal

Establece la instancia de logger global.

```go
func SetGlobal(logger Logger)
```

### GetGlobal

Obtiene la instancia de logger global.

```go
func GetGlobal() Logger
```

### Funciones de Conveniencia Global

```go
func Global() Logger                    // Retorna logger global
func L() Logger                         // Abreviatura de Global()
```

---

## API Legacy v2

La API v2 es completamente compatible hacia atras y se mantiene junto a v3.

### Inicializacion

```go
func Init()
func Close()
```

### Funciones de Logging

```go
func TraceLog(msg string)
func DebugLog(msg string)
func InfoLog(msg string)
func SuccessLog(msg string)
func WarningLog(msg string)
func ErrorLog(msg string)
func FatalLog(msg string)
```

### Logging Formateado

```go
func Tracef(format string, args ...interface{})
func Debugf(format string, args ...interface{})
func Infof(format string, args ...interface{})
func Successf(format string, args ...interface{})
func Warningf(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Fatalf(format string, args ...interface{})
```

---

## Ver Tambien

- [Logging Estructurado](Structured-Logging-es.md) - Documentacion detallada de campos
- [Formateadores](Formatters-es.md) - Configuracion de formateadores
- [Hooks](Hooks-es.md) - Documentacion del sistema de hooks
- [Configuracion](Configuration-es.md) - Variables de entorno
