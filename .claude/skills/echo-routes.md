---
name: echo-routes
description: Esta skill debe usarse cuando el usuario necesite crear, modificar, o diagnosticar rutas en aplicaciones Echo v4. Se activa con peticiones como "crea una nueva ruta en Echo", "añade un endpoint al API", "configura middleware en Echo", o "diagnosticar problemas de rutas".
license: MIT
version: 1.0.0
author: Reverence Hotels Development Team
category: language
tags: [echo, go, http-routes, middleware, rest-api, web-framework]---

# Echo Routes

Skill especializada en el manejo de rutas HTTP con el framework Echo v4 en Go. Proporciona patrones, convenciones y mejores prácticas para la definición, organización y mantenimiento de rutas en aplicaciones que usan Echo, con énfasis en arquitecturas híbridas (MVC + Clean Architecture).

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:
- Se necesite crear nuevos endpoints HTTP en el API
- Se deban organizar rutas en grupos con middleware compartido
- Se requiera configurar middleware (autenticación, autorización, logging, CORS)
- Sea necesario diagnosticar problemas de routing (rutas no encontradas, conflictos de paths)
- Se deban integrar rutas legacy (MVC) con rutas Clean Architecture
- Necesite configuración de rutas para múltiples bases de datos o contextos

Triggers comunes:
- "Crea una nueva ruta POST /api/v1/users"
- "Añade middleware de autenticación a estas rutas"
- "¿Por qué no funciona mi endpoint /api/v1/profiles?"
- "Organiza las rutas del módulo de facturación"
- "Configura ruta para integración PMS con API key"

## Workflow Principal

### 1. Análisis Inicial

Antes de crear o modificar rutas:

1. **Identificar el tipo de ruta**:
   - API restringida (JWT + Authorization) → `/api/v1/*`
   - API pública (login, recovery) → sin JWT
   - API Key protected (integraciones) → `X-API-KEY` header
   - Intranet (portal empleados) → `/intranet/*`

2. **Determinar el patrón arquitectónico**:
   - Legacy MVC → Controller en `controllers/`
   - Clean Architecture → Handler en `internal/{module}/infrastructure/http/`

3. **Verificar prerequisitos**:
   - ¿Existe el controller/handler correspondiente?
   - ¿Está registrado el middleware necesario?
   - ¿El path está disponible (no conflictos)?

### 2. Ejecución

#### Para rutas Legacy MVC:

1. Crear controller en `controllers/{Module}/` si no existe
2. Registrar ruta en `routes/routes.go` usando `e.{Method}()`
3. Añadir middleware específico (JWT, Auth, LevelAccess)
4. Documentar en comentarios el propósito y privilegios requeridos

#### Para rutas Clean Architecture:

1. Crear handler en `internal/{module}/infrastructure/http/handler.go`
2. Registrar ruta en `routes/routes.go` o en `cmd/server/backend/` según módulo
3. Usar middleware de autorización basado en niveles
4. Publicar domain events si aplica

#### Para configuración de middleware:

1. Identificar middleware global en `routes/echo.go`
2. Configurar middleware específico por grupo de rutas
3. Verificar orden de ejecución (middleware se ejecuta en orden inverso al registro)

### 3. Validación

Verificar que:

- [ ] La ruta está registrada correctamente con método HTTP y path
- [ ] El middleware de autenticación/autorización está configurado
- [ ] El handler/controller existe y tiene la firma correcta
- [ ] No hay conflictos de paths con otras rutas
- [ ] La ruta sigue las convenciones de nomenclatura del proyecto
- [ ] Para API restringida: el path tiene un Form asociado en BD

### 4. Output

Presentar resultados:

- Formato: Código Go completo listo para copiar
- Incluir: Comentarios de propósito, parámetros, y privilegios requeridos
- Omitir: Código repetitivo o boilerplate sin contexto

## Patrones de Rutas del Proyecto

### Estructura de Paths

```
/api/v1/{resource}              # API restringida estándar
/api/v1/{resource}/:id          # Recurso específico
/intranet/{feature}             # Portal empleados
/api/v1/pms-{action}            # Integración PMS (API Key)
/external/{action}              # Endpoints externos
```

### Convenciones de Nomenclatura

- **Plural para recursos**: `/api/v1/users`, `/api/v1/profiles`
- **Kebab-case para paths compuestos**: `/api/v1/issued-invoices`, `/api/v1/document-to-sign`
- **Guiones para acciones**: `/api/v1/pms-sync-articles`
- **IDs como parámetros**: `/:id`, `/:profileId`

### Grupos de Rutas por Autenticación

#### 1. API Pública (Sin autenticación)

```go
// Login
e.POST("/login", authController.Login)
e.POST("/recovery-password", authController.RecoveryPassword)
e.POST("/change-password-with-token", authController.ChangePasswordWithToken)
```

#### 2. API Restringida (JWT + Level Authorization)

```go
// Grupo con middleware JWT
restrictedAPI := e.Group("/api/v1")
restrictedAPI.Use(middleware.JWT([]byte(os.Getenv("JWT_SECRET"))))
restrictedAPI.Use(middleware.LevelAccess(db)) // Authorization basado en niveles

// Users
restrictedAPI.GET("/users", userController.GetAll)
restrictedAPI.GET("/users/:id", userController.GetByID)
restrictedAPI.POST("/users", userController.Create)
restrictedAPI.PUT("/users/:id", userController.Update)
restrictedAPI.DELETE("/users/:id", userController.Delete)
```

#### 3. API Key Protected (Integraciones)

```go
// PMS Integration
pmsGroup := e.Group("/api/v1")
pmsGroup.Use(middleware.APIKeyAuth(os.Getenv("PMS_API_KEY")))

pmsGroup.POST("/pms-sync-articles", pmsController.SyncArticles)
pmsGroup.POST("/pms-sync-providers", pmsController.SyncProviders)
pmsGroup.POST("/pms-sync-establishments", pmsController.SyncEstablishments)
```

#### 4. Intranet (Portal Empleados)

```go
intranet := e.Group("/intranet")
intranet.Use(middleware.JWT([]byte(os.Getenv("JWT_SECRET"))))

intranet.GET("/profile-holidays", profileController.GetHolidays)
intranet.GET("/profile-absences", profileController.GetAbsences)
```

## Middleware Disponibles

### JWT Authentication

```go
middleware.JWT([]byte(os.Getenv("JWT_SECRET")))
```

**Propósito**: Validar token JWT en header `Authorization: Bearer <token>`

**Uso**: Aplicar a grupos de rutas que requieren autenticación

**Configuración**: SkipPaths para rutas públicas definidas en `routes/echo.go`

---

### Level Authorization

```go
middleware.LevelAccess(database *gorm.DB)
```

**Propósito**: Verificar privilegios basados en Level del usuario

**Lógica**:
1. Extraer user ID del JWT claim
2. Obtener Level del usuario
3. Buscar LevelPrivileges para el Form asociado al path
4. Verificar Read (GET) o Write (POST/PUT/DELETE)

**Requisito**: El path debe estar asociado a un Form en BD (`forms.PathAPI`)

**Ejemplo de configuración en BD**:
```sql
-- Form para profiles
INSERT INTO forms (Name, PathAPI, Description) 
VALUES ('Profiles', 'profile|profile-holidays|profile-absences', 'Gestión de perfiles');

-- Privilegios para Level 2 (Read + Write)
INSERT INTO level_privileges (LevelID, FormID, CanRead, CanWrite) 
VALUES (2, 1, true, true);
```

---

### API Key Authentication

```go
middleware.APIKeyAuth(expectedKey string)
```

**Propósito**: Validar API key en header `X-API-KEY`

**Uso**: Integraciones externas (PMS, webhooks)

**Ejemplo**:
```go
e.Use(middleware.APIKeyAuth(os.Getenv("PMS_API_KEY")))
```

---

### CORS

```go
middleware.CORSConfig{
    AllowOrigins: []string{"https://*.reverencehotels.com"},
    AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
    AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType},
}
```

**Propósito**: Controlar acceso CORS desde orígenes específicos

**Configuración**: Definida en `routes/echo.go`

---

### Rate Limiting

```go
middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))
```

**Propósito**: Limitar requests por minuto por IP

**Uso**: Prevenir abuse en endpoints públicos

---

### Request Logger

```go
middleware.Logger()
```

**Propósito**: Loggear todas las requests HTTP

**Configuración**: Custom logger en `routes/echo.go`

---

### Request ID

```go
middleware.RequestID()
```

**Propósito**: Añadir ID único a cada request para tracing

**Header**: `X-Request-ID`

---

### Recover

```go
middleware.Recover()
```

**Propósito**: Recuperar gracefully de panics

**Configuración**: Custom panic handler en `routes/echo.go`

---

## Organización de Código

### Estructura de `routes/routes.go`

```go
package routes

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "your-project/controllers"
    "your-project/internal/backend/user/infrastructure/http"
)

func InitRoutes(e *echo.Echo, db *gorm.DB, dbSII *gorm.DB, dbProducts *gorm.DB) {
    // 1. Configurar middleware globales
    setupMiddleware(e)
    
    // 2. Rutas públicas
    setupPublicRoutes(e)
    
    // 3. API restringida
    setupRestrictedAPI(e, db)
    
    // 4. Intranet
    setupIntranetRoutes(e)
    
    // 5. Integraciones (API Key)
    setupIntegrationRoutes(e)
    
    // 6. Handlers Clean Architecture
    setupCleanArchitectureHandlers(e, db, dbSII, dbProducts)
}

func setupMiddleware(e *echo.Echo) {
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS(...))
    e.Use(middleware.RequestID())
    e.Use(middleware.RateLimiter(...))
}

func setupPublicRoutes(e *echo.Echo) {
    e.POST("/login", controllers.Auth.Login)
    e.POST("/recovery-password", controllers.Auth.RecoveryPassword)
}

func setupRestrictedAPI(e *echo.Echo, db *gorm.DB) {
    api := e.Group("/api/v1")
    api.Use(middleware.JWT([]byte(os.Getenv("JWT_SECRET"))))
    api.Use(middleware.LevelAccess(db))
    
    // Users module
    api.GET("/users", controllers.User.GetAll)
    api.GET("/users/:id", controllers.User.GetByID)
    api.POST("/users", controllers.User.Create)
    api.PUT("/users/:id", controllers.User.Update)
    api.DELETE("/users/:id", controllers.User.Delete)
    
    // Profile module
    api.GET("/profiles", controllers.Profile.GetAll)
    api.GET("/profiles/:id", controllers.Profile.GetByID)
    // ... más rutas
}

func setupIntranetRoutes(e *echo.Echo) {
    intranet := e.Group("/intranet")
    intranet.Use(middleware.JWT([]byte(os.Getenv("JWT_SECRET"))))
    
    intranet.GET("/profile-holidays", controllers.Profile.GetHolidays)
    intranet.GET("/profile-absences", controllers.Profile.GetAbsences)
}

func setupIntegrationRoutes(e *echo.Echo) {
    pms := e.Group("/api/v1")
    pms.Use(middleware.APIKeyAuth(os.Getenv("PMS_API_KEY")))
    
    pms.POST("/pms-sync-articles", controllers.PMS.SyncArticles)
    pms.POST("/pms-sync-providers", controllers.PMS.SyncProviders)
}
```

---

### Handlers Clean Architecture

```go
// internal/{module}/infrastructure/http/handler.go

package http

import (
    "github.com/labstack/echo/v4"
    "net/http"
)

type Handler struct {
    service ports.Service
}

func NewHandler(service ports.Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Create(c echo.Context) error {
    var req dto.CreateRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    entity, err := h.service.Create(c.Request().Context(), req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusCreated, entity)
}

func (h *Handler) GetAll(c echo.Context) error {
    entities, err := h.service.GetAll(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, entities)
}

func (h *Handler) GetByID(c echo.Context) error {
    id := c.Param("id")
    entity, err := h.service.GetByID(c.Request().Context(), id)
    if err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
    }
    
    return c.JSON(http.StatusOK, entity)
}

func (h *Handler) Update(c echo.Context) error {
    id := c.Param("id")
    var req dto.UpdateRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    entity, err := h.service.Update(c.Request().Context(), id, req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, entity)
}

func (h *Handler) Delete(c echo.Context) error {
    id := c.Param("id")
    err := h.service.Delete(c.Request().Context(), id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.NoContent(http.StatusNoContent)
}
```

---

## Ejemplos de Uso

### Ejemplo 1: Nueva ruta CRUD estándar

**Input del usuario**:
> "Crea una ruta para gestionar departamentos (departments)"

**Proceso**:

1. Identificar que es CRUD estándar (API restringida)
2. Crear controller en `controllers/Department/`
3. Registrar rutas en `routes/routes.go`
4. Crear Form en BD para authorization

**Output esperado**:

```go
// En routes/routes.go, dentro de setupRestrictedAPI:

// Departments module
api.GET("/departments", departmentController.GetAll)
api.GET("/departments/:id", departmentController.GetByID)
api.POST("/departments", departmentController.Create)
api.PUT("/departments/:id", departmentController.Update)
api.DELETE("/departments/:id", departmentController.Delete)

// En BD (migration seed):
INSERT INTO forms (Name, PathAPI, Description) 
VALUES ('Departments', 'departments', 'Gestión de departamentos organizacionales');

INSERT INTO level_privileges (LevelID, FormID, CanRead, CanWrite) 
VALUES (1, <form_id>, true, false); -- Level 1: Read only
VALUES (2, <form_id>, true, true);  -- Level 2: Read + Write
```

---

### Ejemplo 2: Ruta con múltiples HTTP methods

**Input del usuario**:
> "Necesito un endpoint para facturas emitidas que acepte GET y POST"

**Proceso**:

1. Identificar que es recurso compuesto (issued-invoices)
2. Verificar si existe controller en `controllers/Invoices/`
3. Registrar rutas con mismo path pero métodos diferentes

**Output esperado**:

```go
// En routes/routes.go:
api.GET("/issued-invoices", invoicesController.GetIssuedInvoices)
api.GET("/issued-invoices/:id", invoicesController.GetIssuedInvoiceByID)
api.POST("/issued-invoices", invoicesController.CreateIssuedInvoice)
api.PUT("/issued-invoices/:id", invoicesController.UpdateIssuedInvoice)

// Nota: El POST a /issued-invoices automáticamente envía a SII
// Verificar que el Form en BD incluya "issued-invoices" en PathAPI
```

---

### Ejemplo 3: Ruta con parámetros opcionales y query params

**Input del usuario**:
> "Crea ruta para buscar perfiles con filtros opcionales"

**Proceso**:

1. Usar query params en vez de path params
2. Handler debe extraer params con `c.QueryParam()`

**Output esperado**:

```go
// Ruta en routes/routes.go:
api.GET("/profiles", profileController.SearchProfiles)

// Handler en controllers/Profile/controller.go:
func (pc *ProfileController) SearchProfiles(c echo.Context) error {
    // Query params opcionales
    name := c.QueryParam("name")
    department := c.QueryParam("department")
    level := c.QueryParam("level")
    active := c.QueryParam("active")
    
    // Construir filtros
    filters := map[string]interface{}{}
    if name != "" {
        filters["name"] = name
    }
    if department != "" {
        filters["department"] = department
    }
    if level != "" {
        filters["level"] = level
    }
    if active != "" {
        filters["active"] = active == "true"
    }
    
    profiles, err := pc.service.Search(filters)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, ErrorResponse(err))
    }
    
    return c.JSON(http.StatusOK, profiles)
}
```

---

### Ejemplo 4: Ruta con archivo upload

**Input del usuario**:
> "Necesito endpoint para subir documentos de firma"

**Proceso**:

1. Usar `c.FormFile()` para multipart upload
2. Validar tipo de archivo y tamaño
3. Guardar en storage configurado

**Output esperado**:

```go
// Ruta en routes/routes.go:
api.POST("/document-to-sign", signatureController.CreateDocumentToSign)

// Handler en controllers/Signatures/controller.go:
func (sc *SignatureController) CreateDocumentToSign(c echo.Context) error {
    // Obtener archivo del form
    file, err := c.FormFile("document")
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "document file is required",
        })
    }
    
    // Validar tamaño (max 10MB)
    if file.Size > 10*1024*1024 {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "file too large (max 10MB)",
        })
    }
    
    // Validar tipo (PDF only)
    if !strings.HasSuffix(file.Filename, ".pdf") {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "only PDF files are allowed",
        })
    }
    
    // Abrir archivo
    src, err := file.Open()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to open file",
        })
    }
    defer src.Close()
    
    // Guardar en storage
    filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    dst := filepath.Join("./uploads/signatures", filename)
    
    if err := c.Save(file, dst); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to save file",
        })
    }
    
    // Crear registro en BD
    document := &models.DocumentToSign{
        FileName: filename,
        FilePath: dst,
        Status:   "pending",
    }
    
    if err := sc.db.Create(document).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to create document record",
        })
    }
    
    return c.JSON(http.StatusCreated, document)
}
```

---

### Ejemplo 5: Ruta Clean Architecture

**Input del usuario**:
> "Crea endpoint para gestión de notificaciones usando Clean Architecture"

**Proceso**:

1. Verificar que existe módulo en `internal/notification/`
2. Handler en `infrastructure/http/handler.go`
3. Registrar ruta en `routes/routes.go` o `cmd/server/backend/`

**Output esperado**:

```go
// En internal/notification/infrastructure/http/handler.go:

package http

import (
    "github.com/labstack/echo/v4"
    "net/http"
    "your-project/internal/notification/application"
    "your-project/internal/notification/domain"
)

type Handler struct {
    service *application.Service
}

func NewHandler(service *application.Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) CreateNotification(c echo.Context) error {
    var req domain.CreateNotificationRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    notification, err := h.service.CreateNotification(c.Request().Context(), req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusCreated, notification)
}

func (h *Handler) GetUserNotifications(c echo.Context) error {
    userID := c.Param("userId")
    
    notifications, err := h.service.GetUserNotifications(c.Request().Context(), userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, notifications)
}

// En routes/routes.go:

// Import handler
notificationHandler := http.NewHandler(notificationService)

// Registrar rutas
api.POST("/notifications", notificationHandler.CreateNotification)
api.GET("/notifications/user/:userId", notificationHandler.GetUserNotifications)
api.PUT("/notifications/:id/read", notificationHandler.MarkAsRead)
```

---

## Troubleshooting

### Problema: Ruta retorna 404 Not Found

**Síntoma**: Request a ruta válida retorna 404

**Causas posibles**:

1. **Ruta no registrada**: Olvidó agregar la ruta en `routes/routes.go`
2. **Path incorrecto**: El path del request no coincide exactamente con el registrado
3. **Grupo de rutas**: La ruta está dentro de un grupo (ej: `/api/v1`) pero se hace request sin el prefijo
4. **Método HTTP incorrecto**: Ruta registrada como GET pero se hace POST

**Solución**:

```bash
# 1. Verificar que la ruta está registrada
grep -n "profiles" routes/routes.go

# 2. Verificar el path exacto (incluye prefijos de grupo)
# Si está en api.GET("/profiles"), el path completo es /api/v1/profiles

# 3. Verificar método HTTP
curl -X GET http://localhost:1331/api/v1/profiles  # Correcto
curl -X POST http://localhost:1331/api/v1/profiles  # 404 si no existe POST
```

---

### Problema: Error de autorización "Insufficient privileges"

**Síntoma**: Response 403 con "User does not have required privileges"

**Causa**: Level del usuario no tiene privilegio para el Form asociado al path

**Solución**:

1. **Verificar Form asociado al path**:
```sql
SELECT Name, PathAPI FROM forms WHERE PathAPI LIKE '%profile%';
```

2. **Verificar Level del usuario**:
```sql
SELECT Level FROM users WHERE ID = <user_id>;
```

3. **Verificar privilegios del Level**:
```sql
SELECT lp.*, f.Name, f.PathAPI 
FROM level_privileges lp
JOIN forms f ON lp.FormID = f.ID
WHERE lp.LevelID = <user_level> 
  AND f.PathAPI LIKE '%profile%';
```

4. **Crear privilegio si no existe**:
```sql
INSERT INTO level_privileges (LevelID, FormID, CanRead, CanWrite)
VALUES (<level_id>, <form_id>, true, false); -- Read only
```

---

### Problema: Middleware no se ejecuta

**Síntoma**: Middleware configurado pero no se ejecuta

**Causas**:

1. **Orden de middleware**: Se registró middleware DESPUÉS de las rutas
2. **Scope incorrecto**: Middleware en grupo equivocado

**Solución**:

```go
// ❌ INCORRECTO: Middleware después de rutas
api := e.Group("/api/v1")
api.GET("/users", userController.GetAll)
api.Use(middleware.JWT(...)) // No se ejecutará para GET /users

// ✅ CORRECTO: Middleware antes de rutas
api := e.Group("/api/v1")
api.Use(middleware.JWT(...)) // Se ejecuta antes
api.GET("/users", userController.GetAll)

// ✅ CORRECTO: Middleware en grupo específico
public := e.Group("/public")
public.GET("/login", authController.Login)

restricted := e.Group("/api/v1")
restricted.Use(middleware.JWT(...))
restricted.GET("/users", userController.GetAll) // Solo este tiene JWT
```

---

### Problema: Conflictos de rutas

**Síntoma**: Una ruta "sombrea" a otra, o se ejecuta la incorrecta

**Causa**: Echo usa first-match, ruta más específica debe registrarse primero

**Solución**:

```go
// ❌ INCORRECTO: Ruta genérica primero
api.GET("/users/:id", userController.GetByID)
api.GET("/users/special", userController.GetSpecialUsers) 
// GET /users/special → matchea /users/:id (id = "special")

// ✅ CORRECTO: Ruta específica primero
api.GET("/users/special", userController.GetSpecialUsers)
api.GET("/users/:id", userController.GetByID)
// GET /users/special → matchea ruta específica
// GET /users/123 → matchea /users/:id
```

---

### Problema: Handler no recibe contexto correctamente

**Síntoma**: Timeout o context cancelled en handler

**Causa**: No se está pasando el contexto de Echo al servicio

**Solución**:

```go
// ❌ INCORRECTO: Contexto de background
func (h *Handler) GetByID(c echo.Context) error {
    entity, err := h.service.GetByID(context.Background(), id)
    // Timeout posiblemente no respetado
}

// ✅ CORRECTO: Contexto de Echo
func (h *Handler) GetByID(c echo.Context) error {
    entity, err := h.service.GetByID(c.Request().Context(), id)
    // Timeout y cancelación se propagan correctamente
}
```

---

### Problema: Parámetros de path no se capturan

**Síntoma**: `c.Param("id")` retorna string vacío

**Causa**: Ruta no tiene parámetro definido

**Solución**:

```go
// ❌ INCORRECTO: Ruta sin parámetro
api.GET("/users/:id", userController.GetByID)

// En handler:
func GetUserByEmail(c echo.Context) error {
    email := c.Param("email") // Vacío, ruta no tiene :email
}

// ✅ CORRECTO: Definir parámetro en ruta
api.GET("/users/:email", userController.GetByEmail)

// O usar query param
api.GET("/users", userController.SearchUsers)

// En handler:
func SearchUsers(c echo.Context) error {
    email := c.QueryParam("email") // ?email=test@example.com
}
```

---

### Problema: CORS errors en frontend

**Síntoma**: Browser muestra "CORS policy blocked"

**Causa**: Origen del frontend no está en whitelist de CORS

**Solución**:

```go
// En routes/echo.go, configurar CORS:

e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{
        "http://localhost:3000",           // Desarrollo local
        "https://reverence-hotels.com",    // Producción
        "https://*.reverencehotels.com",   // Subdominios
    },
    AllowMethods: []string{
        echo.GET,
        echo.POST,
        echo.PUT,
        echo.DELETE,
        echo.OPTIONS,
    },
    AllowHeaders: []string{
        echo.HeaderAuthorization,
        echo.HeaderContentType,
        echo.HeaderXRequestID,
    },
    AllowCredentials: true, // Para cookies/auth headers
    MaxAge: 86400,          // 24 horas de cache prefetch
}))
```

---

## Consideraciones Especiales del Proyecto

### Multi-Database Context

Las rutas pueden requerir diferentes conexiones a BD según el módulo:

```go
func InitRoutes(e *echo.Echo, db *gorm.DB, dbSII *gorm.DB, dbProducts *gorm.DB) {
    // Principal DB (users, profiles, etc)
    api.GET("/users", userController.GetAll) // Usa db
    
    // SII DB (invoicing)
    api.GET("/issued-invoices", invoicesController.GetAll) // Usa dbSII
    
    // Products DB (economato)
    api.GET("/articles", articlesController.GetAll) // Usa dbProducts
}
```

**Importante**: Inyectar la DB connection correcta en el controller/handler

---

### Domain Events Publishing

En rutas Clean Architecture, publicar eventos después de mutaciones:

```go
func (h *Handler) Create(c echo.Context) error {
    var req dto.CreateRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    entity, err := h.service.Create(c.Request().Context(), req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    // Publicar domain event
    h.eventBus.Publish(eventDomain.NewDomainEvent(
        eventDomain.GetEventNameToCreate(),
        entity.ID,
        time.Now(),
        entity,
        c.Logger(),
    ))
    
    return c.JSON(http.StatusCreated, entity)
}
```

---

### Level-Based Authorization vs JWT

El proyecto usa **dos capas** de autenticación:

1. **JWT** (primera capa): Verifica que el usuario está autenticado
2. **Level Access** (segunda capa): Verifica qué puede hacer el usuario

```go
restrictedAPI := e.Group("/api/v1")

// Primera capa: JWT
restrictedAPI.Use(middleware.JWT([]byte(os.Getenv("JWT_SECRET"))))

// Segunda capa: Level Access
restrictedAPI.Use(middleware.LevelAccess(db))

// Ahora la ruta está doblemente protegida
restrictedAPI.GET("/profiles", profileController.GetAll)
```

---

### Form-Path Association

Cada ruta restringida debe tener un **Form** asociado en BD:

```sql
-- Form único
INSERT INTO forms (Name, PathAPI, Description) 
VALUES ('Profiles', 'profile', 'Employee profiles');

-- Form con múltiples paths (separados por |)
INSERT INTO forms (Name, PathAPI, Description) 
VALUES ('Profiles Extended', 'profile|profile-holidays|profile-absences', 'Profile management with holidays and absences');
```

**Middleware buscará** coincidencia parcial:
- Request a `/api/v1/profile` → busca Form con `PathAPI LIKE '%profile%'`
- Request a `/api/v1/profile-holidays` → busca Form con `PathAPI LIKE '%profile-holidays%'`

---

## Presentación de Resultados

Al crear o modificar rutas:

1. **Resumir cambios**: "Creadas {n} rutas nuevas para módulo {module}"
2. **Formato de output**: Bloques de código Go completos
3. **Incluir migraciones**: Si requiere cambios en BD (Forms, LevelPrivileges)
4. **Especificar middleware**: Qué middleware se aplica a cada ruta

**Ejemplo de resumen**:

```
Creadas 5 rutas CRUD para módulo Departments:

1. GET /api/v1/departments - Listar todos (Read privilege)
2. GET /api/v1/departments/:id - Obtener por ID (Read privilege)
3. POST /api/v1/departments - Crear (Write privilege)
4. PUT /api/v1/departments/:id - Actualizar (Write privilege)
5. DELETE /api/v1/departments/:id - Eliminar (Write privilege)

Middleware aplicado:
- JWT Authentication
- Level Authorization (Form: departments)

Migración BD requerida:
- INSERT INTO forms (Name, PathAPI) VALUES ('Departments', 'departments');
- INSERT INTO level_privileges (LevelID, FormID, CanRead, CanWrite) VALUES (...);
```

---

## Mejores Prácticas

### 1. Organización de Rutas

- ✅ Agrupar rutas por módulo en funciones separadas (`setupUserRoutes`, `setupProfileRoutes`)
- ✅ Usar constantes para paths comunes
- ✅ Comentar grupos de rutas con encabezados claros

### 2. Nomenclatura

- ✅ Usar plural para recursos (`/users`, no `/user`)
- ✅ Usar kebab-case para paths compuestos (`/issued-invoices`, no `/issuedInvoices`)
- ✅ Ser consistente en nombres (ej: siempre `/:id`, nunca `/:userId` y `/:departmentId` mezclados)

### 3. Middleware

- ✅ Aplicar middleware más específico después del general
- ✅ No repetir middleware innecesariamente (usar grupos)
- ✅ Documentar middleware custom en comentarios

### 4. Handlers

- ✅ Validar input temprano (`c.Bind()` con validación)
- ✅ Usar context.Request().Context() para operaciones
- ✅ Retornar códigos HTTP apropiados (200, 201, 400, 404, 500)
- ✅ Manejar errores consistentemente

### 5. Authorization

- ✅ Siempre asociar rutas a Forms en BD
- ✅ Configurar LevelPrivileges para cada Level que necesite acceso
- ✅ Verificar CanRead para GET y CanWrite para POST/PUT/DELETE
- ✅ Documentar privilegios requeridos en comentarios

---

## Referencias Rápidas

### HTTP Status Codes Comunes

- `200 OK` - GET/PUT exitoso
- `201 Created` - POST exitoso
- `204 No Content` - DELETE exitoso
- `400 Bad Request` - Input inválido
- `401 Unauthorized` - No hay token JWT
- `403 Forbidden` - Token válido pero insuficientes privilegios
- `404 Not Found` - Recurso no existe
- `409 Conflict` - Recurso ya existe
- `500 Internal Server Error` - Error del servidor

### Echo Context Methods Útiles

- `c.Param("name")` - Path param (`:id`)
- `c.QueryParam("name")` - Query string (`?page=1`)
- `c.Get("name")` - Datos del contexto (set por middleware)
- `c.Bind(&obj)`` - Parsear JSON/body a struct
- `c.JSON(code, obj)` - Response JSON
- `c.NoContent(code)` - Response sin body
- `c.Request().Context()` - Contexto para operaciones
- `c.Logger()` - Logger de Echo

### Middleware Order (Ejecución Inversa)

```go
api.Use(middleware1) // 3ro en ejecutarse
api.Use(middleware2) // 2do en ejecutarse
api.Use(middleware3) // 1ro en ejecutarse
api.GET("/path", handler) // Se ejecuta último
```

---

## Roadmap de Mejoras

- [ ] Soporte para versionado de API (`/api/v1/`, `/api/v2/`)
- [ ] Rate limiting por usuario (no solo por IP)
- [ ] Response compression middleware
- [ ] Prometheus metrics middleware
- [ ] Graceful shutdown con active connections drain
- [ ] OpenAPI/Swagger generation automática
- [ ] Request validation middleware con JSON Schema