---
name: technical-writer
description: Esta skill debe usarse cuando el usuario necesite crear, actualizar o mejorar documentación técnica, comentarios de código, guías de usuario o especificaciones. Se activa con peticiones como "documenta este código", "crea una guía de usuario", "mejora los comentarios de esta función", "genera documentación de API", o "escribe especificaciones técnicas".
license: Complete terms in LICENSE.txt
version: 1.0.0
author: platform-team
category: base
tags: [documentation, technical-writing, code-comments, api-docs, user-guides]---

# Technical Writer

Skill especializada en generación y mejora de documentación técnica para el proyecto Reverence Hotels API. Proporciona conocimiento especializado sobre convenciones de documentación del proyecto, estándares de Go, patrones de arquitectura híbrida (MVC + Clean Architecture), y formatos de documentación aceptados.

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:
- Se necesite documentar nuevos módulos o funcionalidades
- Los comentarios de código requieran mejora o estandarización
- Se deban crear guías de usuario para desarrolladores
- Sea necesario documentar endpoints de API
- Requiera especificaciones técnicas para features
- Se deban actualizar documentos existentes con nuevos patrones

Triggers comunes:
- "Documenta este módulo de usuarios"
- "Crea una guía para implementar un nuevo endpoint"
- "Mejora los comentarios de este servicio"
- "Genera documentación de la API de perfiles"
- "Escribe especificaciones técnicas para el módulo de notificaciones"
- "Actualiza la documentación de arquitectura con este nuevo patrón"

## Workflow Principal

### 1. Análisis Inicial

Antes de proceder:
1. Identificar el tipo de documentación requerida:
   - Documentación de código (comentarios, godocs)
   - Documentación de API (endpoints, request/response)
   - Guías de usuario (tutoriales, ejemplos)
   - Especificaciones técnicas (requisitos, arquitectura)
2. Verificar el contexto del proyecto:
   - Módulo específico (controllers, models, services, internal/*)
   - Patrón arquitectónico (Legacy MVC vs Clean Architecture)
   - Base de datos involucrada (Principal, SII, Products)
3. Determinar el formato de salida apropiado:
   - Comentarios inline en código Go
   - Archivos markdown en docs/
   - Actualización de CLAUDE.md
   - Nuevos archivos de especificación

### 2. Ejecución

Según el tipo de documentación:

#### Documentación de Código Go

```go
// Estructura del comentario según convenciones del proyecto:
// 1. Propósito breve (1 línea)
// 2. Descripción detallada si es necesario
// 3. Parámetros con formatos
// 4. Retorno con formatos
// 5. Notas importantes (side effects, consideraciones)

// Ejemplo para servicios Clean Architecture:
```

Consultar `references/go_documentation_standards.md` para patrones detallados.

#### Documentación de API Endpoints

Usar template de `references/api_endpoint_template.md`:

```markdown
## POST /api/v1/resource

**Propósito**: {Descripción clara del propósito}

**Autenticación**: JWT + Authorization Level X

**Request**:
- Headers: ...
- Body: ...

**Response**:
- 200: ...
- 400: ...
- 401: ...
```

#### Guías de Usuario

Consultar `references/user_guide_patterns.md` para estructura estándar.

Incluir siempre:
- Prerrequisitos
- Ejemplos de comandos curl
- Ejemplos de response JSON
- Casos de error comunes

#### Especificaciones Técnicas

Usar estructura de `references/technical_spec_template.md`:

1. Contexto y motivación
2. Requisitos funcionales
3. Requisitos no funcionales
4. Diseño técnico
5. Consideraciones de seguridad
6. Plan de testing

### 3. Validación

Verificar que:

- [ ] La documentación sigue convenciones del proyecto
- [ ] Los ejemplos de código son ejecutables y correctos
- [ ] Los tipos de datos coinciden con los modelos GORM reales
- [ ] Los endpoints mencionados existen en routes/routes.go
- [ ] Se incluyen consideraciones de multi-database si aplica
- [ ] Se mencionan patrones de autorización (Level-based) si aplica
- [ ] La documentación está en español (idioma del proyecto)

### 4. Output

Presentar resultados:

- Formato: Markdown estructurado o comentarios de código formateados
- Incluir: Ejemplos prácticos, comandos, referencias cruzadas
- Omitir: Información obvia o redundante

## Recursos de la Skill

### Referencias (`references/`)

#### `references/go_documentation_standards.md`

**Contenido**: Convenciones de documentación para código Go en el proyecto

**Cuándo consultar**: Al escribir o mejorar comentarios en código Go

**Estructura**:

- Godoc comments format (exported functions/structs)
- Service layer documentation (Legacy MVC)
- Clean Architecture documentation (internal/*/ ports)
- Repository documentation
- Handler documentation
- Model/Entity documentation

**Patrones clave**:

```go
// Service comment structure for Legacy MVC:
// ServiceName provides functionality for...
// Usage:
//   service := services.NewServiceName(db)
//   result, err := service.MethodName(params)
//
// Parameters:
//   - db: *gorm.DB connection to principal database
//
// Returns:
//   - *Result: processed result
//   - error: error if operation fails
```

---

#### `references/api_endpoint_template.md`

**Contenido**: Template estándar para documentar endpoints de API

**Cuándo consultar**: Al documentar nuevos endpoints o actualizar existentes

**Estructura**:

```markdown
### {METHOD} /api/v1/{path}

**Propósito**: {Descripción clara y concisa}

**Autenticación**: 
- Type: {JWT / API Key / Public}
- Level: {N level if applicable}

**Request Headers**:
```json
{
  "Authorization": "Bearer {token}",
  "Content-Type": "application/json"
}
```

**Request Body**:
```json
{
  "field1": "type",
  "field2": "type"
}
```

**Response 200 OK**:
```json
{
  "data": {},
  "message": "Success"
}
```

**Response 400 Bad Request**:
```json
{
  "error": "Invalid input"
}
```

**Response 401 Unauthorized**:
```json
{
  "error": "Unauthorized"
}
```

**Database**: {Principal / SII / Products}

**Examples**:
\`\`\`bash
curl -X POST http://localhost:1331/api/v1/endpoint \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"field":"value"}'
\`\`\`
```

---

#### `references/architecture_patterns_reference.md`

**Contenido**: Patrones de arquitectura del proyecto para documentar correctamente

**Cuándo consultar**: Al documentar módulos que usan patrones específicos del proyecto

**Secciones principales**:

- Hybrid MVC + Clean Architecture pattern
- Multi-database connection pattern
- Level-based authorization system
- Domain event publishing pattern
- Repository pattern implementation
- Service layer patterns (Legacy vs Clean)

**Búsqueda rápida**:

```bash
# Buscar patrón específico
grep -i "repository pattern" references/architecture_patterns_reference.md

# Buscar consideraciones de base de datos
grep -A 15 "multi-database" references/architecture_patterns_reference.md
```

---

#### `references/user_guide_patterns.md`

**Contenido**: Estructuras y patrones para guías de usuario (desarrolladores)

**Cuándo consultar**: Al crear tutoriales o guías prácticas

**Estructura de guía estándar**:

```markdown
# Guía: {Título}

## Prerrequisitos
- Go 1.21+
- Access to databases
- JWT token

## Objetivo
{Qué logrará seguir esta guía}

## Pasos

### 1. {Paso 1}
{Instrucciones claras con ejemplos}

### 2. {Paso 2}
{Instrucciones con ejemplos de código/commands}

## Ejemplo Completo
{Caso de uso end-to-end}

## Troubleshooting
{Problemas comunes y soluciones}
```

---

#### `references/technical_spec_template.md`

**Contenido**: Template para especificaciones técnicas de nuevas features

**Cuándo consultar**: Al documentar requisitos y diseño de nuevas funcionalidades

**Estructura**:

1. **Contexto**: Motivación y problema a resolver
2. **Alcance**: Qué incluye y qué no incluye
3. **Requisitos Funcionales**: Comportamiento esperado
4. **Requisitos No Funcionales**: Performance, seguridad, escalabilidad
5. **Diseño Técnico**:
   - Arquitectura (MVC vs Clean Architecture)
   - Modelos/Entidades
   - Servicios/Repositories
   - Endpoints
   - Database migrations
6. **Integraciones**: Sistemas externos (SII, Porta Sigma, etc.)
7. **Seguridad**: Autorización, validación de datos
8. **Testing**: Estrategia de pruebas
9. **Consideraciones Especiales**: Multi-database, cron jobs, etc.

---

#### `references/module_documentation_checklist.md`

**Contenido**: Checklist para asegurar documentación completa de módulos

**Cuándo consultar**: Al validar que un módulo está completamente documentado

**Checklist por tipo de módulo**:

**Para módulos Clean Architecture (internal/*)**:
- [ ] Domain entity documented
- [ ] Ports (interfaces) documented
- [ ] Application service documented
- [ ] HTTP handler documented with endpoint details
- [ ] Repository implementation documented
- [ ] Domain events published documented
- [ ] Dependencies documented

**Para módulos Legacy MVC**:
- [ ] Model documented with GORM tags
- [ ] Service methods documented
- [ ] Controller endpoints documented
- [ ] Route registration mentioned
- [ ] Database usage specified (Principal/SII/Products)
- [ ] Authorization requirements documented

**Común para todos**:
- [ ] Prerrequisitos documentados
- [ ] Ejemplos de uso proporcionados
- [ ] Casos de error documentados
- [ ] Referencias a módulos relacionados incluidas

---

#### `references/cross_reference_index.md`

**Contenido**: Índice de documentos y módulos del proyecto para referencias cruzadas

**Cuándo consultar**: Para enlazar documentación relacionada

**Estructura por módulo**:

- **Auth**: `services/Auth/`, `controllers/Auth/`, docs/07-seguridad.md
- **Profile**: `controllers/Profile/`, `models/Profile/`
- **Invoices**: `services/Invoices/`, `controllers/Invoices/`, SII database
- **User**: `internal/backend/user/`
- **Forms**: `internal/backend/form/`
- **Levels**: `internal/backend/level/`
- **Events**: `internal/event/`
- **Notifications**: `internal/notification/`

**Búsqueda**:

```bash
# Encontrar docs de un módulo
grep -i "profile" references/cross_reference_index.md

# Encontrar módulos que usan eventos
grep -i "event" references/cross_reference_index.md
```

---

### Assets (`assets/`)

#### `assets/templates/doc_template.md`

**Tipo**: Template de documento

**Uso**: Plantilla base para nuevos documentos de proyecto

**Modificaciones**: Personalizar título, contenido y secciones según necesidad

---

#### `assets/templates/code_comment_template.txt`

**Tipo**: Template de comentarios

**Uso**: Plantilla para comentarios de código Go estandarizados

**Modificaciones**: Adaptar a función/struct específico

---

## Ejemplos de Uso

### Ejemplo 1: Documentar un nuevo servicio Clean Architecture

**Input del usuario**:
> "Documenta el nuevo servicio de notificaciones en internal/notification/"

**Proceso**:

1. Identificar que es un módulo Clean Architecture
2. Revisar estructura: domain, ports, application, infrastructure
3. Consultar `references/go_documentation_standards.md` para patrones Clean Architecture
4. Consultar `references/module_documentation_checklist.md` para verificar completitud
5. Generar documentación para cada capa

**Output esperado**:

```markdown
# Notification Service Documentation

## Overview

El servicio de notificaciones maneja el envío multi-canal (Email, Slack, Push) usando el patrón Strategy.

## Architecture

Módulo ubicado en `internal/notification/` siguiendo Clean Architecture:

```
internal/notification/
├── domain/
│   └── entity.go              # Notification entity
├── ports/
│   ├── service.go             # Service interface
│   └── repository.go          # Repository interface
├── application/
│   └── service.go             # Use cases
└── infrastructure/
    ├── http/
    │   └── handler.go         # HTTP endpoints
    └── db/
        └── gorm_repository.go # Repository implementation
```

## Domain Entity

### Notification

Representa una notificación a enviar con múltiples canales.

```go
// Notification represents a message to be sent through one or more channels.
// It supports Email, Slack, and Push notifications simultaneously.
type Notification struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null;index"`
    Channel   string    `gorm:"type:varchar(20);not null"` // email|slack|push
    Subject   string    `gorm:"type:varchar(255)"`
    Body      string    `gorm:"type:text"`
    Status    string    `gorm:"type:varchar(20);default:'pending'"` // pending|sent|failed
    SentAt    *time.Time
    CreatedAt time.Time
}
```

## Ports (Interfaces)

### NotificationService

```go
// NotificationService defines operations for managing notifications.
// Use cases: Send single notification, send bulk, retry failed notifications.
type NotificationService interface {
    // SendNotification sends a notification through specified channels.
    // Parameters:
    //   - ctx: Context for cancellation and timeouts
    //   - notification: Notification entity with UserID, Channel, Subject, Body
    // Returns:
    //   - error: if notification fails to enqueue
    SendNotification(ctx context.Context, notification *Notification) error
    
    // SendBulk sends multiple notifications efficiently.
    // Uses domain events for async processing.
    SendBulk(ctx context.Context, notifications []*Notification) error
}
```

## Application Service

### NotificationServiceImpl

Implementa use cases de notificaciones:

- **Enviar individual**: `SendNotification()` - Síncrono
- **Envío masivo**: `SendBulk()` - Asíncrono con eventos de dominio
- **Reintentos**: Manejo automático de fallos

Eventos de dominio publicados:
- `eventDomain.GetEventNameToCreate()` - Cuando se crea notificación
- `eventDomain.GetEventNameToUpdate()` - Cuando cambia estado

## HTTP Handler

### POST /api/v1/notifications

**Propósito**: Enviar una notificación

**Autenticación**: JWT + Level based authorization

**Request**:
```json
{
  "user_id": 123,
  "channel": "email",
  "subject": "Bienvenido",
  "body": "Mensaje de bienvenida"
}
```

**Response 200**:
```json
{
  "data": {
    "id": 456,
    "status": "pending"
  },
  "message": "Notificación encolada"
}
```

## Usage Examples

### Send Email Notification

```go
service := notificationService.NewService(db, eventBus)

notif := &notificationDomain.Notification{
    UserID:  123,
    Channel: "email",
    Subject: "Recordatorio de vacaciones",
    Body:    "Tienes vacaciones pendientes",
}

err := service.SendNotification(context.Background(), notif)
```

### Send Bulk Notifications

```go
notifications := []*notificationDomain.Notification{
    {UserID: 1, Channel: "slack", Subject: "Aviso", Body: "..."},
    {UserID: 2, Channel: "slack", Subject: "Aviso", Body: "..."},
}

err := service.SendBulk(context.Background(), notifications)
// Publica domain events para procesamiento asíncrono
```

## Database

- **Database**: Principal (`sensesho_api`)
- **Table**: `notifications`
- **Indexes**: `user_id`, `status`, `created_at`

## Integration with EventBus

Las notificaciones publican eventos de dominio:

```go
service.eventBus.Publish(eventDomain.NewDomainEvent(
    eventDomain.GetEventNameToCreate(),
    notification.ID,
    time.Now(),
    notification,
    log,
))
```

Handlers reaccionan a:
- Logging de notificaciones enviadas
- Auditoría de fallos
- Métricas de delivery rate

## Related Modules

- `internal/event/` - Event bus para domain events
- `internal/backend/user/` - User management para obtener user data
- `task/` - Cron jobs para reintentos de notificaciones fallidas
```

---

### Ejemplo 2: Documentar endpoint de API Legacy

**Input del usuario**:
> "Genera documentación de API para el endpoint de perfiles"

**Proceso**:

1. Identificar que es un módulo Legacy MVC
2. Revisar `controllers/Profile/` para endpoints
3. Verificar routes en `routes/routes.go`
4. Consultar `references/api_endpoint_template.md`
5. Verificar autorización en sistema de niveles

**Output esperado**:

```markdown
# Profile API Documentation

## Overview

API para gestión de perfiles de empleados en el sistema Reverence Hotels.

## Authentication

Todos los endpoints requieren:
- **Header**: `Authorization: Bearer {jwt_token}`
- **Authorization**: Level-based privileges

## Endpoints

### GET /api/v1/profile

**Propósito**: Obtener perfil del usuario autenticado

**Level Required**: Cualquier nivel autenticado

**Request Headers**:
```json
{
  "Authorization": "Bearer eyJhbGc..."
}
```

**Response 200 OK**:
```json
{
  "data": {
    "id": 123,
    "user_id": 456,
    "name": "Juan Pérez",
    "surname": "García",
    "email": "juan.perez@reverencehotels.com",
    "phone": "+34 600 123 456",
    "position": "Recepcionista",
    "department": "Front Office",
    "hotel_id": 1,
    "hire_date": "2020-03-15T00:00:00Z",
    "avatar_url": "https://api.example.com/avatars/123.jpg",
    "created_at": "2020-03-15T10:30:00Z",
    "updated_at": "2024-01-20T14:22:00Z"
  },
  "message": "Profile retrieved successfully"
}
```

**Response 401 Unauthorized**:
```json
{
  "error": "Unauthorized - Invalid or missing token"
}
```

**Response 404 Not Found**:
```json
{
  "error": "Profile not found"
}
```

**Database**: Principal (`sensesho_api`)

**Example**:
```bash
curl -X GET http://localhost:1331/api/v1/profile \\
  -H "Authorization: Bearer $TOKEN"
```

---

### PUT /api/v1/profile

**Propósito**: Actualizar perfil del usuario autenticado

**Level Required**: Level con privilegio Write en Form "profile"

**Request Headers**:
```json
{
  "Authorization": "Bearer eyJhbGc...",
  "Content-Type": "application/json"
}
```

**Request Body**:
```json
{
  "name": "Juan Pedro",
  "surname": "García López",
  "phone": "+34 600 999 888",
  "avatar": "base64_encoded_image..."
}
```

**Response 200 OK**:
```json
{
  "data": {
    "id": 123,
    "name": "Juan Pedro",
    "surname": "García López",
    "phone": "+34 600 999 888",
    "avatar_url": "https://api.example.com/avatars/123-updated.jpg",
    "updated_at": "2024-01-20T15:30:00Z"
  },
  "message": "Profile updated successfully"
}
```

**Response 400 Bad Request**:
```json
{
  "error": "Validation failed",
  "details": {
    "phone": "Invalid phone number format"
  }
}
```

**Response 403 Forbidden**:
```json
{
  "error": "Insufficient privileges - Write access required"
}
```

**Database**: Principal (`sensesho_api`)

**Example**:
```bash
curl -X PUT http://localhost:1331/api/v1/profile \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "Juan Pedro",
    "phone": "+34 600 999 888"
  }'
```

---

### GET /api/v1/profile/holidays

**Propósito**: Obtener balance de vacaciones del empleado

**Level Required**: Level con privilegio Read en Form "profile-holidays"

**Response 200 OK**:
```json
{
  "data": {
    "total_days": 30,
    "used_days": 15,
    "pending_days": 5,
    "available_days": 10,
    "year": 2024
  },
  "message": "Holidays balance retrieved"
}
```

**Authorization Check**:
- Middleware verifica si request path coincide con `Form.PathAPI`
- Path "profile-holidays" requiere privilegio Read
- Si no tiene privilegio, retorna 403 Forbidden

**Example**:
```bash
curl -X GET http://localhost:1331/api/v1/profile/holidays \\
  -H "Authorization: Bearer $TOKEN"
```

## Related Documentation

- **CLAUDE.md**: Authorization system details
- **docs/07-seguridad.md**: Level-based access control
- **services/Profile/**: Business logic implementation
- **models/Profile/**: GORM models
```

---

### Ejemplo 3: Crear guía de usuario para nuevos desarrolladores

**Input del usuario**:
> "Crea una guía para que un nuevo desarrollador añada un endpoint al sistema"

**Proceso**:

1. Determinar si se usará Legacy MVC o Clean Architecture
2. Consultar `references/user_guide_patterns.md` para estructura
3. Consultar `references/architecture_patterns_reference.md` para patrones
4. Incluir ejemplos completos de código
5. Proporcionar checklist de validación

**Output esperado**:

```markdown
# Guía: Añadir un Nuevo Endpoint a Reverence Hotels API

## Objetivo

Esta guía explica cómo añadir un nuevo endpoint API al sistema, siguiendo los patrones de arquitectura del proyecto (Legacy MVC o Clean Architecture).

## Prerrequisitos

- Go 1.21+ instalado
- Conocimiento básico del framework Echo v4
- Familiaridad con GORM v2
- Acceso a las 3 bases de datos (Principal, SII, Products)

## Decisión: ¿Legacy MVC o Clean Architecture?

### Usar Clean Architecture (`internal/`) cuando:
- ✅ Es un módulo de negocio nuevo
- ✅ Requiere testing extenso
- ✅ Tiene lógica de dominio compleja
- ✅ Necesita eventos de dominio
- ✅ Será mantenido a largo plazo

### Usar Legacy MVC cuando:
- ✅ Es CRUD simple
- ✅ Modificación rápida de módulo existente
- ✅ No requiere lógica compleja

---

## Opción 1: Clean Architecture (Recomendado para nuevos módulos)

### Paso 1: Crear estructura de directorios

```bash
mkdir -p internal/{module}/domain
mkdir -p internal/{module}/ports
mkdir -p internal/{module}/application
mkdir -p internal/{module}/infrastructure/http
mkdir -p internal/{module}/infrastructure/db
```

### Paso 2: Definir Domain Entity

Crear `internal/{module}/domain/entity.go`:

```go
package domain

import "time"

// ExampleEntity represents...
type ExampleEntity struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"type:varchar(255);not null"`
    Status    string    `gorm:"type:varchar(20);default:'active'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// TableName specifies the table name for GORM
func (ExampleEntity) TableName() string {
    return "examples"
}
```

### Paso 3: Definir Ports (Interfaces)

Crear `internal/{module}/ports/repository.go`:

```go
package ports

import (
    "context"
    "your-project/internal/{module}/domain"
)

// ExampleRepository defines data access operations
type ExampleRepository interface {
    Create(ctx context.Context, entity *domain.ExampleEntity) error
    GetByID(ctx context.Context, id uint) (*domain.ExampleEntity, error)
    Update(ctx context.Context, entity *domain.ExampleEntity) error
    Delete(ctx context.Context, id uint) error
}
```

Crear `internal/{module}/ports/service.go`:

```go
package ports

import "context"

// ExampleService defines business logic operations
type ExampleService interface {
    CreateExample(ctx context.Context, name string) (*domain.ExampleEntity, error)
    GetExample(ctx context.Context, id uint) (*domain.ExampleEntity, error)
}
```

### Paso 4: Implementar Repository

Crear `internal/{module}/infrastructure/db/gorm_repository.go`:

```go
package db

import (
    "context"
    "gorm.io/gorm"
    "your-project/internal/{module}/domain"
    "your-project/internal/{module}/ports"
)

type gormExampleRepository struct {
    db *gorm.DB
}

// NewExampleRepository creates a new GORM repository
func NewExampleRepository(db *gorm.DB) ports.ExampleRepository {
    return &gormExampleRepository{db: db}
}

func (r *gormExampleRepository) Create(ctx context.Context, entity *domain.ExampleEntity) error {
    return r.db.WithContext(ctx).Create(entity).Error
}

func (r *gormExampleRepository) GetByID(ctx context.Context, id uint) (*domain.ExampleEntity, error) {
    var entity domain.ExampleEntity
    err := r.db.WithContext(ctx).First(&entity, id).Error
    if err != nil {
        return nil, err
    }
    return &entity, nil
}

// ... implementar Update y Delete
```

### Paso 5: Implementar Application Service

Crear `internal/{module}/application/service.go`:

```go
package application

import (
    "context"
    "your-project/internal/{module}/domain"
    "your-project/internal/{module}/ports"
    "your-project/internal/event/domain" as eventDomain
)

type exampleService struct {
    repo     ports.ExampleRepository
    eventBus eventDomain.EventBus
}

// NewExampleService creates a new application service
func NewExampleService(
    repo ports.ExampleRepository,
    eventBus eventDomain.EventBus,
) ports.ExampleService {
    return &exampleService{
        repo:     repo,
        eventBus: eventBus,
    }
}

func (s *exampleService) CreateExample(ctx context.Context, name string) (*domain.ExampleEntity, error) {
    entity := &domain.ExampleEntity{
        Name:   name,
        Status: "active",
    }
    
    if err := s.repo.Create(ctx, entity); err != nil {
        return nil, err
    }
    
    // Publish domain event
    s.eventBus.Publish(eventDomain.NewDomainEvent(
        eventDomain.GetEventNameToCreate(),
        entity.ID,
        time.Now(),
        entity,
        nil,
    ))
    
    return entity, nil
}
```

### Paso 6: Implementar HTTP Handler

Crear `internal/{module}/infrastructure/http/handler.go`:

```go
package http

import (
    "net/http"
    "your-project/internal/{module}/ports"
    "github.com/labstack/echo/v4"
)

type exampleHandler struct {
    service ports.ExampleService
}

// NewExampleHandler creates a new HTTP handler
func NewExampleHandler(service ports.ExampleService) *exampleHandler {
    return &exampleHandler{service: service}
}

// CreateExample handles POST /api/v1/examples
func (h *exampleHandler) CreateExample(c echo.Context) error {
    var req struct {
        Name string `json:"name" validate:"required"`
    }
    
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    entity, err := h.service.CreateExample(c.Request().Context(), req.Name)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create example",
        })
    }
    
    return c.JSON(http.StatusCreated, map[string]interface{}{
        "data":    entity,
        "message": "Example created successfully",
    })
}
```

### Paso 7: Registrar Ruta

Editar `routes/routes.go`:

```go
import (
    examplehttp "your-project/internal/{module}/infrastructure/http"
)

func InitRoutes() *echo.Echo {
    // ... código existente
    
    // Inicializar handler de ejemplo
    exampleHandler := examplehttp.NewExampleHandler(exampleService)
    
    // Registrar ruta
    api.POST("/examples", exampleHandler.CreateExample)
    api.GET("/examples/:id", exampleHandler.GetExample)
    
    return e
}
```

---

## Opción 2: Legacy MVC (Para módulos simples)

### Paso 1: Crear Model

Crear `models/Example/example.go`:

```go
package Example

import (
    "time"
    "gorm.io/gorm"
)

type Example struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"type:varchar(255);not null" json:"name"`
    Status    string    `gorm:"type:varchar(20);default:'active'" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Paso 2: Crear Service

Crear `services/Example/example.go`:

```go
package Example

import (
    "gorm.io/gorm"
    "your-project/models/Example"
)

type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) Create(name string) (*ExampleModel.Example, error) {
    example := &ExampleModel.Example{
        Name:   name,
        Status: "active",
    }
    
    if err := s.db.Create(example).Error; err != nil {
        return nil, err
    }
    
    return example, nil
}
```

### Paso 3: Crear Controller

Crear `controllers/Example/example.go`:

```go
package Example

import (
    "net/http"
    "your-project/services/Example"
    "github.com/labstack/echo/v4"
)

type Controller struct {
    service *Example.Service
}

func NewController(service *Example.Service) *Controller {
    return &Controller{service: service}
}

func (ctrl *Controller) Create(c echo.Context) error {
    var req struct {
        Name string `json:"name" validate:"required"`
    }
    
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request",
        })
    }
    
    example, err := ctrl.service.Create(req.Name)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create",
        })
    }
    
    return c.JSON(http.StatusCreated, map[string]interface{}{
        "data":    example,
        "message": "Created successfully",
    })
}
```

### Paso 4: Registrar Ruta

Editar `routes/routes.go`:

```go
import (
    exampleControllers "your-project/controllers/Example"
    exampleServices "your-project/services/Example"
)

func InitRoutes() *echo.Echo {
    // ... código existente
    
    // Legacy MVC pattern
    exampleService := exampleServices.NewService(db)
    exampleController := exampleControllers.NewController(exampleService)
    
    api.POST("/examples", exampleController.Create)
    
    return e
}
```

---

## Consideraciones Importantes

### Multi-Database

Si tu módulo necesita acceder a bases de datos específicas:

```go
// Para SII (facturación)
dbSII := // conexión a reverence_sii

// Para Products (economato)
dbProducts := // conexión a economato

// Para Principal (default)
db := // conexión a sensesho_api
```

### Authorization

Para proteger endpoints con level-based authorization:

1. **Crear Form** en base de datos (si no existe):
```go
// Form para el módulo
form := &models.Form{
    Name:    "examples",
    PathAPI: "examples|examples-detail",
}
```

2. **Asignar privilegios** a Levels:
```go
privilege := &models.LevelPrivilege{
    LevelID:  levelID,
    FormID:   formID,
    Read:     true,
    Write:    false,
}
```

3. **Middleware verificará automáticamente** si el usuario tiene acceso

### Database Migrations

Si necesitas nuevas tablas, agregar a `migration/base.go`:

```go
// AutoMigrate aggregates all models
db.AutoMigrate(
    // ... modelos existentes
    &ExampleModel.Example{}, // <- agregar aquí
)
```

---

## Testing

### Prueba manual con curl:

```bash
# Obtener token primero
TOKEN=$(curl -X POST http://localhost:1331/api/v1/login \\
  -H "Content-Type: application/json" \\
  -d '{"email":"user@example.com","password":"pass"}' \\
  | jq -r '.data.token')

# Crear ejemplo
curl -X POST http://localhost:1331/api/v1/examples \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"name":"Test Example"}'

# Obtener ejemplo
curl -X GET http://localhost:1331/api/v1/examples/1 \\
  -H "Authorization: Bearer $TOKEN"
```

### Pruebas unitarias:

```go
// internal/{module}/application/service_test.go
func TestCreateExample(t *testing.T) {
    // Mock repository
    mockRepo := &MockRepository{}
    mockEventBus := &MockEventBus{}
    
    service := NewExampleService(mockRepo, mockEventBus)
    
    entity, err := service.CreateExample(context.Background(), "Test")
    
    assert.NoError(t, err)
    assert.Equal(t, "Test", entity.Name)
}
```

---

## Checklist de Validación

Antes de considerar el endpoint completo:

- [ ] **Estructura**: Código sigue patrón elegido (Clean Arch o MVC)
- [ ] **Modelos**: GORM models con tags apropiados
- [ ] **Servicio**: Lógica de negocio separada de handlers
- [ ] **Handlers**: Echo handlers con validación
- [ ] **Rutas**: Registradas en `routes/routes.go`
- [ ] **Base de datos**: Tabla agregada a migrations
- [ ] **Auth**: Form creado en base de datos si requiere autorización
- [ ] **Documentación**: Comentarios godoc incluidos
- [ ] **Testing**: Pruebas manuales ejecutadas exitosamente
- [ ] **Errores**: Manejo apropiado de errores HTTP
- [ ] **Validación**: Request validation implementada
- [ ] **Eventos**: Domain events publicados si aplica (Clean Arch)

---

## Troubleshooting

### Error: "Table doesn't exist"

**Causa**: Migration no ejecutada

**Solución**:
```bash
go run init.go --migrate=yes
```

### Error: "403 Forbidden"

**Causa**: Usuario no tiene privilegio en el Form

**Solución**:
1. Verificar que Form existe en tabla `forms`
2. Verificar que Level tiene `LevelPrivilege` para el Form
3. Verificar que `Form.PathAPI` incluye el path del endpoint

### Error: "GORM v1 migration issues"

**Causa**: Foreign keys inconsistentes

**Solución**: Revisar `migration/base.go` y ejecutar migrate con flags

### Error: "Domain event not published"

**Causa**: EventBus no inyectado en service

**Solución**: Verificar `cmd/server/backend/` wiring

---

## Recursos Adicionales

- **CLAUDE.md**: Arquitectura del proyecto
- **docs/01-arquitectura.md**: Patrones arquitectónicos detallados
- **docs/07-seguridad.md**: Sistema de autorización
- **internal/event/**: Ejemplos de domain events
```

---

## Presentación de Resultados

Al completar documentación:

1. **Formato**: Markdown estructurado con secciones claras
2. **Idioma**: Español (idioma principal del proyecto)
3. **Ejemplos**: Siempre incluir código ejecutable y comandos probados
4. **Referencias**: Enlaces a documentación relacionada
5. **Actualización**: Indicar si actualiza archivo existente o crea nuevo

**Ejemplo de resumen**:

```
Documentación generada para servicio de notificaciones.
- Archivo: docs/notification-service.md
- Secciones: Overview, Architecture, Usage Examples, API Endpoints
- Formato: Markdown con ejemplos de código Go
- Relacionado con: internal/event/, internal/backend/user/
```

## Troubleshooting

### Problema: Documentación inconsistente con código actual

**Síntoma**: Ejemplos de código no funcionan o tienen errores de sintaxis

**Causa**: Código cambió pero documentación no se actualizó

**Solución**:

1. Leer código fuente actual para verificar estructuras
2. Actualizar documentación con firmas correctas
3. Probar ejemplos antes de incluir en docs

---

### Problema: Falta contexto sobre arquitectura híbrida

**Síntoma**: No está claro cuándo usar MVC vs Clean Architecture

**Causa**: Patrones del proyecto no documentados en la skill

**Solución**:

Consultar `references/architecture_patterns_reference.md` y `CLAUDE.md` sección "Adding New Features" para criterios de decisión.

---

### Problema: Referencias a bases de datos incorrectas

**Síntoma**: Documentación menciona base de datos equivocada para el módulo

**Causa**: No se verificó qué base de datos usa el módulo

**Solución**:

Verificar en código:
- Principal DB → `db` variable
- SII DB → `dbSII` variable  
- Products DB → `dbProducts` variable

## Consideraciones Especiales

### Idioma del Proyecto

- **Documentación**: Español
- **Código**: Inglés (variables, funciones, comentarios)
- **Comentarios de código**: Inglés (estándar Go)
- **Documentación de usuario**: Español

### Convenciones de Go

- **Godoc comments**: Comentarios que empiezan con el nombre del elemento
- **Exported names**: PascalCase para público, camelCase para privado
- **Error handling**: Siempre retornar error, nunca nil-check sin log
- **Context**: Usar `context.Context` en todas las operaciones de DB/Servicios

### Multi-Database

Siempre especificar cuál de las 3 bases de datos usa el módulo:
- **Principal** (`sensesho_api`): Users, profiles, HR
- **SII** (`reverence_sii`): Facturación electrónica
- **Products** (`economato`): Artículos, proveedores

## Mejoras Futuras (Roadmap)

- [ ] Incluir generación automática de OpenAPI/Swagger specs
- [ ] Templates para documentación de integraciones externas (SII, Porta Sigma)
- [ ] Scripts para validar que documentación coincide con código actual
- [ ] Ejemplos de documentación para cron jobs (task/)
- [ ] Templates para documentación de migraciones de datos