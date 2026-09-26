---
id: go_logs-instalable-bugs-prod-10
package: go_logs
plan: instalable-bugs-prod
status: pending
priority: 2
depends_on: [go_logs-instalable-bugs-prod-07, go_logs-instalable-bugs-prod-09]
estimate: 3h
actual:
created: 2026-09-26
updated: 2026-09-26
---

# README y wiki con snippets corregidos, `CLAUDE.md` veraz, changelog `v3.1.0` consolidado, GoReleaser v2, verificación e2e

## Description

Cierre del plan: las tres fuentes de documentación (`README.md`, `docs/wiki/`, `website/`) deben contar la misma verdad y compilar; `CLAUDE.md` describe hoy una arquitectura de submódulos que ya no existe y una tabla de rendimiento falsa; el changelog del README anuncia "v3.5 (Actual)" con secciones v3.1-v3.5 mientras los tags paran en `v3.0.4`; y `release.yaml` corre GoReleaser v2 sobre un `.goreleaser.yaml` con sintaxis v1 (fallará en el primer `push --tags`). Decisiones del owner: release `v3.1.0` consolidando el changelog; snippets corregidos también en README y wiki. Ver plan, Objetivos 8 y 9, y la checklist de release.

## Spec

- **Snippets en `README.md` y `docs/wiki/*.md`** (los mismos cuatro patrones y ajustes que la tarea 08 aplicó a `website/`): `logger, err := go_logs.New(...)`; `WithHooks`; `hooks.NewSlackHook(notifier, level)` con `slack.NewNotifier`; `GetMetrics` vía `MetricsGetter`/`*LoggerImpl`; `defer asyncLogger.Close()`; import `/v3`. Verificación: `grep -rnE 'logger := go_logs\.New\(|WithHook\(|NewSlackHook\("' README.md CLAUDE.md docs/wiki --include='*.md'` vacío.
- **README**: badge del sitio / enlace "Documentación completa: https://drossan.github.io/go_logs/"; sección "Instalación" con `go get github.com/drossan/go_logs/v3@v3.1.0`; nota "las versiones v3.0.x no eran instalables con módulos Go (faltaba el sufijo /v3); usa v3.1.0 o superior"; sección Slack con `SetNotifier` + `go_logs/slack/v3`; tabla de rendimiento sustituida por los números medidos en la tarea 04 (benchmark end-to-end) y en la 03 (filtrado), con la fecha y la máquina; changelog: una sola entrada `### v3.1.0` (fecha) que lista lo de este plan y una subentrada "Historia previa (publicada como v3.0.0-v3.0.4, no instalable)" que absorbe las actuales secciones v3.1-v3.5; eliminar "v3.5 (Actual)".
- **`CLAUDE.md`**: eliminar la sección "Arquitectura Híbrida" y el segundo árbol duplicado; nueva estructura real: un módulo `github.com/drossan/go_logs/v3` con subpaquetes `async/`, `hooks/`, `http/`, `otel/`, `signal/`, `domain/` y el submódulo `slack/`; ejemplos con import `/v3`; tabla de rendimiento con los números de este plan; sección "Comandos comunes" con `go test -race ./...`, `(cd slack && go test ./...)`, `pnpm --dir website build`; notas de `Init()` tolerante, `Sync()` real, `Flusher`, nivel compartido, `SetNotifier`, `Close()` de async.
- **`MIGRATION.md`**: comprobar que la sección Slack (tarea 06) y el import `/v3` están; añadir "v3.0.x → v3.1.0" con los cambios de comportamiento: `Init()` ya no termina el proceso; nivel Info por defecto en v2; `Sync()` hace flush + fsync; `SetLevel` afecta a hijos; `async.Close()` en hijos es no-op.
- **`.goreleaser.yaml`**: `version: 2` en la primera clave; `changelog.skip` → `changelog.disable: false`; mantener `builds: [{skip: true}]`. Validar con `goreleaser check` si está instalado (`brew install goreleaser`); si no, revisar contra `goreleaser.com/deprecations`. Documentar en el session log qué se validó.
- **Verificación e2e** (script reproducible en `scripts/e2e-consumer.sh`, opcional pero recomendado):
  1. Módulo temporal en `/tmp` con `replace github.com/drossan/go_logs/v3 => <ruta>` y `replace github.com/drossan/go_logs/slack/v3 => <ruta>/slack`.
  2. `main.go` que importa `v3`, `v3/async`, `slack/v3`; llama a `go_logs.InfoLog("hola v2")`, crea un `New()` con JSON, un hijo con `With`, un async con `Close`, y `SetNotifier(nil)`.
  3. Ejecutar **sin variables de entorno** (`env -i PATH=$PATH HOME=$HOME go run .`): exit 0 y salida que contiene "hola v2".
  4. `go list -deps .` de un módulo que solo importa `v3`: sin `slack-go`.
  5. Anotar en el session log que esto prueba compilación e integración, no instalabilidad desde el proxy (eso va en la checklist post-tag).
- **Checklist de release** copiada **en ambos sitios** (no es una alternativa: el plan y el README): al final del plan y en una sección "Release" del `README.md`: tag raíz, verificación en proxy, tag `slack/v3.1.0`, `release.yaml` en verde, activar Pages.
- Marcar en el plan las casillas de las 10 tareas y añadir la nota retro (estimación vs real, sorpresas).
- **Heredado de la tarea 08 (2026-09-27)**: `website/changelog.md` (y `es/`) ya cuenta la historia real (`v3.0.0`-`v3.0.4` consolidados, `v3.1.0` sin publicar); el changelog del README debe coincidir con él y la entrada `v3.1.0` del sitio se completa con lo que añadan 09-11. `docs/wiki/API-Reference*.md` y `Migration-v2-to-v3*.md` citan `TraceLog`/`DebugLog`/`Tracef`/`Debugf` como API v2 pero no existen (ya retirados en `website/`). `docs/wiki/Installation*.md`, `Configuration*.md`, `Hooks*.md` y `API-Reference.md` (`IsNotifierEnabled`) siguen con el texto de Slack de v3.0: copiar las notas de `SetNotifier`/`slack/v3` de `website/`.

## Fuera de alcance

- README raíz en inglés (se mantiene en español; el sitio es la versión bilingüe).
- `examples/` compilados en CI, `doc.go`, `CHANGELOG.md` separado, `CONTRIBUTING.md` completo.
- Limpieza de `.claude/agents|commands|skills` ajenos, `IMPLEMENTATION_SUMMARY.md`, `PHASE2_SUMMARY.md`, `test.log`.
- Ejecutar `git tag` o activar Pages.
- Retirar `docs/wiki/` o `sync-wiki.yml`.

## Scenarios (Gherkin)

```gherkin
Feature: La documentación compila y cuenta la verdad

  # Tarea mayoritariamente documental. Los escenarios de grep y e2e se verifican con
  # comandos registrados en el session log; el script e2e es ejecutable.

  Scenario: Ningún snippet de README, CLAUDE.md o wiki usa la API rota
    Given README.md, CLAUDE.md y docs/wiki/*.md
    When se buscan "logger := go_logs.New(", "WithHook(" y "NewSlackHook(\""
    Then no hay ninguna coincidencia

  Scenario: Todos los imports de la documentación apuntan a v3
    Given README.md, CLAUDE.md, MIGRATION.md y docs/wiki/*.md
    When se buscan imports de github.com/drossan/go_logs sin /v3
    Then solo aparecen en MIGRATION.md y en la página de migración, citando v2 a propósito

  Scenario: CLAUDE.md describe la estructura real del módulo
    Given CLAUDE.md
    When se busca "Arquitectura Híbrida" o "Submódulo"
    Then no hay coincidencias
    And la tabla de rendimiento cita los benchmarks medidos en este plan con fecha

  Scenario: El changelog tiene una única entrada v3.1.0 y ninguna v3.5
    Given README.md
    When se listan los encabezados "### v3."
    Then hay exactamente uno "### v3.1.0" y uno "### v3.0"
    And ninguno dice "v3.5"

  Scenario: El README cita la tabla de rendimiento medida en este plan
    Given README.md
    When se inspecciona su sección de rendimiento
    Then los números coinciden con los benchmarks de las tareas 03 y 04, con fecha y máquina

  Scenario: MIGRATION.md documenta los cambios de comportamiento v3.0.x a v3.1.0
    Given MIGRATION.md
    When se busca la sección "v3.0.x → v3.1.0"
    Then menciona que Init() ya no termina el proceso
    And menciona el nivel Info por defecto en v2
    And menciona que Sync hace flush y fsync
    And menciona que SetLevel afecta a hijos
    And menciona que Close() en hijos de async es no-op

Feature: Un consumidor externo usa la librería sin configuración

  Scenario: Binario mínimo sin variables de entorno
    Given un módulo temporal que importa v3, v3/async y slack/v3 vía replace
    When se ejecuta con el entorno vacío
    Then termina con código 0
    And la salida contiene "hola v2"
    And la salida contiene una línea JSON del logger v3 con el campo del hijo

  Scenario: Un consumidor que solo importa el core no arrastra Slack
    Given un módulo temporal que importa únicamente v3
    When se listan sus dependencias
    Then no aparece slack-go ni gorilla/websocket

Feature: El release no falla en el workflow

  Scenario: La configuración de GoReleaser es válida para v2
    Given .goreleaser.yaml
    When se valida con goreleaser check (o contra la lista de deprecaciones si no está instalado)
    Then no hay errores ni claves deprecadas
    And la primera clave es version: 2

  Scenario: La checklist de release está documentada en el plan y en el README
    Given el plan y una sección "Release" del README
    When se busca la checklist de release en ambos
    Then los dos contienen, en orden: tag v3.1.0, verificación go list -m -versions, tag slack/v3.1.0, release.yaml en verde, activar Pages
```

## Provides

- — (última tarea del plan; su salida es el estado listo para PR y release).

## Definition of Done

- [ ] Verificación registrada en el session log: greps vacíos, salida del e2e, validación de GoReleaser
- [ ] Todos los tests en verde: `go test -race ./...` en raíz y `slack/`; `pnpm --dir website build` en verde
- [ ] Spec cumplida
- [ ] Lint / format / typecheck OK: `gofmt -l` sin salida; `go vet ./...`; CI en verde en el último push
- [ ] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [ ] Documentación actualizada — tres capas:
  - [ ] **godoc** — sin símbolos nuevos; si el script e2e vive en Go, documentado
  - [ ] **Doc técnica (contexto)** — README, CLAUDE.md, MIGRATION.md según Spec
  - [ ] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-10.md`; nota retro en el plan
- [ ] Commit en la rama del plan: `go_logs-instalable-bugs-prod-10: docs: <descripción>`
- [ ] Plan movido a `completed/` con `status: completed` y PR abierta desde `plan/go_logs/instalable-bugs-prod` a `develop` (commit y PR manuales: `git-automation` off)
