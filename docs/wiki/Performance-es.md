# Rendimiento

go_logs esta disenado para logging de alto rendimiento con sobrecarga minima. Esta pagina cubre benchmarks, caracteristicas de rendimiento y tips de optimizacion.

**[English](Performance.md)** | **Espanol**

## Benchmarks

### Operaciones Core

| Operacion | Rendimiento | Objetivo |
|-----------|-------------|----------|
| Filtrado fast-path | 0.32 ns/op | < 5 ns |
| Creacion de campo | 0.34 ns/op, 0 allocs | < 10 ns |
| TextFormatter | 220.6 ns/op | < 500 ns |
| JSONFormatter | 249.3 ns/op | < 1 us |
| RotatingFileWriter | 16M msg/seg | - |

### Benchmarks Detallados

#### Filtrado de Nivel

```
BenchmarkLevelShouldLog/TraceLevel-8         1000000000    0.32 ns/op    0 B/op    0 allocs/op
BenchmarkLevelShouldLog/DebugLevel-8         1000000000    0.32 ns/op    0 B/op    0 allocs/op
BenchmarkLevelShouldLog/InfoLevel-8          1000000000    0.32 ns/op    0 B/op    0 allocs/op
```

#### Creacion de Campos

```
BenchmarkFieldString-8     1000000000    0.34 ns/op    0 B/op    0 allocs/op
BenchmarkFieldInt-8        1000000000    0.34 ns/op    0 B/op    0 allocs/op
BenchmarkFieldInt64-8      1000000000    0.34 ns/op    0 B/op    0 allocs/op
```

#### Rendimiento de Formateadores

```
BenchmarkTextFormatter/SinCampos-8           5000000    220.6 ns/op    128 B/op    3 allocs/op
BenchmarkTextFormatter/5Campos-8             3000000    385.2 ns/op    256 B/op    5 allocs/op

BenchmarkJSONFormatter/SinCampos-8           5000000    249.3 ns/op    192 B/op    4 allocs/op
BenchmarkJSONFormatter/5Campos-8             3000000    425.6 ns/op    320 B/op    6 allocs/op
```

### Ejecutar Benchmarks

```bash
# Ejecutar todos los benchmarks
go test -bench=. -benchmem ./...

# Ejecutar benchmark especifico
go test -bench=BenchmarkTextFormatter -benchmem ./...

# Ejecutar con profiling de CPU
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

## Caracteristicas de Rendimiento

### Zero Allocations

La creacion de campos y filtrado de nivel son operaciones zero-allocation:

```go
// Estas operaciones no asignan memoria
field := go_logs.String("key", "value")  // 0 allocs
if level.shouldLog(threshold) { }        // 0 allocs
```

### Filtrado Fast-Path

El filtrado de nivel usa comparacion numerica (estilo syslog) para maxima velocidad:

```go
// Esto es una sola comparacion de enteros (< 1ns)
if entry.Level >= threshold {
    // Loguear la entrada
}
```

### I/O con Buffer

RotatingFileWriter usa escrituras con buffer para rendimiento:

```go
// Buffer de 4KB por defecto
writer := bufio.NewWriter(file)

// Flush en Sync() o Close()
writer.Flush()
```

### Thread-Safety

Todas las operaciones son thread-safe usando mutex, pero el camino critico es minimo:

```go
func (l *LoggerImpl) Log(level Level, msg string, fields ...Field) {
    // Fast path: verificar nivel antes de adquirir lock
    if level < l.level {
        return  // No se necesita lock para mensajes filtrados
    }

    l.mu.Lock()
    defer l.mu.Unlock()
    // ... resto del logging
}
```

## Tips de Optimizacion

### 1. Establecer Nivel de Log Apropiado

La optimizacion mas efectiva es filtrar en el nivel correcto:

```go
// Desarrollo
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
)

// Produccion
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
)
```

Los logs filtrados tienen ~0.32 ns de sobrecarga (esencialmente gratis).

### 2. Evitar Operaciones Costosas en Logs Filtrados

Verifica el nivel antes de operaciones costosas:

```go
// Mal: Siempre computa datos costosos
logger.Debug("Volcado de estado", go_logs.Any("state", volcadoCostoso()))

// Bien: Solo computa si debug esta habilitado
if logger.GetLevel() <= go_logs.DebugLevel {
    logger.Debug("Volcado de estado", go_logs.Any("state", volcadoCostoso()))
}
```

### 3. Usar Child Loggers con Moderacion

Los child loggers agregan sobrecarga por copia de campos:

```go
// Bien: Crear una vez por peticion
reqLogger := logger.With(
    go_logs.String("request_id", requestID),
)

// Evitar: Crear en loops ajustados
for i := 0; i < 10000; i++ {
    iterLogger := logger.With(go_logs.Int("iteration", i)) // Sobrecarga!
}
```

### 4. Usar Formateador JSON en Produccion

JSONFormatter esta optimizado para agregadores de logs:

```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### 5. Reducir Cantidad de Campos

Mas campos = mas sobrecarga:

```go
// Mal: Muchos campos
logger.Info("Peticion",
    go_logs.String("campo1", v1),
    go_logs.String("campo2", v2),
    // ... 20 campos mas
)

// Bien: Solo campos esenciales
logger.Info("Peticion",
    go_logs.String("request_id", id),
    go_logs.Int("status", status),
    go_logs.Int64("duration_ms", duration),
)
```

## Comparacion con Otras Bibliotecas

| Biblioteca | Throughput | Asignaciones | Notas |
|------------|------------|--------------|-------|
| go_logs v3 | ~4M logs/seg | 0-8 por log | Campos zero-allocation |
| zap | ~10M logs/seg | 0 por log | Mas optimizado |
| zerolog | ~8M logs/seg | 0 por log | Muy rapido |
| logrus | ~500K logs/seg | 23 por log | Mas lento, mas features |
| standard log | ~1M logs/seg | 2 por log | Simple |

go_logs balancea rendimiento con una API limpia y conjunto completo de features.

## Anti-Patrones de Rendimiento

### 1. Logging en Hot Paths

```go
// Mal: Logging en loop ajustado
for i := 0; i < 1000000; i++ {
    logger.Debug("Procesando", go_logs.Int("i", i))
}

// Bien: Muestrear o agrupar
for i := 0; i < 1000000; i++ {
    if i % 1000 == 0 {
        logger.Debug("Progreso", go_logs.Int("i", i))
    }
}
```

### 2. Formateo de Strings en Mensajes

```go
// Mal: Formatea incluso si esta filtrado
logger.Debug(fmt.Sprintf("Procesando %d items", len(items)))

// Bien: Dejar que el logger formatee
logger.Debug("Procesando items", go_logs.Int("count", len(items)))
```

### 3. Hooks Bloqueantes

```go
// Mal: Bloquea en cada log
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

## Ver Tambien

- [Referencia API](API-Reference-es.md) - Documentacion de API
- [Hooks](Hooks-es.md) - Crear hooks eficientes
- [Ejemplos](Examples-es.md) - Patrones de produccion
