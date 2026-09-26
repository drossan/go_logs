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
