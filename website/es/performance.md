# Rendimiento

go_logs esta disenado para logging de alto rendimiento con sobrecarga minima. Esta pagina cubre benchmarks, caracteristicas de rendimiento y tips de optimizacion.

**[English](/performance)** | **Espanol**

## Benchmarks

### Operaciones Core

| Operacion | Rendimiento |
|-----------|-------------|
| Logger, nivel filtrado (`BenchmarkV3LevelFiltering`) | 2.2 ns/op, 0 allocs |
| Creación de campo | 0.32 ns/op, 0 allocs |
| TextFormatter, entrada simple | 187 ns/op, 4 allocs |
| JSONFormatter, entrada simple | 387 ns/op, 4 allocs |
| Logger → JSON → RotatingFileWriter (de extremo a extremo) | 3.4 µs/op (~300k msg/s), 16 allocs |

Medido en Apple M1 (arm64) el 2026-10-11 con `go test -run xxx -bench . -benchmem .`; tus cifras variarán según hardware y versión de Go. Reprodúcelas con ese comando.

### Benchmarks Detallados

Salida de `go test -run xxx -bench . -benchmem .` (Apple M1, 2026-10-11):

```
BenchmarkV3LevelFiltering-8                532573428     2.224 ns/op      0 B/op     0 allocs/op
BenchmarkV3FieldCreation-8                1000000000     0.3317 ns/op     0 B/op     0 allocs/op
BenchmarkLevelShouldLog/Filtered-8        1000000000     0.3587 ns/op     0 B/op     0 allocs/op
BenchmarkTextFormatter_SimpleEntry-8         6502068     186.6 ns/op    152 B/op     4 allocs/op
BenchmarkTextFormatter_WithFields-8          1706208     704.9 ns/op    616 B/op    21 allocs/op
BenchmarkJSONFormatter_SimpleEntry-8         3068571     387.2 ns/op    352 B/op     4 allocs/op
BenchmarkJSONFormatter_WithFields-8           989826     1243 ns/op     904 B/op    17 allocs/op
BenchmarkV3RotatingFileWriter-8              1000000     1201 ns/op    1284 B/op    16 allocs/op
BenchmarkLoggerToRotatingFile-8               365101     3356 ns/op    1016 B/op    16 allocs/op
BenchmarkV3ConcurrentLogging-8               1430082     833.4 ns/op   1026 B/op    15 allocs/op
BenchmarkV3Redaction-8                        929787     1313 ns/op    1425 B/op    19 allocs/op
```

`BenchmarkV3LevelFiltering` es el coste de llamar a un método del logger cuyo nivel está filtrado. `BenchmarkLevelShouldLog` y la creación de campos miden solo una comparación o construir un struct, por eso bajan de 1 ns.

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
// Esto es una sola comparacion de enteros
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

Los logs filtrados cuestan ~2 ns por llamada y no reservan memoria.

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

- [Referencia API](/es/api-reference) - Documentacion de API
- [Hooks](/es/hooks) - Crear hooks eficientes
- [Ejemplos](/es/examples) - Patrones de produccion
