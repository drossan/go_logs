---
name: multi-database
description: Esta skill debe usarse cuando el usuario necesite trabajar con operaciones que involucren múltiples bases de datos MySQL en el proyecto Reverence Hotels API. Se activa con peticiones como "conecta a las tres bases de datos", "consulta la base de datos del SII", "sincroniza datos entre economato y la base principal", o "crea una query que use db y dbSII".
license: Complete terms in LICENSE.txt
version: 1.0.0
author: Reverence Hotels Development Team
category: language
tags: [mysql, multi-database, gorm, go, database-connections]---

# Multi-Database Skill

Skill especializada en el manejo de la arquitectura de múltiples bases de datos MySQL del proyecto Reverence Hotels API. Proporciona conocimiento técnico sobre las 3 conexiones de bases de datos simultáneas, configuración de GORM, patrones de acceso a datos según la fuente (Principal, SII, Economato), y mejores prácticas para operaciones transversales entre bases de datos.

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:
- Se requiera crear o modificar código que acceda a más de una base de datos
- El usuario necesite entender qué conexión de BD usar para una operación específica
- Se deban implementar queries que involucren múltiples bases de datos
- Sea necesario configurar nuevas conexiones o modificar las existentes
- Se presenten problemas de rendimiento o conexión relacionados con las BDs
- Requiera implementar migraciones o cambios que afecten a las 3 BDs

Triggers comunes:
- "¿Cómo accedo a la base de datos del SII para crear facturas?"
- "Necesito sincronizar artículos del economato con la base principal"
- "¿Qué variable de entorno uso para la BD de productos?"
- "Tengo un error de conexión con dbSII"
- "Crear un endpoint que consulte usuarios y datos de facturación"
- "Optimizar una query que usa las tres bases de datos"

## Arquitectura de Multi-Base de Datos

### Resumen Técnico

El proyecto **Reverence Hotels API** implementa una arquitectura única de **3 bases de datos MySQL simultáneas**, cada una con propósito específico:

1. **Principal** (`sensesho_api`) - Gestión de usuarios, perfiles, RRHH, niveles, formularios
2. **SII** (`reverence_sii`) - Facturación electrónica española (AEAT), XML, certificados
3. **Products/Economato** (`economato`) - Artículos, proveedores, pedidos, almacén

### Conexiones GORM

Las 3 conexiones se inicializan en `routes/InitRoutes()`:

```go
// Base de datos Principal
db, err := configuration.GetDBPrincipal()

// Base de datos SII (Facturación)
dbSII, err := configuration.GetDBSII()

// Base de datos Products (Economato)
dbProducts, err := configuration.GetDBProducts()
```

### Configuración por Variables de Entorno

Cada conexión tiene su propio set de variables:

**Principal** (`DATA_BASE_*`):
- `DATA_BASE_USER`
- `DATA_BASE_PASS`
- `DATA_BASE_HOST`
- `DATA_BASE_NAME`
- `DATA_BASE_PORT`

**SII** (`DATA_BASE_*_SII`):
- `DATA_BASE_USER_SII`
- `DATA_BASE_PASS_SII`
- `DATA_BASE_HOST_SII`
- `DATA_BASE_NAME_SII`
- `DATA_BASE_PORT_SII`

**Products** (`DATA_BASE_*_PRODUCTS`):
- `DATA_BASE_USER_PRODUCTS`
- `DATA_BASE_PASS_PRODUCTS`
- `DATA_BASE_HOST_PRODUCTS`
- `DATA_BASE_NAME_PRODUCTS`
- `DATA_BASE_PORT_PRODUCTS`

### Configuración de Connection Pool

Todas las conexiones usan la misma configuración de pool:

```go
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

## Workflow Principal

### 1. Identificación de Base de Datos

Antes de escribir cualquier código de acceso a datos:

1. **Determinar el dominio de datos**:
   - Usuarios/Perfiles/RRHH → `db` (Principal)
   - Facturas/XML/SII → `dbSII`
   - Artículos/Proveedores/Pedidos → `dbProducts`

2. **Verificar modelos existentes**:
   - Models en `models/` suelen indicar qué BD usan
   - Revisar imports de `configuration/` para ver conexión usada

3. **Confirmar con estructura del proyecto**:
   - Legacy MVC: Revisar `services/` correspondiente
   - Clean Architecture: Revisar `internal/*/infrastructure/db/`

### 2. Selección de Conexión

Según el caso de uso:

#### Acceso a Base Principal

```go
// En Controllers o Services
import (
    "your-project/configuration"
)

func GetUserData(userID uint) (*models.User, error) {
    db := configuration.GetDBPrincipal()
    var user models.User
    err := db.First(&user, userID).Error
    return &user, err
}
```

#### Acceso a Base SII

```go
func CreateInvoice(invoice *models.IssuedInvoice) error {
    dbSII := configuration.GetDBSII()
    return dbSII.Create(invoice).Error
}
```

#### Acceso a Base Products

```go
func SyncArticles() ([]models.Article, error) {
    dbProducts := configuration.GetDBProducts()
    var articles []models.Article
    err := dbProducts.Find(&articles).Error
    return articles, err
}
```

### 3. Operaciones Transversales (Multi-BD)

Para operaciones que requieren múltiples bases de datos:

```go
func ProcessUserWithOrders(userID uint) error {
    // Obtener usuario de BD Principal
    db := configuration.GetDBPrincipal()
    var user models.User
    if err := db.First(&user, userID).Error; err != nil {
        return err
    }
    
    // Obtener pedidos de BD Economato
    dbProducts := configuration.GetDBProducts()
    var orders []models.Order
    if err := dbProducts.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
        return err
    }
    
    // Procesar datos combinados
    // ...
    return nil
}
```

**IMPORTANTE**: No es posible usar transactions GORM entre diferentes bases de datos. Implementar compensating transactions si se requiere ACID.

### 4. Validación

Verificar que:

- [ ] Se usa la conexión correcta para el dominio de datos
- [ ] Las variables de entorno están configuradas para las 3 BDs
- [ ] No hay mixing de conexiones (ej: modelo de SII con db de Products)
- [ ] El código maneja errores de conexión específicos por BD

## Patrones de Acceso por Módulo

### Base Principal (sensesho_api)

**Ubicación**: `db` (GetDBPrincipal)

**Casos de uso típicos**:
- Autenticación y usuarios (`controllers/Auth/`)
- Gestión de perfiles de empleado (`controllers/Profile/`)
- Niveles y privilegios (`controllers/Level/`)
- Formularios dinámicos (`controllers/Form/`)
- Notificaciones (`internal/notification/`)
- Eventos de dominio (`internal/event/`)

**Ejemplo - Clean Architecture**:

```go
// internal/backend/user/infrastructure/db/gorm_repository.go
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
    return &userRepository{db: db}
}

// Uso en cmd/server/backend/
dbPrincipal := configuration.GetDBPrincipal()
userRepo := userRepository.NewUserRepository(dbPrincipal)
```

### Base SII (reverence_sii)

**Ubicación**: `dbSII` (GetDBSII)

**Casos de uso típicos**:
- Facturas emitidas (`services/Invoices/IssuedInvoice`)
- Facturas recibidas (`services/Invoices/ReceivedInvoice`)
- Generación de XML para AEAT
- Gestión de certificados digitales (.pem)
- Códigos de seguimiento CSV

**Ejemplo - Servicio Legacy**:

```go
// services/Invoices/issued_invoice_service.go
func CreateIssuedInvoice(invoice *models.IssuedInvoice) error {
    dbSII := configuration.GetDBSII()
    
    // Crear factura en BD SII
    if err := dbSII.Create(invoice).Error; err != nil {
        return err
    }
    
    // Generar XML y enviar a AEAT
    xmlContent := generateXML(invoice)
    return sendToAEAT(xmlContent)
}
```

### Base Products (economato)

**Ubicación**: `dbProducts` (GetDBProducts)

**Casos de uso típicos**:
- Sincronización de artículos con PMS
- Gestión de proveedores
- Pedidos y albaranes
- Integraciones PMS (`/api/v1/pms-*`)

**Ejemplo - Integración PMS**:

```go
// controllers/PMS/sync_articles.go
func SyncArticlesFromPMS(articles []models.Article) error {
    dbProducts := configuration.GetDBProducts()
    
    for _, article := range articles {
        var existing models.Article
        err := dbProducts.Where("pms_code = ?", article.PMSCode).First(&existing).Error
        
        if err == gorm.ErrRecordNotFound {
            // Nuevo artículo
            dbProducts.Create(&article)
        } else if err == nil {
            // Actualizar existente
            dbProducts.Model(&existing).Updates(article)
        }
    }
    return nil
}
```

## Recursos de la Skill

### Referencias (`references/`)

#### `references/database_schema.md`

**Contenido**: Esquemas detallados de las 3 bases de datos

**Cuándo consultar**:
- Al crear nuevos modelos o migrations
- Para entender relaciones entre tablas
- Al escribir queries complejas

**Estructura**:

```
database_schema.md
├── principal_db_schema.md
│   ├── users
│   ├── profiles
│   ├── levels
│   └── forms
├── sii_db_schema.md
│   ├── issued_invoices
│   ├── received_invoices
│   └── xml_logs
└── products_db_schema.md
    ├── articles
    ├── providers
    └── orders
```

**Búsqueda rápida**:

```bash
# Buscar tabla específica
grep -i "CREATE TABLE.*profiles" references/database_schema.md

# Buscar en BD específica
grep -A 50 "## Base SII" references/database_schema.md
```

---

#### `references/migration_guide.md`

**Contenido**: Guía de migración GORM v1→v2 y manejo de foreign keys

**Cuándo consultar**:
- Al modificar modelos existentes
- Al crear nuevas migraciones
- Si aparecen errores de foreign keys

**Secciones clave**:
- Foreign key handling
- AutoMigrate workflow
- View recreation (view_reduce_profiles)

---

#### `references/connection_troubleshooting.md`

**Contenido**: Diagnóstico y resolución de problemas de conexión

**Cuándo consultar**:
- Errores de conexión a BD
- Timeouts o connection pool exhausted
- Problemas de rendimiento

**Búsqueda rápida**:

```bash
# Error específico
grep -i "connection refused" references/connection_troubleshooting.md

# Problemas de pool
grep -A 10 "max open connections" references/connection_troubleshooting.md
```

### Scripts (`scripts/`)

#### `scripts/test_db_connections.sh`

**Propósito**: Verificar conectividad con las 3 bases de datos

**Uso**:

```bash
bash scripts/test_db_connections.sh [environment]
```

**Parámetros**:

- `environment`: local | pre | pro (default: local)

**Output**: Reporte de estado de conexión para cada BD

**Ejemplo**:

```bash
bash scripts/test_db_connections.sh local
# Output:
# ✓ Principal DB (sensesho_api): Connected
# ✓ SII DB (reverence_sii): Connected
# ✓ Products DB (economato): Connected
```

---

#### `scripts/backup_db.sh`

**Propósito**: Backup individual de cada base de datos

**Uso**:

```bash
bash scripts/backup_db.sh [db_type] [output_dir]
```

**Parámetros**:

- `db_type`: principal | sii | products
- `output_dir`: Directorio de salida (default: ./backups/)

**Ejemplo**:

```bash
bash scripts/backup_db.sh principal ./backups/2025-01-20/
# Output: backups/2025-01-20/sensesho_api_20250120_143022.sql
```

---

#### `scripts/validate_env.sh`

**Propósito**: Validar que todas las variables de entorno de BD están configuradas

**Uso**:

```bash
bash scripts/validate_env.sh
```

**Output**: Lista de variables faltantes o incorrectas

**Ejemplo**:

```bash
bash scripts/validate_env.sh
# Output:
# ✓ DATA_BASE_USER configured
# ✓ DATA_BASE_PASS configured
# ✗ DATA_BASE_HOST_SII not set
# ✓ DATA_BASE_NAME_PRODUCTS configured
```

## Ejemplos de Uso

### Ejemplo 1: Crear nuevo endpoint que usa 2 bases de datos

**Input del usuario**:
> "Necesito un endpoint que devuelva el perfil del empleado con sus pedidos del economato"

**Proceso**:

1. Identificar que requiere BD Principal (perfil) + BD Products (pedidos)
2. Crear handler que use ambas conexiones
3. Implementar lógica de combinación de datos
4. Manejar errores de cada BD independientemente

**Implementación**:

```go
// controllers/Profile/profile_with_orders.go
package Profile

import (
    "your-project/configuration"
    "your-project/models"
    "github.com/labstack/echo/v4"
)

func GetProfileWithOrders(c echo.Context) error {
    userID := c.Param("id")
    
    // Obtener perfil de BD Principal
    db := configuration.GetDBPrincipal()
    var profile models.Profile
    if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
        return c.JSON(500, map[string]string{"error": "Profile not found"})
    }
    
    // Obtener pedidos de BD Products
    dbProducts := configuration.GetDBProducts()
    var orders []models.Order
    if err := dbProducts.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
        return c.JSON(500, map[string]string{"error": "Orders not found"})
    }
    
    // Combinar datos
    response := map[string]interface{}{
        "profile": profile,
        "orders":  orders,
    }
    
    return c.JSON(200, response)
}
```

**Output esperado**:

```
Endpoint creado exitosamente:
- Ruta: GET /api/v1/profile/:id/orders
- Bases de datos: Principal (perfil) + Products (pedidos)
- Manejo de errores: Independiente por BD
- Output: JSON combinado con perfil y array de pedidos
```

---

### Ejemplo 2: Diagnóstico de error de conexión SII

**Input del usuario**:
> "Me da error 'connection refused' al crear facturas en el SII"

**Proceso**:

1. Identificar que el error es en `dbSII` (BD SII)
2. Ejecutar script de test de conexiones
3. Validar variables de entorno SII
4. Verificar configuración de red/firewall

**Diagnóstico**:

```bash
# 1. Test de conexiones
bash scripts/test_db_connections.sh local

# 2. Validar env SII
grep "DATA_BASE.*SII" .env

# 3. Verificar conectividad
mysql -h $DATA_BASE_HOST_SII -P $DATA_BASE_PORT_SII -u $DATA_BASE_USER_SII -p
```

**Output esperado**:

```
Diagnóstico de conexión SII:
✓ Principal DB: Conectada
✗ SII DB: Connection refused
  Causa probable: DATA_BASE_HOST_SII incorrecto o firewall bloqueando
  Host actual: 192.168.1.100
  Puerto: 3306
  
Solución:
1. Verificar que DATA_BASE_HOST_SII sea accesible: telnet 192.168.1.100 3306
2. Confirmar credenciales en .env
3. Verificar reglas de firewall del servidor SII
```

---

### Ejemplo 3: Implementar sincronización entre BDs

**Input del usuario**:
> "Sincroniza los artículos del PMS a la base de products cada hora"

**Proceso**:

1. Crear servicio que use `dbProducts` (destino)
2. Implementar lógica de sincronización
3. Agregar cron job en `task/` según ENV
4. Manejar errores y logging

**Implementación**:

```go
// services/Products/sync_service.go
package Products

import (
    "your-project/configuration"
    "your-project/models"
    "log"
)

func SyncArticlesFromPMS() error {
    dbProducts := configuration.GetDBProducts()
    
    // Llamada a API del PMS
    pmsArticles, err := fetchArticlesFromPMS()
    if err != nil {
        log.Printf("Error fetching from PMS: %v", err)
        return err
    }
    
    // Upsert en BD Products
    for _, article := range pmsArticles {
        var existing models.Article
        err := dbProducts.Where("pms_code = ?", article.PMSCode).First(&existing).Error
        
        if err == gorm.ErrRecordNotFound {
            // Crear nuevo
            if err := dbProducts.Create(&article).Error; err != nil {
                log.Printf("Error creating article %s: %v", article.PMSCode, err)
            }
        } else if err == nil {
            // Actualizar existente
            if err := dbProducts.Model(&existing).Updates(&article).Error; err != nil {
                log.Printf("Error updating article %s: %v", article.PMSCode, err)
            }
        }
    }
    
    log.Printf("Sync completed: %d articles processed", len(pmsArticles))
    return nil
}
```

**Cron job**:

```go
// task/products_sync.go
package task

import (
    "your-project/services/Products"
    "github.com/robfig/cron/v3"
)

func RegisterProductSyncJobs(c *cron.Cron) {
    // Solo en pre y pro
    if os.Getenv("ENV") != "local" {
        cron.AddFunc("0 * * * *", func() {
            Products.SyncArticlesFromPMS()
        })
    }
}
```

**Output esperado**:

```
Sincronización de artículos implementada:
- Servicio: services/Products/sync_service.go
- Base de datos: dbProducts (economato)
- Cron schedule: Cada hora (0 * * * *)
- Entornos: pre y pro (no en local)
- Logging: Errors + contador de artículos procesados
- Patrón: Upsert (insert or update) por PMSCode
```

## Presentación de Resultados

Al completar tareas multi-database:

1. **Resumir configuración**: Listar conexiones usadas y propósito
2. **Especificar BD por operación**: Qué consulta va a qué BD
3. **Incluir validación**: Confirmar conectividad
4. **Documentar trade-offs**: Si no es posible transacción ACID, explicar por qué

**Ejemplo de resumen**:

```
Operación multi-database completada:
- BD Principal (db): Consulta de perfil usuario
- BD Products (dbProducts): Consulta de pedidos
- BD SII (dbSII): No utilizada

Conectividad:
✓ Principal: Connected (127.0.0.1:3306/sensesho_api)
✓ Products: Connected (127.0.0.1:3307/economato)
✗ SII: Not tested (not required)

Limitaciones:
- No es posible transacción ACID entre Principal y Products
- Implementado compensating transaction pattern para rollback manual
```

## Troubleshooting

### Problema: Error "database connection closed"

**Síntoma**: `driver: bad connection` o `database is closed`

**Causa**: Conexión cerrada por timeout o firewall

**Solución**:

1. Verificar `SetConnMaxLifetime` en configuración
2. Revisar logs de MySQL para conexiones cerradas
3. Aumentar `connMaxLifetime` si el timeout de MySQL es mayor

```go
// En configuration/db.go
sqlDB.SetConnMaxLifetime(10 * time.Minute) // Aumentar si MySQL wait_timeout > 5min
```

---

### Problema: Using wrong database connection

**Síntoma**: Table doesn't exist para una tabla que sí existe

**Causa**: Usar `db` en lugar de `dbSII` o `dbProducts`

**Solución**:

Verificar que el modelo corresponde a la conexión correcta:

```go
// ERROR - Usar db para factura de SII
db := configuration.GetDBPrincipal()
var invoice models.IssuedInvoice  // ← Tabla en BD SII
db.First(&invoice)  // ← Error: Table not found

// CORRECTO
dbSII := configuration.GetDBSII()
dbSII.First(&invoice)
```

---

### Problema: Variables de entorno no cargadas

**Síntoma**: Valores vacíos o default en conexión

**Causa**: `.env` no configurado o variables incorrectas

**Solución**:

```bash
# 1. Validar variables
bash scripts/validate_env.sh

# 2. Verificar formato en .env
cat .env | grep DATA_BASE

# 3. Recargar variables
source .env
```

---

### Problema: Connection pool exhausted

**Síntoma**: `too many connections` o `connection pool exhausted`

**Causa**: Más de 25 conexiones simultáneas (`SetMaxOpenConns(25)`)

**Solución**:

1. Identificar si hay connection leaks (conexiones no cerradas)
2. Aumentar `MaxOpenConns` si el tráfico lo justifica
3. Revisar que no se esté creando nueva conexión en cada request

```go
// CORRECTO - Reutilizar conexión existente
var db *gorm.DB  // Variable global o inyectada

// INCORRECTO - Crear nueva conexión cada vez
func handler() {
    db, _ := configuration.GetDBPrincipal()  // ← Nueva conexión cada request
}
```

---

### Problema: Transacción fallida entre bases de datos

**Síntoma**: Error al intentar usar `db.Transaction()` con modelos de diferentes BDs

**Causa**: GORM transactions no funcionan entre diferentes conexiones físicas

**Solución**:

Implementar compensating transaction pattern:

```go
func CreateInvoiceWithProfile(invoice *models.IssuedInvoice, profile *models.Profile) error {
    db := configuration.GetDBPrincipal()
    dbSII := configuration.GetDBSII()
    
    // Paso 1: Crear perfil en BD Principal
    if err := db.Create(&profile).Error; err != nil {
        return err
    }
    
    // Paso 2: Crear factura en BD SII
    if err := dbSII.Create(&invoice).Error; err != nil {
        // Compensating transaction: rollback manual
        db.Delete(&profile)
        return fmt.Errorf("failed to create invoice, profile rolled back: %w", err)
    }
    
    return nil
}
```

## Consideraciones Especiales

### Rendimiento

- **Latencia entre BDs**: Si las BDs están en servidores diferentes, considerar latencia de red
- **N+1 problem**: Evitar queries en bucle que accedan a BDs diferentes
- **Connection pool tuning**: Ajustar según tráfico concurrente esperado

### Seguridad

- **Credenciales separadas**: Cada BD tiene sus propias credenciales (no compartir)
- **Principio de mínimo privilegio**: Cada conexión debe tener solo permisos necesarios
- **No mezclar conexiones**: Nunca usar credenciales de Principal en SII o viceversa

### Migraciones

- **Independent migrations**: Cada BD tiene su propio set de migraciones
- **Order matters**: Migrar Principal antes que SII/Products si hay dependencias
- **Foreign keys cross-BD**: No usar foreign keys físicas entre BDs, implementar a nivel aplicación

### Testing

```go
// Tests deben mockear las 3 conexiones
func TestMultiDatabaseOperation(t *testing.T) {
    // Mock de conexiones
    mockDB := setupMockDB("principal")
    mockDBSII := setupMockDB("sii")
    mockDBProducts := setupMockDB("products")
    
    // Test de operación que usa las 3
    result, err := operationWithThreeDBs(mockDB, mockDBSII, mockDBProducts)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Roadmap de Mejoras

- [ ] Implementar connection pool dinámico según load
- [ ] Agregar health check endpoint para las 3 BDs
- [ ] Implementar circuit breaker para BDs no disponibles
- [ ] Soporte para read replicas en BD Principal
- [ ] Automatic retry logic con exponential backoff
- [ ] Metrics de conexión (Prometheus) para cada BD

## Convenciones del Proyecto

### Nomenclatura de Variables

- `db` → Siempre Principal DB
- `dbSII` → Siempre SII DB
- `dbProducts` → Siempre Products/Economato DB

### Organización de Código

- Legacy MVC: `services/{ModuleName}/` usa la BD correspondiente
- Clean Architecture: `infrastructure/db/gorm_repository.go` inyecta conexión
- Controllers: Deben obtener conexión de `configuration/`, no crear nuevas

### Logging

```go
log.Printf("[DB-PRINCIPAL] Query executed in %v", duration)
log.Printf("[DB-SII] Invoice created: %s", invoiceID)
log.Printf("[DB-PRODUCTS] Sync completed: %d articles", count)
```

Usar prefijos `[DB-XXX]` para identificar qué BD generó el log.