---
id: go_logs-instalable-bugs-prod-04
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: [go_logs-instalable-bugs-prod-01]
estimate: 3h
actual: 0.25h
created: 2026-09-26
updated: 2026-09-26
---

# `Flusher`: sin fsync por entrada, `Flush()` en los cuatro writers, `Logger.Sync()` real con errno ignorados

## Description

`writeEntry()` (`logger_impl.go:351-358`) llama a `output.Sync()` tras cada entrada. `RotatingFileWriter.Sync()` y `EnhancedRotatingFileWriter.Sync()` hacen `bufio.Flush()` **más `file.Sync()`** (fsync): el logging a fichero va a ~260 msg/s cuando el writer crudo hace 9M msg/s (B6 del informe). A la vez `Logger.Sync()` (`logger_impl.go:291-295`) es un no-op documentado como flush: con un `bufio.Writer` como output se pierden los últimos bytes al salir (A1). `MultiWriter` tiene `Sync()` pero no `Flush()` y `SamplingWriter` no tiene ninguno: el design-review demostró que sin cubrirlos `TestWithMultiOutput_WritesToAll` falla. Decisiones del owner: flush por entrada sin fsync; `Sync()` delega en el output y hace fsync; los errores `EINVAL`/`ENOTTY`/`EBADF` se ignoran por errno (stdout, tuberías, stdout envuelto). Ver plan, Objetivo 5.

## Spec

- Interfaz pública, en `formatter.go` o fichero nuevo `writer.go`:
  ```go
  // Flusher es implementado por outputs que bufferizan. El logger llama a Flush
  // DESPUÉS DE CADA ENTRADA para que ninguna línea quede retenida en memoria de
  // usuario; Flush no debe hacer fsync (eso es Sync). Un writer que no quiera flush
  // por entrada simplemente no implementa esta interfaz.
  type Flusher interface{ Flush() error }
  ```
- `RotatingFileWriter.Flush()` y `EnhancedRotatingFileWriter.Flush()`: bajo `w.mu`, `w.writer.Flush()` si `w.writer != nil`. `Sync()` no cambia (flush + fsync).
- `MultiWriter.Flush()`: itera los writers y llama a `Flush()` en los que sean `Flusher`; devuelve el último error. Mismo patrón que su `Sync()`.
- `SamplingWriter.Flush()` y `SamplingWriter.Sync()`: delegan en el writer envuelto si implementa `Flusher`/`Syncer`.
- Aserciones de compilación en tests: `var _ Flusher = (*RotatingFileWriter)(nil)` etc. para los cuatro tipos.
- `writeEntry()`: tras `l.output.Write(formatted)`, `if f, ok := l.output.(Flusher); ok { f.Flush() }`. Eliminar por completo la llamada a `Sync()`.
- `Logger.Sync()`:
  ```go
  func (l *LoggerImpl) Sync() error {
      out := l.getOutput() // bajo RLock
      s, ok := out.(interface{ Sync() error })
      if !ok { return nil }
      err := s.Sync()
      if isIgnorableSyncErr(err) { return nil }
      return err
  }
  ```
  `isIgnorableSyncErr` (no exportado): `errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) || errors.Is(err, syscall.EBADF)`. `os.File.Sync` devuelve `*os.PathError` envolviendo el `syscall.Errno`, así que `errors.Is` funciona. `syscall.ENOTTY` existe en Windows como constante en `syscall`; verificar que compila con `GOOS=windows` y si no, aislar en ficheros con build tag.
- Actualizar el godoc de `Logger.Sync` en `logger.go` (flush + fsync; errores de stdout/tuberías ignorados).
- Writer espía para tests (`spyWriter` en `logger_test.go`): cuenta `Write`, `Flush`, `Sync` y permite inyectar el error de `Sync`.
- Benchmark `BenchmarkLoggerToRotatingFile` en `benchmark_v3_test.go`: Logger → JSONFormatter → `RotatingFileWriter` en `b.TempDir()`, con `-benchmem`. Registrar el resultado en el session log y en CLAUDE.md (tarea 10). Cifra objetivo informativa: < 10 µs/op.
- `gofmt -w` de los ficheros tocados.

## Fuera de alcance

- Eliminar `bufio` del `RotatingFileWriter` o cambiar su tamaño de buffer.
- Añadir `Close()` a la interfaz `Logger`.
- Separar el mutex de escritura del de configuración y el deadlock reentrante (C3 del informe).
- El resto de bugs del rotador enhanced (borrado de ficheros ajenos, `compressFile`): plan aparte.
- `MultiWriter.Write` con solo `RLock` (M12 del informe): plan aparte.

## Scenarios (Gherkin)

```gherkin
Feature: Escribir una entrada no fuerza fsync

  Scenario: Cien entradas no provocan ninguna sincronización a disco
    Given un logger cuyo output es un writer espía que cuenta Write, Flush y Sync
    When se registran 100 mensajes
    Then el contador de Sync es 0
    And el contador de Write es 100
    And el contador de Flush es 100

  Scenario: Cada entrada es legible en el fichero inmediatamente
    Given un logger con RotatingFileWriter como output
    When se registra "línea 1"
    Then el fichero contiene "línea 1" sin haber llamado a Sync ni a Close

  Scenario: Salida múltiple fichero más buffer sigue escribiendo de inmediato
    Given un logger con MultiWriter(RotatingFileWriter, bytes.Buffer) como output
    When se registra "multi"
    Then el fichero contiene "multi"
    And el buffer contiene "multi"

  Scenario: Un writer con muestreo propaga el flush al writer envuelto
    Given un SamplingWriter que envuelve un RotatingFileWriter con tasa 1.0
    When se registra "sampled"
    Then el fichero contiene "sampled"

  Scenario: Un error de Flush por entrada no interrumpe el logging ni se propaga
    Given un logger con un writer espía cuyo Flush devuelve un error
    When se registran 3 mensajes
    Then los 3 Write se ejecutan sin pánico ni bloqueo
    And el error de Flush no se propaga al llamante de Log

  Scenario: Flush y Sync sobre un MultiWriter vacío no fallan
    Given un MultiWriter sin ningún writer
    When se llama a Flush y a Sync
    Then ambos devuelven nil

  Scenario Outline: Los writers del paquete implementan Flusher
    Given el tipo <writer>
    When se comprueba en compilación que implementa Flusher
    Then compila

    Examples:
      | writer                     |
      | RotatingFileWriter         |
      | EnhancedRotatingFileWriter |
      | MultiWriter                |
      | SamplingWriter             |

Feature: Logger.Sync flushea y sincroniza el output

  Scenario: Sync entrega al writer subyacente lo retenido en un bufio.Writer
    Given un logger cuyo output es un bufio.Writer sobre un bytes.Buffer
    When se registra "pendiente" y se llama a Sync
    Then el bytes.Buffer contiene "pendiente"

  Scenario: Sync llama exactamente una vez al Sync del output
    Given un logger con el writer espía como output
    When se registran 10 mensajes y se llama a Sync una vez
    Then el contador de Sync es 1

  Scenario: Sync sobre un output sin Sync ni Flush no falla
    Given un logger cuyo output es un bytes.Buffer
    When se registra "plain" y se llama a Sync
    Then Sync devuelve nil
    And el buffer contiene "plain"

  Scenario Outline: Errores de Sync propios de terminales y tuberías se ignoran
    Given un logger con un writer espía cuyo Sync devuelve <error>
    When se llama a Sync
    Then el resultado es <resultado>

    Examples:
      | error                              | resultado          |
      | PathError{Err: syscall.EINVAL}     | nil                |
      | PathError{Err: syscall.ENOTTY}     | nil                |
      | PathError{Err: syscall.EBADF}      | nil                |
      | io.ErrClosedPipe                   | io.ErrClosedPipe   |
      | errors.New("disk full")            | ese mismo error    |

  Scenario: Sync sobre stdout real no devuelve error
    Given un logger con os.Stdout como output
    When se llama a Sync
    Then el resultado es nil

  Scenario: Sync con salida múltiple ignora solo el error de terminal del escritor que falla
    Given un logger con MultiWriter(RotatingFileWriter real, writer espía cuyo Sync devuelve EBADF)
    When se llama a Sync
    Then el resultado es nil

  Scenario: Sync con salida múltiple propaga errores reales de cualquier escritor
    Given un logger con MultiWriter(writer válido, writer espía cuyo Sync devuelve "disk full")
    When se llama a Sync
    Then el resultado no es nil y menciona "disk full"

  Scenario: Logueo concurrente con Sync periódico no produce carreras
    Given un logger con RotatingFileWriter como output en un directorio temporal
    When 10 goroutines registran mensajes mientras otra goroutine llama a Sync repetidamente
    Then el race detector no reporta ninguna carrera

  Scenario: Sync sobre un fichero real propaga el fsync
    Given un logger con un RotatingFileWriter en un directorio temporal
    When se registra "durable" y se llama a Sync
    Then el resultado es nil
    And el fichero contiene "durable"

Feature: Rendimiento end-to-end honesto a fichero

  Scenario: El benchmark logger → JSON → RotatingFileWriter queda por debajo de 10 µs por entrada
    Given el benchmark BenchmarkLoggerToRotatingFile
    When se ejecuta con -benchmem
    Then el tiempo por operación es inferior a 10 µs (cifra informativa registrada en el session log)
```

## Provides

- Interfaz exportada `Flusher` y métodos `Flush()` en `RotatingFileWriter`, `EnhancedRotatingFileWriter`, `MultiWriter`, `SamplingWriter`.
- `Logger.Sync()` con semántica real (flush + fsync, errno de terminal ignorados): la tarea 05 (`async`) y la e2e de la tarea 10 dependen de ello.
- `spyWriter` de test reutilizable en `logger_test.go`.
- Número de benchmark end-to-end para la tabla de CLAUDE.md (tarea 10).

## Definition of Done

- [x] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor
- [x] Cada escenario Gherkin tiene al menos un test (camino feliz + bordes/errores)
- [x] Todos los tests en verde: `go test -race .` y `go test ./...`; `file_writer_test.go` intacto y en verde
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `gofmt -l` sin salida en los ficheros tocados; `go vet ./...` limpio; `GOOS=windows go build ./...` compila
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **godoc en el código** — `Flusher`, cada `Flush()`, `Logger.Sync` (interfaz e implementación), `isIgnorableSyncErr`
  - [x] **Doc técnica (contexto)** — README ("File Rotation" / "Sync") y CLAUDE.md: flush por entrada, fsync en `Sync()`, errores ignorados
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-04.md` con el número del benchmark
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-04: <conventional commit>`
