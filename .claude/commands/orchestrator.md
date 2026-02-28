---
name: orchestrator
version: 1.0.0
author: Reverence Hotels Development Team
description: Orchestration command for coordinating development tasks in Reverence Hotels API project. Analyzes requests and selects optimal agents and skills for execution.
usage: "orchestrator [task-description] [--priority=normal] [--context=additional-info]"
type: meta
writes_code: false
creates_plan: false
requires_approval: false
dependencies: [planning-agent, go-orchestrator, software-architect-tdd-ddd]
---

# Comando: Orchestrator

## Objetivo

Actuar como punto de entrada centralizado para orquestar tareas de desarrollo en el proyecto Reverence Hotels API. Este comando:

- Analiza la solicitud del usuario
- Identifica el tipo de tarea (nueva feature, refactor, bug fix, migración, documentación, testing)
- Selecciona los agentes y skills más adecuados usando el framework RACI
- Delega la ejecución al agente apropiado
- Coordinar múltiples agentes cuando la tarea requiere fases secuenciales

**No ejecuta trabajo técnico directamente**, sino que coordina y orquesta a los agentes especializados.

## Contexto Requerido del Usuario

- [ ] Descripción clara de la tarea o feature solicitada
- [ ] Tipo de tarea (desarrollo, refactor, bug fix, testing, documentación, migración)
- [ ] Alcance esperado (módulo/s afectados)
- [ ] Prioridad (crítica, alta, media, baja)
- [ ] Restricciones técnicas conocidas (opcional)

## Análisis Inicial (Obligatorio)

Antes de delegar a cualquier agente, el orchestrator debe evaluar:

### Pre-ejecución: Checklist Obligatorio

El command debe verificar:

- [ ] ¿La tarea está bien definida? → Si no, solicitar aclaraciones
- [ ] ¿Existe un agente especializado disponible? → Seleccionar el más adecuado
- [ ] ¿La tarea requiere múltiples fases? → Coordinar secuencia de agentes
- [ ] ¿Existen dependencias o bloqueadores? → Reportar antes de continuar
- [ ] ¿La tarea está dentro del alcance del proyecto? → Validar contra arquitectura existente

**Output esperado**: JSON de análisis antes de continuar.

```json
{
  "task_type": "feature-development | bug-fix | refactoring | migration | testing | documentation",
  "complexity": "low | medium | high",
  "primary_agent": "go-orchestrator | go-debugger | go-reviewer | migration-specialist | planning-agent | software-architect-tdd-ddd | api-integration-expert | go-test-runner | technical-writer",
  "required_skills": ["skill-1", "skill-2"],
  "estimated_phases": 1,
  "blocking_issues": []
}
```

## Selección de Agentes y Skills (Framework RACI)

El orchestrator selecciona dinámicamente los agentes según el tipo de tarea:

### Matriz de Selección de Agentes

| Tipo de Tarea | Agente Primario | Skills Requeridos | Agente Validador |
|---------------|-----------------|-------------------|------------------|
| **Nueva Feature (Clean Architecture)** | `software-architect-tdd-ddd` | `go-clean-architecture`, `echo-routes` | `go-reviewer` |
| **Nueva Feature (MVC Legacy)** | `go-orchestrator` | `echo-routes`, `gorm-models` | `go-reviewer` |
| **Bug Fix / Debugging** | `go-debugger` | `debug-master` | `go-reviewer` |
| **Refactorización** | `migration-specialist` | `go-clean-architecture` | `go-reviewer` |
| **Integración API Externa** | `api-integration-expert` | `jwt-auth`, `multi-database` | `go-reviewer` |
| **Testing / QA** | `go-test-runner` | `go-code-reviewer` | `go-orchestrator` |
| **Planificación Compleja** | `planning-agent` | - | `software-architect-tdd-ddd` |
| **Documentación** | `technical-writer` | `technical-writer` | `go-orchestrator` |
| **Migración MVC → Clean Arch** | `migration-specialist` | `go-clean-architecture` | `software-architect-tdd-ddd` |

### Criterios de Selección

**1. Análisis de Tipo de Tarea**:

```yaml
task_analysis:
  feature_development:
    uses_clean_architecture:
      agent: software-architect-tdd-ddd
      skills: [go-clean-architecture, echo-routes, gorm-models]
    uses_mvc_legacy:
      agent: go-orchestrator
      skills: [echo-routes, gorm-models, jwt-auth]
      
  bug_fix:
    agent: go-debugger
    skills: [debug-master]
    validator: go-reviewer
    
  refactoring:
    agent: migration-specialist
    skills: [go-clean-architecture]
    validator: software-architect-tdd-ddd
    
  integration:
    agent: api-integration-expert
    skills: [jwt-auth, multi-database, sii-invoicing]
    validator: go-reviewer
```

**2. Asignación RACI por Fase** (para tareas complejas):

```yaml
fase_1_analisis:
  responsible: planning-agent
  accountable: software-architect-tdd-ddd
  consulted: []
  informed: [go-orchestrator]

fase_2_implementacion:
  responsible: go-orchestrator
  accountable: go-reviewer
  consulted: [go-clean-architecture, echo-routes]
  informed: [go-test-runner]

fase_3_validacion:
  responsible: go-test-runner
  accountable: go-orchestrator
  consulted: [go-code-reviewer]
  informed: []
```

## Flujo de Trabajo Orquestado

### 1. Análisis de la Solicitud (Orchestrator)

**Objetivo**: Clasificar la tarea y seleccionar el agente óptimo

**Tareas**:

- Leer y entender la solicitud del usuario
- Identificar palabras clave que indiquen el tipo de tarea
- Verificar contexto del proyecto (arquitectura híbrida MVC/Clean Architecture)
- Seleccionar agente y skills según matriz de selección
- Validar que el agente seleccionado esté disponible

**Criterios de Salida**:

- [ ] Tipo de tarea identificado
- [ ] Agente primario seleccionado
- [ ] Skills requeridos identificados
- [ ] Validación JSON generada

---

### 2. Delegación al Agente Seleccionado (Agente Dinámico | Validado por Orchestrator)

**Objetivo**: Ejecutar la tarea con el agente especializado

**Tareas**:

- Invocar al agente seleccionado con el contexto completo
- Proporcionar skills necesarios
- Monitorear progreso de ejecución
- Capturar output y resultados

**Asignación Dinámica**:

- **Agente**: {Según matriz de selección}
- **Skills**: {Según tipo de tarea}
- **Validador**: {Según agente seleccionado}

**Ejemplos de Delegación**:

```yaml
# Ejemplo 1: Nueva feature con Clean Architecture
solicitud: "Crear módulo de gestión de vacaciones usando clean architecture"
agente: software-architect-tdd-ddd
skills: [go-clean-architecture, echo-routes, gorm-models]
validator: go-reviewer

# Ejemplo 2: Bug en producción
solicitud: "El endpoint de perfiles retorna 500 cuando el usuario no tiene nivel asignado"
agente: go-debugger
skills: [debug-master]
validator: go-reviewer

# Ejemplo 3: Integración con SII
solicitud: "Implementar envío de facturas al SII con reintentos automáticos"
agente: api-integration-expert
skills: [sii-invoicing, multi-database]
validator: go-reviewer

# Ejemplo 4: Migración de controller legacy
solicitud: "Refactorizar el controller Profile a clean architecture"
agente: migration-specialist
skills: [go-clean-architecture]
validator: software-architect-tdd-ddd
```

**Criterios de Salida**:

- [ ] Agente ejecutó correctamente
- [ ] Output capturado y formateado
- [ ] Errores reportados si ocurrieron

---

### 3. Validación de Resultados (go-reviewer | Validado por Orchestrator)

**Objetivo**: Asegurar calidad del trabajo realizado

**Tareas**:

- Revisar código generado (si aplica)
- Verificar cumplimiento de convenciones del proyecto
- Validar patrón arquitectónico correcto (MVC vs Clean Architecture)
- Verificar que no se introdujeron regresiones

**Asignación**:

- **Agente**: go-reviewer
- **Skills**: `go-code-reviewer`
- **Dependencias**: Fase 2 completada

**Criterios de Salida**:

- [ ] Código revisado y aprobado
- [ ] Issues documentados si existen
- [ ] Recomendaciones proporcionadas

---

### 4. Testing (go-test-runner | Validado por Orchestrator)

**Objetivo**: Validar funcionalidad y prevenir regresiones

**Tareas**:

- Ejecutar tests del módulo afectado
- Verificar cobertura mínima (80%+)
- Ejecutar tests de integración si aplica
- Reportar tests fallidos

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `go-code-reviewer` (para validar calidad de tests)
- **Dependencias**: Fase 3 completada

**Criterios de Salida**:

- [ ] Tests ejecutados
- [ ] Cobertura medida
- [ ] Tests fallidos identificados

---

### 5. Documentación (technical-writer | Opcional)

**Objetivo**: Documentar cambios realizados

**Tareas**:

- Actualizar CLAUDE.md si hay cambios arquitectónicos
- Documentar nuevos endpoints si aplica
- Crear guías de uso si es una nueva feature
- Actualizar README.md si hay cambios en comandos esenciales

**Asignación**:

- **Agente**: technical-writer
- **Skills**: `technical-writer`
- **Trigger**: Solo si la tarea lo requiere

**Criterios de Salida**:

- [ ] Documentación actualizada
- [ ] Cambios registrados

## Uso de otros Commands y MCPs

El orchestrator puede invocar commands especializados según la situación:

```yaml
commands_invocados:
  - name: plan-manage
    trigger: Cuando la tarea es compleja y requiere planificación estructurada
    output_required: Plan aprobado en .claude/plans/

mcps_utilizados:
  - name: web-reader
    purpose: Consultar documentación de Echo, GORM, o librerías externas
    trigger: Cuando el agente necesita consultar documentación oficial
    
  - name: image-analyzer
    purpose: Analizar diagramas o screenshots de errores
    trigger: Cuando el usuario proporciona evidencia visual

contexto_compartido:
  location: .claude/context/orchestrator-state.json
  format: JSON
  consumers: [todos los agentes]
```

## Output y Artefactos

| Artefacto                | Ubicación                                  | Formato    | Obligatorio     |
|--------------------------|--------------------------------------------|------------|-----------------|
| Log de orquestación      | `.claude/logs/orchestrator-{date}.log`     | Plain text | Sí              |
| Estado de ejecución      | `.claude/context/orchestrator-state.json`  | JSON       | Sí              |
| Reporte de delegación    | `.claude/reports/delegation-{id}.md`       | Markdown   | No              |

## Reglas Críticas

- **No modificación de código**: El orchestrator nunca modifica código directamente
- **Selección basada en contexto**: Siempre analizar el tipo de tarea antes de seleccionar agente
- **Respetar arquitectura híbrida**: Distinguir entre módulos MVC (controllers/) y Clean Architecture (internal/)
- **Delegación explícita**: Siempre documentar qué agente se seleccionó y por qué
- **Validación de resultados**: Siempre verificar calidad antes de dar por completada la tarea
- **Idempotencia**: Múltiples ejecuciones con mismos inputs deben producir mismos resultados
- **Detección de tareas complejas**: Si la tarea requiere múltiples fases, usar planning-agent primero

## Detección de Tareas Complejas

El orchestrator debe detectar cuando una tarea es suficientemente compleja como para requerir planificación previa:

**Indicadores de complejidad**:

- Múltiples módulos afectados
- Cambios en arquitectura (MVC ↔ Clean Architecture)
- Migraciones de base de datos
- Cambios en API pública (endpoints existentes)
- Integraciones con servicios externos
- Refactorizaciones grandes

**Acción**: Si se detectan 2+ indicadores, delegar primero a `planning-agent` para generar un plan estructurado.

---

## Acción del Usuario

Describe la tarea que deseas realizar en el proyecto Reverence Hotels API, incluyendo:

1. **Descripción de la tarea**: ¿Qué necesitas hacer? (ej: "Crear un nuevo endpoint para gestionar vacaciones")
2. **Tipo de tarea**: ¿Es desarrollo, bug fix, refactor, migración, testing o documentación?
3. **Alcance**: ¿Qué módulos o partes del sistema están afectados?
4. **Prioridad**: ¿Crítica, alta, media o baja?
5. **Contexto adicional**: ¿Hay información relevante? (ej: "Usar clean architecture", "Es un bug en producción", etc.)

**Ejemplos de solicitudes válidas**:

> "Crear un nuevo módulo de gestión de ausencias usando clean architecture. Debe tener endpoints para crear, listar y aprobar ausencias. Prioridad: alta."

> "El endpoint GET /api/v1/profile/:id está retornando error 500 cuando el perfil no tiene foto asignada. Necesito que lo investigues. Es un bug en producción."

> "Refactorizar el controller Invoices para que use clean architecture en lugar de MVC. Mantener compatibilidad con los endpoints existentes."

> "Implementar integración con un nuevo servicio web de reservas. La documentación está en https://api.example.com/docs."

> "Escribir tests para el módulo user. Actualmente tiene 0% de cobertura. Necesito al menos 80%."