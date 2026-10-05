# Session log — go_logs-instalable-bugs-prod-07

> Append-only.

## 2026-09-26 23:51 — Arranque

- Rama `plan/go_logs/instalable-bugs-prod` (ya existente, sin upstream). `depends_on: [03, 04, 05, 06]` → las cuatro en `completed/`. Ninguna otra tarea del plan en `active/`.
- Tarea movida a `.claude/tasks/active/go_logs/` (`status: active`).
- Plan: Red (paso gofmt del Spec en rojo; `ci.yml` inexistente) → Green (`ci.yml` según Spec + `gofmt -w` de los 6 ficheros previos en commit aparte) → actionlint → fallo inducido por paso → badge/CLAUDE.md → secuencia local completa → fact-checker → commit → push (autorizado por el owner) y seguimiento del run.
- Decisión: formatear `api.go`, `api_test.go`, `context.go`, `level.go`, `logs.go`, `testing.go` en esta tarea. El Spec fija `gofmt -l $(git ls-files '*.go')` sobre todo el repo y el escenario "la secuencia pasa en local" exige salida vacía: sin formatearlos el primer run saldría en rojo. Commit separado (`style:`) para que el diff de CI quede limpio.

### Red

- `test -z "$(gofmt -l $(git ls-files '*.go'))"` → `exit=1`, lista `api.go api_test.go context.go level.go logs.go testing.go`.
- `ls .github/workflows/ci.yml` → `No such file or directory`.
- Toolchain local: `go1.27.0 darwin/arm64`. `actionlint` no instalado.

### Green

- `gofmt -w` de los 6 ficheros → solo espacios y sangría de comentarios godoc (verificado con el diff filtrando líneas no-comentario: solo realineado de `map`/`struct`). `go build ./...` ok. Commit `825f3db` (`style:`); incluye también el `git mv` de la tarea a `active/` que estaba en el índice.
- `.github/workflows/ci.yml` según Spec. Desviaciones deliberadas, dentro del Spec:
  - `go vet` raíz y `slack/` en **dos pasos** (el Spec los agrupa en el punto 2): el Scenario Outline exige que el job caiga "específicamente" en cada uno.
  - El paso gofmt usa `unformatted=$(gofmt -l …)` + `echo` + `exit 1` en vez de `test -z` a secas: mismo criterio de fallo, pero lista los ficheros (escenario "falla listando ese fichero").
- Sin secretos: el workflow no referencia `secrets.*`; la única variable de entorno es `GOOS=windows` en el paso de compilación cruzada.

### Verificación

| Comando | Resultado |
|---|---|
| `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/ci.yml` (v1.7.12) | exit 0, sin salida. Sanity: una clave `bogus-key` inyectada en una copia da `unexpected key "bogus-key" for "job" section`, exit 1. `shellcheck` no instalado → el `run:` no pasa por shellcheck (revisado a mano). |
| `python3 yaml.safe_load` | ok; `permissions={'contents': 'read'}`, `concurrency.cancel-in-progress=True`, 8 pasos en el orden del Spec. |
| `/tmp/ci_local.sh` (réplica de los pasos del workflow, para en el primer fallo) sobre la rama | los 6 pasos `exit=0`; raíz: 6 paquetes `ok` + `domain` sin tests; `slack/v3` `ok`. |
| Fallos inducidos en worktree temporal de HEAD, uno por vez | gofmt → `FAILED AT: gofmt` listando `zz_badfmt.go`; `Printf %d` con string en raíz → `FAILED AT: go vet (root)`; ídem en `slack/` → `FAILED AT: go vet (slack/)`; `t.Fatal` en raíz → `FAILED AT: go test -race (root)`; ídem en `slack/` → `FAILED AT: go test -race (slack/)`; fichero `//go:build windows` que no compila → `FAILED AT: Cross-compile for Windows`. Worktree eliminado. |

- Limitación: los fallos inducidos prueban la réplica local, no el runner de GitHub. `cancel-in-progress` no se ejercita en local: se verifica por inspección del YAML (y, si coincide, por un run cancelado en Actions).

### Docs

- `README.md`: badge CI debajo de GoDoc/Go Report Card (URL exacta del Spec).
- `CLAUDE.md`: subsección "CI (`.github/workflows/ci.yml`)" dentro de "Comandos Comunes de Desarrollo" con la secuencia a replicar en local.
- Comentarios en `ci.yml` por paso (capa "godoc" del YAML).

### Fact-checker

- Subagente Sonnet (`models.fact-checker: sonnet`), 8 afirmaciones: **8/8 VERIFICADO**, 0 INCORRECTO, 0 NO VERIFICABLE. Comprobó el rojo de gofmt en un worktree de `825f3db^`, re-ejecutó actionlint (v1.7.12, exit 0) y la secuencia completa (gofmt/vet/test -race en raíz y `slack/`, cross-compile Windows) en verde.

### Push y run en GitHub

- `git push -u origin plan/go_logs/instalable-bugs-prod` (autorizado por el owner al arrancar; primera publicación de la rama).
- Run `CI` sobre `596268e`: https://github.com/drossan/go_logs/actions/runs/36274603143 → **success** (21:56:10Z → 21:57:27Z, ~1m17s). Los 9 pasos ✓: Checkout, Set up Go (`stable`), gofmt, go vet (root), go vet (slack/), go test -race (root), go test -race (slack/), Cross-compile for Windows.
- Anotaciones (avisos, no fallos): `actions/checkout@v4` y `actions/setup-go@v5` apuntan a Node.js 20, deprecado (GitHub los fuerza a Node 24); `ubuntu-latest` migra a Ubuntu 26 desde el 2026-10-19.
- `cancel-in-progress`: un solo push, no se observó una cancelación; verificado por inspección del YAML (`concurrency.group: ci-${{ github.ref }}`, `cancel-in-progress: true`).

## 2026-09-26 23:58 — Cierre

**Resumen**: `.github/workflows/ci.yml` añadido: un job en `ubuntu-latest` con Go stable, en push y PR, `contents: read`, cancelación por ref. Ejecuta gofmt (sobre los `.go` trackeados), `go vet` y `go test -race -count=1` en la raíz y en `slack/`, y compilación cruzada a Windows. Seis ficheros previos se formatearon con `gofmt` para que el primer run salga en verde. Badge en el README y secuencia documentada en `CLAUDE.md`. El primer run en GitHub sale en verde.

**Decisiones + porqué**
- Formatear los 6 ficheros en esta tarea (commit `style:` separado): el Spec fija gofmt sobre todos los `.go` trackeados; sin ello el job nacía en rojo.
- vet raíz / vet `slack/` como pasos separados, y gofmt listando los ficheros: el Gherkin exige fallar "específicamente" en cada paso y "listando ese fichero".
- actionlint vía `go run …@latest` en vez de `brew install`: no toca el sistema ni el proyecto. La action no se añade al workflow.

**Verificaciones**: ver tablas de arriba (actionlint, secuencia local, 6 fallos inducidos, run de GitHub). DoD: ✓ en todos los ítems. Mutation: no aplica (`stack.mutation-tool: none`).

**Docs**: `README.md` (badge CI: visibilidad del estado), `CLAUDE.md` (subsección CI: la secuencia local es la del CI), comentarios en `ci.yml`, este log.

**Ficheros / commits**
- `825f3db` — `style: gofmt pre-existing unformatted files` (`api.go`, `api_test.go`, `context.go`, `level.go`, `logs.go`, `testing.go` + `git mv` de la tarea a `active/`).
- `596268e` — `ci: add minimal CI workflow …` (`.github/workflows/ci.yml`, `README.md`, `CLAUDE.md`, este log).
- Commit de cierre (`docs:`): tarea → `completed/`, casilla en el plan, este log.

**Tiempo real**: ~15 min (lectura desde ~23:45, arranque 23:51, cierre 23:58).

**Follow-ups**
- Actualizar `actions/checkout`/`actions/setup-go` a versiones con Node 24 cuando se aborde la matriz de CI (fuera de alcance: el Spec fija `@v4`/`@v5`).
- `TestAsyncLogger_NonBlocking` (< 10 ms con `-race`) podría dar un falso rojo en un runner cargado (señalado en la 05); en el primer run pasó.
- Proteger `main` exigiendo el check `CI / Test` en la PR de cierre del plan (acción del owner en Settings).
