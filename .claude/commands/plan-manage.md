---
name: plan-manage
version: 1.0.0
author: Reverence Hotels Team
description: Meta-command para orquestar la creación, aprobación y ejecución de planes técnicos en el proyecto Reverence Hotels API
usage: "plan-manage [create|approve|execute|list|cancel] [plan-id] [--force]"
type: meta
writes_code: false
creates_plan: false
requires_approval: false
dependencies: []
---

# Comando: Plan Manager

## Objetivo

Meta-command central para orquestar el ciclo de vida completo de los planes técnicos en el proyecto Reverence Hotels API. Este comando es el **único autorizado** para:

1. **Crear** planes a partir de solicitudes de features
2. **Aprobar** planes generados por otros commands
3. **Ejecutar** planes aprobados coordinando agentes
4. **Listar** planes pendientes/en progreso
5. **Cancelar** planes en ejecución

**No genera planes directamente**, sino que orquesta los commands planning y coordina su ejecución.

## Contexto Requerido del Usuario

- [ ] Acción deseada (create, approve, execute, list, cancel)
- [ ] Descripción de la feature (si es create)
- [ ] ID del plan (si es approve, execute o cancel)
- [ ] Confirmación de ejecución (si es execute)
- [ ] Razón de cancelación (si es cancel)

## Análisis Inicial (Obligatorio)

### Validaciones Pre-ejecución

```json
{
  "validation_passed": true,
  "action": "create|approve|execute|list|cancel",
  "risks": [
    "Modifica código en producción",
    "Requiere migración de base de datos",
    "Impacta múltiples módulos"
  ],
  "required_approvals": ["tech-lead"],
  "estimated_complexity": "medium|high",
  "blocking_issues": [],
  "existing_plans": []
}
```

### Checklist por Acción

**Para CREATE:**
- [ ] Verificar que no existe un plan similar pendiente
- [ ] Validar que la solicitud tiene suficiente contexto
- [ ] Identificar el planning command apropiado
- [ ] Confirmar disponibilidad de agentes y skills

**Para APPROVE:**
- [ ] Verificar que el plan existe en `.claude/plans/`
- [ ] Validar que el plan tiene todas las secciones requeridas
- [ ] Confirmar que no hay conflictos con otros planes activos
- [ ] Verificar aprobaciones de tech-lead si es requerida

**Para EXECUTE:**
- [ ] Confirmar que el plan está aprobado
- [ ] Verificar que no hay otro plan en ejecución
- [ ] Validar disponibilidad de agentes
- [ ] Crear snapshot del estado actual

**Para CANCEL:**
- [ ] Verificar que el plan está en ejecución o pendiente
- [ ] Identificar agentes activos a detener
- [ ] Preparar procedimiento de rollback

## Selección de Agentes y Skills

### Acción CREATE

```yaml
fase_1_seleccion_command:
  responsible: go-orchestrator
  accountable: planning-agent
  consulted: [ technical-writer ]
  informed: [ go-reviewer ]
  
fase_2_validacion_plan:
  responsible: go-reviewer
  accountable: software-architect-tdd-ddd
  consulted: [ go-code-reviewer ]
  informed: [ go-orchestrator ]
```

### Acción EXECUTE

```yaml
fase_1_ejecucion_plan:
  responsible: go-orchestrator
  accountable: planning-agent
  consulted: [ technical-writer, go-code-reviewer ]
  informed: [ go-reviewer, go-test-runner ]
  
fase_2_validacion_resultados:
  responsible: go-test-runner
  accountable: go-reviewer
  consulted: [ debug-master ]
  informed: [ go-orchestrator ]
```

### Acción CANCEL

```yaml
fase_1_rollback:
  responsible: go-debugger
  accountable: go-orchestrator
  consulted: [ technical-writer ]
  informed: [ planning-agent ]
```

## Flujo de Trabajo Orquestado

### 1. CREATE: Selección de Planning Command (go-orchestrator | Validado por planning-agent)

**Objetivo**: Analizar la solicitud y seleccionar el command planning apropiado

**Tareas**:

- Analizar el tipo de solicitud (feature, refactor, migration, bug-fix)
- Seleccionar el planning command adecuado:
  - `feature-planner` → Nuevas funcionalidades
  - `refactor-analyzer` → Refactorización de código
  - `migration-coordinator` → Migraciones de datos/BD
  - `bug-fix-planner` → Corrección de bugs
- Invocar el planning command seleccionado
- Recibir el plan generado

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: N/A (solo orquestación)
- **Commands Invocados**: Variable según tipo de solicitud
- **Validador**: planning-agent

**Criterios de Salida**:

- [ ] Planning command identificado correctamente
- [ ] Plan generado en `.claude/plans/`
- [ ] Plan contiene todas las secciones requeridas
- [ ] Usuario notificado del plan creado

---

### 2. APPROVE: Validación de Plan (go-reviewer | Validado por software-architect-tdd-ddd)

**Objetivo**: Revisar el plan técnico y aprobarlo para ejecución

**Tareas**:

- Leer el plan desde `.claude/plans/{plan-id}.md`
- Validar estructura completa del plan
- Verificar que el plan sigue Clean Architecture cuando aplica
- Confirmar que todos los agentes y skills existen
- Validar que no hay conflictos con planes activos
- Solicitar aprobación de tech-lead si es requerida
- Marcar el plan como aprobado

**Asignación**:

- **Agente**: go-reviewer
- **Skills**: `go-code-reviewer`
- **Validador**: software-architect-tdd-ddd

**Criterios de Salida**:

- [ ] Plan validado contra checklist obligatorio (7/7)
- [ ] Agentes y skills verificados como disponibles
- [ ] Sin conflictos con otros planes
- [ ] Plan marcado como aprobado en metadata
- [ ] Tech-lead notificado (si aplica)

---

### 3. EXECUTE: Coordinación de Ejecución (go-orchestrator | Validado por planning-agent)

**Objetivo**: Ejecutar el plan aprobado coordinando los agentes especificados

**Tareas**:

- Leer el plan aprobado
- Crear snapshot del estado actual (`.claude/snapshots/{plan-id}-pre.json`)
- Inicializar log de ejecución (`.claude/logs/execution-{plan-id}.log`)
- Ejecutar cada fase del plan secuencialmente:
  - Invocar el agente responsible de la fase
  - Inyectar los skills especificados
  - Monitorear progreso
  - Validar criterios de salida
- Al completar, notificar al agente accountable
- Marcar plan como completado

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: N/A (solo orquestación)
- **Validador**: planning-agent

**Criterios de Salida**:

- [ ] Snapshot creado antes de ejecución
- [ ] Todas las fases completadas exitosamente
- [ ] Logs de ejecución completos
- [ ] Plan marcado como completado
- [ ] Reporte de resultados generado

---

### 4. VALIDATE: Verificación de Resultados (go-test-runner | Validado por go-reviewer)

**Objetivo**: Validar que la ejecución del plan cumplió todos los criterios

**Tareas**:

- Ejecutar suite de tests del proyecto
- Verificar coverage mínimo (80%)
- Validar que no hay regresiones
- Correr tests de integración afectados
- Revisar logs de ejecución
- Validar artefactos generados
- Generar reporte de validación

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `debug-master`
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Todos los tests pasan
- [ ] Coverage >= 80%
- [ ] Sin regresiones detectadas
- [ ] Artefactos validados
- [ ] Reporte de validación generado

---

### 5. LIST: Estado de Planes (go-orchestrator)

**Objetivo**: Listar todos los planes y su estado actual

**Tareas**:

- Escanear directorio `.claude/plans/`
- Leer metadata de cada plan
- Clasificar por estado:
  - `pending` - Creado, no aprobado
  - `approved` - Aprobado, no ejecutado
  - `executing` - En ejecución
  - `completed` - Ejecutado exitosamente
  - `failed` - Falló durante ejecución
  - `cancelled` - Cancelado por usuario
- Presentar resumen al usuario

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: N/A

**Criterios de Salida**:

- [ ] Lista de planes generada
- [ ] Estado de cada plan identificado
- [ ] Resumen presentado al usuario

---

### 6. CANCEL: Rollback de Plan en Ejecución (go-debugger | Validado por go-orchestrator)

**Objetivo**: Detener la ejecución de un plan y revertir cambios

**Tareas**:

- Identificar agentes activos del plan
- Enviar señal de cancelación a cada agente
- Esperar confirmación de detención
- Restaurar estado desde snapshot:
  - Revertir cambios en código (git checkout)
  - Eliminar archivos temporales
  - Limpiar artefactos parciales
- Registrar cancelación en log
- Notificar commands dependientes
- Marcar plan como cancelado

**Asignación**:

- **Agente**: go-debugger
- **Skills**: `debug-master`, `technical-writer`
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Todos los agentes detenidos
- [ ] Estado restaurado desde snapshot
- [ ] Artefactos parciales eliminados
- [ ] Cancelación registrada
- [ ] Plan marcado como cancelado

## Uso de otros Commands y MCPs

### Commands Invocados por Acción

**CREATE:**
```yaml
comando_seleccionado:
  feature-planner:
    trigger: "Nueva funcionalidad|Crear feature|Implementar"
    
  refactor-analyzer:
    trigger: "Refactorizar|Mejorar código|Reestructurar"
    
  migration-coordinator:
    trigger: "Migración|Cambiar BD|Actualizar esquema"
    
  bug-fix-planner:
    trigger: "Corregir bug|Fix issue|Resolver error"
```

**EXECUTE:**
```yaml
commands_coordinados:
  - feature-planner (si el plan es de feature)
  - refactor-analyzer (si el plan es refactor)
  - migration-coordinator (si el plan es migración)
  - qa-automation (post-ejecución)
```

### MCPs Utilizados

```yaml
mcps_utilizados:
  - name: git-status
    purpose: Verificar estado de repositorio pre/post ejecución
    
  - name: file-system
    purpose: Gestionar snapshots y artefactos
```

### Contexto Compartido

```yaml
contexto_compartido:
  location: .claude/context/plan-state.json
  format: JSON
  content:
    current_plan_id: string
    execution_status: string
    active_agents: string[]
    snapshot_path: string
  consumers:
    - go-orchestrator
    - planning-agent
    - go-reviewer
```

## Output y Artefactos

| Artefacto              | Ubicación                                           | Formato    | Validador          | Obligatorio |
|------------------------|-----------------------------------------------------|------------|--------------------|-------------|
| Plan técnico           | `.claude/plans/{timestamp}-{plan-type}.md`          | Markdown   | `go-reviewer`      | Sí (CREATE) |
| Snapshot pre-ejecución | `.claude/snapshots/{plan-id}-pre.json`              | JSON       | -                  | Sí (EXECUTE)|
| Log de ejecución       | `.claude/logs/execution-{plan-id}.log`              | Plain text | -                  | Sí (EXECUTE)|
| Reporte de validación  | `.claude/reports/validation-{plan-id}.md`           | Markdown   | `go-test-runner`   | Sí (VALIDATE)|
| Estado de planes       | `.claude/context/plan-state.json`                   | JSON       | -                  | Sí          |
| Log de cancelación     | `.claude/logs/cancelled-{plan-id}-{timestamp}.log`  | Plain text | -                  | Sí (CANCEL) |

## Rollback y Cancelación

### Por Acción

**CREATE Rollback:**
1. Eliminar plan generado en `.claude/plans/`
2. Limpiar archivos temporales
3. Notificar al planning command invocado

**EXECUTE Rollback (Procedimiento Completo):**

1. **Detectar agentes activos**:
   ```bash
   ps aux | grep "agent:*" | grep "{plan-id}"
   ```

2. **Enviar señales de terminación**:
   - SIGTERM a cada proceso
   - Esperar hasta 10 segundos
   - Si no responde, SIGKILL

3. **Restaurar desde snapshot**:
   ```bash
   git restore --staged .
   git restore .
   git clean -fd
   ```

4. **Eliminar artefactos parciales**:
   ```bash
   rm -rf .claude/temp/{plan-id}/*
   rm -f .claude/plans/{plan-id}-partial.md
   ```

5. **Revertir contexto compartido**:
   ```bash
   cp .claude/snapshots/{plan-id}-pre.json .claude/context/plan-state.json
   ```

6. **Registrar cancelación**:
   ```
   echo "{timestamp} CANCELLED {plan-id} {reason}" >> .claude/logs/cancelled.log
   ```

7. **Notificar dependientes**:
   - qa-automation: Cancelar tests pendientes
   - deployment-manager: Abortar despliegue

**VALIDATE Rollback:**
1. Si validación falla, marcar plan como `failed`
2. Generar reporte de fallo
3. Ejecutar rollback de EXECUTE

## Reglas Críticas

- **Único meta-command**: No puede existir otro meta-command en el proyecto
- **Prohibición de recurrencia**: Detectar y abortar si se invoca a sí mismo
- **Centralización de ejecución**: Solo este comando puede ejecutar planes
- **Snapshot obligatorio**: Nunca ejecutar sin crear snapshot previo
- **Un plan a la vez**: Bloquear si ya existe un plan en ejecución
- **Validación de agentes**: Verificar disponibilidad antes de ejecutar
- **Rollback garantizado**: Siempre poder revertir al estado anterior
- **Logs inmutables**: Nunca borrar logs de ejecución o cancelación
- **Aprobación explícita**: Nunca ejecutar un plan sin aprobación previa
- **Estado global consistente**: Mantener `.claude/context/plan-state.json` sincronizado

---

## Acción del Usuario

Especifica la acción que deseas realizar:

### CREATE
> "Crear un plan para [descripción de la feature/refactor/migración]"
> 
> Incluye:
> - Qué deseas implementar o cambiar
> - Alcance (módulos afectados)
> - Prioridad (crítica, alta, media, baja)
> - Restricciones técnicas (si aplica)

**Ejemplo**:
> "Crear plan para implementar autenticación OAuth2 en la API. Alcance: módulo Auth y middleware. Prioridad: alta. Restricción: debe usar Echo framework."

### APPROVE
> "Aprobar el plan [plan-id]"
> 
> El comando validará el plan antes de aprobarlo.

### EXECUTE
> "Ejecutar el plan [plan-id]"
> 
> El comando creará un snapshot y ejecutará el plan aprobado.

### LIST
> "Listar todos los planes"
> 
> Muestra estado de todos los planes: pendientes, aprobados, en ejecución, completados.

### CANCEL
> "Cancelar el plan [plan-id] porque [razón]"
> 
> El comando detendrá la ejecución y revertirá cambios.

**Ejemplo de cancelación**:
> "Cancelar el plan 20250120-143022-feature-planner porque cambian los requisitos de negocio."