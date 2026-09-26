# Cómo arrancar una sesión nueva para una tarea — `go_logs`

> Este fichero es **específico del package `go_logs`**, pero su estructura es
> genérica y replicable: cada workspace tiene —o tendrá— su propio
> `HOW-TO-START-A-TASK.md` en `.claude/specs/<package>/`, con el mismo esqueleto y
> solo las **reglas específicas del package** cambiadas (arquitectura, niveles de
> test, gates, comando de filtro). Está referenciado desde el `CLAUDE.md` raíz
> para que se tenga en cuenta en **toda** tarea.

> **¿De dónde salen el plan y las tareas?** Del flujo de la skill **`/plan-task`**
> (plan mode → plan en `.claude/plans/pending/<package>/` → `grilling` →
> descomposición en tareas con escenarios **Gherkin**). Este HOW-TO cubre la
> **ejecución** de cada tarea ya creada. Ver `docs/guides/task-lifecycle.md`.

> ## 📦 Stack de este package
>
> <!-- Refleja `stack:` top-level de `.claude/task-pipeline.yml` (go_logs no tiene entrada propia en
>      `stack.packages`: el repo es Go de punta a punta salvo `website/`, que es documentación, no
>      código del package). El YAML es la FUENTE DE VERDAD; este bloque solo lo refleja para el
>      humano — si divergen, manda el YAML. -->
> - **Lenguaje**: Go (`go 1.27.0` en `go.mod` desde la tarea 01; con `GOTOOLCHAIN=auto` un toolchain local 1.26 descarga 1.27.0)
> - **Gestor de paquetes**: `go mod` (sin gestor de paquetes de terceros; `website/` usa `pnpm`, fuera del
>   alcance de este HOW-TO — ver la tarea 08 del plan `instalable-bugs-prod`)
> - **Runner de tests**: `go test ./...` (con `-race` desde que el módulo es único; `slack/` es submódulo
>   con su propio `go test` — ver tarea 06)
> - **Mutation**: `none` (`stack.mutation-tool: none` en `.claude/task-pipeline.yml`). El gate de mutation
>   testing (Paso 7 del pipeline, skill `/mutation`) queda **desactivado de facto** para este package: cada
>   regresión se defiende con su propio assert concreto (contadores de writers espía, comparación de campos
>   por hijo, tiempos de `Sync()`), no con survivors de mutantes.
>
> Si en el futuro se activa mutation testing para Go, el escape genérico es `mutation-command` (no hay
> tool nombrada shipeada para Go en el plugin; herramientas de referencia: `go-mutesting`, `gremlins`).

> ## 🚦 GATE TDD — IMPERATIVO, NO NEGOCIABLE
>
> Este repo usa `mode: full` con `features.tdd: true` y las tres capas de documentación activas
> (`tsdoc`, `technical-docs`, `context-log`). El único flag que se aparta del preset es
> `stack.mutation-tool: none` (ver arriba): el **gate de mutation testing NO aplica** a `go_logs`; el resto
> del gate (TDD, lint/vet/fmt, `fact-checker`) es exactamente el del preset `full`.
>
> **Idea**: si el test falla **primero** (Red), tenemos la red de seguridad que
> garantiza que la implementación posterior (Green) hace exactamente lo
> especificado y que el refactor no rompe nada. Empezar por el código en vez de
> por el test invalida esa red y está prohibido.
>
> **Antes de tocar una sola línea de código de implementación:**
>
> 1. Confirmar que la tarea está en `.claude/tasks/active/go_logs/<task-id>.md`
>    con `status: active` y que **todas** sus `depends_on` están en `done`. Si una
>    dependencia no está cerrada → **PARAR y avisar al usuario**.
> 2. Confirmar que estamos en la rama del plan `plan/go_logs/<name-plan>` (nunca
>    en `develop` ni en `main`). Si hay otra tarea del mismo plan en
>    `active` → **PARAR**: solo una tarea `active` por plan (comparten rama).
> 3. **Escribir el test que falla ANTES de la implementación** (Red). Los tests
>    salen **1:1 de los escenarios Gherkin** de la sección `## Scenarios (Gherkin)`
>    del task file (cada `Then` es un assert); la(s) spec(s) aplicables en
>    `.claude/specs/` (ver tabla abajo) marcan el contrato y los anti-patrones.
>
>    - Paquete raíz (`logger_impl.go`, `config.go`, `caller.go`, `rotating_writer*.go`, `level.go`,
>      `context.go`, …) → test unitario en el propio paquete (`*_test.go`), sin mocks de infraestructura
>      real: usa `bytes.Buffer`/`bufio.Writer` como output y writers espía (`spyWriter`) para contar
>      `Write`/`Flush`/`Sync`.
>    - `async/`, `hooks/`, `http/`, `otel/`, `signal/` (subpaquetes tras la tarea 01) → test unitario en su
>      propio paquete; los que tocan concurrencia (`async`, `signal`) exigen `go test -race` en verde, no
>      solo `go test`.
>    - `slack/` (submódulo tras la tarea 06) → test unitario dentro de `slack/`, ejecutado con
>      `(cd slack && go test -race ./...)` porque tiene su propio `go.mod`.
>    - Workflows de GitHub Actions (tareas 07, 09) y contenido de `website/` (tarea 08) → sin código Go
>      testeable; la verificación es la que fija el Spec de cada tarea (YAML válido, `pnpm build`, greps),
>      registrada en el session log en vez de en un test automatizado.
>
>    Sin test rojo previo NO se escribe código de producción (salvo en las tareas sin código Go, donde
>    "rojo" es el comando de verificación fallando antes del cambio).
> 4. Implementar lo mínimo para pasar a verde (Green). Refactor con los tests en
>    verde (Refactor).
>
>    Reglas de arquitectura de este repo:
>    - **Dependency Inversion**: el código del paquete raíz depende de interfaces (`Formatter`, `Hook`,
>      `Flusher`, `domain.Notifier`), nunca de implementaciones concretas de terceros. Es justo lo que la
>      tarea 06 restaura al sacar Slack del core.
>    - **Un solo módulo** desde la tarea 01: no reintroducir `go.mod`/`replace` en subdirectorios salvo
>      `slack/` (única excepción deliberada, con `require` real + `replace` de desarrollo).
>    - **Nunca** `log.Fatal*` ni `os.Exit` en código de librería fuera de `Fatal()`/`FatalLog` (que lo hacen
>      a propósito y de forma documentada). Cualquier fallo de configuración se resuelve con un default y un
>      aviso por `stderr` (patrón `envBool` de la tarea 02).
>    - `gofmt -l <ficheros tocados>` sin salida antes de dar una tarea por Green.
> 5. Al cerrar: **todos** los checkboxes de la Definition of Done en verde, incluida
>    la documentación en sus **tres capas obligatorias** —(1) **godoc** en todo
>    símbolo público, (2) **doc técnica/contexto** (README/CLAUDE.md/MIGRATION.md/specs), (3) **histórico
>    de la tarea** (session log)—, con `go vet ./...` y `go test -race ./...` **repo-wide** en verde (sin
>    regresiones; incluye `slack/` desde la tarea 06), **cero secretos en logs**. El **gate de mutation
>    testing NO aplica** a este package (`stack.mutation-tool: none`): no lo busques ni lo bloquees por su
>    ausencia.
> 6. **Gate de `fact-checker`** (no-negociable, sin flag que lo desactive — como
>    `grilling`/aprobación): **tras** verificar tests/lint y **antes** de commit y del
>    resumen final, corre `fact-checker` sobre las afirmaciones factuales de la sesión
>    (incluida «los tests pasan» y «`go vet`/`gofmt` están limpios»). Una afirmación **INCORRECTO bloquea**
>    el cierre hasta corregirla; **NO VERIFICABLE** es un aviso que hay que **reconocer
>    explícitamente**, pero no bloquea; **VERIFICADO** pasa.
>
> Marcar un step de tests como hecho sin haber corrido la suite, o saltarse el
> Red→Green→Refactor, es engañoso y está prohibido. Si la sesión no puede completar
> la tarea, queda en `status: blocked` (o se documenta el progreso parcial en el
> histórico) esperando al usuario — nunca se marca `done` "con nota".

Pega el prompt de abajo (o un equivalente) como **primer mensaje** de cada nueva
sesión de Claude para que arranque con el contexto y el workflow correctos sin
gastar turnos en explicarlo.

---

## Prompt de arranque (copia y pega)

```
Voy a ejecutar la tarea <task-id> del plan <name-plan> (go_logs).

Lee en este orden y repórtame en 3-4 líneas el plan que vas a seguir:

1. `.claude/specs/go_logs/HOW-TO-START-A-TASK.md` (este fichero).
2. `docs/guides/task-lifecycle.md` (flujo canónico: estados, ramas, cierre, DoD).
3. `.claude/plans/active/go_logs/<name-plan>.md` (contexto, objetivos, orden y
   dependencias de las tareas).
4. Las specs aplicables en `.claude/specs/general/` (según el artefacto que toque la tarea — ver tabla
   abajo; este package no tiene specs propias por artefacto más allá de este HOW-TO).
5. `.claude/tasks/active/go_logs/<task-id>.md` (la tarea de esta sesión) y
   `.claude/context/go_logs/<task-id>.md` (histórico de sesión, si existe).

Tras leerlos, ejecuta la tarea siguiendo el ciclo TDD paso a paso.

REGLAS ESTRICTAS DE LA SESIÓN:
- **GATE TDD (imperativo, no negociable, antes de tocar código)**:
  1. `status: active` y todas las `depends_on` en `done`. Si no → PARAR y avisar.
  2. En la rama `plan/go_logs/<name-plan>`. Solo una tarea `active` por plan.
  3. Test rojo ANTES de la implementación (Red → Green → Refactor).
- Un solo módulo Go (`github.com/drossan/go_logs/v3`) salvo `slack/`, que es el único submódulo legítimo.
- Cero `log.Fatal*`/`os.Exit` fuera de `Fatal()`/`FatalLog` documentados.
- `gofmt -l` limpio en los ficheros tocados antes de dar la tarea por Green.
- Cada artefacto sigue las reglas de arquitectura de este HOW-TO. Si introduces o cambias un
  patrón, actualiza este fichero o el `CLAUDE.md` en el mismo cambio.
- Commits en la rama del plan con el formato `<task-id>: <conventional commit>`.
- Documentación actualizada ANTES de declarar la tarea hecha. **Tres capas
  obligatorias**: godoc en todo símbolo público (al crearlo), doc técnica/contexto
  y histórico de la tarea.
- `go vet ./...` en verde ANTES de cerrar (y en `slack/` desde la tarea 06).
- Cero secretos en logs.
- El gate de mutation testing **no aplica** a este package (`stack.mutation-tool: none`): no lo busques.

Al terminar:
- Verifica `go test -race ./...` repo-wide en verde (y en `slack/` desde la tarea 06), sin regresiones.
- **Gate `fact-checker` (no-negociable)**: antes de commit y del resumen, corre
  `fact-checker` sobre las afirmaciones de la sesión. `INCORRECTO` bloquea hasta
  corregir; `NO VERIFICABLE` = aviso a reconocer; `VERIFICADO` pasa.
- Cierra el histórico en `.claude/context/go_logs/<task-id>.md` (resumen,
  decisiones + porqué, tests corridos + resultado, docs actualizadas + motivo,
  ficheros/commits, tiempo real, follow-ups).
- Mueve el task a `.claude/tasks/completed/go_logs/`, `status: done`, rellena
  `actual:` y bump `updated:`.
- Marca la casilla `[ ]` → `[x]` de la tarea en el plan y bump su `updated:`.
- Reporta cambios y siguiente tarea recomendada según el orden del plan.
```

Sustituye `<task-id>` (= `<plan-id>-<nn>`, con `<plan-id> = go_logs-<name-plan>` y
`<nn>` correlativo del plan desde `01`) por el id de la tarea y `<name-plan>` por el plan activo
(`instalable-bugs-prod`). Lista completa de tareas bajo `.claude/plans/active/go_logs/instalable-bugs-prod.md`.

---

## Specs aplicables (paso 3 del gate)

Antes de escribir el test, abre la(s) spec(s) del artefacto que toques. Son el
contrato del que se derivan los tests:

| Si la tarea crea / toca… | Lee la spec |
|---|---|
| El propio task file (`Spec` y `## Scenarios (Gherkin)`) | Es el contrato primario: cada tarea trae su propio contrato detallado, no hay specs por artefacto separadas para `go_logs` todavía |
| Cualquier código (transversal) | `.claude/specs/general/coding-standards.md` (si existe; no es drift si el repo no lo materializó — es user-owned) |
| Cualquier test (transversal) | `.claude/specs/general/testing.md` (si existe) |
| Modelado de errores | `.claude/specs/general/error-handling.md` (si existe) |
| Seguridad (rutas, validación, datos) | `.claude/specs/general/security.md` (si existe) |
| Rama o commit | `.claude/specs/general/git-workflow.md` (si existe) |
| Arquitectura general del proyecto | `CLAUDE.md` (raíz) |
| Análisis y hallazgos que motivan el plan actual | `.claude/reports/analisis-completo-20260926.md` |

---

## Qué hace la sesión por sí sola

Al recibir el prompt, la sesión debería:

1. **Leer los ficheros** en orden y construir el contexto (HOW-TO →
   task-lifecycle → plan → specs → task + histórico).
2. **Crear un plan de pasos** con el ciclo TDD: test rojo → implementación mínima
   → refactor → `gofmt`/`go vet` → docs → cierre del histórico → mover task +
   tick en el plan.
3. **Empezar por el gate**: comprobar `depends_on`, rama del plan, y escribir el
   primer test que falla.
4. **Avanzar Red → Green → Refactor** por cada escenario Gherkin de la tarea.
5. **No declarar la tarea `done`** hasta que **todos** los checkboxes de la DoD
   estén en verde (recordando que el ítem de mutation testing no aplica a este package).

## Si la sesión se queda sin contexto a mitad

- **Guardar progreso parcial** en el histórico `.claude/context/go_logs/<task-id>.md`
  (append-only): qué se ha hecho, qué queda, decisiones, próximos pasos.
- Tarea bloqueada por algo externo: `status: blocked` + motivo (ver `task-lifecycle.md`).
  Ninguna otra tarea del plan pasa a `active` hasta resolverlo.
- Abrir nueva sesión con el prompt + nota de que se retoma; la nueva sesión lee el
  histórico para no rehacer trabajo.

## Reglas anti-context-bloat

- Evitar leer ficheros completos; usar `grep -n` para localizar líneas concretas.
- Delegar exploración amplia a la skill `Explore` (corre en sub-proceso y solo
  devuelve la conclusión).
- Limitar output verbose: pasar `2>&1 | tail -N` a comandos de test/build.
- No re-leer un fichero recién editado solo para verificar: si el edit no falló,
  el cambio se aplicó.

## Replicar este HOW-TO en otro package

1. Copia este fichero a `.claude/specs/<otro-package>/HOW-TO-START-A-TASK.md`.
2. Cambia el título, el comando de filtro, la tabla de
   specs y las reglas específicas del gate (stack real, arquitectura del workspace).
3. Referencia el nuevo fichero desde el `CLAUDE.md` del workspace.
