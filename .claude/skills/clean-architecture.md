---
name: go-clean-architecture
description: Esta skill debe usarse cuando el usuario necesite implementar o refactorizar módulos usando Clean Architecture/Hexagonal en Go. Se activa con peticiones como "crea un módulo con clean architecture", "refactoriza este controller a clean architecture", "implementa un nuevo feature usando hexagonal architecture", o "agrega un repositorio GORM siguiendo clean architecture".
license: MIT
version: 1.0.0
author: Reverence Hotels Development Team
category: language
tags: [go, clean-architecture, hexagonal, gorm, echo, design-patterns]---

# Go Clean Architecture

Skill especializada en la implementación y refactorización de módulos usando Clean Architecture (Hexagonal) en Go para el proyecto Reverence Hotels API. Proporciona patrones estructurales, convenciones de nomenclatura y workflows para crear código mantenible y testeable siguiendo principios de separación de dependencias.

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:

- Se deba crear un nuevo módulo de negocio en el proyecto
- Se requiera refactorizar código legacy MVC a Clean Architecture
- Se necesite implementar nuevos features con arquitectura hexagonal
- Se deban agregar repositorios, handlers o servicios siguiendo patrones establecidos
- Se requiera mantener coherencia con módulos existentes en `internal/`

Triggers comunes:

- "Crea un módulo de gestión de vacaciones con clean architecture"
- "Refactoriza este controller para usar repositorios"
- "Implementa un endpoint para notificaciones usando hexagonal architecture"
- "Agrega un nuevo caso de uso al módulo user"
- "Crea un repositorio GORM para la entidad Document"

## Workflow Principal

### 1. Análisis Inicial

Antes de proceder:

1. Identificar si el módulo debe usar Clean Architecture o Legacy MVC
   - **Clean Architecture**: Para features nuevos, lógica de negocio compleja, o módulos que requieren testing
   - **Legacy MVC**: Para CRUD simple, modificaciones menores en módulos estabilizados
2. Verificar módulos existentes en `internal/` para seguir convenciones
3. Determinar entidades de dominio y sus relaciones
4. Identificar dependencias externas (bases de datos, APIs, servicios)

### 2. Estructura del Módulo

Para nuevos módulos con Clean Architecture, crear la siguiente estructura en `internal/{module-name}/`:

```
internal/{module-name}/
├── domain/                    # Capa de dominio (core business logic)
│   └── entity.go             # Entidades puras, sin dependencias externas
├── ports/                     # Contratos/interfaces
│   ├── service.go            # Interfaz del servicio de aplicación
│   └── repository.go         # Interfaz del repositorio (puerto secundario)
├── application/              # Casos de uso (orchestration layer)
│   └── service.go            # Implementación del puerto primario
└── infrastructure/           # Implementaciones técnicas
    ├── http/                 # Handlers/adapters (primary adapters)
    │   └── handler.go
    └── db/                   # Repositorios (secondary adapters)
        └── gorm_repository.go
```

### 3. Implementación por Capas

#### Capa de Dominio (`domain/`)

Definir entidades puras con lógica de negocio:

```go
package domain

import (
    "time"
)

type Entity struct {
    ID        uint
    CreatedAt time.Time
    UpdatedAt time.Time
    // Campos de negocio
}

func (e *Entity) Validate() error {
    // Lógica de validación de negocio
}
```

**Reglas**:
- Sin importaciones de paquetes externos (excepto `time`, `errors`, etc.)
- Lógica de negocio encapsulada en métodos
- Validaciones de dominio aquí, no en infrastructure

#### Capa de Puertos (`ports/`)

Definir interfaces que desacoplan capas:

```go
package ports

type Repository interface {
    Create(ctx context.Context, entity *domain.Entity) error
    GetByID(ctx context.Context, id uint) (*domain.Entity, error)
    Update(ctx context.Context, entity *domain.Entity) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, filter Filter) ([]*domain.Entity, error)
}

type Service interface {
    CreateEntity(ctx context.Context, req CreateRequest) (*domain.Entity, error)
    GetEntity(ctx context.Context, id uint) (*domain.Entity, error)
    UpdateEntity(ctx context.Context, id uint, req UpdateRequest) (*domain.Entity, error)
}
```

**Reglas**:
- Interfaces en `ports/`, implementaciones en `infrastructure/`
- Dependencias siempre apuntan hacia adentro (dependency rule)
- Usar `context.Context` en todos los métodos

#### Capa de Aplicación (`application/`)

Implementar casos de uso con orquestación:

```go
package application

type service struct {
    repo   ports.Repository
    eventBus event.EventBus
    log     *log.Logger
}

func NewService(repo ports.Repository, eventBus event.EventBus, log *log.Logger) ports.Service {
    return &service{
        repo:      repo,
        eventBus:  eventBus,
        log:       log,
    }
}

func (s *service) CreateEntity(ctx context.Context, req CreateRequest) (*domain.Entity, error) {
    // 1. Validar request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // 2. Crear entidad de dominio
    entity := &domain.Entity{
        // ...
    }
    
    if err := entity.Validate(); err != nil {
        return nil, err
    }
    
    // 3. Persistir usando repositorio
    if err := s.repo.Create(ctx, entity); err != nil {
        s.log.Error(err.Error())
        return nil, err
    }
    
    // 4. Publicar evento de dominio
    s.eventBus.Publish(eventDomain.NewDomainEvent(
        eventDomain.GetEventNameToCreate(),
        entity.ID,
        time.Now(),
        entity,
        s.log,
    ))
    
    return entity, nil
}
```

**Reglas**:
- Publicar eventos de dominio después de operaciones write
- Manejar transacciones cuando sea necesario
- Lógica de orquestación, no detalles técnicos

#### Capa de Infraestructura - Repositorio (`infrastructure/db/`)

Implementar acceso a datos con GORM:

```go
package db

import (
    "context"
    "gorm.io/gorm"
)

type gormRepository struct {
    db *gorm.DB
}

func NewGormRepository(db *gorm.DB) ports.Repository {
    return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, entity *domain.Entity) error {
    return r.db.WithContext(ctx).Create(entity).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
    var entity domain.Entity
    err := r.db.WithContext(ctx).First(&entity, id).Error
    if err != nil {
        return nil, err
    }
    return &entity, nil
}

func (r *gormRepository) Update(ctx context.Context, entity *domain.Entity) error {
    return r.db.WithContext(ctx).Save(entity).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&domain.Entity{}, id).Error
}

func (r *gormRepository) List(ctx context.Context, filter Filter) ([]*domain.Entity, error) {
    var entities []*domain.Entity
    query := r.db.WithContext(ctx)
    
    // Aplicar filtros si existen
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    
    err := query.Find(&entities).Error
    return entities, err
}
```

**Reglas**:
- Usar `WithContext(ctx)` para queries GORM
- Usar la conexión de base de datos correcta (`db`, `dbSII`, `dbProducts`)
- Mapear entre modelos de dominio y GORM si es necesario
- Manejar errores apropiadamente

#### Capa de Infraestructura - Handler (`infrastructure/http/`)

Implementar handler Echo:

```go
package http

import (
    "net/http"
    "github.com/labstack/echo/v4"
)

type handler struct {
    service ports.Service
}

func NewHandler(service ports.Service) *handler {
    return &handler{service: service}
}

func (h *handler) RegisterRoutes(e *echo.Echo) {
    restricted := e.Group("/api/v1")
    restricted.POST("/{module}", h.create)
    restricted.GET("/{module}/:id", h.getByID)
    restricted.PUT("/{module}/:id", h.update)
    restricted.DELETE("/{module}/:id", h.delete)
}

func (h *handler) create(c echo.Context) error {
    ctx := c.Request().Context()
    
    var req CreateRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid request",
        })
    }
    
    entity, err := h.service.CreateEntity(ctx, req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": err.Error(),
        })
    }
    
    return c.JSON(http.StatusCreated, entity)
}

func (h *handler) getByID(c echo.Context) error {
    ctx := c.Request().Context()
    
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid id",
        })
    }
    
    entity, err := h.service.GetEntity(ctx, uint(id))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return c.JSON(http.StatusNotFound, map[string]string{
                "error": "not found",
            })
        }
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": err.Error(),
        })
    }
    
    return c.JSON(http.StatusOK, entity)
}
```

**Reglas**:
- Extraer contexto de `c.Request().Context()`
- Validar parámetros antes de pasar al servicio
- Mapear errores a códigos HTTP apropiados
- Usar middleware de autorización (level-based) en `routes/routes.go`

### 4. Registrar Rutas

Agregar rutas en `routes/routes.go`:

```go
// Importar handler
"{module}Http": internal{module}http,

// En InitRoutes()
{module}Handler := {module}Http.NewHandler({module}Service)
{module}Handler.RegisterRoutes(e)
```

### 5. Wire Up en cmd/server/backend/

Inyectar dependencias en `cmd/server/backend/main.go`:

```go
// Repositorios
{module}Repository := {module}Db.NewGormRepository(db)

// Servicio de aplicación
{module}Service := {module}Application.NewService(
    {module}Repository,
    eventBus,
    logger,
)
```

### 6. Validación

Verificar que:

- [ ] Estructura de directorios sigue el patrón hexagonal
- [ ] No hay dependencias de capas externas hacia internas (dependency rule)
- [ ] Entidades de dominio no importan paquetes de infraestructura
- [ ] Interfaces están en `ports/`, implementaciones en `infrastructure/`
- [ ] Context se pasa correctamente a través de todas las capas
- [ ] Eventos de dominio se publican en operaciones write
- [ ] Tests cubren lógica de negocio y casos de uso

## Patrones y Convenciones del Proyecto

### Nomenclatura

**Directorios**: kebab-case (`user-management`, `document-to-sign`)

**Archivos Go**: snake_case (`entity.go`, `gorm_repository.go`)

**Paquetes Go**: sin prefijos, nombre descriptivo
- `package domain`
- `package ports`
- `package application`
- `package db`
- `package http`

**Interfaces**: sufijo `-er` o nombre descriptivo
```go
type Repository interface { }
type Service interface { }
type Validator interface { }
```

**Estructuras**: PascalCase
```go
type UserService struct { }
type CreateRequest struct { }
```

### Inyección de Dependencias

**Constructor con dependencias explícitas**:

```go
// ✅ Correcto
func NewService(repo ports.Repository, eventBus event.EventBus, log *log.Logger) ports.Service {
    return &service{
        repo:     repo,
        eventBus: eventBus,
        log:      log,
    }
}

// ❌ Incorrecto - dependencias ocultas
func NewService() *service {
    return &service{
        repo: NewDefaultRepository(), // acoplado
    }
}
```

### Manejo de Errores

**Wrappear errores con contexto**:

```go
if err := s.repo.Create(ctx, entity); err != nil {
    s.log.Error("failed to create entity: %w", err)
    return fmt.Errorf("create entity: %w", err)
}
```

**Usar errors.Is() para comparar**:

```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, ErrEntityNotFound
}
```

### Testing

**Test structure**:

```
internal/{module}/
├── application/
│   ├── service.go
│   └── service_test.go      # Tests de casos de uso
├── infrastructure/
│   ├── db/
│   │   ├── gorm_repository.go
│   │   └── gorm_repository_test.go  # Tests de integración con DB
│   └── http/
│       ├── handler.go
│       └── handler_test.go  # Tests HTTP handlers
```

**Tests de servicio (unit tests)**:

```go
func TestService_CreateEntity(t *testing.T) {
    // Setup
    mockRepo := &MockRepository{}
    mockEventBus := &MockEventBus{}
    service := NewService(mockRepo, mockEventBus, log.NewNop())
    
    tests := []struct {
        name    string
        req     CreateRequest
        setup   func()
        wantErr bool
    }{
        {
            name: "success",
            req:  CreateRequest{Name: "test"},
            setup: func() {
                mockRepo.CreateFunc = func(ctx context.Context, entity *domain.Entity) error {
                    return nil
                }
            },
            wantErr: false,
        },
        // más casos...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            entity, err := service.CreateEntity(context.Background(), tt.req)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateEntity() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !tt.wantErr && entity == nil {
                t.Error("expected entity, got nil")
            }
        })
    }
}
```

### Eventos de Dominio

**Publicar eventos en operaciones write**:

```go
s.eventBus.Publish(eventDomain.NewDomainEvent(
    eventDomain.GetEventNameToCreate(),
    entity.ID,
    time.Now(),
    entity,
    s.log,
))
```

**Nombres de eventos**: seguir convención del módulo
- `user_created`, `user_updated`
- `document_signed`, `signature_completed`

## Ejemplos de Uso

### Ejemplo 1: Crear módulo completo

**Input del usuario**:
> "Crea un módulo de gestión de documentos con clean architecture"

**Proceso**:

1. Analizar requisitos:
   - Entidad: `Document` con campos (id, name, type, url, status)
   - Casos de uso: create, get, list, update, delete, change_status
   - Base de datos: usar `db` (principal)
   - Necesita eventos: document_created, document_status_changed

2. Crear estructura:
   ```
   internal/document/
   ├── domain/
   │   └── entity.go
   ├── ports/
   │   ├── service.go
   │   └── repository.go
   ├── application/
   │   └── service.go
   └── infrastructure/
       ├── http/
       │   └── handler.go
       └── db/
           └── gorm_repository.go
   ```

3. Implementar por capas (domain → ports → application → infrastructure)

4. Registrar rutas en `routes/routes.go`

5. Wire up en `cmd/server/backend/main.go`

**Output esperado**:

```
Módulo Document creado con Clean Architecture:
✓ domain/entity.go - Entidad Document con validaciones
✓ ports/service.go - Interfaz con 6 métodos (Create, GetByID, List, Update, Delete, ChangeStatus)
✓ ports/repository.go - Interfaz de repositorio GORM
✓ application/service.go - Casos de uso con eventos de dominio
✓ infrastructure/http/handler.go - 6 endpoints Echo
✓ infrastructure/db/gorm_repository.go - Repositorio con GORM
✓ Rutas registradas en /api/v1/documents
✓ Dependencias inyectadas en cmd/server/backend/

Endpoints disponibles:
- POST /api/v1/documents (crear documento)
- GET /api/v1/documents/:id (obtener por ID)
- GET /api/v1/documents (listar con filtros)
- PUT /api/v1/documents/:id (actualizar)
- DELETE /api/v1/documents/:id (eliminar)
- PATCH /api/v1/documents/:id/status (cambiar estado)

Eventos de dominio:
- document_created
- document_updated
- document_status_changed
- document_deleted
```

---

### Ejemplo 2: Refactorizar controller legacy

**Input del usuario**:
> "Refactoriza el controller Signature a clean architecture"

**Proceso**:

1. Analizar código legacy en `controllers/Signatures/`:
   - Identificar lógica de negocio
   - Extraer entidades de dominio
   - Identificar dependencies

2. Crear nueva estructura en `internal/signature/`

3. Migrar gradualmente:
   - Mover lógica de negocio a `domain/entity.go`
   - Crear interfaces en `ports/`
   - Implementar service en `application/service.go`
   - Reemplazar controller con handler en `infrastructure/http/`

4. Actualizar rutas para apuntar al nuevo handler

5. Mantener compatibilidad con código existente

**Output esperado**:

```
Refactorización completada de Signature a Clean Architecture:

Migraciones realizadas:
✓ Lógica de negocio extraída a domain/entity.go
✓ Interfaces definidas en ports/
✓ Casos de uso implementados en application/service.go
✓ Handler Echo creado en infrastructure/http/handler.go
✓ Repositorio GORM en infrastructure/db/gorm_repository.go
✓ Eventos de dominio agregados (signature_created, signature_completed)

Beneficios:
- Testabilidad: Service ahora puede testearse sin DB
- Desacoplamiento: Infraestructura aislada de dominio
- Mantenibilidad: Lógica de negocio centralizada

Compatibilidad:
✓ Rutas legacy redirigen a nuevo handler
✓ Responses mantienen mismo formato JSON
✓ Tests legacy adaptados a nueva estructura

Siguiente paso: Actualizar clientes internos para usar nuevos endpoints
```

---

### Ejemplo 3: Agregar feature a módulo existente

**Input del usuario**:
> "Agrega un endpoint para aprobar documentos al módulo Document"

**Proceso**:

1. Verificar estructura existente en `internal/document/`

2. Agregar caso de uso en `ports/service.go`:
   ```go
   ApproveDocument(ctx context.Context, id uint) (*domain.Document, error)
   ```

3. Implementar lógica en `application/service.go`:
   - Validar que documento existe
   - Verificar estado actual (debe ser "pending")
   - Cambiar estado a "approved"
   - Publicar evento `document_approved`
   - Manejar concurrencia (optimistic locking)

4. Agregar método en `infrastructure/http/handler.go`:
   ```go
   func (h *handler) approve(c echo.Context) error {
       // implementation
   }
   ```

5. Registrar ruta en `RegisterRoutes()`:
   ```go
   restricted.PATCH("/documents/:id/approve", h.approve)
   ```

**Output esperado**:

```
Feature agregada al módulo Document:

✓ Caso de uso: ApproveDocument en ports/service.go
✓ Implementación: Lógica de aprobación en application/service.go
  - Valida estado previo (debe ser 'pending')
  - Cambia estado a 'approved'
  - Publica evento document_approved
  - Maneja errores (documento no existe, estado inválido)

✓ Endpoint: PATCH /api/v1/documents/:id/approve
✓ Autorización: Requiere privilegio 'Write' en formulario 'documents'

Ejemplo de uso:
PATCH /api/v1/documents/123/approve
Status: 200 OK
{
  "id": 123,
  "status": "approved",
  "approved_at": "2025-01-20T10:30:00Z",
  "approved_by": 5
}

Evento publicado:
document_approved (entity_id: 123, timestamp: 2025-01-20T10:30:00Z)
```

## Consideraciones Específicas del Proyecto

### Multi-Database Architecture

**Usar la conexión correcta según el módulo**:

```go
// Base de datos principal (users, profiles, HR)
func NewGormRepository(db *gorm.DB) ports.Repository {
    return &gormRepository{db: db}
}

// SII Database (invoicing)
func NewGormSIIRepository(dbSII *gorm.DB) ports.Repository {
    return &gormSIIRepository{db: dbSII}
}

// Products Database (economato)
func NewGormProductsRepository(dbProducts *gorm.DB) ports.Repository {
    return &gormProductsRepository{db: dbProducts}
}
```

**Regla**: Si el módulo trabaja con facturación → usar `dbSII`. Si es productos → usar `dbProducts`. Default → `db`.

### Authorization Integration

**Level-based authorization se maneja en middleware**, no en servicios:

```go
// En routes/routes.go
restricted := e.Group("/api/v1")
restricted.Use(middleware.Authorization(forms.FormDocuments, "write"))
restricted.POST("/documents", handler.create) // Requiere privilegio Write

restricted.Use(middleware.Authorization(forms.FormDocuments, "read"))
restricted.GET("/documents", handler.list) // Requiere privilegio Read
```

**Servicios NO deben validar authorization**, eso es responsabilidad del middleware.

### Domain Events Integration

**Usar EventBus del proyecto**:

```go
import (
    eventDomain "gitlab.com/reverence/hotels/api/internal/event/domain"
)

// Publicar evento después de operación write
s.eventBus.Publish(eventDomain.NewDomainEvent(
    "document_created", // Nombre del evento
    entity.ID,          // ID de la entidad
    time.Now(),         // Timestamp
    entity,             // Payload completo
    s.log,              // Logger
))
```

**Eventos existentes que se pueden emitar**:
- `user_created`, `user_updated`
- `profile_created`, `profile_updated`
- `document_created`, `document_status_changed`
- `invoice_issued`, `invoice_received`

### Cron Jobs Integration

**Si el módulo necesita tareas programadas**, agregar en `task/`:

```go
// task/documents_cron.go
func ProcessPendingDocuments() {
    // Lógica que llama al servicio de aplicación
    service := getDocumentService() // inyectado
    service.ProcessPending(context.Background())
}
```

**Recordar**: Cron jobs solo ejecutan en `pre` y `pro`, no en `local`.

## Troubleshooting

### Problema: Circular dependency

**Síntoma**: Error de importación circular al crear módulo

**Causa**: Una capa interna importa una capa externa

**Solución**:

1. Revisar dirección de dependencias (debe ser hacia adentro)
2. Mover tipos compartidos a `domain/` si ambas capas los necesitan
3. Usar interfaces en `ports/` para romper dependencias directas

**Ejemplo de error**:
```go
// ❌ Error: domain/ importa infrastructure/
package domain
import "myproject/infrastructure/db" // VIOLA dependency rule

// ✅ Correcto: infrastructure/ importa domain/
package db
import "myproject/domain"
```

---

### Problema: GORM model vs domain entity

**Síntoma**: Necesito métodos GORM en la entidad de dominio

**Causa**: Mezclar responsabilidades de persistencia con dominio

**Solución**: Crear modelo GORM separado en infrastructure/db/

```go
// domain/entity.go - Entidad pura
type Document struct {
    ID      uint
    Name    string
    Content string
}

// infrastructure/db/gorm_model.go - Modelo GORM
type GormDocument struct {
    gorm.Model
    DocumentID uint   `gorm:"column:document_id"`
    Name       string `gorm:"column:name"`
    Content    string `gorm:"column:content"`
}

// Mapeo en repositorio
func (r *gormRepository) toDomain(g *GormDocument) *domain.Document {
    return &domain.Document{
        ID:      g.DocumentID,
        Name:    g.Name,
        Content: g.Content,
    }
}
```

---

### Problema: Context no se propaga

**Síntoma**: Timeouts o cancelaciones no funcionan

**Causa**: Context no se pasa a través de todas las capas

**Solución**:

```go
// ✅ Correcto: Context en todas las capas
func (h *handler) create(c echo.Context) error {
    ctx := c.Request().Context() // Extraer context
    entity, err := h.service.CreateEntity(ctx, req) // Pasar context
}

func (s *service) CreateEntity(ctx context.Context, req CreateRequest) (*domain.Entity, error) {
    return s.repo.Create(ctx, entity) // Pasar context
}

func (r *repository) Create(ctx context.Context, entity *domain.Entity) error {
    return r.db.WithContext(ctx).Create(entity).Error // Usar context
}
```

---

### Problema: Tests no compilan

**Síntoma**: Error "undefined" al correr tests

**Causa**: Paquetes no se encuentran o paths incorrectos

**Solución**:

```bash
# Verificar estructura de directorios
ls -la internal/{module}/

# Asegurar que go.mod está actualizado
go mod tidy

# Correr tests con verbose para ver detalles
go test -v ./internal/{module}/...

# Si es problema de imports, verificar package names
head -5 internal/{module}/domain/entity.go
```

---

### Problema: Eventos no se publican

**Síntoma**: EventBus.Publish se llama pero nadie recibe el evento

**Causa**: EventBus no está inicializado o no hay suscriptores

**Solución**:

1. Verificar inyección de EventBus en el servicio
2. Agregar suscriptor si es necesario
3. Verificar que el nombre del evento coincide

```go
// Verificar inyección
service := application.NewService(repo, eventBus, log)

// Suscriptor en cmd/server/backend/
eventBus.Subscribe("document_created", func(e event.DomainEvent) {
    log.Info("Document created: %d", e.EntityID)
})
```

---

### Problema: Handler retorna 500 pero no hay logs

**Síntoma**: Error interno del servidor sin información útil

**Causa**: Error no se está loggeando apropiadamente

**Solución**:

```go
// ✅ Correcto: Loggear antes de retornar error
entity, err := h.service.CreateEntity(ctx, req)
if err != nil {
    h.log.Error("failed to create entity: %w", err)
    return c.JSON(http.StatusInternalServerError, map[string]string{
        "error": "failed to create entity",
    })
}

// O usar middleware de error handling de Echo
// que loggea automáticamente
```

## Referencias Rápidas

### Comandos útiles

```bash
# Crear estructura de módulo
mkdir -p internal/{module}/{domain,ports,application,infrastructure/{http,db}}

# Inicializar módulo
go mod init gitlab.com/reverence/hotels/api
go mod tidy

# Correr tests del módulo
go test -v ./internal/{module}/...

# Formatear código
go fmt ./internal/{module}/...

# Linter
golangci-lint run ./internal/{module}/...
```

### Plantilla de archivo domain/entity.go

```go
package domain

import (
    "errors"
    "time"
)

var (
    ErrEntityNotFound = errors.New("entity not found")
    ErrInvalidInput   = errors.New("invalid input")
)

type Entity struct {
    ID        uint
    CreatedAt time.Time
    UpdatedAt time.Time
    // Campos específicos
}

func (e *Entity) Validate() error {
    if e.ID == 0 {
        return ErrInvalidInput
    }
    return nil
}
```

### Plantilla de archivo ports/repository.go

```go
package ports

import "context"

type Repository interface {
    Create(ctx context.Context, entity *domain.Entity) error
    GetByID(ctx context.Context, id uint) (*domain.Entity, error)
    Update(ctx context.Context, entity *domain.Entity) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, filter Filter) ([]*domain.Entity, error)
}

type Filter struct {
    // Campos de filtro
}
```

### Plantilla de archivo infrastructure/db/gorm_repository.go

```go
package db

import (
    "context"
    "gorm.io/gorm"
)

type gormRepository struct {
    db *gorm.DB
}

func NewGormRepository(db *gorm.DB) ports.Repository {
    return &gormRepository{db: db}
}

// Implementar métodos de ports.Repository
func (r *gormRepository) Create(ctx context.Context, entity *domain.Entity) error {
    return r.db.WithContext(ctx).Create(entity).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
    var entity domain.Entity
    err := r.db.WithContext(ctx).First(&entity, id).Error
    if err != nil {
        return nil, err
    }
    return &entity, nil
}

// ... otros métodos
```

## Consideraciones Especiales

### Rendimiento

- **Optimistic Locking**: Usar `version` field para concurrencia
- **Batch Operations**: Agregar métodos `CreateBatch`, `UpdateBatch` en repositorio
- **Caching**: Considerar caché en capa de application para reads frecuentes

### Seguridad

- **Input Validation**: Validar en handlers ANTES de pasar a servicio
- **SQL Injection**: GORM ya previene esto, pero cuidado con queries raw
- **Sensitive Data**: No loggear datos sensibles (passwords, tokens)

### Compatibilidad

- **Go Version**: Go 1.21+
- **GORM**: v2.52.0
- **Echo**: v4.11.0
- **MySQL**: 8.0+

## Mejoras Futuras (Roadmap)

- [ ] Generador de código para crear scaffolding de módulos automáticamente
- [ ] Soporte para transacciones distribuidas entre múltiples bases de datos
- [ ] Integration con OpenAPI/Swagger para documentación automática de endpoints
- [ ] Metrics y tracing con OpenTelemetry en todas las capas
- [ ] Soporte para gRPC como alternativa a HTTP REST
- [ ] Domain events con mensajería asíncrona (RabbitMQ/Redis Streams)