---
name: pre-flight
version: 1.0.0
author: Reverence Hotels API Team
description: Comando de pre-flight para validar el estado del proyecto antes de iniciar desarrollo, testing o despliegue en Reverence Hotels API
usage: "pre-flight [--environment=local|pre|pro] [--check=deps|tests|db|all]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: [go-expert]
---

# Comando: Pre-Flight Check

## Objetivo

Ejecutar una validación completa del estado del proyecto Reverence Hotels API antes de iniciar sesiones de desarrollo, testing o despliegue. Este comando realiza checks críticos sobre:

- Estado de dependencias de Go
- Conectividad con las 3 bases de datos MySQL
- Estado de los tests unitarios y de integración
- Configuración de variables de entorno
- Estado de servicios externos (SII, Porta Sigma, Slack)
- Integridad de la arquitectura híbrida (MVC + Clean Architecture)
- Validaciones de seguridad críticas

**Output**: Reporte detallado de estado con indicadores de salud (health score) y recomendaciones.

**No genera plan** ni modifica código, solo valida y reporta.

---

## Contexto Requerido del Usuario

- [ ] Ambiente a validar (local, pre-production, production)
- [ ] Tipo de check (deps, tests, db, security, all)
- [ ] Opcional: Reporte previo para comparar regresiones

---

## Análisis Inicial

### Pre-ejecución: Checklist Obligatorio

El command debe verificar:

- [ ] ¿El ambiente especificado corresponde a ENV variable?
- [ ] ¿Existe archivo .env para el ambiente?
- [ ] ¿Las credenciales de base de datos están configuradas?
- [ ] ¿Los servicios externos son alcanzables?

**Output esperado**: JSON de validación antes de continuar.

```json
{
  "validation_passed": true,
  "environment": "pre",
  "checks_to_execute": ["dependencies", "databases", "tests", "external_services"],
  "estimated_duration_seconds": 45,
  "blocking_issues": []
}
```

---

## Selección de Agentes y Skills

### Framework RACI

```yaml
fase_1_validacion_dependencias:
  responsible: go-debugger
  accountable: software-architect-tdd-ddd
  consulted: [ go-expert, debug-master ]
  informed: [ planning-agent ]

fase_2_validacion_databases:
  responsible: go-debugger
  accountable: api-integration-expert
  consulted: [ multi-database, gorm-models ]
  informed: [ migration-specialist ]

fase_3_validacion_tests:
  responsible: go-test-runner
  accountable: go-reviewer
  consulted: [ go-expert, go-code-reviewer ]
  informed: [ planning-agent ]

fase_4_validacion_seguridad:
  responsible: go-reviewer
  accountable: software-architect-tdd-ddd
  consulted: [ jwt-auth, go-code-reviewer ]
  informed: [ go-orchestrator ]
```

---

## Flujo de Trabajo Orquestado

### 1. Validación de Dependencias y Configuración (go-debugger | Validado por software-architect-tdd-ddd)

**Objetivo**: Verificar que el proyecto tiene todas las dependencias necesarias y configuración válida

**Tareas**:

- Ejecutar `go mod verify` para verificar integridad de dependencias
- Ejecutar `go mod tidy` y reportar cambios necesarios
- Validar versión de Go (>= 1.21)
- Verificar que los archivos .env existan para el ambiente especificado
- Validar variables de entorno críticas (DATA_BASE_*, API_URL, ENV)
- Verificar existencia de certificados .pem para SII
- Validar configuración de Echo framework

**Asignación**:

- **Agente**: go-debugger
- **Skills**: `go-expert`, `debug-master`
- **Validador**: software-architect-tdd-ddd

**Criterios de Salida**:

- [ ] Integridad de dependencias confirmada
- [ ] Variables de entorno críticas presentes
- [ ] Versión de Go válida
- [ ] Certificados digitales localizados

---

### 2. Validación de Conectividad de Bases de Datos (go-debugger | Validado por api-integration-expert)

**Objetivo**: Verificar conectividad y estado de las 3 bases de datos MySQL

**Tareas**:

- Intentar conexión a base de datos Principal (sensesho_api)
- Intentar conexión a base de datos SII (reverence_sii)
- Intentar conexión a base de datos Products (economato)
- Verificar que las tablas críticas existen
- Validar configuración de GORM (connection pool, timeouts)
- Verificar que vistas (view_reduce_profiles) son accesibles
- Reportar latencias de conexión

**Asignación**:

- **Agente**: go-debugger
- **Skills**: `multi-database`, `gorm-models`
- **Dependencias**: Fase 1 completada
- **Validador**: api-integration-expert

**Criterios de Salida**:

- [ ] Las 3 bases de datos son accesibles
- [ ] Tablas críticas verificadas
- [ ] Latencias dentro de SLA (<100ms)
- [ ] Configuración de GORM válida

---

### 3. Validación de Tests (go-test-runner | Validado por go-reviewer)

**Objetivo**: Ejecutar suite de tests y reportar estado de calidad

**Tareas**:

- Ejecutar `go test ./...` con flags de coverage
- Validar coverage mínimo (80% para módulos críticos)
- Ejecutar tests de módulos Clean Architecture en internal/
- Ejecutar tests de controllers legacy
- Verificar tests de integración con bases de datos
- Identificar tests flaky o con race conditions
- Comparar con ejecución previa (si existe reporte)

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `go-expert`, `go-code-reviewer`
- **Dependencias**: Fase 2 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Todos los tests pasan
- [ ] Coverage >= 80% en módulos críticos
- [ ] No race conditions detectadas
- [ ] No tests flaky identificados

---

### 4. Validación de Seguridad y Arquitectura (go-reviewer | Validado por software-architect-tdd-ddd)

**Objetivo**: Verificar compliance de seguridad y patrones arquitectónicos

**Tareas**:

- Verificar configuración de JWT (secret keys, expiration)
- Validar middleware de autorización por niveles
- Verificar que APIs sensibles requieren autenticación
- Validar configuración de rate limiting
- Verificar que módulos en internal/ siguen Clean Architecture
- Validar que no hay hardcoded credentials
- Verificar configuración de CORS
- Validar que APIs de PMS usan API Key authentication

**Asignación**:

- **Agente**: go-reviewer
- **Skills**: `jwt-auth`, `go-code-reviewer`
- **Dependencias**: Fase 3 completada
- **Validador**: software-architect-tdd-ddd

**Criterios de Salida**:

- [ ] No hardcoded credentials detectados
- [ ] Autenticación JWT configurada correctamente
- [ ] APIs críticas protegidas adecuadamente
- [ ] Clean Architecture respetada en módulos internal/
- [ ] Rate limiting activo

---

### 5. Validación de Servicios Externos (api-integration-expert | Validado por go-orchestrator)

**Objetivo**: Verificar conectividad con servicios externos críticos

**Tareas**:

- Verificar configuración de Slack (token y channel)
- Validar conectividad con API de Porta Sigma
- Verificar configuración de endpoints del SII (AEAT)
- Validar que URLs de SII corresponden al ambiente (pre/pro)
- Verificar configuración de SMTP para emails
- Validar timeouts de conexiones externas

**Asignación**:

- **Agente**: api-integration-expert
- **Skills**: `multi-database`, `debug-master`
- **Dependencias**: Fase 4 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Configuración de Slack válida
- [ ] API de Porta Sigma accesible
- [ ] Endpoints SII configurados correctamente
- [ ] SMTP configurado para envío de emails

---

## Uso de otros Commands y MCPs

Este comando es autónomo y no invoca otros commands ni MCPs del proyecto.

---

## Output y Artefactos

| Artefacto                | Ubicación                                    | Formato    | Validador         | Obligatorio |
|--------------------------|----------------------------------------------|------------|-------------------|-------------|
| Reporte de Pre-flight    | `.claude/reports/pre-flight-{timestamp}.md`  | Markdown   | -                 | Sí          |
| Health Score             | Incluido en reporte                          | JSON       | -                 | Sí          |
| Log de ejecución         | `.claude/logs/pre-flight-{date}.log`         | Plain text | -                 | Sí          |
| Comparación con previo   | Incluido en reporte si existe previo         | Diff       | -                 | No          |

### Estructura del Reporte

```markdown
# Pre-Flight Check Report

**Environment**: pre
**Timestamp**: 2025-01-20 15:30:00
**Duration**: 42 seconds
**Health Score**: 85/100

## Resumen Ejecutivo

✅ Dependencies: PASS
✅ Databases: PASS (3/3 connected)
⚠️  Tests: WARN (78% coverage, target: 80%)
✅ Security: PASS
⚠️  External Services: WARN (Slack token expiring soon)

## Detalles por Categoría

### Dependencies
- Go version: 1.21.5 ✅
- Dependencies verified ✅
- Environment variables: All present ✅

### Databases
- Principal: Connected (12ms) ✅
- SII: Connected (18ms) ✅
- Products: Connected (9ms) ✅

### Tests
- Total tests: 342
- Passed: 340
- Failed: 2 (legacy/Invoices)
- Coverage: 78%
- Race conditions: 0

### Security
- JWT configured ✅
- No hardcoded secrets ✅
- APIs protected ✅

### External Services
- Slack: Configured ⚠️ (token expires in 3 days)
- Porta Sigma: Connected ✅
- SII endpoints: Configured for pre-prod ✅

## Recomendaciones

1. Update Slack token before 2025-01-23
2. Fix failing tests in legacy/Invoices
3. Improve coverage in internal/backend/form/

## Health Score Breakdown

- Dependencies: 20/20
- Databases: 20/20
- Tests: 15/20
- Security: 20/20
- External Services: 10/20
```

---

## Rollback y Cancelación

Este comando es de solo lectura y no modifica archivos, por lo que no requiere rollback.

Si el usuario cancela durante la ejecución:

1. **Detener ejecución inmediata**
2. **Generar reporte parcial** con resultados obtenidos hasta el momento
3. **Registrar cancelación** en `.claude/logs/pre-flight-cancelled-{timestamp}.log`
4. **Notificar estado parcial** al usuario

---

## Reglas Críticas

- **Solo lectura**: Este comando no puede crear, modificar o eliminar archivos
- **No ejecuta código**: Solo valida estado, no corre la aplicación
- **Reporte obligatorio**: Siempre genera reporte en `.claude/reports/`
- **Idempotencia**: Ejecutar múltiples veces produce mismo resultado
- **Timeout global**: Si ejecución excede 2 minutos, cancelar y reportar timeout
- **Comparación inteligente**: Si existe reporte previo, incluir comparación de regresiones
- **Health score**: Siempre calcular y reportar score de salud del proyecto
- **No bypass de errores**: Cualquier check crítico fallido debe reportarse como WARNING o ERROR

---

## Acción del Usuario

Para ejecutar el pre-flight check, proporciona:

1. **Ambiente**: ¿Qué ambiente validar? (`local`, `pre`, `pro`)
2. **Tipo de check**: ¿Qué validar? 
   - `deps` - Solo dependencias y configuración
   - `db` - Solo bases de datos
   - `tests` - Solo suite de tests
   - `security` - Solo validaciones de seguridad
   - `all` - Todos los checks (default)

**Ejemplos de uso**:

```
# Check completo en ambiente local
pre-flight local all

# Solo tests en pre-producción
pre-flight pre tests

# Solo bases de datos en producción
pre-flight pro db
```

**Output esperado**:
- Reporte detallado en `.claude/reports/pre-flight-{timestamp}.md`
- Health score del proyecto (0-100)
- Recomendaciones accionables
- Comparación con ejecución previa (si existe)