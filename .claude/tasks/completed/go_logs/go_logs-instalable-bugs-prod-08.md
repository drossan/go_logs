---
id: go_logs-instalable-bugs-prod-08
package: go_logs
plan: instalable-bugs-prod
status: done
priority: 2
depends_on: [go_logs-instalable-bugs-prod-06]
estimate: 5h
actual: 0.4h
created: 2026-09-26
updated: 2026-09-27
---

# Sitio VitePress en `website/` con i18n y contenido migrado del wiki con snippets corregidos

## Description

El proyecto no tiene sitio de documentación propio; la documentación vive en `docs/wiki/` (14 páginas en inglés, 14 en español, más `Home`, `README` y `_Sidebar`) y se sincroniza al wiki de GitHub. Decisiones del owner: sitio nuevo con VitePress en `website/`, contenido **copiado** del wiki (el wiki sigue existiendo), inglés por defecto y español bajo `/es/`, desplegado en GitHub Pages (tarea 09). Al copiar se corrigen los cuatro patrones de snippets que no compilan y el import path pasa a `/v3`. Es la tarea más larga del plan: el revisor de diseño advirtió que corregir los snippets de 28 páginas es el grueso del trabajo. Ver plan, Objetivo 8.

## Spec

- `website/package.json` (pnpm, `"private": true`): dependencia de desarrollo `vitepress` (última estable) y scripts `dev`, `build`, `preview` (`vitepress dev|build|preview`). `website/pnpm-lock.yaml` commiteado. Node 22.
- `website/.vitepress/config.ts`:
  - `base: '/go_logs/'`, `title: 'go_logs'`, `lastUpdated: true`, `cleanUrls: true`.
  - `locales`: `root` (`label: 'English'`, `lang: 'en'`) y `es` (`label: 'Español'`, `lang: 'es'`, `link: '/es/'`), cada uno con su `themeConfig.nav` y `themeConfig.sidebar` derivados de `docs/wiki/_Sidebar.md`.
  - `themeConfig.search: { provider: 'local' }`, `socialLinks` a GitHub, `editLink` al repo.
- Contenido: cada `docs/wiki/<Page>.md` → `website/<page>.md` (kebab-case minúsculas) y cada `docs/wiki/<Page>-es.md` → `website/es/<page>.md`. `Home.md` → `website/index.md` con layout `home` (hero + features); `Home-es.md` → `website/es/index.md`. `README.md` y `_Sidebar.md` del wiki no se copian (su función la cumple la config).
- Correcciones obligatorias en la copia (y solo en la copia; README y `docs/wiki/` los corrige la tarea 10):
  1. `github.com/drossan/go_logs` → `github.com/drossan/go_logs/v3` en imports y `go get` (excepto en la página de migración donde se cita v2 a propósito, marcado como tal).
  2. `logger := go_logs.New(...)` → `logger, err := go_logs.New(...)` con manejo de `err` (o `logger, _ :=` en snippets ilustrativos, con comentario).
  3. `go_logs.WithHook(h)` → `go_logs.WithHooks(h)`.
  4. `hooks.NewSlackHook("xoxb-token", "C123456")` → construcción con `slack.NewNotifier(token, channel)` + `hooks.NewSlackHook(notifier, go_logs.ErrorLevel)`.
  5. `logger.GetMetrics()` sobre la interfaz → type assertion a `go_logs.MetricsGetter` o `*go_logs.LoggerImpl`, con nota.
  6. `defer asyncLogger.Sync()` → `defer asyncLogger.Close()` (tarea 05); nota de `SetNotifier` (tarea 06) en las páginas de configuración y hooks.
- Enlaces internos relativos (`[Formatters](Formatters)`) → rutas del sitio (`/formatters`, `/es/formatters`).
- `website/README.md`: cómo desarrollar (`pnpm install && pnpm dev`), estructura, y la declaración: "Este sitio es la fuente canónica de la documentación de usuario a partir de v3.1.0; `docs/wiki/` se mantiene sincronizado con el wiki de GitHub pero no se edita primero".
- `.gitignore` raíz: `website/node_modules/`, `website/.vitepress/dist/`, `website/.vitepress/cache/`.
- Verificación de snippets: `grep -rnE 'logger := go_logs\.New\(|WithHook\(|NewSlackHook\("' website --include='*.md'` vacío; `grep -rn 'drossan/go_logs"' website --include='*.md'` vacío (todos con `/v3`).
- `pnpm --dir website build` termina sin errores ni enlaces muertos (VitePress falla el build con dead links por defecto; no desactivar `ignoreDeadLinks`).
- No tocar `docs/wiki/` en esta tarea.

## Fuera de alcance

- Retirar `docs/wiki/` o `sync-wiki.yml`: plan aparte; el wiki sigue.
- Reescribir o ampliar el contenido: solo copia + correcciones de snippets e import path.
- Corregir los snippets en `README.md` y `docs/wiki/`: tarea 10.
- Dominio propio; el sitio vive en `drossan.github.io/go_logs`.
- Workflow de despliegue: tarea 09.
- Traducir páginas que hoy no existan en uno de los dos idiomas: si falta, se enlaza a la versión disponible.

## Scenarios (Gherkin)

```gherkin
Feature: Sitio de documentación VitePress bilingüe

  # Tarea sin código Go testeable. Verificación: build de VitePress, recuentos de
  # páginas y greps de snippets, registrados en el session log.

  Scenario: El sitio compila
    Given el directorio website/ con dependencias instaladas
    When se ejecuta pnpm build
    Then termina sin errores
    And no reporta enlaces muertos

  Scenario: Un enlace interno roto hace fallar el build
    Given una página con un enlace a una página inexistente
    When se ejecuta pnpm build
    Then el build falla señalando el enlace muerto

  Scenario Outline: Todas las páginas del wiki tienen su equivalente en el sitio
    Given la página <wiki> en docs/wiki/
    When se busca su copia en website/
    Then existe <sitio>

    Examples:
      | wiki                        | sitio                          |
      | Getting-Started.md          | getting-started.md             |
      | Getting-Started-es.md       | es/getting-started.md          |
      | Home.md                     | index.md                       |
      | Home-es.md                  | es/index.md                    |
      | Optional-Modules.md         | optional-modules.md            |
      | Migration-v2-to-v3-es.md    | es/migration-v2-to-v3.md       |

  Scenario: Recuento de páginas por idioma
    Given el sitio construido
    When se cuentan los .md de website/ (sin es/) y de website/es/ (Home.md/Home-es.md cuentan como index.md/es/index.md, no aparte)
    Then hay 14 ficheros en inglés (incluido index.md) y 14 en español (incluido es/index.md)

  Scenario: Ningún snippet usa la API rota
    Given todas las páginas de website/
    When se buscan los patrones "logger := go_logs.New(", "WithHook(" y "NewSlackHook(\""
    Then no hay ninguna coincidencia

  Scenario: Todos los imports apuntan a v3
    Given todas las páginas de website/
    When se buscan imports de "github.com/drossan/go_logs" sin sufijo /v3
    Then no hay ninguna coincidencia fuera de la página de migración, donde v2 se cita a propósito

  Scenario: Navegación por idioma
    Given el sitio en desarrollo
    When se abre la raíz y se cambia el idioma a Español
    Then la URL pasa a /es/ y el sidebar muestra los títulos en español

  Scenario: La búsqueda local encuentra contenido
    Given el sitio construido y servido con preview
    When se busca "RotatingFileWriter"
    Then aparece la página de rotación de ficheros
```

## Provides

- `website/` construible con `pnpm --dir website build`, con `base: '/go_logs/'`. La tarea 09 despliega su `dist`; la tarea 10 enlaza el sitio desde el README.

## Definition of Done

- [x] Verificación registrada en el session log: salida de `pnpm build`, recuentos y greps
- [x] Todos los tests en verde: `go test -race ./...` sigue en verde (esta tarea no toca Go)
- [x] Spec cumplida; lo declarado en `Provides` queda realmente disponible para las tareas dependientes
- [x] Lint / format / typecheck OK: `pnpm build` sin warnings de TypeScript en `config.ts`
- [x] Gate de `fact-checker` superado — afirmaciones de la sesión verificadas (INCORRECTO bloquea; NO VERIFICABLE = aviso a reconocer), antes de commit/resumen  · no-negociable
- [x] Documentación actualizada — tres capas:
  - [x] **Comentarios en `config.ts`** (locales, base, por qué no se ignoran dead links)
  - [x] **Doc técnica (contexto)** — `website/README.md`; CLAUDE.md menciona `website/` en la estructura
  - [x] **Histórico de la tarea** — session log en `.claude/context/go_logs/go_logs-instalable-bugs-prod-08.md`
- [x] Commit en la rama del plan: `go_logs-instalable-bugs-prod-08: docs: <descripción>`
