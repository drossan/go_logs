---
name: go-code-reviewer
description: Esta skill debe usarse cuando el usuario necesite revisar código Go para identificar problemas de calidad, bugs potenciales, incumplimiento de mejores prácticas y desviaciones de los patrones arquitectónicos del proyecto. Se activa con peticiones como "revisa este código", "encuentra bugs en este handler", "evalúa la calidad de este servicio" o "revisa si sigue las convenciones del proyecto".
license: Complete terms in LICENSE.txt
version: 1.0.0
author: reverence-hotels-dev-team
category: base
tags: [go, code-review, clean-architecture, quality, best-practices, echo-framework]---

# Go Code Reviewer

Skill especializada para revisión exhaustiva de código Go en el contexto del proyecto Reverence Hotels API. Proporciona conocimiento profundo sobre los patrones arquitectónicos híbridos (MVC + Clean Architecture), convenciones específicas del proyecto, mejores prácticas de Go, y estándares de calidad para el framework Echo v4 y GORM v2.

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:
- Se requiera revisar código Go nuevo o modificado
- Necesite identificarse bugs potenciales o problemas de concurrencia
- Deba verificarse el cumplimiento de patrones Clean Architecture
- Se detecten desviaciones de las convenciones del proyecto
- Requiera evaluarse la calidad de handlers, services, repositories o models
- Necesite validarse la correcta implementación de middleware o autorización

Triggers comunes:
- "Revisa este handler de Echo"
- "Encuentra problemas en este servicio"
- "¿Este código sigue Clean Architecture?"
- "Evalúa la calidad de este repository"
- "Revisa si hay race conditions en este código"
- "¿Cumple este código las convenciones del proyecto?"
- "Revisa la seguridad de este endpoint"

## Workflow Principal

### 1. Análisis Contextual

Antes de revisar el código:
1. Identificar el tipo de módulo (Legacy MVC vs Clean Architecture)
2. Verificar la ubicación del archivo (`controllers/`, `internal/`, `models/`, etc.)
3. Determinar el patrón arquitectónico que debe seguir según el módulo
4. Identificar dependencias externas (GORM, Echo, EventBus, etc.)

### 2. Revisión Arquitectónica

Evaluar adherencia a patrones del proyecto:

**Para módulos Clean Architecture (`internal/*/`)**:
- [ ] Domain entity en `domain/` sin dependencias externas
- [ ] Interfaces definidas en `ports/`
- [ ] Service en `application/` orquestando use cases
- [ ] Repository en `infrastructure/db/` implementando puerto
- [ ] Handler en `infrastructure/http/` como primary adapter
- [ ] Inyección de dependencias manual, sin frameworks

**Para módulos Legacy MVC**:
- [ ] Model en `models/` con tags GORM correctos
- [ ] Service en `services/` con lógica de negocio
- [ ] Controller en `controllers/` usando Echo context
- [ ] Separación clara de responsabilidades

### 3. Revisión de Calidad de Código

Evaluar aspectos específicos de Go:

**Manejo de Errores**:
- [ ] Todos los errores están capturados y retornados
- [ ] Se usa `errors.Wrap()` o `errors.WithMessage()` para contexto
- [ ] No hay `panic()` en código de producción
- [ ] Errores se validan inmediatamente después de operaciones críticas

**Gestión de Contextos**:
- [ ] `context.Context` se propaga adecuadamente
- [ ] Timeouts configurados para operaciones externas
- [ ] Context cancelado ante errores o finalización
- [ ] No se usa `context.Background()` en handlers

**Concurrencia**:
- [ ] No hay race conditions (uso de `go vet --race`)
- [ ] Mutex o canales para compartir estado
- [ ] WaitGroups para esperar goroutines
- [ ] No hay leaks de goroutines

**Manejo de Base de Datos**:
- [ ] Uso correcto de GORM v2 (no métodos v1 deprecados)
- [ ] Transacciones para operaciones atómicas
- [ ] `defer db.Rollback()` en transacciones
- [ ] Conexión correcta a DB (db, dbSII, dbProducts según módulo)

### 4. Revisión de Seguridad

Verificar vulnerabilidades comunes:

**SQL Injection**:
- [ ] Siempre usar GORM con parámetros (nunca concatenar strings)
- [ ] `?` placeholders para queries raw
- [ ] Validación de input antes de queries

**Autorización**:
- [ ] Middleware de autenticación JWT en endpoints `/api/v1/*`
- [ ] Verificación de privilegios de Level para path específico
- [ ] API key validation para endpoints PMS (`X-API-KEY` header)
- [ ] No hay hardcoded credentials

**Data Sanitization**:
- [ ] Validación de structs con tags `validate`
- [ ] Sanitización de input de usuario
- [ ] No hay exposición de datos sensibles en errores

### 5. Revisión de Integraciones

Evaluar puntos de integración específicos:

**SII (Facturación Electrónica)**:
- [ ] XML firmado con certificado `.pem`
- [ ] Comunicación con endpoint AEAT correcto
- [ ] Tracking codes CSV generados correctamente

**Porta Sigma (Firmas Digitales)**:
- [ ] Transaction creada en API Porta Sigma
- [ ] Email enviado al empleado
- [ ] Status polling en intervalos correctos (15 min)

**EventBus (Domain Events)**:
- [ ] Events publicados con `eventBus.Publish()`
- [ ] Structs con ID, Timestamp, Data, Log
- [ ] Event handlers registrados correctamente

**Notifications (Multi-channel)**:
- [ ] Strategy pattern para notifiers (Email, Slack, Push)
- [ ] Fallback configurado para canales alternativos

### 6. Revisión de Performance

Identificar problemas de rendimiento:

**Database Queries**:
- [ ] No hay N+1 queries (usar `Preload` de GORM)
- [ ] Índices usados en filtros comunes
- [ ] `Select` específico para evitar fetch de columnas innecesarias
- [ ] `Limit` y `Offset` para paginación

**Memory Management**:
- [ ] No hay leaks de recursos (defer Close/CloseAllWindows)
- [ ] Buffers y readers cerrados adecuadamente
- [ ] Cantidad de datos acotada (no fetch sin límite)

**Goroutines**:
- [ ] Pool de goroutines para tareas concurrentes
- [ ] Límite de concurrencia configurado
- [ ] Context cancellation para cancelar trabajo

### 7. Validación de Convenciones del Proyecto

Verificar adherencia a estándares específicos:

**Nomenclatura**:
- [ ] Nombres de archivos en snake_case: `user_service.go`
- [ ] Nombres de funciones exportadas en PascalCase: `GetUserByID`
- [ ] Nombres de funciones privadas en camelCase: `validateInput`
- [ ] Constantes en PascalCase o UPPER_CASE

**Estructura de Paquetes**:
- [ ] `internal/{module}/domain/` para entidades
- [ ] `internal/{module}/ports/` para interfaces
- [ ] `internal/{module}/application/` para servicios
- [ ] `internal/{module}/infrastructure/` para implementaciones

**Configuración**:
- [ ] Variables de entorno usadas para configuración
- [ ] No hay hardcoded values (URLs, ports, paths)
- [ ] Environment-based behavior (local/pre/pro)

**Testing**:
- [ ] Tests en archivo `_test.go` junto al código
- [ ] Table-driven tests para múltiples casos
- [ ] Mocks para dependencias externas
- [ ] Coverage mínimo 70% para código crítico

## Recursos de la Skill

### Referencias (`references/`)

#### `references/architecture_patterns.md`

**Contenido**: Patrones arquitectónicos del proyecto, estructura de módulos Clean Architecture vs Legacy MVC, diagramas de capas.

**Cuándo consultar**: Al revisar estructura de módulos nuevos o verificar adherencia a patrones existentes.

**Estructura**:

- Sección 1: Clean Architecture en `internal/` (domain, ports, application, infrastructure)
- Sección 2: Legacy MVC en controllers/models/services
- Sección 3: Criterios para elegir patrón según tipo de módulo
- Sección 4: Diagramas de flujo de datos entre capas

**Búsqueda rápida**:

```bash
# Buscar patrón para módulo específico
grep -A 15 "### User Module" references/architecture_patterns.md

# Buscar criterios Clean Architecture
grep -B 5 -A 10 "When to use Clean Architecture" references/architecture_patterns.md
```

---

#### `references/go_best_practices.md`

**Contenido**: Mejores prácticas específicas de Go aplicadas al proyecto (error handling, context management, concurrency).

**Cuándo consultar**: Al validar calidad de código Go, identificar antipatterns.

**Estructura**:

- Error Handling con wrapping
- Context propagation patterns
- Concurrency patterns (worker pools, fan-out/fan-in)
- Memory management
- Interface design principles

**Búsqueda rápida**:

```bash
# Buscar manejo de errores
grep -A 20 "## Error Handling" references/go_best_practices.md

# Buscar patrones de concurrencia
grep -B 3 -A 15 "### Worker Pool" references/go_best_practices.md
```

---

#### `references/gorm_v2_patterns.md`

**Contenido**: Patrones de uso de GORM v2, migración de v1 a v2, queries optimizadas.

**Cuándo consultar**: Al revisar repositorios, models, queries crudas.

**Estructura**:

- Configuración de conexiones múltiples (db, dbSII, dbProducts)
- Queries con Preload para evitar N+1
- Transacciones con Context
- Migraciones AutoMigrate vs manuales
- Hooks de GORM (BeforeCreate, AfterUpdate)

**Búsqueda rápida**:

```bash
# Buscar patrón de transacciones
grep -A 25 "## Transaction Pattern" references/gorm_v2_patterns.md

# Buscar queries N+1
grep -B 5 -A 20 "### Avoiding N+1" references/gorm_v2_patterns.md
```

---

#### `references/echo_framework_patterns.md`

**Contenido**: Patrones de handlers Echo, middleware, routing, binding, validation.

**Cuándo consultar**: Al revisar controllers, handlers HTTP, middleware personalizado.

**Estructura**:

- Handler function signatures
- Context extraction and response
- Middleware chain order
- Request validation with echo's validator
- Error handling middleware

**Búsqueda rápida**:

```bash
# Buscar patrones de handlers
grep -A 30 "## Handler Pattern" references/echo_framework_patterns.md

# Buscar middleware ordering
grep -B 5 -A 15 "### Middleware Chain" references/echo_framework_patterns.md
```

---

#### `references/security_checklist.md`

**Contenido**: Checklist de seguridad específico del proyecto (autorización, SQL injection, sanitización).

**Cuándo consultar**: Al revisar handlers, endpoints públicos/privados, manejo de credenciales.

**Estructura**:

- JWT authentication requirements
- Level-based authorization checks
- API key validation for PMS endpoints
- Input validation requirements
- Sensitive data handling

**Búsqueda rápida**:

```bash
# Buscar checks de autorización
grep -A 20 "## Authorization" references/security_checklist.md

# Buscar validación de input
grep -B 5 -A 15 "### Input Validation" references/security_checklist.md
```

---

#### `references/coding_conventions.md`

**Contenido**: Convenciones de código específicas del proyecto (nomenclatura, estructura, comentarios).

**Cuándo consultar**: Al verificar consistencia con estándares del equipo.

**Estructura**:

- Nomenclatura de archivos y paquetes
- Funciones exportadas vs privadas
- Comentarios y godoc formatting
- Constantes y configuración
- Test naming conventions

**Búsqueda rápida**:

```bash
# Buscar convenciones de nomenclatura
grep -A 15 "## Naming Conventions" references/coding_conventions.md

# Buscar formato de tests
grep -B 3 -A 10 "### Test Conventions" references/coding_conventions.md
```

---

#### `references/integration_patterns.md`

**Contenido**: Patrones de integración con sistemas externos (SII, Porta Sigma, EventBus, Notifications).

**Cuándo consultar**: Al revisar código de integraciones, eventos, notificaciones.

**Estructura**:

- SII XML generation and signing
- Porta Sigma transaction flow
- EventBus publish/subscribe
- Notification strategy pattern
- Cron job scheduling

**Búsqueda rápida**:

```bash
# Buscar flujo de SII
grep -A 25 "## SII Integration" references/integration_patterns.md

# Buscar eventos de dominio
grep -B 5 -A 20 "### Domain Events" references/integration_patterns.md
```

---

#### `references/common_gotchas.md`

**Contenido**: Problemas comunes específicos del proyecto y cómo detectarlos.

**Cuándo consultar**: Al investigar bugs sutiles o comportamientos inesperados.

**Estructura**:

- GORM v1 → v2 migration issues
- Multi-database query mistakes
- Authorization path matching edge cases
- Cron job environment differences
- Foreign key handling

**Búsqueda rápida**:

```bash
# Buscar problemas de migración GORM
grep -A 20 "## GORM Migration" references/common_gotchas.md

# Buscar edge cases de autorización
grep -B 5 -A 15 "### Authorization Edge Cases" references/common_gotchas.md
```

---

### Scripts (`scripts/`)

#### `scripts/analyze_goroutines.sh`

**Propósito**: Detectar potenciales race conditions y leaks de goroutines

**Uso**:

```bash
./scripts/analyze_goroutines.sh <path-to-go-files>
```

**Parámetros**:

- `path-to-go-files`: Directorio con archivos .go a analizar (obligatorio)

**Output**: Reporte de issues de concurrencia detectados

**Ejemplo**:

```bash
./scripts/analyze_goroutines.sh ./internal/user/
```

---

#### `scripts/check_db_queries.sh`

**Propósito**: Identificar queries N+1 y falta de preloads en código GORM

**Uso**:

```bash
./scripts/check_db_queries.sh <path-to-go-files>
```

**Parámetros**:

- `path-to-go-files`: Directorio con archivos .go (obligatorio)

**Output**: Lista de queries potencialmente optimizables

**Ejemplo**:

```bash
./scripts/check_db_queries.sh ./controllers/Profile/
```

---

#### `scripts/validate_auth_patterns.sh`

**Propósito**: Verificar que los endpoints tienen middleware de autenticación/autorización correcto

**Uso**:

```bash
./scripts/validate_auth_patterns.sh <routes-file>
```

**Parámetros**:

- `routes-file`: Archivo de definición de rutas (ej: routes/routes.go)

**Output**: Reporte de endpoints sin protección adecuada

**Ejemplo**:

```bash
./scripts/validate_auth_patterns.sh ./routes/routes.go
```

---

#### `scripts/check_error_handling.sh`

**Propósito**: Detectar errores no manejados o uso incorrecto de panic()

**Uso**:

```bash
./scripts/check_error_handling.sh <path-to-go-files>
```

**Parámetros**:

- `path-to-go-files`: Directorio con archivos .go (obligatorio)

**Output**: Lista de errores no manejados y usos de panic

**Ejemplo**:

```bash
./scripts/check_error_handling.sh ./services/Invoices/
```

---

#### `scripts/verify_gorm_v2_migration.sh`

**Propósito**: Verificar que el código no usa métodos deprecados de GORM v1

**Uso**:

```bash
./scripts/verify_gorm_v2_migration.sh <path-to-go-files>
```

**Parámetros**:

- `path-to-go-files`: Directorio con archivos .go (obligatorio)

**Output**: Métodos v1 detectados que deben migrarse a v2

**Ejemplo**:

```bash
./scripts/verify_gorm_v2_migration.sh ./models/
```

---

### Assets (`assets/`)

#### `assets/review_report_template.md`

**Tipo**: Template

**Uso**: Plantilla para generar reportes estructurados de code review

**Modificaciones**: Personalizar secciones según tipo de revisión (arquitectura, seguridad, performance)

---

#### `assets/checklist_architecture.md`

**Tipo**: Checklist interactivo

**Uso**: Verificar adherencia a Clean Architecture o MVC según módulo

**Modificaciones**: Adaptar items según módulo específico bajo revisión

---

#### `assets/security_audit_template.md`

**Tipo**: Template

**Uso**: Documentar findings de seguridad durante revisión

**Modificaciones**: Completar con vulnerabilidades específicas encontradas

---

## Ejemplos de Uso

### Ejemplo 1: Revisión de Handler Echo Nuevo

**Input del usuario**:
> "Revisa este nuevo handler para crear usuarios y busca problemas"

**Proceso**:

1. Identificar que es un handler en `internal/backend/user/infrastructure/http/`
2. Verificar que sigue Clean Architecture (handler → service → repository)
3. Revisar middleware de autenticación JWT
4. Validar manejo de errores con wrapping
5. Verificar que valida input con struct tags
6. Revisar que se publiquen eventos de dominio
7. Buscar problemas de seguridad (SQL injection, autorización)

**Output esperado**:

```
## Code Review: CreateUserHandler

### ✅ Fortalezas
- Clean Architecture bien implementada (handler → service → repository)
- Error wrapping adecuado con context
- Validación de input con struct tags
- Domain event publicado correctamente

### ⚠️ Issues Críticos
1. **Missing authorization check**
   - Ubicación: `internal/backend/user/infrastructure/http/handler.go:45`
   - Problema: No verifica nivel de usuario antes de crear
   - Impacto: Cualquier usuario autenticado podría crear usuarios
   - Solución: Agregar middleware de nivel o verificar LevelPrivileges
   ```go
   // Verificar nivel antes de crear
   if !h.authService.CanCreateUser(c.Get("userLevel")) {
       return echo.NewHTTPError(http.StatusForbidden, "insufficient privileges")
   }
   ```

2. **SQL injection risk in raw query**
   - Ubicación: `internal/backend/user/infrastructure/db/gorm_repository.go:78`
   - Problema: Concatenación de string en query raw
   - Impacto: Posible SQL injection si email viene malicioso
   - Solución: Usar placeholder de GORM
   ```go
   // Incorrecto
   db.Raw("SELECT * FROM users WHERE email = '" + email + "'")
   
   // Correcto
   db.Raw("SELECT * FROM users WHERE email = ?", email)
   ```

### 📊 Issues Menores
3. **Missing timeout in DB operation**
   - Ubicación: `gorm_repository.go:78`
   - Problema: Query sin timeout podría colgar
   - Solución: Agregar context con timeout
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   result := r.db.WithContext(ctx).Where(...)
   ```

4. **Goroutine leak potential**
   - Ubicación: `handler.go:62`
   - Problema: Goroutine lanzada sin WaitGroup
   - Solución: Usar WaitGroup o async job con cleanup

### 🏗️ Arquitectura
- ✅ Sigue Clean Architecture correctamente
- ✅ Dependency injection manual (sin framework)
- ✅ Interfaces bien definidas en ports/
- ✅ Repository implementa puerto correctamente

### 🔒 Seguridad
- ❌ Falta verificación de autorización (crítico)
- ✅ JWT validation presente
- ⚠️ Input validation básica, podría ser más robusta
- ✅ No hay hardcoded credentials

### ⚡ Performance
- ⚠️ Posible N+1 query si se crean relaciones sin Preload
- ✅ Paginación implementada
- ✅ Timeout faltante en DB query

### 📝 Convenciones
- ✅ Nomenclatura correcta (snake_case archivos, PascalCase exportadas)
- ✅ Estructura de paquetes sigue estándar del proyecto
- ⚠️ Falta godoc en funciones exportadas

### 🧪 Testing
- ❌ No hay tests en el archivo (debería haber user_handler_test.go)
- Recomendación: Agregar table-driven tests para:
   - Creación exitosa
   - Validación de input
   - Usuario sin privilegios
   - Email duplicado

### 📊 Resumen
- **Críticos**: 2 (requiere fix inmediato)
- **Moderados**: 3 (debería corregirse pronto)
- **Menores**: 2 (mejora sugerida)
- **Fortalezas**: 6

**Recomendación**: NO mergear sin corregir issues críticos de seguridad.
```

---

### Ejemplo 2: Revisión de Servicio Legacy

**Input del usuario**:
> "Revisa este servicio de facturación y busca bugs"

**Proceso**:

1. Identificar que es módulo Legacy MVC en `services/Invoices/`
2. Verificar patrón service → models
3. Revisar comunicación con API SII
4. Validar manejo de certificados .pem
5. Buscar problemas de concurrencia en envío de XMLs
6. Revisar transacciones de base de datos

**Output esperado**:

```
## Code Review: InvoiceService

### ⚠️ Issues Críticos
1. **Race condition en contador de facturas**
   - Ubicación: `services/Invoices/service.go:112`
   - Problema: Múltiples goroutines accediendo a contador sin mutex
   - Impacto: Números de factura duplicados
   - Solución: Proteger acceso con mutex
   ```go
   var (
       invoiceCounter int
       counterMutex   sync.Mutex
   )
   
   func getNextInvoiceNumber() int {
       counterMutex.Lock()
       defer counterMutex.Unlock()
       invoiceCounter++
       return invoiceCounter
   }
   ```

2. **Transaction not rolled back on SII failure**
   - Ubicación: `services/Invoices/service.go:156`
   - Problema: Si API SII falla, factura queda en BD pero no enviada
   - Impacto: Inconsistencia entre BD y SII
   - Solución: Usar transacción y rollback en error
   ```go
   tx := db.Begin()
   if err := tx.Create(&invoice).Error; err != nil {
       tx.Rollback()
       return err
   }
   
   if err := sendToSII(invoice); err != nil {
       tx.Rollback()
       return err
   }
   
   tx.Commit()
   ```

### 📊 Issues Moderados
3. **GORM v1 syntax detected**
   - Ubicación: `models/Invoice/invoice.go:45`
   - Problema: Usa `Find()` sin `Error` check (patrón v1)
   - Impacto: Errores silenciosos si query falla
   - Solución: Migrar a sintaxis v2
   ```go
   // GORM v1 (deprecated)
   db.Find(&invoices)
   
   // GORM v2 (correct)
   result := db.Find(&invoices)
   if result.Error != nil {
       return result.Error
   }
   ```

4. **Certificate hardcoded path**
   - Ubicación: `services/Invoices/sii_client.go:23`
   - Problema: Path del certificado .pem hardcoded
   - Impacto: No funciona en diferentes entornos
   - Solución: Usar variable de entorno
   ```go
   certPath := os.Getenv("SII_CERT_PATH")
   if certPath == "" {
       return errors.New("SII_CERT_PATH not set")
   }
   ```

### 🔍 Issues Menores
5. **Missing context timeout in SII API call**
   - Ubicación: `services/Invoices/sii_client.go:78`
   - Problema: HTTP request sin timeout podría colgar
   - Solución: Agregar context con timeout

6. **Godoc missing in exported functions**
   - Ubicación: Todo el archivo
   - Problema: Funciones exportadas sin documentación
   - Solución: Agregar godoc comments

### 🏗️ Arquitectura
- ✅ Sigue patrón Legacy MVC correctamente
- ✅ Service → Model separación clara
- ⚠️ Podría beneficiarse de migración a Clean Architecture
- ✅ Usa DB correcta (dbSII para facturas)

### 🔒 Seguridad
- ✅ Certificado .pem manejado correctamente
- ✅ XML firmado antes de enviar a SII
- ⚠️ API key de SII debería rotarse
- ✅ No hay exposición de datos sensibles

### ⚡ Performance
- ❌ Posible N+1 si se fetchan líneas de factura
- ✅ Índices usados en filtros comunes
- ⚠️ Timeout faltante en HTTP client

### 📝 Convenciones
- ✅ Nomenclatura correcta
- ✅ Ubicación correcta en Legacy MVC
- ⚠️ Falta documentación

### 🧪 Testing
- ❌ Sin tests (crítico para módulo de facturación)
- Recomendación: Agregar tests para:
   - Generación de XML
   - Firma de documentos
   - Manejo de errores SII
   - Rollback de transacciones

### 📊 Resumen
- **Críticos**: 2
- **Moderados**: 2
- **Menores**: 2
- **Fortalezas**: 5

**Recomendación**: Corregir race condition y issue de transacciones ANTES de mergear.
```

---

### Ejemplo 3: Revisión de Arquitectura de Nuevo Módulo

**Input del usuario**:
> "Este nuevo módulo de notificaciones sigue Clean Architecture?"

**Proceso**:

1. Examinar estructura en `internal/notification/`
2. Verificar que domain/ no tiene dependencias externas
3. Validar que interfaces están en ports/
4. Revisar que application/ orquesta use cases
5. Verificar que infrastructure/db implementa repository
6. Revisar que infrastructure/http implementa handler
7. Validar inyección de dependencias en cmd/server/

**Output esperado**:

```
## Code Review: Notification Module Architecture

### ✅ Clean Architecture Adherencia

**Domain Layer** (`internal/notification/domain/`)
- ✅ Entity `Notification` sin dependencias externas
- ✅ Lógica de dominio pura (no depende de Go std lib ni frameworks)
- ✅ Estructura simple con ID, Type, Content, Status, Timestamps

**Ports Layer** (`internal/notification/ports/`)
- ✅ Interfaces bien definidas:
   - `NotificationService` con métodos de negocio
   - `NotificationRepository` para persistencia
- ✅ Interfaces sin implementaciones (correcto)
- ✅ Nombres descriptivos y cohesivos

**Application Layer** (`internal/notification/application/`)
- ✅ `NotificationServiceImpl` implementa puerto
- ✅ Orquesta use cases (Create, Send, MarkAsRead)
- ✅ Publica domain events correctamente
- ✅ No depende de infraestructura (solo interfaces)
- ⚠️ Falta validación de business rules antes de crear

**Infrastructure Layer** (`internal/notification/infrastructure/`)
- ✅ `db/gorm_repository.go` implementa `NotificationRepository`
- ✅ `http/handler.go` implementa endpoint HTTP
- ✅ Strategy pattern para notifiers (Email, Slack, Push)
- ✅ Adapters externos aislados

**Wiring** (`cmd/server/backend/`)
- ✅ Inyección manual de dependencias
- ✅ Instanciación de repository con DB connection
- ✅ Instanciación de service con repository y EventBus
- ✅ Handler registrado con service inyectado

### 🔍 Issues Encontrados

**Moderados**:
1. **Domain entity con JSON tags**
   - Ubicación: `internal/notification/domain/entity.go:15`
   - Problema: Domain no debería tener detalles de serialización
   - Solución: Mover JSON tags a DTO en infrastructure/http
   ```go
   // domain/entity.go (limpio)
   type Notification struct {
       ID       int
       Type     string
       Content  string
       Status   string
       SentAt   time.Time
   }
   
   // infrastructure/http/dto.go (con serialización)
   type NotificationDTO struct {
       ID       int    `json:"id"`
       Type     string `json:"type"`
       Content  string `json:"content"`
       Status   string `json:"status"`
       SentAt   time.Time `json:"sent_at"`
   }
   ```

2. **Service usando struct de DB directamente**
   - Ubicación: `internal/notification/application/service.go:45`
   - Problema: Service retorna domain entity pero la usa el handler
   - Solución: Crear DTOs para layer de presentación
   ```go
   // application/service.go
   func (s *NotificationServiceImpl) CreateNotification(ctx context.Context, req CreateNotificationRequest) (*domain.Notification, error) {
       // ... lógica de negocio
       return notification, nil
   }
   
   // infrastructure/http/handler.go
   func (h *NotificationHandler) CreateNotification(c echo.Context) error {
       // ... obtener de service
       entity, err := h.service.CreateNotification(ctx, req)
       dto := toDTO(entity)  // Convertir a DTO
       return c.JSON(http.StatusCreated, dto)
   }
   ```

**Menores**:
3. **Falta validación de input en service**
   - Ubicación: `application/service.go:32`
   - Problema: No valida que Type sea válido antes de crear
   - Solución: Agregar validación de enum o const

4. **Repository sin manejo de context**
   - Ubicación: `infrastructure/db/gorm_repository.go:28`
   - Problema: Queries sin context timeout
   - Solución: Aceptar context en métodos de repository

### 🏗️ Diagrama de Capas (Actual)

```
┌─────────────────────────────────────────────┐
│ HTTP Handler (Echo)                         │ ← Presentación
├─────────────────────────────────────────────┤
│ Application Service                         │ ← Casos de uso
├─────────────────────────────────────────────┤
│ Domain Entity (Notification)                │ ← Dominio
├─────────────────────────────────────────────┤
│ Repository (GORM)                           │ ← Persistencia
└─────────────────────────────────────────────┘
```

**Flujo de Datos**: ✅ Correcto (dependencias apuntan hacia adentro)

### 📊 Resumen

**Adherencia a Clean Architecture**: 85%

- ✅ Dominio aislado de infraestructura
- ✅ Puertos bien definidos
- ✅ Inyección de dependencias manual
- ⚠️ Mejoras: Separar DTOs de domain entities
- ⚠️ Mejoras: Agregar validaciones en service

**Recomendación**: 
- Aprobar arquitectura general
- Corregir issues moderados antes de producción
- Agregar tests de integración entre capas
- Documentar patrón de DTOs para futuros módulos
```

---

### Ejemplo 4: Detección de Vulnerabilidad de Seguridad

**Input del usuario**:
> "Encuentra vulnerabilidades de seguridad en este endpoint de perfil"

**Proceso**:

1. Identificar endpoint en `controllers/Profile/profile_controller.go`
2. Revisar middleware de autenticación y autorización
3. Buscar validación de input
4. Detectar posibles SQL injections
5. Verificar que no exponga datos sensibles
6. Revisar autorización level-based

**Output esperado**:

```
## Security Audit: Profile Endpoint

### 🔴 Vulnerabilidades Críticas

**1. Broken Authorization - IDOR (Insecure Direct Object Reference)**
- **Ubicación**: `controllers/Profile/profile_controller.go:78`
- **Endpoint**: `GET /api/v1/profile/:id`
- **Vulnerabilidad**: Cualquier usuario puede ver perfil de otro cambiando el ID
- **Impacto**: Exposición de datos personales de empleados
- **Código vulnerable**:
  ```go
  func GetProfile(c echo.Context) error {
      id := c.Param("id")  // ❌ No verifica que el ID pertenezca al usuario
      profile, err := service.GetProfileByID(id)
      // ...
  }
  ```
- **Solución**:
  ```go
  func GetProfile(c echo.Context) error {
      currentUserID := c.Get("userID").(string)  // ✅ Del JWT
      requestedID := c.Param("id")
      
      // Verificar que solo puede ver su propio perfil
      if currentUserID != requestedID {
           // O verificar si tiene nivel de admin
           if !hasAdminLevel(c.Get("userLevel")) {
               return echo.NewHTTPError(http.StatusForbidden, "access denied")
           }
      }
      
      profile, err := service.GetProfileByID(requestedID)
      // ...
  }
  ```

**2. SQL Injection en búsqueda**
- **Ubicación**: `controllers/Profile/profile_controller.go:145`
- **Endpoint**: `POST /api/v1/profile/search`
- **Vulnerabilidad**: Query raw con string concatenation
- **Impacto**: Posible exfiltración de base de datos completa
- **Código vulnerable**:
  ```go
  query := "SELECT * FROM profiles WHERE name LIKE '%" + searchTerm + "%'"
  db.Raw(query).Find(&profiles)  // ❌ SQL injection
  ```
- **Solución**:
  ```go
  query := "SELECT * FROM profiles WHERE name LIKE ?"
  db.Raw(query, "%"+searchTerm+"%").Find(&profiles)  // ✅ Safe
  ```

### 🟡 Vulnerabilidades Moderadas

**3. Missing Level-Based Authorization**
- **Ubicación**: `controllers/Profile/profile_controller.go:210`
- **Endpoint**: `PUT /api/v1/profile/:id/holidays`
- **Problema**: No verifica LevelPrivileges para modificar vacaciones
- **Impacto**: Usuario sin privilegios podría modificar vacaciones ajenas
- **Solución**:
  ```go
  func UpdateHolidays(c echo.Context) error {
      level := c.Get("userLevel").(int)
      formPath := "profile-holidays"
      
      // Verificar privilegio de Write
      if !authService.HasWritePrivilege(level, formPath) {
          return echo.NewHTTPError(http.StatusForbidden, "insufficient privileges")
      }
      // ...
  }
  ```

**4. Sensitive Data in Error Messages**
- **Ubicación**: `controllers/Profile/profile_controller.go:67`
- **Problema**: Error messages incluyen detalles de base de datos
- **Impacto**: Information disclosure para atacantes
- **Código vulnerable**:
  ```go
  if err != nil {
      return c.JSON(500, map[string]string{
          "error": "Database error: " + err.Error(),  // ❌ Expone DB internals
      })
  }
  ```
- **Solución**:
  ```go
  if err != nil {
      log.Error().Err(err).Msg("failed to get profile")  // ✅ Log completo
      return c.JSON(500, map[string]string{
          "error": "Internal server error",  // ✅ Mensaje genérico
      })
  }
  ```

### 🟢 Fortalezas de Seguridad

- ✅ JWT validation presente en middleware
- ✅ HTTPS obligatorio en producción
- ✅ Passwords hasheados con bcrypt
- ✅ No hay hardcoded credentials visibles
- ✅ CORS configurado correctamente

### 📋 Recomendaciones Prioritarias

1. **INMEDIATO**: Corregir IDOR en endpoint de perfil (crítico)
2. **INMEDIATO**: Corregir SQL injection en búsqueda (crítico)
3. **CORTO PLAZO**: Implementar verificación de LevelPrivileges en todos endpoints
4. **CORTO PLAZO**: Sanitizar mensajes de error
5. **MEDIO PLAZO**: Agregar rate limiting para prevenir brute force
6. **MEDIO PLAZO**: Implementar audit logging para accesos a perfiles ajenos

### 📊 Score de Seguridad

**Actual**: 4/10 (vulnerabilidades críticas presentes)
**Tras correcciones**: 8/10 (con recomendaciones implementadas)

**NO RECOMENDADO** desplegar sin corregir vulnerabilidades críticas.
```

---

## Presentación de Resultados

Al completar una revisión de código:

1. **Estructurar el reporte** con secciones claras:
   - Resumen ejecutivo (fortalezas, issues críticos, score)
   - Issues detallados por severidad (críticos, moderados, menores)
   - Análisis arquitectónico
   - Revisión de seguridad
   - Evaluación de performance
   - Verificación de convenciones
   - Estado de testing
   - Recomendaciones priorizadas

2. **Formato de issues**:
   ```markdown
   ### [Severidad] Título del issue
   - **Ubicación**: `archivo:línea`
   - **Problema**: Descripción clara
   - **Impacto**: Qué podría causar
   - **Código vulnerable**: Snippet del problema
   - **Solución**: Código corregido con explicación
   ```

3. **Incluir código de ejemplo** para issues críticos y moderados

4. **Priorizar correcciones**:
   - 🔴 Críticos: Seguridad, datos, crashes (bloquean merge)
   - 🟡 Moderados: Performance, mantenibilidad (deberían corregirse)
   - 🟢 Menores: Estilo, documentación (mejoras sugeridas)

5. **Métricas de calidad**:
   - Adherencia a Clean Architecture (si aplica): X%
   - Coverage de tests: X%
   - Score de seguridad: X/10
   - Performance: ⚠️ issues identificados
   - Convenciones: ✅ cumple / ❌ no cumple

6. **Veredicto final**:
   - **APPROVED**: Merge seguro
   - **APPROVED WITH MINOR SUGGESTIONS**: Merge ok, mejoras opcionales
   - **NEEDS REVISION**: Corregir moderados antes de merge
   - **REJECTED**: Críticos deben corregirse obligatoriamente

---

## Troubleshooting

### Problema: Dificultad para determinar patrón arquitectónico

**Síntoma**: No está claro si módulo debe seguir Clean Architecture o Legacy MVC

**Causa**: El proyecto está en transición híbrida

**Solución**:

1. Consultar `references/architecture_patterns.md` sección "Criterios de elección"
2. Verificar si el módulo ya existe en `internal/` (Clean Arch) o `controllers/` (Legacy)
3. Para módulos nuevos:
   - Si es complejo con múltiples integraciones → Clean Architecture
   - Si es CRUD simple → Legacy MVC
4. Cuando en duda, preferir Clean Architecture para nueva funcionalidad

---

### Problema: No detectar race conditions en código concurrente

**Síntoma**: Revisión manual no encuentra issues de concurrencia

**Solución**:

1. Ejecutar `go run -race init.go` en desarrollo
2. Usar script `scripts/analyze_goroutines.sh` para análisis estático
3. Buscar patrones problemáticos:
   - Variables compartidas sin mutex
   - Goroutines sin WaitGroup
   - Channels sin buffer que causan deadlocks
4. Consultar `references/go_best_practices.md` sección "Concurrency"

---

### Problema: Distinguir entre GORM v1 y v2 syntax

**Síntoma**: No claro si código usa métodos deprecados

**Solución**:

1. Ejecutar script `scripts/verify_gorm_v2_migration.sh`
2. Consultar `references/gorm_v2_patterns.md` para comparación
3. Principales diferencias:
   - v1: `db.Find(&items)` (ignora error)
   - v2: `result := db.Find(&items); if result.Error != nil {...}`
4. Buscar `?` placeholders vs string concatenation en queries

---

### Problema: Verificar autorización level-based correctamente

**Síntoma**: No claro cómo validar privilegios de nivel

**Solución**:

1. Consultar `references/security_checklist.md` sección "Authorization"
2. Verificar que el endpoint tenga middleware de JWT
3. Buscar verificación de `LevelPrivileges` en service o controller:
   ```go
   // Obtener nivel del usuario
   level := c.Get("userLevel").(int)
   
   // Obtener path del formulario (desde DB o config)
   formPath := "profile-holidays"
   
   // Verificar privilegio según método HTTP
   if c.Request().Method == "GET" {
       hasPrivilege := authService.HasReadPrivilege(level, formPath)
   } else {
       hasPrivilege := authService.HasWritePrivilege(level, formPath)
   }
   ```
4. Validar que `Form.PathAPI` incluya el path actual (puede ser múltiple con `|`)

---

### Problema: Identificar N+1 queries en código GORM

**Síntoma**: Performance issues pero no claro su origen

**Solución**:

1. Ejecutar script `scripts/check_db_queries.sh`
2. Buscar patrones de N+1:
   ```go
   // ❌ N+1: Fetch users, luego query por cada user
   users := []User{}
   db.Find(&users)
   for _, user := range users {
       db.Where("user_id = ?", user.ID).Find(&profile)  // Query extra por cada user
   }
   
   // ✅ Correcto: Preload relations
   users := []User{}
   db.Preload("Profile").Find(&users)  // Solo 2 queries
   ```
3. Consultar `references/gorm_v2_patterns.md` sección "Avoiding N+1"
4. Habilitar log de queries en desarrollo para detectar:

```go
db.LogMode(true)  // Log todas las queries
```

---

### Problema: Determinar si módulo debe publicar domain events

**Síntoma**: No claro cuándo usar EventBus

**Solución**:

1. Consultar `references/integration_patterns.md` sección "Domain Events"
2. Publicar eventos cuando:
   - Se crea una entidad aggregate root (User, Invoice, Notification)
   - Se modifica estado crítico (Profile.Status, Invoice.SentAt)
   - Se requiere notificación a otros módulos
3. NO publicar para:
   - Queries de lectura
   - Operaciones internas del módulo
   - Transient state changes
4. Eventos típicos del proyecto:
   - `user.created`, `user.updated`
   - `invoice.created`, `invoice.sent_to_sii`
   - `notification.created`, `notification.sent`

---

## Consideraciones Especiales

### Híbrida Arquitectónica

El proyecto está en transición de MVC a Clean Architecture:
- **Módulos nuevos**: Preferir Clean Architecture (`internal/`)
- **Módulos legacy**: Mantener MVC (`controllers/`, `models/`, `services/`)
- **No mezclar**: Un módulo debe ser claramente uno u otro

### Multi-Base de Datos

Verificar siempre que el código use la conexión correcta:
- `db`: Base de datos principal (users, profiles, HR)
- `dbSII`: Base de datos SII (facturación electrónica)
- `dbProducts`: Base de datos Economato (artículos, proveedores)

### Environment-Specific Behavior

Algunos comportamientos cambian según `ENV` variable:
- **local**: Sin cron jobs, logs en debug
- **pre**: Cron jobs limitados, logs en info
- **pro**: Todos los cron jobs, logs en warn/error

Verificar que código funcione correctamente en los 3 entornos.

### GORM v2 Migration

El proyecto migró recientemente de GORM v1 a v2:
- Revisar que no haya sintaxis v1 residual
- Foreign keys tuvieron issues (ver `migration/base.go`)
- Verificar que `AutoMigrate` se ejecuta con `--migrate=yes`

### Authorization Path-Based

La autorización se basa en **paths de API**, no en controllers:
- Cada `Form` tiene un `PathAPI` (ej: `"profile|profile-holidays|profile-absences"`)
- El middleware verifica si el path del request coincide con `Form.PathAPI`
- Un path puede corresponder a múltiples forms (separados por `|`)

### Testing Culture

El proyecto valora testing pero coverage varía por módulo:
- **Mínimo aceptable**: 70% para código crítico
- **Módulos nuevos**: Deben incluir tests desde el inicio
- **Preferencia**: Table-driven tests para múltiples casos
- **Integración**: Tests end-to-end para endpoints críticos

---

## Mejoras Futuras (Roadmap)

- [ ] Agregar detector automático de code smells
- [ ] Integrar con static analysis tools (golangci-lint)
- [ ] Generar métricas de complejidad ciclomática
- [ ] Agregar diff-based review (solo líneas cambiadas)
- [ ] Crear checklist interactivo de PR reviewers
- [ ] Integrar con seguridad de dependencias (go.sum validation)
- [ ] Agregar detector de deuda técnica acumulada
- [ ] Crear reportes automatizados de calidad de código