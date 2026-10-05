---
id: go_logs-instalable-bugs-prod-07
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 1
depends_on: [go_logs-instalable-bugs-prod-03, go_logs-instalable-bugs-prod-04, go_logs-instalable-bugs-prod-05, go_logs-instalable-bugs-prod-06]
estimate: 1h
actual: 15m
created: 2026-09-26
updated: 2026-09-27
---

# `ci.yml` mínimo: gofmt, vet, test -race en raíz y `slack/`

## Description

No existe ningún workflow que ejecute tests (solo `release.yaml` y `sync-wiki.yml`). Tras las tareas 03-06 el repo tiene por primera vez un `go test -race ./...` significativo y cinco fixes de concurrencia y durabilidad que nada impide volver a romper. Esta tarea añade el CI mínimo en Linux; la matriz completa (versiones de Go, macOS/Windows, lint, cobertura) queda para el punto 3 del roadmap. Decisión B3 del design-review. Ver plan, Objetivo 7.

## Spec

- Fichero `.github/workflows/ci.yml`:
  - `on: push` (todas las ramas) y `pull_request`.
  - `permissions: contents: read`.
  - Un job `test` en `ubuntu-latest`: `actions/checkout@v4`, `actions/setup-go@v5` con `go-version: stable` y `cache: true`.
  - Pasos, en este orden, cada uno fallando el job:
    1. `gofmt`: `test -z "$(gofmt -l $(git ls-files '*.go'))"` (excluye `website/` y `vendor/` por construcción al listar solo `.go` trackeados).
    2. `go vet ./...` en la raíz y `go vet ./...` en `slack/`.
    3. `go test -race -count=1 ./...` en la raíz.
    4. `go test -race -count=1 ./...` en `slack/` (con `working-directory: slack`).
    5. `GOOS=windows go build ./...` en la raíz (compilación cruzada, sin tests).
  - `concurrency` con `cancel-in-progress: true` por ref.
- Badge en el README, justo debajo de los existentes: `[![CI](https://github.com/drossan/go_logs/actions/workflows/ci.yml/badge.svg)](https://github.com/drossan/go_logs/actions/workflows/ci.yml)`.
- Validar el YAML con `actionlint` si está instalado (`brew install actionlint`); si no, con `python3 -c 'import yaml,sys; yaml.safe_load(open(sys.argv[1]))'` y revisión manual.
- Ejecutar localmente la misma secuencia de comandos y registrar la salida en el session log.

## Fuera de alcance

- Matriz de versiones de Go, macOS/Windows como runners, `golangci-lint`, cobertura y Codecov.
- Tocar `release.yaml` o `.goreleaser.yaml` (tarea 10).
- Workflow de Pages (tarea 09).

## Scenarios (Gherkin)

```gherkin
Feature: Integración continua mínima para el módulo Go

  # Esta tarea no produce código Go testeable. Verificación: el YAML es válido, la
  # secuencia de comandos del workflow pasa en local, y el primer push a la rama del
  # plan muestra el job en verde en GitHub Actions.

  Scenario: El workflow es sintácticamente válido
    Given el fichero .github/workflows/ci.yml
    When se valida con actionlint (o un parser YAML si actionlint no está disponible)
    Then no hay errores

  Scenario: La secuencia del workflow pasa en local sobre la rama del plan
    Given la rama del plan con las tareas 03-06 cerradas
    When se ejecutan en local los mismos comandos del workflow en el mismo orden
    Then gofmt no lista ningún fichero
    And go vet termina sin error en raíz y en slack/
    And go test -race termina en verde en raíz y en slack/
    And la compilación cruzada a Windows termina sin error

  Scenario: Un fichero sin formatear hace fallar el job
    Given un fichero .go con formato incorrecto introducido temporalmente
    When se ejecuta el paso de gofmt del workflow
    Then el paso falla listando ese fichero

  Scenario Outline: Un fallo en cualquier otro paso hace fallar el job en ese paso
    Given un fallo introducido en <paso>
    When se ejecuta el workflow
    Then el job termina en rojo específicamente en <paso>

    Examples:
      | paso                                          |
      | go vet en la raíz                              |
      | go vet en slack/                               |
      | go test -race en la raíz                       |
      | go test -race en slack/ (working-directory)    |
      | compilación cruzada GOOS=windows                |

  Scenario: Un segundo push cancela la ejecución anterior del mismo ref
    Given una ejecución del workflow en curso para una rama
    When se hace un segundo push a la misma rama
    Then la ejecución anterior se cancela y solo continúa la nueva

  Scenario: Los permisos del workflow son mínimos
    Given el workflow ci.yml
    When se inspeccionan sus permisos declarados
    Then solo declara contents: read

  Scenario: El primer push dispara el job y termina en verde
    Given el workflow commiteado en la rama del plan
    When se hace push a origin
    Then GitHub Actions ejecuta "CI" sobre ese commit y termina con éxito
    And el badge del README apunta a ese workflow
```

## Provides

- Workflow `CI` que protege los fixes de las tareas 02-06 en cada push/PR. La PR de cierre del plan exige este job en verde.

## Definition of Done

- [x] Verificación registrada en el session log: YAML válido, secuencia local en verde, job en verde en GitHub tras el push (con enlace a la ejecución)
- [x] Todos los tests en verde: `go test -race ./...` en raíz y `slack/`
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: el propio job lo comprueba
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **Comentarios en el workflow** explicando cada paso (equivalente a godoc para YAML)
  - [x] **Doc técnica (contexto)** — badge en README; sección "Comandos comunes" de CLAUDE.md menciona que CI ejecuta la misma secuencia
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-07.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-07: ci: <descripción>`
