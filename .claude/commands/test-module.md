---
name: test-module
version: 1.0.0
author: development-team
description: Comando para test-module en Reverence Hotels API
usage: "test-module [module-path] [--coverage] [--verbose]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: []
---

# Comando: Test Module

## Objetivo

Ejecutar tests de módulos específicos en el proyecto Reverence Hotels API, proporcionando análisis detallado de cobertura, rendimiento y resultados de pruebas.

Este comando es de tipo **executable**, lo que significa que:
- Ejecuta validaciones directamente
- No genera plan de trabajo
- No puede modificar archivos
- Retorna resultados inmediatos en formato de reporte

**Si durante la ejecución se detecta la necesidad de modificar código (tests fallidos, refactorización necesaria), el comando debe abortar y recomendar usar un comando de planificación.**

## Contexto Requerido del Usuario

Antes de ejecutar el comando, se debe proporcionar:

- [ ] **Módulo objetivo**: Ruta del módulo a testear (ej: `./internal/backend/user/...`, `./controllers/Profile/...`)
- [ ] **Tipo de tests**: Unitarios, integración, o todos (por defecto: todos)
- [ ] **Nivel de cobertura**: Cobertura mínima esperada (opcional, por defecto: 80%)
- [ ] **Verbosidad**: Nivel de detalle en salida (standard, verbose)
- [ ] **Timeout**: Tiempo máximo de ejecución en segundos (opcional, por defecto: 120s)

## Análisis Inicial (Obligatorio)

### Validaciones Pre-ejecución

El comando debe evaluar antes de ejecutar:

```json
{
  "validation_passed": true,
  "risks": [
    "Tests pueden fallar si dependencias externas no están disponibles",
    "Tests de integración requieren bases de datos configuradas"
  ],
  "required_approvals": [],
  "estimated_complexity": "low",
  "blocking_issues": [],
  "execution_mode": "direct"
}
```

### Checklist de Pre-ejecución

El comando debe verificar:

- [ ] ¿Existe el módulo especificado? → Si no, fallar con error claro
- [ ] ¿Hay archivos de tests en el módulo? → Si no, advertir y continuar
- [ ] ¿Es un módulo Clean Architecture o Legacy MVC? → Determina estrategia de testing
- [ ] ¿Requiere bases de datos? → Configurar conexiones apropiadas
- [ ] ¿Hay variables de entorno necesarias? → Validar antes de ejecutar
- [ ] ¿El contexto del usuario es suficiente? → Solicitar aclaraciones si falta algo

**Output esperado**: JSON de validación antes de continuar.

```json
{
  "module_exists": true,
  "module_type": "clean-architecture",
  "test_files_found": 5,
  "requires_database": true,
  "databases_needed": ["principal", "sii"],
  "env_vars_valid": true,
  "ready_to_execute": true
}
```

## Selección de Agentes y Skills (Framework RACI)

### Agente Principal

```yaml
fase_ejecucion_tests:
  responsible: go-test-runner
  accountable: go-reviewer
  consulted: [ go-code-reviewer, debug-master ]
  informed: [ go-orchestrator ]
```

**Justificación de selección**:

- **go-test-runner**: Descripción dice "Senior QA Engineer especializado en razonamiento sobre testing, validación de calidad y ejecución de tests en proyectos Go". Es el agente perfecto para esta tarea.
- **go-reviewer**: Descripción dice "Senior Go Code Reviewer especializado en razonamiento sobre calidad de código, mejores prácticas y convenciones del proyecto". Validará que los tests sigan convenciones.
- **Skills**: Requiere skills para análisis de código y debugging si hay fallos.

### Skills Requeridas

Analizando las skills disponibles:

- **debug-master**: Propósito "Provides debug-master-related expertise and capabilities". Necesario para analizar stack traces si hay tests fallidos.
- **go-code-reviewer**: Propósito "Provides code-reviewer-related expertise and capabilities". Útil para validar calidad de los tests ejecutados.

## Flujo de Trabajo Orquestado

### 1. Preparación y Validación del Entorno (go-test-runner)

**Objetivo**: Verificar que el entorno está configurado correctamente para ejecutar tests

**Tareas**:

- Verificar que el módulo existe en la ruta especificada
- Identificar si es módulo Clean Architecture (`internal/`) o Legacy MVC (`controllers/`, `services/`, `models/`)
- Contar archivos de test disponibles (`*_test.go`)
- Validar que las variables de entorno necesarias estén configuradas (`DATA_BASE_*`, `ENV`, etc.)
- Verificar disponibilidad de bases de datos si el módulo las requiere
- Compilar el proyecto para detectar errores de compilación antes de ejecutar tests

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `debug-master` (para diagnosticar problemas de compilación)
- **Validador**: Auto-validación

**Criterios de Salida**:

- [ ] Módulo localizado y tipificado (Clean Architecture vs Legacy)
- [ ] Número de archivos de test identificados
- [ ] Variables de entorno validadas
- [ ] Dependencias externas verificadas (BD, APIs)
- [ ] Proyecto compila sin errores

---

### 2. Ejecución de Tests (go-test-runner)

**Objetivo**: Ejecutar la suite de tests del módulo especificado

**Tareas**:

- Construir comando `go test` apropiado según el tipo de módulo
- Incluir flags de cobertura (`-cover`) si se solicita
- Configurar timeout apropiado (por defecto: 120s)
- Ejecutar tests y capturar output completo (stdout y stderr)
- Parsear resultados para identificar PASS/FAIL/SKIP
- Extraer métricas:覆盖率, número de tests, tiempo de ejecución

**Comandos típicos a ejecutar**:

```bash
# Tests unitarios con cobertura
go test -v -cover -coverprofile=coverage.out ./internal/backend/user/...

# Tests de integración (requieren BD)
go test -v -cover ./controllers/Profile/...

# Con timeout personalizado
go test -v -timeout=180s ./services/Invoices/...
```

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `debug-master` (para analizar si hay fallos)
- **Dependencias**: Fase 1 completada
- **Validador**: go-reviewer (valida que resultados sean interpretables)

**Criterios de Salida**:

- [ ] Comando `go test` ejecutado sin errores de sintaxis
- [ ] Output capturado completamente (stdout + stderr)
- [ ] Resultados parseados (PASS/FAIL/SKIP por test)
- [ ] Métricas extraídas (cobertura, timing)

---

### 3. Análisis de Resultados (go-test-runner | Validado por go-reviewer)

**Objetivo**: Analizar los resultados de los tests y generar reporte detallado

**Tareas**:

- Clasificar resultados: éxito, fallo, skip, panic
- Analizar tests fallidos: extraer stack traces, mensajes de error
- Calcular cobertura real vs esperada
- Identificar tests lentos (>5s)
- Detectar race conditions si se ejecutó con `-race`
- Comparar con umbrales de calidad del proyecto
- Generar resumen ejecutivo

**Análisis de fallos** (si los hay):

- Extraer stack trace completo
- Identificar línea de código que falló
- Analizar causa raíz (assert error, panic, timeout)
- Proponer hipótesis de causa raíz

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `debug-master`, `go-code-reviewer`
- **Dependencias**: Fase 2 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Todos los resultados clasificados
- [ ] Tests fallidos analizados con stack trace
- [ ] Cobertura calculada y comparada con umbral
- [ ] Anomalías detectadas (race conditions, timeouts)
- [ ] Recomendaciones generadas si hay fallos

---

### 4. Generación de Reporte (go-test-runner)

**Objetivo**: Generar reporte estructurado con resultados y métricas

**Tareas**:

- Crear reporte en Markdown con secciones:
  - Resumen ejecutivo (status global, cobertura, tiempo)
  - Detalle de tests por archivo
  - Tests fallidos con stack traces
  - Métricas de cobertura (por paquete si se generó profile)
  - Recomendaciones si aplica
- Guardar reporte en `.claude/reports/test-module-{timestamp}.md`
- Guardar salida cruda en `.claude/logs/test-module-{date}.log`
- Generar JSON con métricas para consumo de otros tools

**Formato de reporte**:

```markdown
# Test Report: {module-path}

**Ejecutado**: 2025-01-20 15:30:22
**Duración**: 45.2s
**Status**: ✅ PASS | ❌ FAIL | ⚠️ PARTIAL

## Resumen

- Tests ejecutados: 42
- Pasaron: 38
- Fallaron: 4
- Skip: 0
- Cobertura: 87.3% (objetivo: 80%)

## Tests Fallidos

### TestCreateUser (user_service_test.go:45)

```
Error: assertion failed: expected user.ID > 0, got 0
Stack trace:
...
```

**Hipótesis**: Database connection not initialized

## Cobertura por Paquete

| Paquete | Cobertura | Archivos |
|---------|-----------|----------|
| domain  | 95.2%     | 5/5      |
| application | 82.1% | 3/3      |
| infrastructure/http | 78.4% | 2/2 |

## Recomendaciones

1. [Recomendación 1]
2. [Recomendación 2]
```

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `technical-writer` (para redacción clara del reporte)
- **Dependencias**: Fase 3 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Reporte Markdown generado en `.claude/reports/`
- [ ] Log crudo guardado en `.claude/logs/`
- [ ] JSON de métricas generado
- [ ] Formato validado y legible

---

### 5. Validación de Calidad (go-reviewer)

**Objetivo**: Validar que los tests cumplen con estándares del proyecto

**Tareas**:

- Verificar que los tests siguen convenciones de Go (table-driven tests)
- Validar nombramiento de tests (`TestFunctionName_State`)
- Revisar que haya tests para casos edge
- Verificar uso de mocks apropiados
- Validar que tests de integración aíslen dependencies externas
- Comprobar que hay tests positivos y negativos
- Verificar_documentación de tests complejos

**Criterios de calidad**:

- ✅ Tests tienen nombres descriptivos
- ✅ Tests son independientes (no comparten estado)
- ✅ Tests usan setUp/tearDown apropiadamente
- ✅ Tests tienen asserts claros
- ✅ Tests de integración usan fixtures o test DB

**Asignación**:

- **Agente**: go-reviewer
- **Skills**: `go-code-reviewer`
- **Dependencias**: Fase 4 completada
- **Validador**: Auto-validación

**Criterios de Salida**:

- [ ] Checklist de calidad completado
- [ ] Issues de calidad identificados
- [ ] Recomendaciones de mejora documentadas

---

### 6. Reporte de Estado Final (go-test-runner)

**Objetivo**: Presentar resumen final al usuario con actionable next steps

**Tareas**:

- Sintetizar resultados de todas las fases
- Determinar estado global (PASS/FAIL/PARTIAL)
- Generar next steps según estado:
  - **PASS**: Felicitar, no acción requerida
  - **FAIL**: Proponer diagnosticar o crear plan de fixes
  - **PARTIAL**: Identificar tests que necesitan atención
- Presentar métricas clave en formato legible
- Ofrecer comandos útiles para debugging si hay fallos

**Formato de presentación**:

```
📊 RESULTADO FINAL: ✅ PASS | ❌ FAIL | ⚠️ PARTIAL

Módulo: ./internal/backend/user/...
Tests: 38/42 pasaron (90.5%)
Cobertura: 87.3% (objetivo: 80%) ✅
Tiempo: 45.2s

🔍 Tests fallidos: 4
  - TestCreateUser (user_service_test.go:45)
  - TestUpdateProfile (profile_handler_test.go:78)
  - TestDeleteUser (user_service_test.go:120)
  - TestAuthMiddleware (auth_middleware_test.go:34)

💡 Next steps:
  1. Revisar logs: .claude/logs/test-module-20250120.log
  2. Analizar stack traces en reporte: .claude/reports/test-module-20250120-153022.md
  3. Diagnosticar con: /debug-module ./internal/backend/user/... --focus=TestCreateUser
  4. Si requieres fixes: /plan-fix-tests ./internal/backend/user/...
```

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `technical-writer`
- **Dependencias**: Fase 5 completada
- **Validador**: Auto-validación

**Criterios de Salida**:

- [ ] Resumen ejecutivo presentado
- [ ] Estado claramente identificado (PASS/FAIL/PARTIAL)
- [ ] Next steps accionables definidos
- [ ] Links a artefactos generados (reportes, logs)

## Uso de otros Commands y MCPs

### Commands Invocados

Este comando **no invoca otros commands** directamente, pero puede recomendar su uso según resultados:

```yaml
recomendaciones_por_resultado:
  pass:
    next_action: null
    message: "Todos los tests pasaron. No se requiere acción adicional."

  fail:
    next_command: "/debug-module [module-path] --focus=[failing-test]"
    reason: "Analizar stack traces y causas raíz de tests fallidos"

  partial:
    next_command: "/plan-fix-tests [module-path]"
    reason: "Planificar corrección sistemática de tests fallidos"
```

### MCPs Utilizados

No se requieren MCPs externos para este comando. Todo el análisis se realiza con:

- Go toolchain nativo (`go test`)
- Parser de output de tests
- Análisis estático de código Go

### Contexto Compartido

Si se generan reportes, otros commands pueden consumirlos:

```yaml
contexto_generado:
  reports:
    location: ".claude/reports/"
    format: "Markdown"
    consumers:
      - "/debug-module" (para analisis profundo de fallos)
      - "/code-coverage-report" (para analisis de cobertura)

  metrics:
    location: ".claude/reports/metrics-{timestamp}.json"
    format: "JSON"
    consumers:
      - "/quality-dashboard" (dashboard de metricas)
      - "/ci-gatekeeper" (validacion en pipeline CI/CD)
```

## Output y Artefactos

| Artefacto | Ubicación | Formato | Validador | Obligatorio |
|-----------|-----------|---------|-----------|-------------|
| Reporte de tests | `.claude/reports/test-module-{timestamp}.md` | Markdown | go-reviewer | Sí |
| Log crudo de ejecución | `.claude/logs/test-module-{date}.log` | Plain text | - | Sí |
| Métricas JSON | `.claude/reports/metrics-{timestamp}.json` | JSON | schema-validator | Sí |
| Cobertura HTML (opcional) | `.claude/reports/coverage-{timestamp}.html` | HTML | - | No |
| Perfil de cobertura | `.claude/reports/coverage-{timestamp}.out` | Binary | - | No |

### Estructura del JSON de Métricas

```json
{
  "module_path": "./internal/backend/user/...",
  "timestamp": "2025-01-20T15:30:22Z",
  "duration_seconds": 45.2,
  "status": "partial",
  "summary": {
    "total": 42,
    "passed": 38,
    "failed": 4,
    "skipped": 0,
    "coverage_percentage": 87.3,
    "coverage_target": 80.0,
    "coverage_met": true
  },
  "failing_tests": [
    {
      "name": "TestCreateUser",
      "file": "user_service_test.go",
      "line": 45,
      "error": "assertion failed: expected user.ID > 0, got 0",
      "stack_trace": "..."
    }
  ],
  "performance": {
    "slowest_tests": [
      {"name": "TestIntegrationEndToEnd", "duration_seconds": 12.3}
    ]
  },
  "quality_check": {
    "naming_convention": "pass",
    "test_independence": "pass",
    "assert_clarity": "pass",
    "edge_cases_covered": "warning",
    "documentation_adequate": "pass"
  }
}
```

## Rollback y Cancelación

### Procedimiento de Rollback

Dado que este comando no modifica código, el rollback es simple:

1. **Eliminar artefactos parciales**:
   ```bash
   rm -f .claude/reports/test-module-*.md
   rm -f .claude/reports/metrics-*.json
   rm -f .claude/logs/test-module-*.log
   ```

2. **Limpiar archivos temporales de Go**:
   ```bash
   go clean -testcache
   ```

3. **Registrar cancelación**:
   ```
   echo "Cancelled at $(date)" >> .claude/logs/cancelled-test-module.log
   ```

4. **Notificar al usuario**: Confirmar que no se realizó ningún cambio en el código

### Estados Finales Posibles

- **`completed`**: Tests ejecutados exitosamente, reportes generados
- **`failed`**: Error irrecuperable (módulo no existe, Go no instalado)
- **`cancelled`**: Cancelado por usuario durante ejecución
- **`partial`**: Tests ejecutados pero algunos fallaron (estado válido, no es error)

## Reglas Críticas

1. **No modificación de código**: Este comando solo ejecuta tests y genera reportes. Nunca debe modificar archivos `*_test.go` ni código fuente.

2. **Ejecución restringida**: Si durante el análisis se detecta que hay tests fallidos que requieren fixes, el comando debe:
   - ❌ NO intentar corregirlos
   - ✅ Generar reporte detallado con análisis
   - ✅ Recomendar usar `/plan-fix-tests` o `/debug-module`

3. **Aislamiento de tests**: Tests que modifiquen estado (BD, filesystem) deben:
   - Usar bases de datos de test separadas
   - Hacer cleanup en `tearDown()`
   - No afectar tests posteriores

4. **Transparencia en fallos**: Si un test falla:
   - Incluir stack trace completo
   - Proponer hipótesis de causa raíz
   - No especular sin evidencia

5. **Idempotencia**: Ejecutar el comando múltiples veces con los mismos parámetros debe producir resultados consistentes (siempre asumiendo que el código no cambió).

6. **Respeto de timeouts**: Si un test excede el timeout configurado:
   - Marcarlo como FAIL
   - Documentar el timeout en el reporte
   - Recomendar aumentar timeout o investigar test lento

7. **Validación de entorno**: Antes de ejecutar tests, verificar:
   - Go está instalado (`go version`)
   - Dependencias disponibles (`go mod verify`)
   - Variables de entorno configuradas
   - Bases de datos accesibles (si se requieren)

8. **Manejo de panics**: Si un test causa panic:
   - Capturar stack trace completo
   - Analizar si es panic conocido o desconocido
   - Recomendar uso de `/debug-module` con flags de race detector

## Manejo de Errores Comunes

### Error: Module Not Found

```
❌ Error: El módulo "./nonexistent/module" no existe
📁 Módulos disponibles:
  - ./internal/backend/user/...
  - ./controllers/Profile/...
  - ./services/Invoices/...
💡 Usa "go list ./..." para ver todos los paquetes
```

**Acción**: Listar módulos disponibles y terminar gracefully.

### Error: No Test Files

```
⚠️ Advertencia: No se encontraron archivos *_test.go en "./some/legacy/code"
📦 Este módulo no tiene tests aún
💡 Considera agregar tests o usar otro módulo
```

**Acción**: Advertir pero no fallar. Generar reporte indicando ausencia de tests.

### Error: Database Connection Failed

```
❌ Error: No se puede conectar a la base de datos de tests
🔍 Causa: DATA_BASE_* variables no configuradas o BD no accesible
💡 Solución:
  1. Verificar que MySQL está corriendo
  2. Configurar variables de entorno en .env.test
  3. Ejecutar: go run init.go --migrate=yes --env=test
```

**Acción**: Fallar temprano con instrucciones claras de resolución.

### Error: Compilation Error

```
❌ Error: El código no compila
📝 Errores de compilación:
  models/user.go:45: undefined: UserFactory
💡 Corrige los errores de compilación antes de ejecutar tests
```

**Acción**: Fallar con los errores de compilación específicos. No intentar ejecutar tests.

## Ejemplos de Uso

### Ejemplo 1: Tests Exitosos

```
$ test-module ./internal/backend/user/

📊 RESULTADO FINAL: ✅ PASS

Módulo: ./internal/backend/user/...
Tests: 42/42 pasaron (100%)
Cobertura: 87.3% (objetivo: 80%) ✅
Tiempo: 45.2s

✅ Todos los tests pasaron. No se requiere acción adicional.

📁 Artefactos generados:
  - Reporte: .claude/reports/test-module-20250120-153022.md
  - Métricas: .claude/reports/metrics-20250120-153022.json
```

### Ejemplo 2: Tests Fallidos

```
$ test-module ./controllers/Profile/

📊 RESULTADO FINAL: ❌ FAIL

Módulo: ./controllers/Profile/...
Tests: 15/18 pasaron (83.3%)
Cobertura: 72.1% (objetivo: 80%) ❌
Tiempo: 38.7s

🔍 Tests fallidos: 3
  - TestUpdateProfile (profile_controller_test.go:78)
    Error: "expected status 200, got 500"
    Hipótesis: Database constraint violation
  
  - TestDeleteProfile (profile_controller_test.go:120)
    Error: "timeout after 30s"
    Hipótesis: External API not responding
  
  - TestGetProfileHolidays (profile_controller_test.go:45)
    Panic: "invalid memory address"
    Hipótesis: Nil pointer dereference

💡 Next steps:
  1. Ver stack traces completas: .claude/reports/test-module-20250120-161245.md
  2. Diagnosticar con race detector: /debug-module ./controllers/Profile/ --race
  3. Si requieres fixes planificados: /plan-fix-tests ./controllers/Profile/
```

### Ejemplo 3: Tests Parciales con Cobertura Baja

```
$ test-module ./internal/backend/form/ --coverage=90

📊 RESULTADO FINAL: ⚠️ PARTIAL

Módulo: ./internal/backend/form/...
Tests: 28/30 pasaron (93.3%)
Cobertura: 82.4% (objetivo: 90%) ❌
Tiempo: 52.1s

🔍 Tests fallidos: 2
  - TestValidateForm (form_service_test.go:67)
  - TestUpdateFormFields (form_handler_test.go:134)

📉 Cobertura insuficiente:
  - domain/form.go: 95.2% ✅
  - application/service.go: 78.3% ⚠️
  - infrastructure/db/repository.go: 65.7% ❌

💡 Next steps:
  1. Revisar tests fallidos en reporte
  2. Aumentar cobertura en repository.go (branch coverage bajo)
  3. Considerar agregar tests para casos edge en service.go
  4. Para planificar mejoras: /plan-improve-coverage ./internal/backend/form/
```

## Acción del Usuario

Para ejecutar este comando, proporciona:

**Requerido**:
1. **Ruta del módulo**: ¿Qué módulo deseas testear? (ej: `./internal/backend/user/...`, `./controllers/Profile/...`)

**Opcional**:
2. **Cobertura mínima**: ¿Qué porcentaje de cobertura esperas? (por defecto: 80%)
3. **Verbosidad**: ¿Nivel de detalle? (standard, verbose)
4. **Timeout**: ¿Tiempo máximo de ejecución en segundos? (por defecto: 120)
5. **Ejecutar con race detector**: ¿Detectar condiciones de carrera? (por defecto: no)

**Ejemplo de solicitud válida**:
> "Ejecuta tests del módulo ./internal/backend/user/... con cobertura mínima 85%, timeout 180s, y verbosity verbose."

**Ejemplo rápido**:
> "Testea ./controllers/Profile/"