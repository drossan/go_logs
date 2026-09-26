---
id: go_logs-instalable-bugs-prod-02
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: [go_logs-instalable-bugs-prod-01]
estimate: 2h
actual: 0.1h
created: 2026-09-26
updated: 2026-09-26
---

# `Init()` tolerante: sin `log.Fatalf` en `config.go`, avisos por stderr, nivel Info por defecto

## Description

`Init()` (`config.go:92-125`) llama a `log.Fatalf` cuando `SAVE_LOG_FILE` o `NOTIFICATIONS_SLACK_ENABLED` están vacías porque `strconv.ParseBool("")` devuelve error. Como `save.go:10` auto-llama a `Init()`, un binario que solo hace `go_logs.InfoLog("hola")` sin variables de entorno termina con `exit status 1`. Hay 7 `log.Fatalf` en `Init()`/`loadNotificationsConfig()` y uno en `initPersistentLogFile()` (`config.go:299`). Además, con el entorno vacío `logLevel` queda en 0 y `getNotificationSettings()` (`config.go:264-270`) devuelve `false` para todo, así que sin el crash la API v2 no escribiría nada. Decisiones del owner: valor vacío = default en silencio; valor inválido = default + aviso por stderr; nivel `Info` por defecto. Ver plan, Objetivo 3.

## Spec

- Nuevo helper no exportado en `config.go`:
  ```go
  // envBool lee una variable de entorno booleana. Vacía o ausente devuelve def en
  // silencio; un valor no reconocido por strconv.ParseBool devuelve def y escribe un
  // aviso en stderr con el nombre de la variable y el valor recibido.
  func envBool(key string, def bool) bool
  ```
  El aviso va a un `io.Writer` de paquete (`warnOutput io.Writer = os.Stderr`) para poder capturarlo en tests. Formato: `go_logs: variable de entorno KEY="valor" no válida, se usa el valor por defecto (false)`.
- Sustituir los 7 `strconv.ParseBool` + `log.Fatalf` de `Init()` y `loadNotificationsConfig()` por `envBool(..., false)`.
- `initPersistentLogFile()`: si `os.OpenFile` falla, escribir aviso en `warnOutput` (`go_logs: no se puede abrir el fichero de log <ruta>: <err>; guardado en fichero desactivado`), poner `saveLogFile = false` y `logWriter = nil`. Sin `log.Fatalf`.
- `loadLogLevel()`: cuando `LOG_LEVEL` está vacío o no es un nivel reconocido, `logLevel = LevelInfo`; si es inválido (no vacío) emitir aviso por `warnOutput`. Comprobar que `getNotificationSettings()` con el default resultante devuelve `true` para `info`, `warning`, `error`, `fatal` y `success`, y `false` para `debug`/`trace`.
- `Init()` sigue sin devolver error (compatibilidad v2). Añadir en godoc que nunca termina el proceso.
- Verificación final: `grep -n 'log.Fatal' config.go` vacío. `logs.go:29` (`FatalLog`) **no se toca**.
- Tests en `config_test.go`: usar `t.Setenv` y reset de las globales (`isInit`, `saveLogFile`, `logFileOnce`…) mediante un helper `resetConfigForTest(t)` si no existe ya; capturar `warnOutput` con un `bytes.Buffer`.
- `gofmt -w config.go` (hoy no está formateado).

## Fuera de alcance

- Sacar `adapters/` y `loadSlackConfig()` del core: tarea 06. Aquí `NOTIFICATIONS_SLACK_ENABLED` se lee con `envBool` y el resto sigue igual.
- Otros globales v2 con races distintas de `Init()` (A6 del informe): plan aparte.
- Cambiar la semántica de `Fatal`/`FatalLog`: siguen terminando el proceso.
- Reescribir la documentación de configuración más allá de la nota sobre `Init()` tolerante (tarea 10 hace el resto).

## Scenarios (Gherkin)

```gherkin
Feature: La configuración por entorno nunca termina el proceso del consumidor

  Scenario Outline: Lectura tolerante de variables booleanas
    Given la variable de entorno "X" con el valor <valor>
    When se lee como booleano con valor por defecto false
    Then el resultado es <resultado>
    And el aviso en stderr <aviso>

    Examples:
      | valor    | resultado | aviso                                     |
      | ausente  | false     | no se emite                               |
      | ""       | false     | no se emite                               |
      | "1"      | true      | no se emite                               |
      | "true"   | true      | no se emite                               |
      | "TRUE"   | true      | no se emite                               |
      | "0"      | false     | no se emite                               |
      | "false"  | false     | no se emite                               |
      | "quizas" | false     | contiene X y "quizas" y "valor por defecto" |
      | " true " | false     | contiene X (espacios no son válidos)      |

  Scenario: Inicialización con el entorno vacío
    Given ninguna variable de entorno de go_logs definida
    When se inicializa la configuración
    Then el proceso sigue vivo
    And el guardado en fichero está desactivado
    And las notificaciones Slack están desactivadas
    And el nivel de log efectivo es Info
    And no se emite ningún aviso

  Scenario: El quickstart v2 escribe algo sin configuración
    Given ninguna variable de entorno de go_logs definida
    And la salida estándar capturada
    When se registra "hola" con InfoLog
    Then la salida contiene "hola"

  Scenario Outline: Nivel de log por defecto y niveles inválidos
    Given la variable LOG_LEVEL con el valor <valor>
    When se inicializa la configuración
    Then el nivel efectivo es <nivel>
    And el aviso en stderr <aviso>

    Examples:
      | valor    | nivel | aviso          |
      | ausente  | Info  | no se emite    |
      | "debug"  | Debug | no se emite    |
      | "ERROR"  | Error | no se emite    |
      | "loud"   | Info  | contiene "loud" |

  Scenario Outline: getNotificationSettings con el nivel Info por defecto
    Given el entorno vacío tras inicializar la configuración
    When se consulta getNotificationSettings para <nivel>
    Then el resultado es <resultado>

    Examples:
      | nivel   | resultado |
      | info    | true      |
      | warning | true      |
      | error   | true      |
      | fatal   | true      |
      | success | true      |
      | debug   | false     |
      | trace   | false     |

  Scenario: Reinicialización con LOG_LEVEL distinto actualiza el nivel efectivo
    Given Init() ya se llamó con LOG_LEVEL=info
    When se cambia LOG_LEVEL a "debug" y se vuelve a llamar a Init()
    Then el nivel efectivo pasa a ser Debug

  Scenario: El fichero de log no se puede abrir
    Given SAVE_LOG_FILE=1 y LOG_FILE_PATH apuntando a un directorio inexistente
    When se inicializa la configuración y se registra "hola" con InfoLog
    Then el proceso sigue vivo
    And el aviso en stderr contiene la ruta y "desactivado"
    And el guardado en fichero queda desactivado
    And la salida estándar contiene "hola"

  Scenario: La configuración no contiene terminaciones de proceso
    Given el fichero config.go
    When se buscan llamadas a log.Fatal
    Then no hay ninguna
```

## Provides

- `envBool(key, def)` y `warnOutput` (no exportados) reutilizables por la tarea 06 al leer `NOTIFICATIONS_SLACK_ENABLED`.
- `Init()` garantizado sin `os.Exit`: la e2e de la tarea 10 (binario sin env termina con exit 0) depende de esto.

## Definition of Done

- [x] Tests escritos ANTES de la implementación (TDD) — Red → Green → Refactor
- [x] Cada escenario Gherkin tiene al menos un test (camino feliz + bordes/errores)
- [x] Todos los tests en verde: `go test ./...`
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `gofmt -l config.go config_test.go` sin salida; `go vet ./...` limpio
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **godoc en el código** — `envBool`, `Init()` (nunca termina el proceso; defaults), `initPersistentLogFile`
  - [x] **Doc técnica (contexto)** — sección "Variables de entorno" de README y CLAUDE.md: vacío = default, inválido = aviso, nivel Info por defecto
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-02.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-02: <conventional commit>`
