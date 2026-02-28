---
name: debug-master
description: Esta skill debe usarse cuando el usuario necesite investigar y resolver errores complejos, analizar stack traces, identificar race conditions, diagnosticar memory leaks, o proponer soluciones a bugs en código Go. Se activa con peticiones como "investiga este error", "¿por qué falla este código?", "tengo un panic en producción", o "ayuda a debuggear este problema".
license: Complete terms in LICENSE.txt
version: 1.0.0
author: Reverence Hotels Development Team
category: base
tags: [debugging, go, error-analysis, troubleshooting, performance]---

# Debug Master

Skill especializada en diagnóstico y resolución de problemas complejos en el proyecto Reverence Hotels API. Proporciona conocimiento profundo sobre patrones de error específicos del proyecto, arquitectura híbrida MVC/Clean Architecture, multi-database, y Common gotchas con GORM v2, Echo framework, y integraciones externas (SII, Porta Sigma, PMS).

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:
- Se produzca un error o panic en el código Go del proyecto
- Un test falle de forma inesperada
- Haya problemas de rendimiento (memory leaks, slow queries)
- Se detecte una race condition en código concurrente
- Una integración externa (SII, Porta Sigma, Slack) falle
- Haya problemas con las 3 conexiones a bases de datos MySQL
- El middleware de autorización no funcione como esperado
- Los cron jobs no se ejecuten según el entorno (local/pre/pro)

Triggers comunes:
- "Investiga este error: [stack trace]"
- "¿Por qué falla el endpoint /api/v1/users?"
- "Tengo un panic en producción, ayuda"
- "Este test pasa en local pero falla en CI"
- "La conexión a la base de datos SII falla"
- "El middleware de autorización rechaza peticiones válidas"
- "Memory leak en el servicio de notificaciones"

## Workflow Principal

### 1. Análisis Inicial del Problema

Antes de profundizar:

1. **Identificar el tipo de error**:
   - Panic/runtime error
   - Error de lógica de negocio
   - Error de infraestructura (DB, red)
   - Error de integración externa
   - Race condition o concurrencia

2. **Recopilar contexto**:
   - Stack trace completo
   - Logs alrededor del error
   - Estado del sistema (entorno, carga)
   - Configuración relevante (ENV, variables)

3. **Localizar en la arquitectura**:
   - ¿Es código legacy MVC o Clean Architecture?
   - ¿Qué módulo está afectado?
   - ¿Qué base de datos está involucrada?

### 2. Diagnóstico Específico del Proyecto

Aplicar conocimiento del proyecto Reverence Hotels:

#### 2.1 Identificar Patrón de Error

Consultar `references/common_errors.md` para errores típicos del proyecto:
- Errores de migración GORM v1→v2
- Problemas con foreign keys
- Conexiones a las 3 bases de datos
- Authorization middleware path matching
- Cron jobs por entorno

#### 2.2 Análisis de Stack Trace

Descomponer el stack trace identificando:
1. **Capa afectada**: Controller/Handler, Service, Repository, Model
2. **Framework**: Echo middleware vs aplicación
3. **Librería**: GORM, database driver, librería externa
4. **Goroutine**: Si es una race condition

#### 2.3 Verificar Configuración

Validar aspectos comunes que causan problemas:
- Variable `ENV` (local/pre/pro) para cron jobs
- Credenciales de las 3 BD (DATA_BASE_*, DATA_BASE_*_SII, DATA_BASE_*_PRODUCTS)
- Tokens de integración (SLACK_TOKEN, API_PORTA_SIGMA_*)
- Configuración de GORM (MaxOpenConns, ConnMaxLifetime)

### 3. Ejecución de Scripts de Diagnóstico

Usar scripts según el tipo de problema:

#### Race Conditions

```bash
go test -race ./path/to/module/...
```

#### Memory Leaks

```bash
go test -memprofile=mem.prof ./path/to/module/
go tool pprof mem.prof
```

#### CPU Profiling

```bash
go test -cpuprofile=cpu.prof ./path/to/module/
go tool pprof cpu.prof
```

#### Análisis de dependencias

```bash
go mod graph | grep <problematic-package>
```

### 4. Propuesta de Solución

Generar propuesta structured:

1. **Raíz del problema**: Explicación clara del por qué
2. **Solución propuesta**: Código o configuración específica
3. **Validación**: Cómo verificar que funciona
4. **Prevención**: Cómo evitar recurrencia

### 5. Validación de la Solución

Verificar que:

- [ ] El error está reproducido y entendido
- [ ] La solución propuesta es específica al proyecto
- [ ] Se considera el impacto en módulos relacionados
- [ ] Se respetan patrones de arquitectura (MVC vs Clean)
- [ ] No se introducen nuevos problemas
- [ ] Los tests existentes siguen pasando
- [ ] Se agregan tests si el error no estaba cubierto

## Recursos de la Skill

### Scripts (`scripts/`)

#### `scripts/analyze_stack_trace.sh`

**Propósito**: Extraer información estructurada de un stack trace de Go

**Uso**:

```bash
bash scripts/analyze_stack_trace.sh <stack-trace-file>
```

**Parámetros**:

- `stack-trace-file`: Archivo de texto con el stack trace completo

**Output**: Análisis estructurado con:
- Función donde ocurrió el error
- Archivo y línea
- Secuencia de llamadas
- Goroutine ID si aplica

**Ejemplo**:

```bash
bash scripts/analyze_stack_trace.sh error.log
```

---

#### `scripts/check_race_conditions.sh`

**Propósito**: Ejecutar tests con detector de race conditions en módulos específicos

**Uso**:

```bash
bash scripts/check_race_conditions.sh <module-path>
```

**Parámetros**:

- `module-path`: Ruta del módulo a testear (ej: ./internal/backend/user/)

**Output**: Reporte de races detectadas con goroutines implicadas

**Ejemplo**:

```bash
bash scripts/check_race_conditions.sh ./internal/notification/
```

---

#### `scripts/analyze_memory.sh`

**Propósito**: Generar y analizar profile de memoria para detectar leaks

**Uso**:

```bash
bash scripts/analyze_memory.sh <module-path> [test-name]
```

**Parámetros**:

- `module-path`: Ruta del módulo a analizar
- `test-name`: Nombre específico del test (opcional, corre todos si se omite)

**Output**: Reporte de allocations y objetos retidos en memoria

**Ejemplo**:

```bash
bash scripts/analyze_memory.sh ./services/Invoices/ TestSendInvoice
```

---

#### `scripts/verify_db_connections.sh`

**Propósito**: Verificar conectividad a las 3 bases de datos del proyecto

**Uso**:

```bash
bash scripts/verify_db_connections.sh
```

**Output**: Estado de conexión para:
- Base de datos Principal (sensesho_api)
- Base de datos SII (reverence_sii)
- Base de datos Products (economato)

**Ejemplo**:

```bash
bash scripts/verify_db_connections.sh
# Output:
# ✓ Principal DB: Connected (127.0.0.1:3306)
# ✗ SII DB: Connection refused
# ✓ Products DB: Connected (127.0.0.1:3306)
```

---

#### `scripts/check_env_config.sh`

**Propósito**: Validar que todas las variables de entorno requeridas están configuradas

**Uso**:

```bash
bash scripts/check_env_config.sh <environment>
```

**Parámetros**:

- `environment`: Entorno a validar (local, pre, pro)

**Output**: Reporte de variables faltantes o con valores inválidos

**Ejemplo**:

```bash
bash scripts/check_env_config.sh pro
# Output:
# ✗ Missing: SLACK_TOKEN
# ✗ Invalid: DATA_BASE_PORT (must be numeric)
# ✓ ENV configuration complete
```

---

### Referencias (`references/`)

#### `references/common_errors.md`

**Contenido**: Catálogo de errores frecuentes específicos del proyecto Reverence Hotels, organizados por categoría:
- Errores de GORM v2 migration
- Problemas con las 3 BD
- Errores del middleware de autorización
- Issues con cron jobs y entornos
- Errores de integraciones (SII, Porta Sigma)

**Cuándo consultar**: Al identificar un error para verificar si es un patrón conocido

**Estructura**:
```
## GORM Migration Errors
### Error: "Error 1213: Deadlock found when trying to get lock"
**Cause**: Foreign key recreation during migration
**Solution**: [...]

## Authorization Errors
### Error: "403 Forbidden on valid user request"
**Cause**: Form.PathAPI doesn't match request path
**Solution**: [...]
```

**Búsqueda rápida**:

```bash
# Buscar por mensaje de error
grep -i "Error 1213" references/common_errors.md

# Buscar por categoría
grep -A 20 "## Authorization Errors" references/common_errors.md
```

---

#### `references/architecture_gotchas.md`

**Contenido**: Lista de "gotchas" específicos de la arquitectura híbrida del proyecto

**Cuándo consultar**: Cuando el error parece relacionado con la estructura del proyecto

**Secciones clave**:
- Diferencias entre MVC y Clean Architecture modules
- Cuándo usar cada patrón
- Common mistakes en el boundary entre ambos
- Interacciones con EventBus y notificaciones

**Ejemplo de contenido**:
```markdown
## Gotcha: Direct GORM access in Clean Architecture modules

**Problem**: Using `db.Where()` directly in `internal/` handlers instead of repositories

**Impact**: Bypasses domain events, breaks architecture principles

**Signs in errors**:
- "no such table" (wrong DB connection used)
- Tests passing in dev, failing in CI (different DB setup)

**Solution**: Always use repositories from `infrastructure/db/`
```

---

#### `references/debugging_by_layer.md`

**Contenido**: Estrategias de debugging según la capa de la arquitectura afectada

**Cuándo consultar**: Para enfocar el análisis en la capa correcta

**Estructura**:
```
## Presentation Layer (Controllers/Handlers)
- [ ] Echo middleware chain
- [ ] JWT validation
- [ ] Request/Response binding
- [ ] Path matching

## Business Logic Layer (Services/Application Services)
- [ ] Domain events published?
- [ ] Transaction boundaries
- [ ] Error handling consistency

## Data Access Layer (Repositories/Models)
- [ ] Correct DB connection (db vs dbSII vs dbProducts)
- [ ] GORM queries optimized?
- [ ] N+1 queries
- [ ] Connection pool exhaustion

## Infrastructure Layer
- [ ] External service availability
- [ ] Network timeouts
- [ ] API credentials valid
```

---

#### `references/integration_errors.md`

**Contenido**: Errores específicos de integraciones externas del proyecto

**Cuándo consultar**: Si el error involucra servicios externos

**Secciones**:
- **SII (AEAT)**: XML generation, certificate signing, API communication
- **Porta Sigma**: Transaction creation, signature status checking
- **Slack**: Webhook posting, rate limiting
- **PMS**: API key authentication, data synchronization

**Ejemplo**:
```markdown
## SII Integration: Certificate Signing Error

**Error**: `x509: certificate signed by unknown authority`

**Cause**: 
- .pem certificate file missing or corrupted
- Expired certificate
- Wrong certificate file (test vs production)

**Diagnosis steps**:
1. Verify certificate file exists
2. Check certificate expiration: `openssl x509 -in cert.pem -noout -dates`
3. Validate certificate chain

**Solution**:
[...]
```

---

#### `references/middleware_authorization_debug.md`

**Contenido**: Guía detallada del sistema de autorización del proyecto para debugging

**Cuándo consultar**: Errores 401/403 o problemas de permisos

**Contenido**:
- Cómo funciona el path-based authorization
- Mapping de Level → LevelPrivileges → Form → PathAPI
- Common pitfalls en PathAPI matching
- Cómo debugguear permisos denegados

**Snippet de diagnóstico**:
```markdown
## Debugging 403 Errors

### Step 1: Verify User Level
```sql
SELECT ID, Level FROM users WHERE ID = <user_id>;
```

### Step 2: Check Level Privileges
```sql
SELECT * FROM level_privileges 
WHERE LevelID = <level_id>
AND FormID IN (
  SELECT ID FROM forms WHERE PathAPI LIKE '%<request_path>%'
);
```

### Step 3: Validate PathAPI Format
- PathAPI uses `|` for multiple paths
- Example: `"profile|profile-holidays|profile-absences"`
- Request path must match one of the pipes
```

---

#### `references/testing_troubleshooting.md`

**Contenido**: Problemas comunes en tests del proyecto y cómo resolverlos

**Cuándo consultar**: Tests que fallan inesperadamente

**Secciones**:
- Test setup para 3 databases
- Mocking external services (SII, Porta Sigma)
- Test isolation y cleanup
- Tests que pasan en local pero fallan en CI
- Testing Clean Architecture modules

---

### Assets (`assets/`)

#### `assets/debug_checklist.md`

**Tipo**: Checklist interactivo para diagnóstico sistemático

**Uso**: Usar como guía durante investigación para no pasar por alto pasos comunes

**Secciones**:
- [ ] Recopilación de información inicial
- [ ] Análisis de stack trace
- [ ] Verificación de configuración
- [ ] Reproducción del error
- [ ] Aislamiento de la causa
- [ ] Propuesta de solución
- [ ] Validación
- [ ] Prevención futura

---

#### `assets/error_report_template.md`

**Tipo**: Template para reportes estructurados de errores

**Uso**: Generar reportes consistentes de bugs o problemas

**Modificaciones**: Personalizar con información específica del error encontrado

```markdown
# Error Report: [Título descriptivo]

## Resumen Ejecutivo
[2-3 frases describiendo el error y su impacto]

## Detalles del Error

### Información Básica
- **Fecha/Hora**: [timestamp]
- **Entorno**: [local/pre/pro]
- **Módulo afectado**: [nombre del módulo]
- **Severidad**: [crítica/alta/media/baja]

### Stack Trace
\`\`\`
[stack trace completo]
\`\`\`

### Contexto
- **Último cambio**: [qué se modificó recientemente]
- **Logs relevantes**: [extracto de logs]
- **Configuración**: [variables de entorno relevantes]

## Análisis

### Raíz del Problema
[Explicación técnica del por qué ocurre]

### Impacto
- Qué funcionalidades están afectadas
- Qué usuarios/roles impacta
- Datos comprometidos (si aplica)

## Solución Propuesta

### Cambios Necesarios
- [ ] [Archivo]: [descripción del cambio]
- [ ] [Archivo]: [descripción del cambio]

### Testing Plan
- [ ] Test que reproduce el error
- [ ] Test que valida la solución
- [ ] Tests de regresión

### Validación
- [ ] Verificado en [entorno]
- [ ] Tests pasan: [porcentaje]%
- [ ] Sin efectos colaterales

## Prevención
- Qué cambios hacer para evitar recurrencia
- Mejoras en testing o monitoreo
```

---

## Ejemplos de Uso

### Ejemplo 1: Error en producción con integración SII

**Input del usuario**:
> "Tenemos este error en producción en el servicio de facturación SII: 'x509: certificate signed by unknown authority'. ¿Qué puede ser?"

**Proceso**:

1. Identificar que es un error de integración externa (SII)
2. Consultar `references/integration_errors.md` sección "SII Certificate Errors"
3. Ejecutar verificación de certificado:

```bash
openssl x509 -in configuration/cert_sii.pem -noout -dates
```

4. Determinar causa (certificado expirado, archivo equivocado, etc.)
5. Proporcionar solución específica según diagnóstico

**Output esperado**:

```
## Diagnóstico

**Error**: `x509: certificate signed by unknown authority`

**Causa identificada**: El certificado SII ha expirado o es inválido

**Verificación ejecutada**:
\`\`\`bash
openssl x509 -in configuration/cert_sii.pem -noout -dates
notBefore=Jan 15 12:00:00 2024 GMT
notAfter=Jan 15 12:00:00 2025 GMT  ← EXPIRADO
\`\`\`

**Impacto**: 
- No se pueden enviar facturas al SII
- Las facturas creadas se quedan en estado "pending"
- Potencial incumplimiento de plazos legales

## Solución Propuesta

1. Obtener nuevo certificado vigente de la AEAT
2. Reemplazar archivo `configuration/cert_sii.pem`
3. Verificar nuevo certificado:
\`\`\`bash
openssl x509 -in configuration/cert_sii.pem -noout -dates
openssl x509 -in configuration/cert_sii.pem -noout -subject
\`\`\`
4. Reiniciar el servicio: `go run init.go --migrate=yes --pro=yes`

## Validación

- [ ] Certificado nuevo tiene fecha de vigencia futura
- [ ] Subject del certificado corresponde a la empresa
- [ ] Test de envío de factura es exitoso
- [ ] No hay errores en logs del servicio SII

## Prevención

- Agregar alerta de monitoreo 30 días antes de expiración
- Documentar proceso de renovación en runbook
- Agregar validación de certificación al startup del servidor
```

---

### Ejemplo 2: Race condition en handler de Clean Architecture

**Input del usuario**:
> "El test TestCreateUserConcurrent está fallando intermitentemente con un data race. Ayuda a identificar el problema."

**Proceso**:

1. Ejecutar test con detector de races:

```bash
go test -race ./internal/backend/user/... -run TestCreateUserConcurrent
```

2. Analizar output del race detector
3. Identificar variables compartidas sin protección
4. Consultar `references/common_errors.md` sobre races en Clean Architecture
5. Proponer solución con sincronización adecuada

**Output esperado**:

```
## Análisis de Data Race

**Test afectado**: `TestCreateUserConcurrent`

**Race detectada**:
\`\`\`
WARNING: DATA RACE
Read at 0x00c0000b4010 by goroutine 8:
  internal/backend/user/application/service.go:127 user.(*service).Create()
  [... stack trace ...]

Previous write at 0x00c0000b4010 by goroutine 7:
  internal/backend/user/application/service.go:127 user.(*service).Create()
  [... stack trace ...]
\`\`\`

**Línea problemática**: `service.go:127`
\`\`\`go
counter := s.counter  ← Race aquí
s.counter = counter + 1
\`\`\`

**Causa**: Campo `counter` compartido entre goroutines sin sincronización (mutex o atomic)

## Solución Propuesta

Reemplazar contador no-safe por versión thread-safe:

**Opción A: sync.Mutex**
\`\`\`go
type service struct {
    counter int
    mu      sync.Mutex
    // ... otros campos
}

func (s *service) incrementCounter() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.counter++
}
\`\`\`

**Opción B: sync/atomic (preferido para contadores simples)**
\`\`\`go
type service struct {
    counter int64
    // ... otros campos
}

func (s *service) incrementCounter() {
    atomic.AddInt64(&s.counter, 1)
}
\`\`\`

**Recomendación**: Usar **Opción B** (atomic) porque es más eficiente para contadores numéricos simples.

## Validación

1. Aplicar el cambio en `internal/backend/user/application/service.go`
2. Ejecutar test repetidamente para confirmar que no hay races:

\`\`\`bash
# Ejecutar 50 veces para asegurar que es determinista
for i in {1..50}; do
  go test -race ./internal/backend/user/... -run TestCreateUserConcurrent || exit 1
done
echo "All runs passed - no race condition detected"
\`\`\`

## Prevención

- Agregar check `-race` al pipeline de CI/CD
- Documentar práctica: "Siempre usar atomic o mutex para estado compartido"
- Considerar hacer `service.counter` privado y acceder solo vía método thread-safe
```

---

### Ejemplo 3: Memory leak en servicio de notificaciones

**Input del usuario**:
> "El servicio de notificaciones está consumiendo cada vez más memoria y eventualmente hace OOM. ¿Cómo lo investigo?"

**Proceso**:

1. Ejecutar script de análisis de memoria:

```bash
bash scripts/analyze_memory.sh ./internal/notification/ TestSendBatchNotifications
```

2. Analizar output de pprof para identificar objetos retiros
3. Buscar patrones en código que causen retención de memoria
4. Consultar `references/common_errors.md` sobre memory leaks en goroutines

**Output esperado**:

```
## Diagnóstico de Memory Leak

**Síntoma**: Crecimiento continuo de memoria hasta OOM kill

**Profile de memoria**:
\`\`\`bash
go tool pprof mem.prof

(pprof) top10
Showing nodes accounting for 850MB, 98% of 865MB total
Dropped 25 nodes (cum <= 4.32MB)
      flat  flat%   sum%        cum   cum%
     450MB 52.02% 52.02%      450MB 52.02%  github.com/gorilla/websocket.Conn
     300MB 34.68% 86.70%      750MB 86.70%  internal/notification/notifier.slackNotifier
      50MB  5.78% 92.48%      800MB 92.48%  internal/event/eventBus.subscribers
\`\`\`

**Objetos problemáticos identificados**:
1. **websocket.Conn**: 450MB sin liberarse (52% del total)
2. **slackNotifier**: 300MB acumulado (34%)
3. **eventBus.subscribers**: 50MB creciendo continuamente (6%)

**Causa raíz**: 
Goroutines que publican eventos no se limpian del `eventBus.subscribers` después de completar, y conexiones WebSocket no se cierran en `slackNotifier`.

## Solución Propuesta

### 1. Cleanup de EventBus subscribers

**Archivo**: `internal/event/event_bus.go`

**Agregar**:
\`\`\`go
func (eb *eventBus) Cleanup() {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    // Remover subscribers cerrados
    for topic, subscribers := range eb.subscribers {
        active := make([]chan Event, 0, len(subscribers))
        for _, ch := range subscribers {
            // Si canal está cerrado, len() retorna 0 al intentar enviar
            select {
            case ch <- Event{}: // Test send
                active = append(active, ch)
            default:
                close(ch) // Cleanup
            }
        }
        eb.subscribers[topic] = active
    }
}
\`\`\`

**Llamar periodicamente**:
\`\`\`go
// Agregar al cron job de production
task.Add("@every 5m", func() {
    eventBus.Cleanup()
})
\`\`\`

### 2. Cerrar conexiones WebSocket en SlackNotifier

**Archivo**: `internal/notification/infrastructure/http/slack_notifier.go`

**Agregar defer**:
\`\`\`go
func (n *slackNotifier) Notify(ctx context.Context, notification Notification) error {
    conn, err := websocket.Dial(n.url, "", "http://localhost")
    if err != nil {
        return err
    }
    defer conn.Close()  ← AGREGAR ESTO
    
    // ... resto del código
}
\`\`\`

### 3. Verificar con profile

Después de cambios:
\`\`\`bash
go test -memprofile=mem_new.prof ./internal/notification/...
go tool pprof mem_new.mem.prof

# Comparar con baseline
go tool pprof -base mem.prof mem_new.prof
\`\`\`

## Validación

- [ ] Profile post-fix muestra reducción de retención de memoria
- [ ] No hay goroutines泄漏 (verificar con `runtime.NumGoroutine()`)
- [ ] Servicio corre 24h sin OOM en staging
- [ ] Metricas de memoria se estabilizan (no crecimiento lineal)

## Monitoreo

Agregar métricas para detección temprana:
\`\`\`go
import "runtime"

var memStats runtime.MemStats

func monitorMemory() {
    for {
        runtime.ReadMemStats(&memStats)
        log.Printf("HeapAlloc: %dMB, HeapInuse: %dMB, NumGoroutine: %d",
            memStats.HeapAlloc/1024/1024,
            memStats.HeapInuse/1024/1024,
            runtime.NumGoroutine())
        time.Sleep(5 * time.Minute)
    }
}
\`\`\`

Alertar si:
- HeapInuse crece > 100MB en 1 hora
- NumGoroutine > 1000
```

---

## Presentación de Resultados

Al completar el debugging:

1. **Resumen ejecutivo**: Una frase clara con el diagnóstico
   - Ejemplo: "Error causado por [causa raíz]. Solución: [acción específica]."

2. **Estructura del reporte**:
   - Diagnóstico (qué está mal)
   - Análisis (por qué está mal)
   - Solución propuesta (cómo arreglarlo)
   - Validación (cómo verificar que funciona)
   - Prevención (cómo evitar que recurra)

3. **Incluir siempre**:
   - Stack trace completo (si aplica)
   - Líneas de código específicas (archivo:linea)
   - Comandos ejecutados para diagnóstico
   - Output de herramientas (pprof, race detector, etc.)

4. **Formato de código**:
   - Mostrar código antes (problemático)
   - Mostrar código después (corregido)
   - Explicar los cambios línea por línea

5. **No incluir**:
   - Especulación sin evidencia
   - Soluciones genéricas no aplicables al proyecto
   - Información irrelevante que distrae

## Troubleshooting de la Skill

### Problema: Stack trace incompleto o truncado

**Síntoma**: Stack trace no muestra toda la secuencia de llamadas

**Causa**: Logs muy largos truncados por sistema de logging

**Solución**:

1. Revisar logs crudos (sin formatear)
2. Aumentar nivel de log a DEBUG
3. Usar `dlv` debugger para attach al proceso:

```bash
dlv attach <pid>
```

---

### Problema: Error no reproducible en local

**Síntoma**: Error ocurre solo en producción/pre-production

**Causa**: Diferencias en configuración, datos, o carga

**Solución**:

1. Consultar `references/testing_troubleshooting.md` sección "Environment-specific bugs"
2. Revisar differences en:
   - Variables de entorno (`scripts/check_env_config.sh`)
   - Datos de producción vs seed data
   - Volumen de carga/concurrencia
3. Agregar logging adicional en producción
4. Usar remote debugging si es posible

---

### Problema: Race condition intermitente

**Síntoma**: Test falla aleatoriamente, no siempre reproduce el error

**Causa**: Timing de goroutines no determinístico

**Solución**:

1. Ejecutar test muchas veces con `-race`:

```bash
for i in {1..100}; do
  go test -race ./... || exit 1
done
```

2. Si aún no reproduce, agregar `runtime.Gosched()` para exponer race
3. Usar `-parallel=1` para aislar si es causado por concurrencia de tests

---

### Problema: Memory leak difícil de identificar

**Síntoma**: pprof muestra muchos allocations pero no claro qué los retiene

**Solución**:

1. Usar `go tool pprof -alloc_space` en vez de `-inuse_space`
2. Ver gráfico de llamadas:

```bash
go tool pprof -web mem.prof
```

3. Buscar "trabajo pendiente" (goroutines bloqueadas, conexiones no cerradas)
4. Revisar patrones en `references/common_errors.md` sección "Memory Leaks"

---

### Problema: Error en integración externa (SII, Slack, etc.)

**Síntoma**: Error al comunicar con servicio externo, no claro si es problema del código o del servicio

**Solución**:

1. Consultar `references/integration_errors.md` para el servicio específico
2. Verificar conectividad de red:

```bash
curl -v <external-service-url>
```

3. Verificar credenciales/tokens:

```bash
# Ejemplo para Slack
curl -X POST https://slack.com/api/auth.test \
  -H "Authorization: Bearer $SLACK_TOKEN"
```

4. Revisar status del servicio externo (si está disponible)
5. Validar que el certificado (SII) es válido y no expiró

---

### Problema: Test de Clean Architecture falla por DB

**Síntoma**: Test en módulo `internal/` falla con error de base de datos

**Causa**: Test no está mockeando el repository correctamente

**Solución**:

1. Verificar que el test está en `infrastructure/db/` y NO en `application/`
2. Asegurar que usa mock de repository, no DB real
3. Consultar `references/testing_troubleshooting.md` sección "Testing Clean Architecture"
4. Pattern correcto:

```go
// ✓ Correcto - Test en infrastructure/db con mock DB
func TestGormRepository_Create(t *testing.T) {
    db := setupTestDB() // DB en memoria o mock
    repo := NewGormRepository(db)
    // ...
}

// ✗ Incorrecto - Test en application con DB real
func TestService_Create(t *testing.T) {
    db := getRealDB() // NO HACER ESTO
    service := NewService(repo)
    // ...
}
```

---

## Consideraciones Especiales del Proyecto

### Arquitectura Híbrida

**CRÍTICO**: El proyecto tiene DOS patrones arquitectónicos coexistiendo:

1. **MVC Legacy** (`controllers/`, `models/`, `services/`)
   - Stabilized modules
   - Acceso directo a GORM
   - Sin repository abstraction

2. **Clean Architecture** (`internal/{module}/`)
   - New modules
   - Repository pattern obligatorio
   - Domain events obligatorios

**Gotcha**: NO mezclar patrones. Si estás en módulo `internal/`, seguir Clean Architecture estrictamente.

---

### Multi-Database Architecture

El proyecto usa **3 bases de datos MySQL diferentes**:

1. **Principal** (`sensesho_api`): Users, profiles, HR
2. **SII** (`reverence_sii`): Electronic invoicing
3. **Products** (`economato`): Articles, providers, orders

**Gotchas comunes**:

- Usar `db` cuando se necesita `dbSII` o viceversa
- Foreign keys entre BDs (NO son posibles, misma instancia pero diferente DB)
- Connection pool exhaustion si se abren muchas transacciones

**Verificar**: `scripts/verify_db_connections.sh`

---

### Authorization por Path

El middleware de autorización es **path-based**, no controller-based:

```
User → Level → LevelPrivileges → Form → PathAPI
```

**Gotcha**: `Form.PathAPI` puede contener múltiples rutas separadas por `|`

Ejemplo: `"profile|profile-holidays|profile-absences"`

Si el request path NO está en `PathAPI`, se retorna 403 aunque el user tenga el Level correcto.

**Debugging**: Consultar `references/middleware_authorization_debug.md`

---

### Cron Jobs por Entorno

Los cron jobs se activan según la variable `ENV`:

- **local**: No cron jobs (desactivados)
- **pre**: Subset limitado (SII invoices, emails)
- **pro**: Todos los cron jobs (SII, signatures, notifications, emails)

**Gotcha**: Tests que asumen que un cron job se ejecutó fallan en local porque `ENV=local`.

**Verificar**: `scripts/check_env_config.sh`

---

### GORM v2 Migration Issues

El proyecto migró de GORM v1 a v2, lo que causó problemas con foreign keys:

- Migration code drops FKs, runs AutoMigrate, then recreates
- Views are recreated after each migration
- Some models have inconsistent tags

**Gotcha**: Si agregas un nuevo modelo con FK, seguir el pattern de `migration/base.go`.

**Referencia**: `migration/base.go` para ver cómo se hace correctamente.

---

## Mejoras Futuras (Roadmap)

- [ ] Script para análisis automático de logs y extracción de errores
- [ ] Integración con Sentry para tracking automático de errores
- [ ] Base de datos de errores históricos para búsqueda de patrones
- [ ] Script de reproducción automática de errores (replay request)
- [ ] Dashboard de métricas de errores por módulo
- [ ] Herramienta de visualización de stack traces
- [ ] Template de runbooks para common incidents