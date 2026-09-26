---
id: go_logs-instalable-bugs-prod-09
package: go_logs
plan: instalable-bugs-prod
status: pending
priority: 2
depends_on: [go_logs-instalable-bugs-prod-08]
estimate: 1h
actual:
created: 2026-09-26
updated: 2026-09-26
---

# Workflow de despliegue del sitio a GitHub Pages

## Description

El sitio VitePress de la tarea 08 necesita publicarse en `https://drossan.github.io/go_logs/`. Decisión del owner: despliegue con GitHub Actions (`actions/deploy-pages`), no con rama `gh-pages`. GitHub Pages no está activado en el repo (verificado con `gh api repos/drossan/go_logs/pages` → 404): activarlo con source "GitHub Actions" es un paso manual del owner en Settings que esta tarea documenta. Ver plan, Objetivo 8.

## Spec

- Fichero `.github/workflows/deploy-pages.yml`:
  - `on: push` a `main` con `paths: ['website/**']`, y `workflow_dispatch`.
  - `permissions: { contents: read, pages: write, id-token: write }`.
  - `concurrency: { group: pages, cancel-in-progress: false }`.
  - Job `build` (`ubuntu-latest`): `actions/checkout@v4`, `pnpm/action-setup@v4`, `actions/setup-node@v4` con `node-version: 22` y `cache: pnpm` (`cache-dependency-path: website/pnpm-lock.yaml`), `actions/configure-pages@v5`, `pnpm install --frozen-lockfile` y `pnpm build` con `working-directory: website`, `actions/upload-pages-artifact@v3` con `path: website/.vitepress/dist`.
  - Job `deploy` (`needs: build`, `environment: { name: github-pages, url: ${{ steps.deployment.outputs.page_url }} }`): `actions/deploy-pages@v4` con `id: deployment`.
- `website/README.md`: sección "Despliegue" con el paso manual: Settings → Pages → Build and deployment → Source: **GitHub Actions**; y cómo lanzar el workflow a mano la primera vez (`gh workflow run deploy-pages.yml`).
- Validar con `actionlint` si está disponible; si no, parser YAML y revisión manual.
- No hay push a `main` en esta tarea (el plan se mergea por PR): la ejecución real se verifica lanzando el workflow con `workflow_dispatch` desde la rama del plan **solo si** el owner ha activado Pages; si no, se registra como pendiente de la checklist de release.

## Fuera de alcance

- Activar Pages en Settings: acción del owner.
- Dominio propio / `CNAME`.
- Desplegar previews por PR.
- Tocar `ci.yml`, `release.yaml` o `sync-wiki.yml`.

## Scenarios (Gherkin)

```gherkin
Feature: Publicación del sitio en GitHub Pages

  # Tarea sin código Go testeable. Verificación: YAML válido y, si Pages está
  # activado, una ejecución manual en verde con la URL publicada.

  Scenario: El workflow es sintácticamente válido
    Given el fichero .github/workflows/deploy-pages.yml
    When se valida con actionlint (o un parser YAML)
    Then no hay errores

  Scenario: Solo se despliega cuando cambia el sitio
    Given el workflow
    When se inspecciona su disparador
    Then se ejecuta en push a main únicamente si cambian ficheros bajo website/
    And puede lanzarse manualmente

  Scenario: Los permisos son los mínimos para Pages
    Given el workflow
    When se inspeccionan sus permisos
    Then contents es read, pages es write e id-token es write

  Scenario: El despliegue no ocurre si el build falla
    Given el job build falla (p. ej. pnpm build con errores)
    Then el job deploy no se ejecuta

  Scenario: Despliegues simultáneos se encolan en vez de cancelarse
    Given un despliegue en curso
    When se dispara otro despliegue mientras el primero corre
    Then el segundo espera en cola sin cancelar al primero (grupo de concurrencia "pages")

  Scenario: El artefacto publicado es la salida de VitePress
    Given el workflow
    When se inspecciona el paso de subida del artefacto
    Then la ruta es website/.vitepress/dist

  Scenario: Ejecución manual con Pages activado
    Given Pages activado por el owner con source "GitHub Actions"
    When se lanza el workflow a mano desde la rama del plan
    Then termina en verde
    And la URL https://drossan.github.io/go_logs/ sirve la portada del sitio

  Scenario: Ejecución manual sin Pages activado
    Given Pages no activado
    When se lanza el workflow a mano
    Then el job deploy falla con un mensaje que indica que Pages no está configurado
    And la tarea registra el paso pendiente en el session log y en la checklist de release
```

## Provides

- Workflow `deploy-pages.yml` listo; URL pública del sitio para enlazar desde el README (tarea 10).

## Definition of Done

- [ ] Verificación registrada en el session log: YAML válido; resultado de la ejecución manual o el motivo de no haberla hecho
- [ ] Todos los tests en verde: `go test -race ./...` sigue en verde (esta tarea no toca Go)
- [ ] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [ ] Lint / format / typecheck OK: `actionlint` o parser YAML sin errores
- [ ] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [ ] Documentación actualizada — tres capas:
  - [ ] **Comentarios en el workflow** explicando permisos, concurrencia y disparadores
  - [ ] **Doc técnica (contexto)** — `website/README.md` sección "Despliegue" con el paso manual de Settings
  - [ ] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-09.md`
- [ ] Commit en la rama del plan: `go_logs-instalable-bugs-prod-09: ci: <descripción>`
