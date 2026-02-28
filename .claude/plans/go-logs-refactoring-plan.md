# Plan Integral de Refactorización: go_logs

**Proyecto:** go_logs (biblioteca de logging en Go)
**Fecha:** 2026-02-28
**Metodología:** TDD + GitFlow
**Duración estimada:** 3-4 semanas

---

## Resumen Ejecutivo

Este plan aborda los **16 problemas identificados** en la biblioteca go_logs mediante un enfoque estructurado en **6 fases incrementales** que priorizan la corrección de bugs críticos, mejoras de seguridad y optimizaciones de rendimiento. Cada fase sigue metodología **TDD** y **GitFlow**, garantizando que el código esté siempre testeado y que cada deliverable sea mergeable a la rama `develop`. El plan incluye estrategias de rollback, cobertura objetivo del 80%+, y transformación gradual hacia una arquitectura más robusta con soporte para contextos, formato de mensajes y mejor documentación.

---

## Matriz de Fases

| Fase | Nombre | Issues | Esfuerzo | Dependencias | Mergeable | Prioridad |
|------|--------|--------|----------|--------------|-----------|-----------|
| F1 | Corrección de Bugs Críticos | #1, #2, #3, #4, #5 | Alto | Ninguna | ✅ Sí | P0 |
| F2 | Hardening de Seguridad | #6, #7, #8 | Medio | F1 | ✅ Sí | P0 |
| F3 | Optimización de Rendimiento | #9, #10 | Medio | F1 | ✅ Sí | P1 |
| F4 | Tests y Cobertura | #12 | Alto | F1, F2, F3 | ✅ Sí | P1 |
| F5 | Mejoras de API y Formato | #11, #13, #15 | Medio | F4 | ✅ Sí | P2 |
| F6 | Limpieza y Documentación | #14, #16 | Bajo | F5 | ✅ Sí | P3 |

---

## Fase 1: Corrección de Bugs Críticos

**Branch:** `feature/fix-critical-bugs`
**Duración:** 2-3 días
**Prioridad:** P0 - BLOCKING

### Issues a Resolver

#### #1: Race condition en notificationSettings
- **Severidad:** CRÍTICA
- **Archivo:** config.go (líneas 24-30, 93-99)
- **Solución:** Implementar sync.RWMutex para proteger acceso concurrente
- **Tests:**
  - TestNotificationSettingsConcurrency()
  - Ejecutar `go test -race` para verificar

#### #2: Inicialización incorrecta del mapa de notificaciones
- **Severidad:** CRÍTICA
- **Archivo:** config.go (línea 24)
- **Solución:** Implementar inicialización lazy con sync.Once
- **Tests:**
  - TestInitNotCalled()
  - TestErrorLogWithoutInit()

#### #3: Variable global err compartida
- **Severidad:** CRÍTICA
- **Archivo:** config.go (línea 34)
- **Solución:** Eliminar variable global, usar returns locales
- **Tests:**
  - Verificar con `go vet` que no hay shadowing

#### #4: Error en ruta del archivo de log
- **Severidad:** ALTA
- **Archivo:** config.go (línea 107)
- **Solución:** Usar filepath.Join() para construcción portable
- **Tests:**
  - TestLogFilePathConstruction()

#### #5: Chequeo inútil de nil en logger
- **Severidad:** MEDIA
- **Archivo:** save.go (línea 20)
- **Solución:** Eliminar chequeo innecesario
- **Tests:**
  - TestNilLoggerCheck()

### Criterios de Aceptación
- ✅ Todos los tests de concurrencia pasan (`go test -race`)
- ✅ No hay data races detectadas
- ✅ Cobertura de tests > 85% para archivos modificados
- ✅ `go vet` no reporta warnings
- ✅ Benchmarks muestran < 5% degradación de performance

### Riesgos y Mitigación
- **Riesgo:** Cambios en estado global pueden afectar a consumers existentes
- **Mitigación:**
  1. Mantener backward compatibility en API pública
  2. Atributos internos pueden cambiar sin afectar API
  3. Crear tests de regresión antes de cambios
  4. Usar feature flag si es necesario

---

## Fase 2: Hardening de Seguridad

**Branch:** `feature/security-hardening`
**Duración:** 1-2 días
**Prioridad:** P0 - BLOCKING
**Dependencias:** Fase 1

### Issues a Resolver

#### #6: Permisos de archivo demasiado permisivos
- **Severidad:** ALTA
- **Archivo:** save.go
- **Solución:** Cambiar permisos de 0666 a 0600
- **Tests:**
  - TestLogFilePermissions()
  - Verificar con `ls -l` que permisos son correctos

#### #7: Sin validación de credenciales de Slack
- **Severidad:** ALTA
- **Archivo:** adapters/slack_notifier.go
- **Solución:** Validar token y channel en constructor
- **Tests:**
  - TestSlackNotifierInvalidToken()
  - TestSlackNotifierInvalidChannel()

#### #8: Error tipográfico en variable de entorno
- **Severidad:** MEDIA
- **Archivo:** config.go
- **Solución:** Soportar SLACK_CHANNEL_ID y SLACK_CHANEL_ID (backward compat)
- **Tests:**
  - TestSlackChannelNameTypo()
  - TestSlackChannelNameCorrect()

### Criterios de Aceptación
- ✅ Archivos de log tienen permisos 0600 por defecto
- ✅ Credenciales de Slack validadas antes de usar
- ✅ Ambas variables de entorno funcionan (correcta + typo)
- ✅ Warning emitido si se usa variable con typo
- ✅ Cobertura de tests > 90% para slack_notifier.go

---

## Fase 3: Optimización de Rendimiento

**Branch:** `feature/performance-optimization`
**Duración:** 1-2 días
**Prioridad:** P1 - ALTA
**Dependencias:** Fase 1

### Issues a Resolver

#### #9: Abrir/cerrar archivo en cada escritura
- **Severidad:** MEDIA
- **Archivo:** save.go (registerMessage())
- **Solución:** Implementar archivo persistente con bufio.Writer
- **Benchmarks:**
  - BenchmarkLogOperations() - antes vs después
  - Objetivo: > 50% mejora en throughput

#### #10: Lectura repetida de variables de entorno
- **Severidad:** MEDIA
- **Archivo:** config.go
- **Solución:** Cachear env vars en struct Config durante Init()
- **Tests:**
  - TestEnvVarsCached()
  - Verificar que no se llaman os.Getenv() en hot path

### Criterios de Aceptación
- ✅ Benchmark muestra > 50% mejora en throughput de logging
- ✅ No hay file descriptor leaks en tests de larga duración
- ✅ Variables de entorno cacheadas (no llamadas a os.Getenv en hot path)
- ✅ bufio.Writer se flushea correctamente en FatalLog()
- ✅ Cobertura de tests > 85%

---

## Fase 4: Tests y Cobertura

**Branch:** `feature/test-coverage`
**Duración:** 2-3 días
**Prioridad:** P1 - ALTA
**Dependencias:** Fases 1, 2, 3

### Issues a Resolver

#### #12: No hay tests unitarios (0% cobertura)
- **Severidad:** ALTA
- **Archivos:** Todos
- **Solución:** Crear suite completa de tests

### Archivos de Tests a Crear

- **config_test.go**
  - TestInitWithDefaults()
  - TestInitWithEnvVars()
  - TestNotificationSettingsEnabled()
  - TestNotificationSettingsConcurrency()
  - TestEnvVarCaching()

- **logs_test.go**
  - TestFatalLog()
  - TestErrorLog()
  - TestInfoLog()
  - TestSuccessLog()
  - TestLogConcurrent()

- **save_test.go**
  - TestRegisterMessageToFile()
  - TestRegisterMessageToSlack()
  - TestLogFileCreation()
  - TestLogFilePermissions()

- **slack_notifier_test.go**
  - TestNewSlackNotifierValid()
  - TestSlackNotifierSend()
  - TestSlackNotifierDisabled()

- **integration_test.go**
  - TestFullLoggingFlow()
  - TestConcurrentLogging()
  - TestHighVolumeLogging()

- **benchmark_test.go**
  - BenchmarkErrorLog()
  - BenchmarkConcurrentLogs()
  - BenchmarkFileWrite()

### Criterios de Aceptación
- ✅ Cobertura >= 80% (`go test -cover ./...`)
- ✅ Todos los tests pasan (`go test -v ./...`)
- ✅ No data races (`go test -race ./...`)
- ✅ Benchmarks ejecutan sin errores
- ✅ `go vet` no reporta warnings

---

## Fase 5: Mejoras de API y Formato

**Branch:** `feature/api-improvements`
**Duración:** 1-2 días
**Prioridad:** P2 - MEDIA
**Dependencias:** Fase 4

### Issues a Resolver

#### #11: Falta de soporte para context.Context
- **Severidad:** MEDIA
- **Archivos:** logs.go, save.go
- **Solución:** Agregar funciones *ctx (ErrorLogctx, InfoLogctx, etc.)
- **Tests:**
  - TestErrorLogctx()
  - TestContextCancellation()

#### #13: Falta de manejo de formato de mensajes
- **Severidad:** MEDIA
- **Archivo:** logs.go
- **Solución:** Agregar funciones *f (ErrorLogf, InfoLogf, etc.)
- **Tests:**
  - TestErrorLogf()
  - TestLogFormattingVerbs()

#### #15: Error tipográfico en log de Slack
- **Severidad:** BAJA
- **Archivo:** adapters/slack_notifier.go
- **Solución:** Corregir typo en mensaje

### Criterios de Aceptación
- ✅ Funciones *ctx implementadas y testadas
- ✅ Funciones *f implementadas y testadas
- ✅ Typos corregidos
- ✅ Cobertura de tests > 85% para logs.go
- ✅ Documentación godoc actualizada con ejemplos
- ✅ Benchmarks muestran overhead mínimo (< 10%)

---

## Fase 6: Limpieza y Documentación

**Branch:** `feature/cleanup-documentation`
**Duración:** 1 día
**Prioridad:** P3 - BAJA
**Dependencias:** Fase 5

### Issues a Resolver

#### #14: Variable no utilizada notificationLogWarning
- **Severidad:** BAJA
- **Archivo:** config.go
- **Solución:** Eliminar variable no utilizada

#### #16: Falta de documentación godoc
- **Severidad:** MEDIA
- **Archivos:** Todos
- **Solución:** Agregar comentarios godoc completos

### Criterios de Aceptación
- ✅ Sin variables no utilizadas (`go vet` clean)
- ✅ Documentación godoc completa para paquetes y exports
- ✅ `go doc` genera output completo y claro
- ✅ README.md actualizado con ejemplos de nuevas features
- ✅ CHANGELOG.md documentando todos los cambios
- ✅ `golint` no reporta faltantes de documentación

---

## Estrategia GitFlow

### Estructura de Branches

```bash
main              # Producción (releases solo)
└── develop       # Desarrollo principal
    ├── feature/fix-critical-bugs         # Fase 1
    ├── feature/security-hardening        # Fase 2
    ├── feature/performance-optimization  # Fase 3
    ├── feature/test-coverage             # Fase 4
    ├── feature/api-improvements          # Fase 5
    └── feature/cleanup-documentation     # Fase 6
```

### Flujo de Trabajo por Fase

```bash
# 1. Crear feature branch
git checkout develop
git pull origin develop
git checkout -b feature/NOMBRE-FASE

# 2. Desarrollo y commits
git add .
git commit -m "feat: descripción del cambio"

# 3. Merge a develop
git checkout develop
git pull origin develop
git merge feature/NOMBRE-FASE --no-ff -m "Merge feature/NOMBRE-FASE"
git push origin develop

# 4. Borrar feature branch (opcional)
git branch -d feature/NOMBRE-FASE
```

### Política de Commits

- `feat:` Nueva funcionalidad
- `fix:` Bug fix
- `refactor:` Refactorización
- `test:` Agregar/modificar tests
- `docs:` Cambios de documentación
- `perf:` Mejoras de performance
- `security:` Mejoras de seguridad

---

## Estrategia TDD

### Ciclo Red-Green-Refactor

```
┌─────────────────────────────────────────┐
│  RED: Escribir test que falle           │
│  ├─ Crear archivo *_test.go            │
│  ├─ Escribir test case específico      │
│  └─ Ejecutar: go test -v               │
│                                         │
│  GREEN: Escribir código mínimo         │
│  ├─ Implementar lógica mínima          │
│  ├─ NO escribir código extra           │
│  └─ Ejecutar: go test -v               │
│                                         │
│  REFACTOR: Limpiar código              │
│  ├─ Eliminar duplicación               │
│  ├─ Mejorar nombres                    │
│  └─ Ejecutar: go test -race -cover     │
└─────────────────────────────────────────┘
```

### Cobertura Objetivo por Fase

| Fase | Cobertura Mínima | Comando |
|------|------------------|---------|
| Fase 1 | 85% (config.go, save.go) | `go test -cover ./...` |
| Fase 2 | 90% (adapters/) | `go test -cover ./adapters/...` |
| Fase 3 | 85% (save.go, config.go) | `go test -cover ./...` |
| Fase 4 | 80% (global) | `go test -cover ./...` |
| Fase 5 | 85% (logs.go) | `go test -cover ./...` |
| Fase 6 | N/A | N/A |

---

## Cronograma Estimado

### Sprint 1 (Semana 1)
- **Día 1-3:** Fase 1 - Bugs Críticos
- **Día 4-5:** Fase 2 - Seguridad
- **Meta:** develop estable sin bugs críticos

### Sprint 2 (Semana 2)
- **Día 1-2:** Fase 3 - Performance
- **Día 3-5:** Fase 4 - Tests (parte 1)
- **Meta:** Performance optimizado

### Sprint 3 (Semana 3)
- **Día 1-2:** Fase 4 - Tests (completar)
- **Día 3-4:** Fase 5 - API
- **Día 5:** Fase 6 - Documentación
- **Meta:** Cobertura 80%+, API mejorada

### Sprint 4 (Semana 4)
- **Día 1-2:** Review y pruebas finales
- **Día 3:** Release v2.0.0
- **Día 4-5:** Buffer y post-release

```
Semana 1: [F1====][F2==]
Semana 2: [F3==][F4====]
Semana 3: [F4==][F5==][F6=][Review]
Semana 4: [Release=][Buffer=]
```

---

## Métricas de Éxito

### Métricas Técnicas

| Métrica | Línea Base | Objetivo | Medición |
|---------|------------|----------|----------|
| Cobertura de Tests | 0% | 80%+ | `go test -cover` |
| Data Races | Detectadas | 0 | `go test -race` |
| go vet Warnings | Varias | 0 | `go vet ./...` |
| Throughput Logging | Baseline | +50% | Benchmarks |
| Permisos Archivos | 0666 | 0600 | `ls -l` |
| Env Vars Leídas | Por llamada | 1 vez | Code review |

### Métricas de Calidad

| Métrica | Objetivo |
|---------|----------|
| Bugs Críticos Resueltos | 5/5 (100%) |
| Issues Seguridad Resueltos | 3/3 (100%) |
| Mejoras Performance | 2/2 (100%) |
| Mejoras Diseño | 6/6 (100%) |
| Documentación Completa | 100% |

---

## Estrategia de Rollback

### Por Fase

**Fase 1:** Revertir commits si data races persisten o performance degrada > 20%
**Fase 2:** Mantener compatibilidad con typo SLACK_CHANEL_ID
**Fase 3:** Feature flag USE_PERSISTENT_FILE=false si hay file leaks
**Fase 4:** N/A (tests no rompen nada)
**Fase 5:** Mantener API original + nuevas funciones
**Fase 6:** Revertir cambios si documentación incorrecta

### Rollback General

```bash
# Si algo sale mal:
git checkout main
git reset --hard <pre-release-commit>
git tag -a v2.0.1 -m "Rollback v2.0.0"

# O revertir merge
git checkout develop
git revert -m 1 <merge-commit-hash>
```

---

## Conclusión

Este plan proporciona una **ruta estructurada e incremental** para resolver los 16 problemas identificados en go_logs. La priorización por severidad, enfoque TDD y estrategia GitFlow garantizan entregas incrementales con baja probabilidad de regresiones.

**Estimación total:** 3-4 semanas de desarrollo, entregando una biblioteca robusta, testeada, documentada y con mejoras significativas de seguridad y performance.

---

**Plan generado por:** planning-agent
**Orquestado por:** Orchestrator
**Fecha:** 2026-02-28
