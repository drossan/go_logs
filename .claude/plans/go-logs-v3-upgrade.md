# Go Logs v3 Upgrade Plan

**Project**: go_logs v3 - Modernización de Biblioteca de Logging
**Author**: Planning Agent
**Date**: 2025-02-28
**Status**: Draft

---

## Resumen Ejecutivo

Este plan detalla la modernización de `go_logs` v2 a v3, transformándolo de una biblioteca de logging básica con funciones globales a una solución moderna comparable a zap, zerolog o logrus. El upgrade mantiene **backward compatibility total** con v2 mientras introduce 10 mejoras arquitectónicas significativas.

### Objetivos Principales

1. **Interfaz Logger** - Permitir inyección de dependencias y múltiples instancias
2. **Campos Estructurados** - Logging tipado con Fields minimizando allocations
3. **Child Loggers** - Logger con contexto inyectado (request_id, user_id, etc.)
4. **Contexto Real** - Extraer trace_id, span_id de context.Context automáticamente
5. **Dual Formatter** - TextFormatter (dev) y JSONFormatter (prod)
6. **Rotación de Archivos** - RotatingFileWriter sin dependencias externas
7. **Hooks Extensibles** - Slack como hook más, permitir Sentry, métricas, etc.
8. **Redactor** - Enmascarar datos sensibles automáticamente
9. **Filtrado por Nivel** - SetLevel() en runtime + LOG_LEVEL env var
10. **Compatibilidad v2** - Funciones globales siguen funcionando

### Estimación Global

- **Tiempo Total Estimado**: 40-60 horas de desarrollo
- **Fases**: 7 fases secuenciales
- **Riesgo**: Medio (mitigado con backward compatibility y testing exhaustivo)
- **Compatibilidad**: 100% backward compatible con v2

---

## Matriz de Dependencias

### Gráfico de Dependencias

```
1. Logger Interface (BLOCKING)
   ↓
2. Structured Fields (BLOCKING)
   ↓
3. Filtering by Level (BLOCKING)
   ↓
4. Dual Formatter (DEPENDS: 1,2,3)
   ↓
5. Child Loggers (DEPENDS: 1,2)
   ↓
6. Real Context (DEPENDS: 5)
   ↓
7. Hooks System (DEPENDS: 1)
   ↓
8. RotatingFileWriter (DEPENDS: 4)
   ↓
9. Redactor (DEPENDS: 2)
   ↓
10. Backward Compatibility Layer (DEPENDS: ALL)
```

### Análisis de Bloqueos

| Mejora | Bloquea | Depende De | Puede Paralelizarse Con |
|--------|---------|------------|------------------------|
| 1. Logger Interface | 2,3,4,5,6,7 | - | - |
| 2. Structured Fields | 4,5,9 | 1 | 3 |
| 3. Filtering by Level | 4 | 1,2 | - |
| 4. Dual Formatter | 8 | 1,2,3 | 5,6,7,9 |
| 5. Child Loggers | 6 | 1,2 | 3,4,7,9 |
| 6. Real Context | - | 5 | 7,8,9 |
| 7. Hooks System | - | 1 | 4,5,6,9 |
| 8. RotatingFileWriter | - | 4 | 5,6,7,9 |
| 9. Redactor | - | 2 | 4,5,6,7,8 |
| 10. Backward Compatibility | - | ALL | - |

### Identificación de Fases

Basado en dependencias, agrupamos en 7 fases secuenciales:

**Fase 1**: Foundation (1,2,3) - *Blocking para todo*
**Fase 2**: Output Layer (4) - *Depende de F1*
**Fase 3**: Context & Propagation (5,6) - *Depende de F1*
**Fase 4**: Extensibility (7) - *Depende de F1*
**Fase 5**: File Management (8) - *Depende de F2*
**Fase 6**: Security (9) - *Depende de F1*
**Fase 7**: Compatibility (10) - *Depende de todas*

---

## Fase 1: Foundation - Core Interfaces y Structured Fields

**Objetivo**: Establecer la arquitectura base con interfaces, campos estructurados y filtrado por nivel.

**Mejoras Incluidas**: #1 (Logger Interface), #2 (Structured Fields), #3 (Filtering by Level)

### Archivos a Crear

```
go_logs/
├── logger.go              # Nueva interfaz Logger
├── field.go               # Struct Field y helpers (String, Int, Err, etc.)
├── entry.go               # Entry representa un log entry con fields
├── level.go               # Level type con métodos (shouldLog, String, etc.)
├── options.go             # Option pattern para configuración
└── internal/
    └── logger_impl.go     # Implementación interna de Logger
```

### Archivos a Modificar

```
go_logs/
├── config.go              # Agregar SetLevel(), level parsing mejorado
└── logs.go                # Migrar a usar nueva interfaz interna
```

### Especificación Detallada

#### 1.1 Interfaz Logger

**Archivo**: `logger.go`

```go
package go_logs

import "context"

// Logger define la interfaz principal de logging
type Logger interface {
    // Logging básico
    Log(level Level, msg string, fields ...Field)

    // Convenience methods
    Trace(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)

    // Logging con contexto
    LogCtx(ctx context.Context, level Level, msg string, fields ...Field)

    // Child logger con campos pre-pendidos
    With(fields ...Field) Logger

    // Control de nivel
    SetLevel(level Level)
    GetLevel() Level

    // IO
    Sync() error
}

// New crea un nuevo Logger con opciones
func New(opts ...Option) (Logger, error)
```

#### 1.2 Structured Fields

**Archivo**: `field.go`

```go
package go_logs

// Field representa un campo estructurado key-value
type Field struct {
    key       string
    valueType FieldType
    value     interface{}
}

type FieldType int

const (
    StringType FieldType = iota
    IntType
    Int64Type
    Float64Type
    BoolType
    ErrorType
    AnyType
)

// Helpers para crear Fields sin allocations
func String(key, val string) Field {
    return Field{key: key, valueType: StringType, value: val}
}

func Int(key string, val int) Field {
    return Field{key: key, valueType: IntType, value: val}
}

func Int64(key string, val int64) Field {
    return Field{key: key, valueType: Int64Type, value: val}
}

func Float64(key string, val float64) Field {
    return Field{key: key, valueType: Float64Type, value: val}
}

func Bool(key string, val bool) Field {
    return Field{key: key, valueType: BoolType, value: val}
}

func Err(err error) Field {
    return Field{key: "error", valueType: ErrorType, value: err}
}

func Any(key string, val interface{}) Field {
    return Field{key: key, valueType: AnyType, value: val}
}
```

**Rationale**: Fields son structs simples, no interfaces, para minimizar allocations. Se pre-crean y se reusan.

#### 1.3 Level Type con Filtering

**Archivo**: `level.go`

```go
package go_logs

import "strings"

// Level representa el nivel de log (syslog-style)
type Level int

const (
    TraceLevel  Level = 10
    DebugLevel  Level = 20
    InfoLevel   Level = 30
    WarnLevel   Level = 40
    ErrorLevel  Level = 50
    FatalLevel  Level = 60
    SilentLevel Level = 0
)

// shouldLog implementa fast-path filtering
func (l Level) shouldLog(threshold Level) bool {
    return l >= threshold
}

// String devuelve representación del nivel
func (l Level) String() string {
    switch l {
    case TraceLevel:
        return "TRACE"
    case DebugLevel:
        return "DEBUG"
    case InfoLevel:
        return "INFO"
    case WarnLevel:
        return "WARN"
    case ErrorLevel:
        return "ERROR"
    case FatalLevel:
        return "FATAL"
    case SilentLevel:
        return "SILENT"
    default:
        return "UNKNOWN"
    }
}

// ParseLevel convierte string a Level
func ParseLevel(levelStr string) Level {
    switch strings.ToUpper(levelStr) {
    case "TRACE":
        return TraceLevel
    case "DEBUG":
        return DebugLevel
    case "INFO":
        return InfoLevel
    case "WARN", "WARNING":
        return WarnLevel
    case "ERROR":
        return ErrorLevel
    case "FATAL":
        return FatalLevel
    case "SILENT", "NONE", "DISABLE":
        return SilentLevel
    default:
        return InfoLevel // Default
    }
}
```

#### 1.4 Entry para Log Entries

**Archivo**: `entry.go`

```go
package go_logs

import "time"

// Entry representa un log entry completo
type Entry struct {
    Level     Level
    Message   string
    Fields    []Field
    Timestamp time.Time
}

// addFields agrega campos al entry
func (e *Entry) addFields(fields ...Field) {
    e.Fields = append(e.Fields, fields...)
}
```

#### 1.5 Option Pattern para Configuración

**Archivo**: `options.go`

```go
package go_logs

import (
    "io"
    "os"
)

// Option define un funcional option para configurar Logger
type Option func(*loggerImpl)

// WithOutput sets the output destination
func WithOutput(w io.Writer) Option {
    return func(l *loggerImpl) {
        l.output = w
    }
}

// WithLevel sets the minimum log level
func WithLevel(level Level) Option {
    return func(l *loggerImpl) {
        l.level = level
    }
}

// WithOutputFlags configura flags de output (color, timestamp, etc.)
func WithOutputFlags(flags int) Option {
    return func(l *loggerImpl) {
        l.flags = flags
    }
}

// WithFormatter sets the formatter (Text or JSON)
func WithFormatter(f Formatter) Option {
    return func(l *loggerImpl) {
        l.formatter = f
    }
}

// WithHooks agrega hooks al logger
func WithHooks(hooks ...Hook) Option {
    return func(l *loggerImpl) {
        l.hooks = append(l.hooks, hooks...)
    }
}

// WithRedactor habilita redacción de campos sensibles
func WithRedactor(keys []string) Option {
    return func(l *loggerImpl) {
        l.redactor = NewRedactor(keys...)
    }
}
```

#### 1.6 Implementación Interna

**Archivo**: `internal/logger_impl.go`

```go
package internal

import (
    "context"
    "fmt"
    "io"
    "os"
    "sync"

    "github.com/drossan/go_logs"
)

type loggerImpl struct {
    mu        sync.Mutex
    level     go_logs.Level
    output    io.Writer
    formatter go_logs.Formatter
    hooks     []go_logs.Hook
    redactor  *go_logs.Redactor

    // Child logger support
    parent    *loggerImpl
    fields    []go_logs.Field
}

func New(opts ...go_logs.Option) (*loggerImpl, error) {
    l := &loggerImpl{
        level:     go_logs.InfoLevel,
        output:    os.Stdout,
        formatter: go_logs.NewTextFormatter(),
        hooks:     []go_logs.Hook{},
        fields:    []go_logs.Field{},
    }

    for _, opt := range opts {
        opt(l)
    }

    return l, nil
}

func (l *loggerImpl) Log(level go_logs.Level, msg string, fields ...go_logs.Field) {
    // Fast-path: filtrar por nivel antes de cualquier trabajo
    if !level.shouldLog(l.level) {
        return
    }

    // Crear entry
    entry := &go_logs.Entry{
        Level:     level,
        Message:   msg,
        Fields:    append(l.fields, fields...), // Heredar campos del padre
        Timestamp: time.Now(),
    }

    // Aplicar redactor si existe
    if l.redactor != nil {
        l.redactor.Redact(entry)
    }

    // Ejecutar hooks
    for _, hook := range l.hooks {
        if err := hook.Run(entry); err != nil {
            fmt.Fprintf(os.Stderr, "Hook error: %v\n", err)
        }
    }

    // Formatear y escribir
    l.mu.Lock()
    defer l.mu.Unlock()

    formatted, err := l.formatter.Format(entry)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Format error: %v\n", err)
        return
    }

    l.output.Write(formatted)
}

func (l *loggerImpl) With(fields ...go_logs.Field) go_logs.Logger {
    return &loggerImpl{
        level:     l.level,
        output:    l.output,
        formatter: l.formatter,
        hooks:     l.hooks,
        redactor:  l.redactor,
        parent:    l,
        fields:    append(l.fields, fields...),
    }
}
```

### Criterios de Éxito

- [ ] Interfaz `Logger` definida con todos los métodos requeridos
- [ ] `Field` struct minimiza allocations (no interfaces, solo tipos primitivos)
- [ ] `Level.shouldLog()` implementa fast-path filtering (< 5ns)
- [ ] `loggerImpl` es thread-safe (mutex protege output write)
- [ ] `With()` crea child logger que hereda campos del padre sin mutar al padre
- [ ] `go test ./...` pasa sin errores
- [ ] `go build ./...` compila sin warnings

### Testing Plan

```go
// logger_test.go
func TestLogger_Interface(t *testing.T) {
    logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

    // Test que implementa interfaz
    var _ go_logs.Logger = logger

    // Test logging básico
    logger.Info("test message", go_logs.String("key", "value"))
}

func TestLevel_ShouldLog(t *testing.T) {
    tests := []struct {
        level     go_logs.Level
        threshold go_logs.Level
        expected  bool
    }{
        {go_logs.InfoLevel, go_logs.InfoLevel, true},
        {go_logs.DebugLevel, go_logs.InfoLevel, false},
        {go_logs.ErrorLevel, go_logs.InfoLevel, true},
    }

    for _, tt := range tests {
        got := tt.level.shouldLog(tt.threshold)
        if got != tt.expected {
            t.Errorf("Level.shouldLog() = %v, want %v", got, tt.expected)
        }
    }
}

func TestLogger_With(t *testing.T) {
    parent, _ := go_logs.New()
    child := parent.With(go_logs.String("parent_field", "parent_value"))

    // Child no afecta al padre
    parent.Info("parent msg")
    child.Info("child msg", go_logs.String("child_field", "child_value"))
}

func TestField_Allocations(t *testing.T) {
    // Benchmark para asegurar mínimo allocations
    f := go_logs.String("key", "value")

    allocs := testing.AllocsPerRun(1000, func() {
        _ = go_logs.String("test", "value")
    })

    if allocs > 1 {
        t.Errorf("String() allocated too much: %f", allocs)
    }
}
```

### Esfuerzo Estimado

- **Complexity**: Alta (arquitectura base afecta todo)
- **Time**: 12-15 horas
- **Risk**: Medio (error en diseño bloquea todo v3)

### Riesgos y Mitigaciones

| Riesgo | Impacto | Probabilidad | Mitigación |
|--------|---------|--------------|------------|
| Interface mal diseñada | Alto | Media | Revisar interfaces de zap/zerolog para reference |
| Field allocations | Alto | Baja | Benchmark desde el inicio |
| Thread-safety issues | Alto | Media | Usar mutex solo en writes, no en reads |
| Performance regression | Alto | Media | Comparar benchmarks con v2 |

---

## Fase 2: Output Layer - Dual Formatter

**Objetivo**: Implementar formateadores intercambiables (Text para dev, JSON para prod).

**Mejoras Incluidas**: #5 (Dual Formatter)

### Archivos a Crear

```
go_logs/
├── formatter.go           # Interfaz Formatter
├── text_formatter.go      # TextFormatter con colores
└── json_formatter.go      # JSONFormatter para producción
```

### Archivos a Modificar

```
go_logs/
├── config.go              # Agregar LOG_FORMAT env var parsing
└── internal/logger_impl.go # Usar formatter inyectado
```

### Especificación Detallada

#### 2.1 Interfaz Formatter

**Archivo**: `formatter.go`

```go
package go_logs

// Formatter define cómo formatear un Entry en bytes
type Formatter interface {
    Format(entry *Entry) ([]byte, error)
}

// FormatterConfig configura opciones de formateo
type FormatterConfig struct {
    EnableColors    bool
    EnableTimestamp bool
    EnableLevel     bool
    TimestampFormat string // ej: "2006/01/02 15:04:05"
}
```

#### 2.2 TextFormatter

**Archivo**: `text_formatter.go`

```go
package go_logs

import (
    "bytes"
    "fmt"
    "strings"
    "time"

    "github.com/fatih/color"
)

// TextFormatter formatea logs como texto legible con colores
type TextFormatter struct {
    config FormatterConfig

    // Colores por nivel
    levelColors map[Level]*color.Color
}

// NewTextFormatter crea un TextFormatter con configuración default
func NewTextFormatter() *TextFormatter {
    return &TextFormatter{
        config: FormatterConfig{
            EnableColors:    true,
            EnableTimestamp: true,
            EnableLevel:     true,
            TimestampFormat: "2006/01/02 15:04:05",
        },
        levelColors: map[Level]*color.Color{
            TraceLevel:  color.New(color.FgCyan),
            DebugLevel:  color.New(color.FgHiBlue),
            InfoLevel:   color.New(color.FgYellow),
            WarnLevel:   color.New(color.FgHiYellow),
            ErrorLevel:  color.New(color.FgRed),
            FatalLevel:  color.New(color.FgHiRed),
        },
    }
}

func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
    var buf bytes.Buffer

    // Timestamp
    if f.config.EnableTimestamp {
        buf.WriteString(entry.Timestamp.Format(f.config.TimestampFormat))
        buf.WriteString(" ")
    }

    // Level con color
    if f.config.EnableLevel {
        levelColor := f.levelColors[entry.Level]
        levelColor.Fprint(&buf, entry.Level.String())
        buf.WriteString(" ")
    }

    // Message
    buf.WriteString(entry.Message)

    // Fields
    if len(entry.Fields) > 0 {
        buf.WriteString(" ")
        f.formatFields(&buf, entry.Fields)
    }

    buf.WriteString("\n")
    return buf.Bytes(), nil
}

func (f *TextFormatter) formatFields(buf *bytes.Buffer, fields []Field) {
    var parts []string
    for _, field := range fields {
        parts = append(parts, fmt.Sprintf("%s=%v", field.key, field.value))
    }
    buf.WriteString(strings.Join(parts, " "))
}
```

#### 2.3 JSONFormatter

**Archivo**: `json_formatter.go`

```go
package go_logs

import (
    "encoding/json"
    "time"
)

// JSONFormatter formatea logs como JSON para ingestión en logs aggregators
type JSONFormatter struct {
    config FormatterConfig
}

// NewJSONFormatter crea un JSONFormatter
func NewJSONFormatter() *JSONFormatter {
    return &JSONFormatter{
        config: FormatterConfig{
            EnableTimestamp: true,
            EnableLevel:     true,
        },
    }
}

type JSONLogEntry struct {
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Timestamp string                 `json:"timestamp,omitempty"`
    Fields    map[string]interface{} `json:"fields,omitempty"`
}

func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
    jsonEntry := JSONLogEntry{
        Level:   entry.Level.String(),
        Message: entry.Message,
    }

    if f.config.EnableTimestamp {
        jsonEntry.Timestamp = entry.Timestamp.Format(time.RFC3339)
    }

    if len(entry.Fields) > 0 {
        jsonEntry.Fields = make(map[string]interface{})
        for _, field := range entry.Fields {
            jsonEntry.Fields[field.key] = field.value
        }
    }

    return json.Marshal(jsonEntry)
}
```

#### 2.4 Integración con Config

**Modificación en `config.go`**:

```go
// En Init(), agregar:
func loadLogFormat() Formatter {
    format := os.Getenv("LOG_FORMAT")
    switch strings.ToLower(format) {
    case "json":
        return NewJSONFormatter()
    case "text", "":
        return NewTextFormatter()
    default:
        log.Printf("Warning: Unknown LOG_FORMAT '%s', using text", format)
        return NewTextFormatter()
    }
}

// Y al crear logger default:
var defaultFormatter = loadLogFormat()
```

### Criterios de Éxito

- [ ] `TextFormatter` produce salida legible con colores
- [ ] `JSONFormatter` produce JSON válido parseable
- [ ] `LOG_FORMAT=json` cambia output a JSON automáticamente
- [ ] Ambos formatters soportan todos los Field types
- [ ] `go test ./...` pasa
- [ ] Integración con `loggerImpl` funciona

### Testing Plan

```go
func TestTextFormatter_Format(t *testing.T) {
    formatter := NewTextFormatter()

    entry := &Entry{
        Level:     InfoLevel,
        Message:   "test message",
        Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
        Fields:    []Field{String("key", "value")},
    }

    output, err := formatter.Format(entry)
    assert.NoError(t, err)
    assert.Contains(t, string(output), "test message")
    assert.Contains(t, string(output), "key=value")
}

func TestJSONFormatter_Format(t *testing.T) {
    formatter := NewJSONFormatter()

    entry := &Entry{
        Level:     ErrorLevel,
        Message:   "error occurred",
        Fields:    []Field{String("user_id", "123"), Int("count", 42)},
    }

    output, err := formatter.Format(entry)
    assert.NoError(t, err)

    var parsed map[string]interface{}
    err = json.Unmarshal(output, &parsed)
    assert.NoError(t, err)
    assert.Equal(t, "ERROR", parsed["level"])
    assert.Equal(t, "error occurred", parsed["message"])
    assert.Equal(t, "123", parsed["fields"].(map[string]interface{})["user_id"])
}
```

### Esfuerzo Estimado

- **Complexity**: Media
- **Time**: 6-8 horas
- **Risk**: Bajo

---

## Fase 3: Context & Propagation - Child Loggers y Contexto Real

**Objetivo**: Implementar child loggers y extracción automática de trace_id de context.Context.

**Mejoras Incluidas**: #3 (Child Loggers), #4 (Real Context)

### Archivos a Crear

```
go_logs/
├── context.go              # Utilidades para extracción de contexto
└── internal/
    └── context_keys.go     # Keys para context.Context
```

### Archivos a Modificar

```
go_logs/
├── api.go                  # Implementar Ctx functions que usan contexto real
└── internal/logger_impl.go # Agregar LogCtx que extrae del contexto
```

### Especificación Detallada

#### 3.1 Context Keys

**Archivo**: `internal/context_keys.go`

```go
package internal

import "context"

// Keys para almacenar valores en context.Context
type contextKey struct{}

var (
    TraceIDKey   = contextKey{}
    SpanIDKey    = contextKey{}
    UserIDKey    = contextKey{}
    RequestIDKey = contextKey{}
)

// WithTraceID agrega trace_id al contexto
func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSpanID agrega span_id al contexto
func WithSpanID(ctx context.Context, spanID string) context.Context {
    return context.WithValue(ctx, SpanIDKey, spanID)
}

// WithRequestID agrega request_id al contexto
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetTraceID extrae trace_id del contexto
func GetTraceID(ctx context.Context) string {
    if v := ctx.Value(TraceIDKey); v != nil {
        if traceID, ok := v.(string); ok {
            return traceID
        }
    }
    return ""
}

// GetSpanID extrae span_id del contexto
func GetSpanID(ctx context.Context) string {
    if v := ctx.Value(SpanIDKey); v != nil {
        if spanID, ok := v.(string); ok {
            return spanID
        }
    }
    return ""
}
```

#### 3.2 Funciones Helper para Context

**Archivo**: `context.go`

```go
package go_logs

import "context"

// WithTraceID crea un child logger con trace_id inyectado
// y también lo agrega al contexto
func WithTraceID(ctx context.Context, traceID string) context.Context {
    return internal.WithTraceID(ctx, traceID)
}

// WithRequestID crea un child logger con request_id inyectado
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return internal.WithRequestID(ctx, requestID)
}

// ExtractFieldsFromContext extrae campos del contexto
// para incluirlos automáticamente en logs
func ExtractFieldsFromContext(ctx context.Context) []Field {
    var fields []Field

    if traceID := internal.GetTraceID(ctx); traceID != "" {
        fields = append(fields, String("trace_id", traceID))
    }

    if spanID := internal.GetSpanID(ctx); spanID != "" {
        fields = append(fields, String("span_id", spanID))
    }

    if requestID := internal.GetRequestID(ctx); requestID != "" {
        fields = append(fields, String("request_id", requestID))
    }

    return fields
}
```

#### 3.3 Implementación de LogCtx en loggerImpl

**Modificación en `internal/logger_impl.go`**:

```go
func (l *loggerImpl) LogCtx(ctx context.Context, level go_logs.Level, msg string, fields ...go_logs.Field) {
    // Extraer campos del contexto automáticamente
    contextFields := go_logs.ExtractFieldsFromContext(ctx)

    // Combinar: campos del contexto + campos explícitos
    allFields := append(contextFields, fields...)

    // Logear con campos combinados
    l.Log(level, msg, allFields...)
}

func (l *loggerImpl) WithTraceID(traceID string) go_logs.Logger {
    return l.With(go_logs.String("trace_id", traceID))
}
```

#### 3.4 Actualizar API Global

**Modificación en `api.go`**:

```go
// Remplazar las funciones stub que no hacían nada con:

func InfoLogCtx(ctx context.Context, message string) {
    defaultLogger.LogCtx(ctx, InfoLevel, message)
}

func ErrorLogCtx(ctx context.Context, message string) {
    defaultLogger.LogCtx(ctx, ErrorLevel, message)
}

// ... etc para todos los niveles
```

### Criterios de Éxito

- [ ] `LogCtx()` extrae trace_id, span_id, request_id del contexto automáticamente
- [ ] Child logger creado con `With()` hereda campos sin mutar al padre
- [ ] `WithTraceID(ctx, "xyz")` agrega trace_id al contexto Y al logger
- [ ] Funciones globales `InfoLogCtx()` ahora usan el contexto real
- [ ] Thread-safe: context extraction no causa race conditions

### Testing Plan

```go
func TestLogCtx_ExtractsTraceID(t *testing.T) {
    logger, buf := newTestLogger()

    traceID := "abc-123-def-456"
    ctx := WithTraceID(context.Background(), traceID)

    logger.LogCtx(ctx, InfoLevel, "message", String("key", "value"))

    output := buf.String()
    assert.Contains(t, output, "trace_id=abc-123-def-456")
    assert.Contains(t, output, "key=value")
}

func TestLogger_With_ChildLogger(t *testing.T) {
    parent, buf := newTestLogger()

    child := parent.With(String("service", "my-service"))

    parent.Info("parent message")
    child.Info("child message", String("request_id", "123"))

    output := buf.String()
    assert.NotContains(t, output, "service=my-service parent message") // Parent no tiene field
    assert.Contains(t, output, "service=my-service child message")     // Child sí tiene
    assert.Contains(t, output, "request_id=123")
}

func TestWithTraceID_AggregatesToContext(t *testing.T) {
    ctx := context.Background()
    ctx = WithTraceID(ctx, "trace-xyz")

    traceID := GetTraceID(ctx)
    assert.Equal(t, "trace-xyz", traceID)
}
```

### Esfuerzo Estimado

- **Complexity**: Media-Alta
- **Time**: 8-10 horas
- **Risk**: Medio (context propagation puede ser tricky)

---

## Fase 4: Extensibility - Hooks System

**Objetivo**: Transformar Slack de código especial a un hook más, permitiendo añadir Sentry, métricas, etc.

**Mejoras Incluidas**: #7 (Hooks Extensibles)

### Archivos a Crear

```
go_logs/
├── hook.go                 # Interfaz Hook
├── hooks/
│   ├── slack_hook.go       # Mover Slack a hook
│   └── func_hook.go        # Hook genérico para funciones
└── adapters/
    └── slack_notifier.go   # Mover a hooks/slack_hook.go
```

### Archivos a Modificar

```
go_logs/
├── config.go              # Eliminar código especial de Slack
├── internal/logger_impl.go # Agregar hooks execution
└── save.go                # Eliminar notificación directa a Slack
```

### Especificación Detallada

#### 4.1 Interfaz Hook

**Archivo**: `hook.go`

```go
package go_logs

// Hook define un punto de extensión para procesar log entries
type Hook interface {
    // Run procesa un log entry. Retorna error si falla.
    Run(entry *Entry) error
}

// HookFunc es un adapter para usar funciones como hooks
type HookFunc func(entry *Entry) error

func (f HookFunc) Run(entry *Entry) error {
    return f(entry)
}

// NewFuncHook crea un hook desde una función simple
func NewFuncHook(fn func(entry *Entry) error) Hook {
    return HookFunc(fn)
}
```

#### 4.2 Slack Hook

**Archivo**: `hooks/slack_hook.go`

```go
package hooks

import (
    "github.com/drossan/go_logs"
    "github.com/drossan/go_logs/adapters"
)

// SlackHook envía logs a Slack basado en nivel
type SlackHook struct {
    notifier    *adapters.SlackNotifier
    level       go_logs.Level // Enviar a Slack solo >= este nivel
}

// NewSlackHook crea un nuevo Slack hook
func NewSlackHook(notifier *adapters.SlackNotifier, level go_logs.Level) *SlackHook {
    return &SlackHook{
        notifier: notifier,
        level:    level,
    }
}

func (h *SlackHook) Run(entry *go_logs.Entry) error {
    // Solo enviar si el nivel está por encima del threshold
    if !entry.Level.shouldLog(h.level) {
        return nil
    }

    // Formatear mensaje para Slack
    message := h.formatMessage(entry)

    // Enviar
    return h.notifier.SendNotification(message)
}

func (h *SlackHook) formatMessage(entry *go_logs.Entry) string {
    return fmt.Sprintf("[%s] %s", entry.Level.String(), entry.Message)
}
```

#### 4.3 Eliminar Código Especial de Slack

**Eliminaciones**:

- `config.go`: Eliminar variables `notificationsEnabled`, `notificationLogFatal`, etc.
- `config.go`: Eliminar `loadNotificationsConfig()`, `loadSlackConfig()`, `getNotificationSettings()`
- `save.go`: Eliminar `registerMessage()` - ahora es hook
- `adapters/slack_notifier.go`: Mover contenido a `hooks/slack_hook.go` y eliminar archivo

#### 4.4 Ejecución de Hooks en loggerImpl

**Modificación en `internal/logger_impl.go`**:

```go
func (l *loggerImpl) Log(level go_logs.Level, msg string, fields ...go_logs.Field) {
    if !level.shouldLog(l.level) {
        return
    }

    entry := &go_logs.Entry{
        Level:     level,
        Message:   msg,
        Fields:    append(l.fields, fields...),
        Timestamp: time.Now(),
    }

    // Aplicar redactor
    if l.redactor != nil {
        l.redactor.Redact(entry)
    }

    // Ejecutar hooks ANTES de escribir (permite modificar entry o abortar)
    for _, hook := range l.hooks {
        if err := hook.Run(entry); err != nil {
            // Loggear error pero no fallar el log original
            fmt.Fprintf(os.Stderr, "Hook error: %v\n", err)
        }
    }

    // Formatear y escribir
    l.mu.Lock()
    defer l.mu.Unlock()

    formatted, err := l.formatter.Format(entry)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Format error: %v\n", err)
        return
    }

    l.output.Write(formatted)
}
```

### Criterios de Éxito

- [ ] `Hook` interface definida con método `Run()`
- [ ] `SlackHook` implementa `Hook` y envía notificaciones
- [ ] `NewFuncHook()` permite crear hooks desde funciones simples
- [ ] Slack ya NO es código especial, es un hook más
- [ ] Hooks se ejecutan en orden y errors no abortan el log
- [ ] Configurar Slack ahora es: `New(SlackHook(...))`

### Testing Plan

```go
func TestSlackHook_Run(t *testing.T) {
    // Mock slack notifier
    notifier := &mockSlackNotifier{}
    hook := hooks.NewSlackHook(notifier, go_logs.ErrorLevel)

    entry := &go_logs.Entry{
        Level:   go_logs.ErrorLevel,
        Message: "test error",
    }

    err := hook.Run(entry)
    assert.NoError(t, err)
    assert.True(t, notifier.called)
}

func TestFuncHook(t *testing.T) {
    called := false
    hook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
        called = true
        return nil
    })

    logger, _ := go_logs.New(go_logs.WithHooks(hook))
    logger.Info("test")

    assert.True(t, called)
}
```

### Ejemplos de Uso

```go
// Slack hook
slackNotifier, _ := adapters.NewSlackNotifier()
slackHook := hooks.NewSlackHook(slackNotifier, go_logs.ErrorLevel)

// Sentry hook
sentryHook := NewFuncHook(func(entry *go_logs.Entry) error {
    if entry.Level >= go_logs.ErrorLevel {
        sentry.CaptureException(entry.Fields[0].value.(error))
    }
    return nil
})

// Métricas hook
metricsHook := NewFuncHook(func(entry *go_logs.Entry) error {
    metrics.IncCounter(fmt.Sprintf("logs.%s", entry.Level.String()))
    return nil
})

logger, _ := New(
    WithHooks(slackHook, sentryHook, metricsHook),
)
```

### Esfuerzo Estimado

- **Complexity**: Media
- **Time**: 6-8 horas
- **Risk**: Bajo-Medio (migración de Slack requiere cuidado)

---

## Fase 5: File Management - Rotación de Archivos

**Objetivo**: Implementar rotación de archivos por tamaño sin dependencias externas.

**Mejoras Incluidas**: #6 (Rotación de archivos)

### Archivos a Crear

```
go_logs/
├── writer.go               # Interfaz Writer
└── writers/
    ├── rotating_writer.go  # RotatingFileWriter
    └── buffer_writer.go    # Wrapper con buffering
```

### Archivos a Modificar

```
go_logs/
├── config.go              # Agregar LOG_MAX_SIZE, LOG_MAX_BACKUPS
└── internal/logger_impl.go # Usar RotatingFileWriter
```

### Especificación Detallada

#### 5.1 RotatingFileWriter

**Archivo**: `writers/rotating_writer.go`

```go
package writers

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "sync"
)

// RotatingFileWriter implementa rotación por tamaño
type RotatingFileWriter struct {
    filename    string
    maxSize     int64  // Max bytes antes de rotar
    maxBackups  int    // Max archivos de backup a mantener
    currentSize int64

    file    *os.File
    writer  *bufio.Writer
    mu      sync.Mutex
}

// NewRotatingFileWriter crea un nuevo writer con rotación
func NewRotatingFileWriter(filename string, maxSizeMB int, maxBackups int) (*RotatingFileWriter, error) {
    w := &RotatingFileWriter{
        filename:   filename,
        maxBackups: maxBackups,
        maxBytes:   int64(maxSizeMB) * 1024 * 1024,
    }

    if err := w.openFile(); err != nil {
        return nil, err
    }

    return w, nil
}

func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
    w.mu.Lock()
    defer w.mu.Unlock()

    writeLen := int64(len(p))

    // Check si necesitamos rotar ANTES de escribir
    if w.currentSize + writeLen > w.maxSize {
        if err := w.rotate(); err != nil {
            return 0, err
        }
    }

    n, err = w.writer.Write(p)
    if err != nil {
        return n, err
    }

    w.currentSize += writeLen
    return n, nil
}

func (w *RotatingFileWriter) rotate() error {
    // Cerrar archivo actual
    if err := w.writer.Flush(); err != nil {
        return err
    }
    if err := w.file.Close(); err != nil {
        return err
    }

    // Rotar backups: app.log.3 -> app.log.4, app.log.2 -> app.log.3, etc.
    for i := w.maxBackups - 1; i >= 1; i-- {
        oldBackup := fmt.Sprintf("%s.%d", w.filename, i)
        newBackup := fmt.Sprintf("%s.%d", w.filename, i+1)

        if _, err := os.Stat(oldBackup); err == nil {
            os.Rename(oldBackup, newBackup)
        }
    }

    // Mover actual a app.log.1
    backupName := fmt.Sprintf("%s.1", w.filename)
    os.Rename(w.filename, backupName)

    // Abrir nuevo archivo
    return w.openFile()
}

func (w *RotatingFileWriter) openFile() error {
    file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
    if err != nil {
        return err
    }

    w.file = file
    w.writer = bufio.NewWriter(file)

    // Obtener tamaño actual si el archivo ya existía
    if info, err := file.Stat(); err == nil {
        w.currentSize = info.Size()
    } else {
        w.currentSize = 0
    }

    return nil
}

func (w *RotatingFileWriter) Close() error {
    w.mu.Lock()
    defer w.mu.Unlock()

    if err := w.writer.Flush(); err != nil {
        return err
    }
    return w.file.Close()
}
```

#### 5.2 Integración con Config

**Modificación en `config.go`**:

```go
func loadRotatingFileConfig() (*writers.RotatingFileWriter, error) {
    if !saveLogFile {
        return nil, nil
    }

    maxSizeStr := os.Getenv("LOG_MAX_SIZE")
    maxSize := 100 // Default 100MB
    if maxSizeStr != "" {
        if val, err := strconv.Atoi(maxSizeStr); err == nil {
            maxSize = val
        }
    }

    maxBackupsStr := os.Getenv("LOG_MAX_BACKUPS")
    maxBackups := 3 // Default 3 backups
    if maxBackupsStr != "" {
        if val, err := strconv.Atoi(maxBackupsStr); err == nil {
            maxBackups = val
        }
    }

    fullPath := filepath.Join(logFilePath, logFileName)
    return writers.NewRotatingFileWriter(fullPath, maxSize, maxBackups)
}
```

### Criterios de Éxito

- [ ] `RotatingFileWriter` rota archivos cuando alcanza `maxSize`
- [ ] Mantiene máximo `maxBackups` archivos rotados
- [ ] Rotación es thread-safe (mutex protege todo)
- [ ] Archivos rotados siguen patrón: `app.log.1`, `app.log.2`, etc.
- [ ] Funciona sin dependencias externas (solo stdlib)
- [ ] `LOG_MAX_SIZE` y `LOG_MAX_BACKUPS` configuran rotación

### Testing Plan

```go
func TestRotatingFileWriter_Rotate(t *testing.T) {
    // Crear archivo temporal
    tmpDir := t.TempDir()
    filename := filepath.Join(tmpDir, "test.log")

    writer, _ := writers.NewRotatingFileWriter(filename, 1, 3) // 1MB, 3 backups

    // Escribir más de 1MB
    largeData := make([]byte, 1024*1024) // 1MB
    for i := 0; i < 5; i++ {
        writer.Write(largeData)
    }

    // Verificar que existen backups
    assert.FileExists(t, filename)        // Current
    assert.FileExists(t, filename+".1")   // Backup 1
    assert.FileExists(t, filename+".2")   // Backup 2
    assert.FileExists(t, filename+".3")   // Backup 3
    assert.NoFileExists(t, filename+".4") // No existe (max 3)
}
```

### Esfuerzo Estimado

- **Complexity**: Media
- **Time**: 6-8 horas
- **Risk**: Medio (rotación mal implementada puede perder logs)

---

## Fase 6: Security - Redactor de Datos Sensibles

**Objetivo**: Enmascarar automáticamente campos sensibles (password, token, secret).

**Mejoras Incluidas**: #8 (Redactor)

### Archivos a Crear

```
go_logs/
├── redactor.go             # Redactor principal
└── patterns.go             # Patrones comunes de sensibles
```

### Archivos a Modificar

```
go_logs/
├── options.go              # Agregar WithRedactor()
└── internal/logger_impl.go # Aplicar redactor antes de formatear
```

### Especificación Detallada

#### 6.1 Redactor

**Archivo**: `redactor.go`

```go
package go_logs

import (
    "strings"
)

// Redactor enmascara campos sensibles en log entries
type Redactor struct {
    sensitiveKeys map[string]bool
    maskValue     string
}

// NewRedactor crea un nuevo redactor con las keys sensibles
func NewRedactor(keys ...string) *Redactor {
    sensitiveKeys := make(map[string]bool)
    for _, key := range keys {
        sensitiveKeys[key] = true
    }

    return &Redactor{
        sensitiveKeys: sensitiveKeys,
        maskValue:     "***",
    }
}

// Redact modifica el entry en su lugar, enmascarando valores sensibles
func (r *Redactor) Redact(entry *Entry) {
    for i := range entry.Fields {
        if r.sensitiveKeys[entry.Fields[i].key] {
            entry.Fields[i].value = r.maskValue
        }
    }
}

// AddKey agrega una key sensible al redactor
func (r *Redactor) AddKey(key string) {
    r.sensitiveKeys[key] = true
}

// RemoveKey elimina una key de la lista de sensibles
func (r *Redactor) RemoveKey(key string) {
    delete(r.sensitiveKeys, key)
}
```

#### 6.2 Patrones Comunes

**Archivo**: `patterns.go`

```go
package go_logs

// CommonSensitiveKeys devuelve una lista de keys comúnmente sensibles
func CommonSensitiveKeys() []string {
    return []string{
        "password",
        "passwd",
        "pwd",
        "token",
        "api_key",
        "apikey",
        "api-key",
        "secret",
        "authorization",
        "auth",
        "cookie",
        "session",
        "credit_card",
        "ssn",
        "social_security",
    }
}
```

#### 6.3 Integración en loggerImpl

**Ya implementado en Fase 1**:

```go
// En Log():
if l.redactor != nil {
    l.redactor.Redact(entry)
}
```

### Criterios de Éxito

- [ ] `Redactor` enmascara valores de keys sensibles
- [ ] `WithRedactor()` habilita redacción en logger
- [ ] `CommonSensitiveKeys()` provee lista de keys comunes
- [ ] Redacción ocurre ANTES de hooks y formateo
- [ ] No afecta performance (< 100ns por entry)

### Testing Plan

```go
func TestRedactor_Redact(t *testing.T) {
    redactor := go_logs.NewRedactor("password", "token")

    entry := &go_logs.Entry{
        Fields: []go_logs.Field{
            go_logs.String("username", "john"),
            go_logs.String("password", "secret123"),
            go_logs.String("token", "xyz789"),
            go_logs.String("email", "john@example.com"),
        },
    }

    redactor.Redact(entry)

    assert.Equal(t, "john", entry.Fields[0].value)
    assert.Equal(t, "***", entry.Fields[1].value) // Masked
    assert.Equal(t, "***", entry.Fields[2].value) // Masked
    assert.Equal(t, "john@example.com", entry.Fields[3].value)
}

func TestLogger_WithRedactor(t *testing.T) {
    redactor := go_logs.NewRedactor("password")
    logger, buf := newTestLogger(go_logs.WithRedactor(redactor))

    logger.Info("login",
        go_logs.String("user", "john"),
        go_logs.String("password", "secret123"),
    )

    output := buf.String()
    assert.NotContains(t, output, "secret123")
    assert.Contains(t, output, "***")
}
```

### Esfuerzo Estimado

- **Complexity**: Baja
- **Time**: 3-4 horas
- **Risk**: Bajo

---

## Fase 7: Compatibility - Backward Compatibility Layer

**Objetivo**: Mantener 100% compatibilidad con v2 mientras se expone la nueva API v3.

**Mejoras Incluidas**: #10 (Compatibilidad con v2)

### Archivos a Crear

```
go_logs/
├── v2_compat.go            # Wrapper para funciones globales v2
└── example_v3_test.go      # Ejemplos de uso v3
```

### Archivos a Modificar

```
go_logs/
├── logs.go                 # Renombrar a v2_compat.go
├── api.go                  # Actualizar para usar defaultLogger v3
├── config.go               # Agregar Init() que crea defaultLogger
└── save.go                 # Eliminar (funcionalidad migrada a hooks)
```

### Especificación Detallada

#### 7.1 Default Logger Global

**Modificación en `config.go`**:

```go
var (
    defaultLogger Logger
    defaultOnce   sync.Once
)

// Init inicializa el logger default con configuración de entorno
// Mantiene backward compatibility con v2
func Init() {
    defaultOnce.Do(func() {
        // Cargar configuración de entorno
        loadLogLevel()

        // Crear formatter
        var formatter Formatter
        if os.Getenv("LOG_FORMAT") == "json" {
            formatter = NewJSONFormatter()
        } else {
            formatter = NewTextFormatter()
        }

        // Crear output
        var output io.Writer = os.Stdout
        if saveLogFile {
            if rotatingWriter, err := loadRotatingFileConfig(); err == nil {
                output = rotatingWriter
            }
        }

        // Crear hooks (Slack si está configurado)
        var hooks []Hook
        if notificationsEnabled {
            if slackNotifier, err := adapters.NewSlackNotifier(); err == nil {
                slackHook := hooks.NewSlackHook(slackNotifier, logLevel)
                hooks = append(hooks, slackHook)
            }
        }

        // Crear redactor si hay keys sensibles
        var redactor *Redactor
        if sensitiveKeys := os.Getenv("LOG_REDACT_KEYS"); sensitiveKeys != "" {
            keys := strings.Split(sensitiveKeys, ",")
            redactor = NewRedactor(keys...)
        }

        // Crear logger default con todas las opciones
        opts := []Option{
            WithOutput(output),
            WithLevel(logLevel),
            WithFormatter(formatter),
            WithHooks(hooks...),
        }

        if redactor != nil {
            opts = append(opts, WithRedactor(redactor))
        }

        defaultLogger, _ = New(opts...)
    })
}
```

#### 7.2 Funciones Globales v2 (Compatibilidad)

**Archivo**: `v2_compat.go` (renombrar `logs.go`)

```go
package go_logs

import (
    "log"
)

// FatalLog logs a fatal message and terminates the program.
// v2 compatibility function - now uses defaultLogger v3
func FatalLog(message string) {
    defaultLogger.Log(FatalLevel, message)
    log.Fatal(" 💣  " + message)
}

// ErrorLog logs an error message.
// v2 compatibility function - now uses defaultLogger v3
func ErrorLog(message string) {
    defaultLogger.Log(ErrorLevel, message)
}

// InfoLog logs an informational message.
// v2 compatibility function - now uses defaultLogger v3
func InfoLog(message string) {
    defaultLogger.Log(InfoLevel, message)
}

// SuccessLog logs a success message.
// v2 compatibility function - now uses defaultLogger v3
func SuccessLog(message string) {
    defaultLogger.Log(SuccessLevel, message) // Necesitamos agregar SuccessLevel
}

// WarningLog logs a warning message.
// v2 compatibility function - now uses defaultLogger v3
func WarningLog(message string) {
    defaultLogger.Log(WarnLevel, message)
}
```

#### 7.3 Close() para v2

**Modificación en `config.go`**:

```go
// Close flushes and closes the logger.
// v2 compatibility - now calls Sync() on defaultLogger
func Close() {
    if defaultLogger != nil {
        _ = defaultLogger.Sync()
    }
}
```

#### 7.4 SuccessLevel (Agregar Nivel Faltante)

**Modificación en `level.go`**:

```go
const (
    TraceLevel     Level = 10
    DebugLevel     Level = 20
    InfoLevel      Level = 30
    WarnLevel      Level = 40
    ErrorLevel     Level = 50
    FatalLevel     Level = 60
    SuccessLevel   Level = 25  // NUEVO: Entre Debug e Info
    SilentLevel    Level = 0
)

// Agregar en String():
func (l Level) String() string {
    switch l {
    // ... existing cases ...
    case SuccessLevel:
        return "SUCCESS"
    // ...
    }
}
```

#### 7.2 Ejemplos de Uso v3

**Archivo**: `example_v3_test.go`

```go
package go_logs_test

import (
    "context"
    "os"

    "github.com/drossan/go_logs"
)

func Example_v3_structuredLogging() {
    // Crear logger con campos estructurados
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Info("Server started",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
        go_logs.Bool("debug", false),
    )
}

func Example_v3_childLogger() {
    // Logger padre con contexto de servicio
    baseLogger, _ := go_logs.New()

    // Child logger con request_id inyectado
    reqLogger := baseLogger.With(
        go_logs.String("service", "api"),
        go_logs.String("version", "1.0"),
    )

    // Todos los logs de reqLogger tendrán service y version
    reqLogger.Info("Handling request",
        go_logs.String("endpoint", "/users"),
        go_logs.String("method", "GET"),
    )
}

func Example_v3_context() {
    logger, _ := go_logs.New()

    // Agregar trace_id al contexto
    ctx := go_logs.WithTraceID(context.Background(), "trace-abc-123")

    // Log con contexto - trace_id se extrae automáticamente
    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received")
}

func Example_v3_jsonFormat() {
    // Configurar con JSON formatter para producción
    logger, _ := go_logs.New(
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Error("Database connection failed",
        go_logs.String("host", "db.example.com"),
        go_logs.Int("port", 5432),
        go_logs.Err(err),
    )
    // Output: {"level":"ERROR","message":"Database connection failed","fields":{"host":"db.example.com","port":5432,"error":"..."}}
}

func Example_v2_compatibility() {
    // Código v2 sigue funcionando sin cambios
    go_logs.Init()
    defer go_logs.Close()

    go_logs.InfoLog("This still works exactly like v2")
    go_logs.Infof("User %s logged in", username)
}

func Example_v3_multipleInstances() {
    // Diferentes loggers con distintas configuraciones
    accessLogger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(createFile("access.log")),
    )

    errorLogger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.ErrorLevel),
        go_logs.WithFormatter(go_logs.NewTextFormatter()),
        go_logs.WithOutput(createFile("error.log")),
    )

    accessLogger.Info("API request")
    errorLogger.Error("Database error")
}
```

### Criterios de Éxito

- [ ] Código v2 funciona sin cambios (100% backward compatible)
- [ ] Funciones globales (`InfoLog`, `Errorf`, etc.) usan `defaultLogger` v3
- [ ] `Init()` configura `defaultLogger` con todas las opciones de entorno
- [ ] `Close()` llama `Sync()` en `defaultLogger`
- [ ] Tests v2 pasan sin modificaciones
- [ ] Nueva API v3 está disponible y documentada

### Testing Plan

```go
func TestV2Compatibility_GlobalFunctions(t *testing.T) {
    Init()
    defer Close()

    // Estas funciones deben seguir funcionando
    InfoLog("test info")
    ErrorLog("test error")
    WarningLog("test warning")
    SuccessLog("test success")

    // Con formato
    Infof("User %s logged in", "john")
    Errorf("Failed to connect: %v", err)

    // Con contexto
    ctx := context.Background()
    InfoLogCtx(ctx, "with context")
    InfoLogCtxf(ctx, "User %s", "john")
}

func TestV3_NewAPI(t *testing.T) {
    logger, buf := newTestLogger()

    // Nueva API con campos estructurados
    logger.Info("test message",
        String("key1", "value1"),
        Int("key2", 42),
    )

    output := buf.String()
    assert.Contains(t, output, "test message")
    assert.Contains(t, output, "key1=value1")
    assert.Contains(t, output, "key2=42")
}
```

### Esfuerzo Estimado

- **Complexity**: Media
- **Time**: 8-10 horas
- **Risk**: Alto (cualquier break en compatibility es crítico)

---

## Plan de Testing Global

### Estrategia de Testing

1. **Unit Tests**: Cada componente individualmente
2. **Integration Tests**: Intercación entre componentes
3. **V2 Compatibility Tests**: Asegurar que nada se rompe
4. **Benchmark Tests**: Performance no debe degradarse
5. **Example Tests**: Godoc examples deben compilar y ejecutarse

### Cobertura Objetivo

- **Core packages**: > 85%
- **Adapters/Hooks**: > 80%
- **Formatters**: > 90%
- **Global**: > 83% (mantener cobertura actual)

### Test Suite Structure

```
go_logs/
├── logger_test.go          # Tests de interfaz Logger
├── field_test.go           # Tests de Field
├── level_test.go           # Tests de Level
├── formatter_test.go       # Tests de formatters
├── hook_test.go            # Tests de hooks
├── redactor_test.go        # Tests de redactor
├── rotating_test.go        # Tests de rotación
├── v2_compat_test.go       # Tests de compatibilidad v2
├── integration_test.go     # Tests end-to-end
├── benchmark_test.go       # Benchmarks
└── example_v3_test.go      # Godoc examples
```

### Benchmarks Críticos

```go
// Comparar v2 vs v3
func BenchmarkInfoLogV2(b *testing.B) {
    Init()
    defer Close()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        InfoLog("test message")
    }
}

func BenchmarkInfoLogV3(b *testing.B) {
    logger, _ := New(WithOutput(io.Discard))

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("test message")
    }
}

func BenchmarkFieldsCreation(b *testing.B) {
    b.Run("String", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = String("key", "value")
        }
    })

    b.Run("Multiple", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = []Field{
                String("key1", "value1"),
                Int("key2", 42),
                Err(errors.New("test")),
            }
        }
    })
}

func BenchmarkChildLogger(b *testing.B) {
    parent, _ := New(WithOutput(io.Discard))
    child := parent.With(String("service", "test"))

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        child.Info("message")
    }
}

func BenchmarkJSONFormatter(b *testing.B) {
    formatter := NewJSONFormatter()
    entry := &Entry{
        Level:   InfoLevel,
        Message: "test",
        Fields:  []Field{String("key", "value")},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = formatter.Format(entry)
    }
}
```

### Objetivos de Performance

- **Field creation**: < 10ns por field
- **Fast-path filtering (shouldLog)**: < 5ns
- **TextFormatter**: < 1µs por entry simple
- **JSONFormatter**: < 2µs por entry simple
- **Child logger creation**: < 100ns
- **No regression vs v2**: v3 debe ser ≤ 20% más lento que v2 para caso simple

---

## Riesgos y Mitigaciones

### Riesgos Críticos

| Riesgo | Impacto | Probabilidad | Mitigación |
|--------|---------|--------------|------------|
| **Backward incompatibility** | Crítico | Media | - Suite exhaustiva de tests v2<br>- Mantener functions globales como wrappers<br>- Changelog claro con breaking changes |
| **Performance regression** | Alto | Media | - Benchmarks en cada commit<br>- Fast-path filtering en shouldLog()<br>- Minimizar allocations en Fields<br>- Profile antes/despues |
| **Thread-safety violations** | Alto | Baja | - Mutex solo en writes<br>- Usar sync.Once para init<br>- Tests con -race en CI |
| **Over-engineering** | Medio | Alta | - Mantener simple: Fields son structs, no interfaces<br>- No abstraer demasiado<br>- Seguir patrones de zap/zerolog |
| **Rotación de archivos mal implementada** | Medio | Baja | - Tests exhaustivos de rotación<br>- Validar tamaño y backups<br>- Manejo de errores robusto |
| **Slack migration rompe notificaciones** | Alto | Baja | - Test SlackHook con mock<br>- Mantener compatibilidad de env vars<br>- Migration guide para usuarios |

### Plan de Contingencia

**Si backward compatibility se rompe**:
1. Identificar qué función v2 falla
2. Agregar wrapper temporal con deprecation warning
3. Fix en next patch release
4. Actualizar migration guide

**Si performance degrada > 20%**:
1. Profile con pprof
2. Identificar hotspot
3. Optimizar o eliminar abstraction
4. Considerar que "good enough" es aceptable si < 50% slower

**Si thread-safety issues**:
1. Ejecutar tests con `go test -race`
2. Revisar todos los mutex
3. Aumentar cobertura de tests concurrentes

---

## Estrategia de Migración para Usuarios

### Migración Gradual en 3 Fases

#### Fase 1: Drop-in Replacement (Sin Cambios)

**Usuarios pueden seguir usando v2 tal cual**:

```go
// Código v2 - SIGUE FUNCIONANDO
import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()
    defer go_logs.Close()

    go_logs.InfoLog("Server started")
    go_logs.Errorf("Error: %v", err)
}
```

**Cambios requeridos**: Ninguno. Solo actualizar dependency a v3.

#### Fase 2: Adoptar Nueva API Progresivamente

**Migrar código nuevo a v3 mientras v2 sigue funcionando**:

```go
// Código mixto v2 + v3
func handleRequest() {
    // Logger v2 existente
    go_logs.InfoLog("Request received")

    // Nuevo logger v3 con campos estructurados
    logger, _ := go_logs.New(
        go_logs.WithFields(
            go_logs.String("service", "api"),
            go_logs.String("version", "2.0"),
        ),
    )

    logger.Info("Processing request",
        go_logs.String("request_id", reqID),
        go_logs.Int("user_id", userID),
    )
}
```

#### Fase 3: Migración Completa a v3

**Reemplazar todas las llamadas v2 con v3**:

```go
// ANTES (v2)
go_logs.InfoLog("User logged in")
go_logs.Infof("User %s logged in from %s", username, ip)
go_logs.ErrorLog("Database connection failed")

// DESPUÉS (v3)
logger.Info("user_logged_in",
    go_logs.String("username", username),
    go_logs.String("ip", ip),
)

logger.Error("database_connection_failed",
    go_logs.String("host", dbHost),
    go_logs.Err(err),
)
```

### Guía de Migración Detallada

#### 1. Reemplazar Funciones Globales

```go
// v2
go_logs.InfoLog("message")
go_logs.Infof("formatted %s", arg)
go_logs.InfoLogCtx(ctx, "with context")

// v3 - Crear logger y usarlo
logger, _ := go_logs.New()
logger.Info("message")
logger.Info("formatted", go_logs.String("arg", arg))
logger.LogCtx(ctx, go_logs.InfoLevel, "with context")
```

#### 2. Migrar de Printf-Style a Structured Fields

```go
// v2 - Printf style
go_logs.Infof("User %s (ID: %d) logged in from %s", username, userID, ip)

// v3 - Structured fields
logger.Info("user_logged_in",
    go_logs.String("username", username),
    go_logs.Int("user_id", userID),
    go_logs.String("ip", ip),
)
```

**Beneficios**: Fields son queryeables en log aggregators (ELK, Splunk, etc.)

#### 3. Migrar Contexto

```go
// v2 - Contexto aceptado pero ignorado
ctx := context.Background()
go_logs.InfoLogCtx(ctx, "message") // Contexto no usado

// v3 - Contexto extrae trace_id automáticamente
ctx := go_logs.WithTraceID(context.Background(), "trace-123")
logger.LogCtx(ctx, go_logs.InfoLevel, "message") // Incluye trace_id
```

#### 4. Migrar Configuración de Entorno

```bash
# v2 - Variables de entorno
SAVE_LOG_FILE=1
NOTIFICATION_ERROR_LOG=1
NOTIFICATIONS_SLACK_ENABLED=1

# v3 - Mismas variables + nuevas
SAVE_LOG_FILE=1
LOG_LEVEL=error              # NUEVO: Reemplaza NOTIFICATION_*_LOG
LOG_FORMAT=json              # NUEVO: Text o JSON
LOG_MAX_SIZE=100             # NUEVO: Rotación por MB
LOG_MAX_BACKUPS=3            # NUEVO: Max archivos rotados
LOG_REDACT_KEYS=password,token,secret  # NUEVO: Keys sensibles
```

### Migration Script

```bash
#!/bin/bash
# migrate_v2_to_v3.sh

echo "Migrando go_logs de v2 a v3..."

# Actualizar go.mod
go get -u github.com/drossan/go_logs@v3
go mod tidy

# Opción 1: Sin cambios (código v2 sigue funcionando)
echo "✅ Código v2 sigue funcionando. Actualiza dependency a v3."

# Opción 2: Migrar gradualmente
echo "📝 Migra gradualmente:"
echo "  1. Reemplazar InfoLog() -> logger.Info() en código nuevo"
echo "  2. Agregar campos estructurados: String(), Int(), Err()"
echo "  3. Usar child loggers: logger.With()"
echo "  4. Extraer contexto: LogCtx()"

# Opción 3: Migración completa
echo "🚀 Para migración completa:"
echo "  1. Reemplazar todas las funciones globales"
echo "  2. Usar campos estructurados en lugar de Printf"
echo "  3. Agregar LOG_LEVEL, LOG_FORMAT a .env"
echo "  4. Configurar hooks: Slack, Sentry, etc."

echo "✨ Consulta MIGRATION.md para detalles"
```

---

## Cronograma de Implementación

### Estimación por Fase

| Fase | Duración | Start | End | Dependencies |
|------|----------|-------|-----|--------------|
| **F1: Foundation** | 12-15h | Week 1 | Week 2 | - |
| **F2: Output Layer** | 6-8h | Week 2 | Week 2 | F1 |
| **F3: Context** | 8-10h | Week 3 | Week 3 | F1 |
| **F4: Hooks** | 6-8h | Week 3 | Week 4 | F1 |
| **F5: File Rotation** | 6-8h | Week 4 | Week 4 | F2 |
| **F6: Redactor** | 3-4h | Week 4 | Week 4 | F1 |
| **F7: Compatibility** | 8-10h | Week 5 | Week 5 | F1-F6 |
| **Testing & Docs** | 10-12h | Week 5 | Week 6 | F1-F7 |
| **Buffer for Issues** | 8-10h | Week 6 | Week 6 | - |
| **TOTAL** | **67-93h** | **Week 1** | **Week 6** | **8 semanas** |

### Milestones

**Milestone 1: Foundation Complete (End of Week 2)**
- Logger interface definida
- Structured fields implementados
- Filtering por nivel funcional
- Tests de core pasando

**Milestone 2: Feature Complete (End of Week 4)**
- Dual formatter (Text + JSON)
- Child loggers con contexto
- Hooks system con Slack migrado
- RotatingFileWriter funcional
- Redactor de datos sensibles

**Milestone 3: Production Ready (End of Week 6)**
- 100% backward compatible con v2
- Tests suite completa (> 83% cobertura)
- Benchmarks satisfactorios (≤ 20% slower)
- Documentación completa (godoc + MIGRATION.md)
- Changelog preparado

### Sprints Recomendados

**Sprint 1 (Week 1-2): Foundation**
- Logger interface + Level + Field
- Implementación base de loggerImpl
- Tests de interfaz y fields

**Sprint 2 (Week 2-3): Output & Context**
- TextFormatter + JSONFormatter
- Child loggers
- Context extraction

**Sprint 3 (Week 3-4): Extensibility & Files**
- Hooks system
- Slack hook migration
- RotatingFileWriter

**Sprint 4 (Week 4-5): Security & Compatibility**
- Redactor
- Backward compatibility layer
- Migración de tests v2

**Sprint 5 (Week 5-6): Polish & Docs**
- Testing exhaustivo
- Benchmarks
- Documentación
- MIGRATION guide

---

## Checklist Final de Release

### Pre-Release

- [ ] Todos los tests pasan (`go test ./...`)
- [ ] Tests con race detector pasan (`go test -race ./...`)
- [ ] Cobertura > 83%
- [ ] Benchmarks satisfacen objetivos
- [ ] Godoc completa en todas las funciones exportadas
- [ ] Examples en godoc compilan y ejecutan
- [ ] Changelog preparado
- [ ] MIGRATION.md escrito
- [ ] README.md actualizado con ejemplos v3
- [ ] CLAUDE.md actualizado con nueva arquitectura

### Backward Compatibility Verification

- [ ] Tests v2 pasan sin modificaciones
- [ ] Funciones globales (InfoLog, Errorf, etc.) funcionan
- [ ] Init() y Close() funcionan
- [ ] Variables de entorno v2 siguen funcionando
- [ ] Slack notifications funcionan como en v2

### Performance Verification

- [ ] BenchmarkV2 vs BenchmarkV3: ≤ 20% slower
- [ ] Field allocation: < 10ns
- [ ] shouldLog: < 5ns
- [ ] No memory leaks (check con pprof)

### Documentation

- [ ] README.md con ejemplos v2 y v3
- [ ] MIGRATION.md con guía paso a paso
- [ ] Godoc en todos los exports
- [ ] Examples ejecutables en godoc
- [ ] Changelog con v3.0.0 features

### Release

- [ ] Tag v3.0.0 en Git
- [ ] GoReleaser build exitoso
- [ ] Publicar release en GitHub con notas
- [ ] Anuncio en blog/post/git

---

## Apéndice: Comparativa v2 vs v3

### API Comparison

| v2 | v3 | Notas |
|----|----|-------|
| `InfoLog("msg")` | `logger.Info("msg")` | Mismo uso, con logger instance |
| `Infof("formatted %s", arg)` | `logger.Info("formatted", String("arg", arg))` | Printf → Structured fields |
| `InfoLogCtx(ctx, "msg")` | `logger.LogCtx(ctx, InfoLevel, "msg")` | Context ahora extrae trace_id |
| - | `logger.With(String("key", val))` | NUEVO: Child logger |
| - | `go_logs.New(WithLevel(...))` | NUEVO: Configuración programática |
| `Init()` con env vars | `Init()` + `New(opts...)` | Ambos coexisten |
| `Close()` | `logger.Sync()` | Close llama Sync() en defaultLogger |

### Feature Comparison

| Feature | v2 | v3 |
|---------|----|----|
| Niveles de log | ✅ Fatal, Error, Warning, Info, Success | ✅ + Trace, Debug |
| Formato | Text plano | Text + JSON |
| Colores | ✅ Sí | ✅ Sí (configurable) |
| Archivo persistente | ✅ Sí | ✅ Sí + rotación |
| Slack notifications | ✅ Sí (hardcoded) | ✅ Sí (hook) |
| Context support | ⚠️ Acepta pero ignora | ✅ Extrae trace_id |
| Fields estructurados | ❌ No | ✅ String, Int, Err, etc. |
| Child loggers | ❌ No | ✅ With(fields) |
| Hooks | ❌ No (solo Slack) | ✅ Sistema de hooks extensible |
| Redactor | ❌ No | ✅ Enmascara sensibles |
| Nivel runtime | ❌ No | ✅ SetLevel() |
| Multiple instances | ❌ No (global solo) | ✅ New() crea instancias |
| Zero dependencies | ⚠️ fatih/color, slack-go | ✅ Mismo (solo para color/Slack) |
| Thread-safe | ✅ Sí | ✅ Sí |

---

## Conclusión

Este plan transforma `go_logs` v2 en una biblioteca moderna y competitiva v3 manteniendo total backward compatibility. Las 10 mejoras solicitadas se implementan en 7 fases secuenciales con dependencias claras, testing exhaustivo y mitigación de riesgos.

**Puntos Clave**:

1. **Foundation First**: Interfaz Logger, Fields, y Filtering son la base
2. **Backward Compatible**: Código v2 funciona sin cambios
3. **Performance Fast-path**: shouldLog() filtra antes de allocations
4. **Extensible**: Hooks system permite añadir funcionalidad sin tocar core
5. **Production Ready**: Rotación de archivos, JSON formatter, redactor

**Next Steps**:

1. Revisar y aprobar este plan
2. Crear branch `feature/v3-upgrade`
3. Comenzar con Fase 1: Foundation
4. Progress tracking con milestones en GitHub/GitLab

**Estimación Final**: 67-93 horas de desarrollo (8-10 semanas para un developer a tiempo completo, o 4-5 semanas con dedicación parcial).
