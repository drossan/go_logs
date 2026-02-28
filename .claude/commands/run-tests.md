---
name: run-tests
version: 1.0.0
author: danielrossellosanchez
description: Ejecuta tests en el proyecto Reverence Hotels API con opciones de cobertura, filtrado por módulo y análisis de resultados
usage: "run-tests [--coverage] [--module=path] [--verbose]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: []
---

# Comando: Run Tests

## Objetivo

Ejecutar la suite de tests del proyecto Reverence Hotels API de manera estructurada, proporcionando:

- Ejecución de tests con opciones de cobertura
- Filtrado por módulos específicos (legacy MVC o clean architecture)
- Análisis y reporte de resultados
- Detección de tests con problemas o coverage insuficiente
- Validación de que no hay regresiones

**No modifica código**, solo ejecuta tests y genera reportes de resultados.

## Contexto Requerido del Usuario

- [ ] Ámbito de los tests (todos o módulo específico)
- [ ] ¿Requiere reporte de cobertura? (opcional)
- [ ] ¿Nivel de detalle requerido? (básico/verbose)
- [ ] ¿Umbral mínimo de coverage esperado? (opcional, default: 80%)
- [ ] ¿Existe algún test específico que deba ejecutarse? (opcional)

## Análisis Inicial (Obligatorio)

Antes de cualquier acción, el command debe evaluar:

- Estado actual del proyecto (¿compila? ¿hay tests?)
- Módulos afectados por el cambio actual
- Historial de ejecuciones de tests anteriores
- Disponibilidad de dependencias de测试
- Riesgos identificados en el código a测试

### Pre-ejecución: Checklist Obligatorio

El command debe verificar:

- [ ] ¿El proyecto compila sin errores? → Si no, fallar temprano
- [ ] ¿Existen tests en el ámbito solicitado? → Si no, advertir
- [ ] ¿Las dependencias de测试 están disponibles? → Verificar go.mod
- [ ] ¿Hay suficientes recursos del sistema? → Verificar memoria/CPU
- [ ] ¿El usuario especificó un módulo válido? → Validar ruta

**Output esperado**: JSON de validación antes de continuar.

```json
{
  "validation_passed": true,
  "project_state": "ready",
  "available_tests": {
    "total_modules": 42,
    "legacy_controllers": 28,
    "clean_architecture": 14
  },
  "risks": [
    "Algunos módulos legacy pueden tener tests desactualizados"
  ],
  "estimated_duration": "3-5 minutos",
  "blocking_issues": []
}
```

## Selección de Agentes y Skills

### Fase 1: Ejecución de Tests

```yaml
responsible: go-test-runner
accountable: go-reviewer
consulted: [ go-code-reviewer ]
informed: [ go-orchestrator ]
```

**Justificación**: 
- `go-test-runner` es el agente especializado en testing, QA y validación de calidad en proyectos Go
- Su descripción menciona explícitamente "especializado en razonamiento sobre testing, validación de calidad y ejecución de tests"
- Es el agente más adecuado para ejecutar, analizar y reportar resultados de tests

### Fase 2: Análisis de Resultados

```yaml
responsible: go-test-runner
accountable: go-reviewer
consulted: [ debug-master ]
informed: [ technical-writer ]
```

**Justificación**:
- `go-test-runner` continúa siendo responsable por su conocimiento del framework de testing
- `go-reviewer` valida la calidad de los resultados y detecta posibles problemas
- `debug-master` proporciona expertise en análisis de failures si los hay

## Flujo de Trabajo Orquestado

### 1. Preparación y Validación del Entorno (go-test-runner | Validado por go-reviewer)

**Objetivo**: Asegurar que el entorno esté listo para ejecutar tests

**Tareas**:

- Verificar que el proyecto compila (`go build ./...`)
- Validar que `go.mod` y `go.sum` están actualizados
- Identificar módulos disponibles para测试
- Verificar acceso a las 3 bases de datos de测试 (si aplica)
- Limpiar caché de测试 anterior si es necesario
- Preparar directorio de reportes `.claude/reports/testing/`

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: N/A (ejecución directa de comandos Go)
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Proyecto compila sin errores
- [ ] Dependencias actualizadas
- [ ] Entorno validado y listo

---

### 2. Ejecución de Tests (go-test-runner | Validado por go-reviewer)

**Objetivo**: Ejecutar la suite de tests según el ámbito especificado

**Tareas**:

- Ejecutar tests con el comando apropiado:
  - Todos: `go test ./... -v -coverprofile=coverage.out`
  - Por módulo: `go test ./internal/backend/user/... -v`
  - Con coverage: `go test -coverprofile=coverage.out ./...`
- Capturar output completo (stdout/stderr)
- Registrar tiempo de ejecución
- Identificar tests que fallan o pasan con warnings
- Recopilar métricas de coverage por paquete

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: N/A
- **Dependencias**: Fase 1 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Tests ejecutados completamente
- [ ] Output capturado en archivo
- [ ] Métricas recopiladas

---

### 3. Análisis y Reporte de Resultados (go-test-runner | Validado por go-reviewer)

**Objetivo**: Generar reporte detallado de resultados y detectar problemas

**Tareas**:

- Parsear output de tests para extraer:
  - Número de tests pasados/fallados/skipped
  - Coverage por paquete y total
  - Tests que fallaron con sus mensajes de error
  - Tests lentos (>5 segundos)
  - Race conditions detectadas
- Generar reporte en Markdown con:
  - Resumen ejecutivo (pasaron/total, coverage %)
  - Detalle de failures con stack traces
  - Análisis de coverage (paquetes por debajo del umbral)
  - Recomendaciones para mejorar
- Guardar reporte en `.claude/reports/testing/{timestamp}-run-tests.md`
- Si hay failures, analizar causas raíz

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `debug-master` (si hay failures para análisis)
- **Dependencias**: Fase 2 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Reporte generado y guardado
- [ ] Tests con problemas identificados
- [ ] Coverage analizado
- [ ] Recomendaciones documentadas

---

### 4. Validación de Calidad (go-reviewer | Validado por go-orchestrator)

**Objetivo**: Validar que los resultados cumplan los estándares de calidad

**Tareas**:

- Revisar reporte generado en fase 3
- Verificar que coverage mínimo sea alcanzado (default: 80%)
- Validar que no hay regresiones (tests que antes pasaban y ahora fallan)
- Identificar tests problemáticos recurrentes
- Proponer acciones correctivas si es necesario
- Actualizar métricas históricas de测试

**Asignación**:

- **Agente**: go-reviewer
- **Skills**: `go-code-reviewer`
- **Dependencias**: Fase 3 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Calidad de tests validada
- [ ] Regresiones detectadas (si las hay)
- [ ] Acciones correctivas propuestas (si aplica)

---

### 5. Notificación de Resultados (go-orchestrator | Validado por usuario)

**Objetivo**: Comunicar resultados al usuario de manera clara y accionable

**Tareas**:

- Presentar resumen ejecutivo al usuario
- Indicar ubicación del reporte detallado
- Si hay failures: preguntar si desea análisis más profundo
- Si coverage es bajo: sugerir estrategias para mejorar
- Ofrecer opciones siguientes (re-ejecutar, debuggear, etc.)

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: N/A
- **Dependencias**: Fase 4 completada
- **Validador**: Usuario

**Criterios de Salida**:

- [ ] Usuario notificado con resultados
- [ ] Reporte accesible
- [ ] Siguientes pasos claros

## Uso de otros Commands y MCPs

No se invocan otros commands durante la ejecución de este command.

Sin embargo, este command puede ser invocado por:
- `pre-flight` - Como parte del checklist previo a desarrollo/deploys
- `plan-manager` - Como paso de validación en ejecución de planes
- Comandos de deployment - Como validación previa al despliegue

## Output y Artefactos

| Artefacto                | Ubicación                                          | Formato  | Validador        | Obligatorio |
|--------------------------|----------------------------------------------------|----------|------------------|-------------|
| Reporte de测试          | `.claude/reports/testing/{timestamp}-run-tests.md` | Markdown | go-reviewer      | Sí          |
| Coverage profile         | `.claude/reports/testing/coverage.out`             | Binary   | go-test-runner   | Sí          |
| HTML coverage report     | `.claude/reports/testing/coverage.html`            | HTML     | -                | No          |
| Log de ejecución         | `.claude/logs/run-tests-{timestamp}.log`           | Plain    | -                | Sí          |
| JSON de métricas         | `.claude/metrics/testing/{timestamp}.json`         | JSON     | schema-validator | No          |

## Rollback y Cancelación

Si el command falla o el usuario cancela durante la ejecución:

### Procedimiento de Rollback

1. **Detener tests en curso**: Enviar `Ctrl+C` al proceso de `go test`
2. **Eliminar artefactos parciales**:
   - Borrar reportes incompletos en `.claude/reports/testing/`
   - Limpiar archivos temporales
3. **Registrar cancelación**:
   ```
   .claude/logs/cancelled-run-tests-{timestamp}.log
   ```
4. **Restaurar estado previo**: Limpiar caché de测试 si es necesario
5. **Notificar al usuario**: Confirmar cancelación y estado de limpieza

### Estados Finales Posibles

- `completed`: Tests ejecutados exitosamente (incluso si algunos fallaron)
- `failed`: Error irrecuperable (ej: proyecto no compila)
- `cancelled`: Cancelado por usuario
- `partial`: Ejecución interrumpida pero con resultados parciales útiles

## Reglas Críticas

- **No modificación de código**: Este command solo ejecuta tests y genera reportes
- **No auto-corrección**: Si tests fallan, no intentar arreglarlos automáticamente
- **Coverage mínimo**: Advertir si coverage < 80% (configurable)
- **Detección de regresiones**: Identificar tests que antes pasaban y ahora fallan
- **Respeto al ámbito**: Ejecutar solo los tests solicitados por el usuario
- **Transparencia en reportes**: Incluir tanto successes como failures
- **Idempotencia**: Ejecutar múltiples veces produce resultados consistentes
- **Validación de entorno**: Verificar prerequisitos antes de ejecutar

---

## Acción del Usuario

Especifica cómo deseas ejecutar los tests:

### Opciones Básicas

**Ejecutar todos los tests**:
> "run-tests"

**Ejecutar con coverage**:
> "run-tests --coverage"

**Ejecutar un módulo específico**:
> "run-tests --module=./internal/backend/user"
> "run-tests --module=./controllers/Profile"

**Ejecutar con detalle verbose**:
> "run-tests --verbose"

### Opciones Avanzadas

**Ejecutar con umbral de coverage específico**:
> "run-tests --coverage --threshold=90"

**Ejecutar solo tests de una capa arquitectónica**:
> "run-tests --module=./internal/... --type=clean-architecture"
> "run-tests --module=./controllers/... --type=legacy-mvc"

**Ejecutar y generar reporte HTML**:
> "run-tests --coverage --html-report"

### Ejemplos Completos

```bash
# Ejecutar todos los tests con coverage
run-tests --coverage

# Ejecutar solo tests del módulo user
run-tests --module=./internal/backend/user

# Validar antes de un deploy
run-tests --coverage --threshold=85 --verbose
```

**Output esperado**:

El command generará un reporte en `.claude/reports/testing/{timestamp}-run-tests.md` con:

- ✅ Número de tests que pasaron
- ❌ Número de tests que fallaron (si los hay)
- 📊 Coverage total y por paquete
- ⚠️ Advertencias (tests lentos, coverage bajo, etc.)
- 🔍 Análisis de failures (si los hay)
- 💡 Recomendaciones para mejorar