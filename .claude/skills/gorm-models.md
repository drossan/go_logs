---
name: gorm-models
description: Esta skill debe usarse cuando el usuario necesite crear, modificar o refactorizar modelos GORM en el proyecto Reverence Hotels API. Se activa con peticiones como "crea un modelo GORM para", "añade un campo al modelo", "refactoriza el modelo Profile", o "necesito un nuevo modelo con relaciones".
license: Complete terms in LICENSE.txt
version: 1.0.0
author: Reverence Hotels Development Team
category: language
tags: [gorm, go, models, database, mysql, clean-architecture]---

# GORM Models

Skill especializada en el diseño, creación y refactorización de modelos GORM para el proyecto Reverence Hotels API. Proporciona patrones, convenciones y mejores prácticas específicas para la arquitectura híbrida del proyecto (MVC + Clean Architecture).

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:

- Se necesite crear un nuevo modelo GORM para una entidad de dominio
- Se requiera añadir campos o relaciones a modelos existentes
- Se deba refactorizar un modelo siguiendo las convenciones del proyecto
- Sea necesario migrar modelos entre arquitecturas (MVC → Clean Architecture)
- Se requiera establecer relaciones entre modelos (Has One, Has Many, Many2Many)
- Sea necesaria la configuración de hooks, índices o constraints en modelos

Triggers comunes:

- "Crea un modelo GORM para gestionar [entidad]"
- "Añade el campo [campo] al modelo [Modelo]"
- "Refactoriza el modelo [Modelo] para usar convenciones GORM v2"
- "Necesito una relación muchos a muchos entre [ModeloA] y [ModeloB]"
- "El modelo [Modelo] necesita índices por [campo]"
- "Implementa hooks para antes de crear/actualizar en [Modelo]"

## Workflow Principal

### 1. Análisis Inicial

Antes de crear o modificar un modelo:

1. **Identificar el módulo objetivo**:
   - ¿Es un módulo Legacy MVC? → Ubicar en `models/`
   - ¿Es un módulo Clean Architecture? → Ubicar en `internal/{module}/domain/`

2. **Verificar convenciones del proyecto**:
   - Revisar `references/gorm_conventions.md` para patrones existentes
   - Consultar modelos similares en el código base
   - Confirmar la base de datos destino (Principal, SII, o Products)

3. **Determinar requisitos del modelo**:
   - Campos requeridos y su tipo
   - Relaciones con otros modelos
   - Índices necesarios para rendimiento
   - Hooks o callbacks requeridos
   - Soft delete vs hard delete

### 2. Diseño del Modelo

Aplicar estructura según arquitectura:

#### Para Legacy MVC (`models/`):

```go
package Profile

import (
    "time"
    "gorm.io/gorm"
)

type Profile struct {
    gorm.Model
    ID              uint   `gorm:"primarykey" json:"id"`
    UserID          uint   `gorm:"not null;index" json:"user_id"`
    Name            string `gorm:"type:varchar(255);not null" json:"name"`
    // ... más campos
}
```

#### Para Clean Architecture (`internal/{module}/domain/`):

```go
package domain

import (
    "time"
    "gorm.io/gorm"
)

type Entity struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    // Campos de negocio
    Field     string         `gorm:"type:varchar(255)" json:"field"`
}
```

### 3. Definición de Campos

Seguir convenciones de nomenclatura:

**Tipos de datos comunes**:

```go
// Strings - siempre especificar longitud
Name string `gorm:"type:varchar(255)" json:"name"`

// Enteros - usar uint para PKs y FKs
ID   uint   `gorm:"primarykey" json:"id"`
FKID uint   `gorm:"not null;index" json:"fk_id"`

// Booleanos
IsActive bool `gorm:"default:true" json:"is_active"`

// Time - usar time.Time
CreatedAt time.Time `json:"created_at"`

// Textos largos
Description string `gorm:"type:text" json:"description"`

// Decimales para dinero/medidas
Price float64 `gorm:"type:decimal(10,2)" json:"price"`
```

**Etiquetas GORM comunes**:

```go
`gorm:"primarykey"`              // Primary key
`gorm:"not null"`                 // No permite nulos
`gorm:"unique"`                   // Valor único
`gorm:"index"`                    // Índice simple
`gorm:"uniqueIndex"`              // Índice único
`gorm:"index:idx_name"`           // Índice con nombre
`gorm:"type:varchar(255)"`        // Tipo específico
`gorm:"default:0"`                // Valor por defecto
`gorm:"comment:Descripción"`      // Comentario en tabla
`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"` // Foreign key constraints
```

### 4. Relaciones entre Modelos

Consultar `references/relationship_patterns.md` para patrones detallados.

**Has One (Uno a Uno)**:

```go
// User tiene un Profile
type User struct {
    gorm.Model
    Profile Profile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Profile struct {
    gorm.Model
    UserID uint `gorm:"uniqueIndex;not null"`
}
```

**Has Many (Uno a Muchos)**:

```go
// Level tiene muchos LevelPrivileges
type Level struct {
    gorm.Model
    Name            string            `gorm:"type:varchar(100);not null;unique"`
    LevelPrivileges []LevelPrivilege `gorm:"foreignKey:LevelID;constraint:OnDelete:CASCADE"`
}

type LevelPrivilege struct {
    gorm.Model
    LevelID uint `gorm:"not null;index"`
}
```

**Many to Many (Muchos a Muchos)**:

```go
// User ↔ Roles
type User struct {
    gorm.Model
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    gorm.Model
    Users []User `gorm:"many2many:user_roles;"`
}
```

**Polimórficas**:

```go
// Document puede pertenecer a User o Company
type Document struct {
    gorm.Model
    OwnerID   uint
    OwnerType string `gorm:"type:varchar(50)"` // "User" o "Company"
}
```

### 5. Hooks y Callbacks

```go
// Hook antes de crear
func (p *Profile) BeforeCreate(tx *gorm.DB) error {
    // Validaciones o transformaciones antes de insertar
    if p.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

// Hook antes de actualizar
func (p *Profile) BeforeUpdate(tx *gorm.DB) error {
    // Lógica pre-actualización
    return nil
}

// Hook después de crear
func (p *Profile) AfterCreate(tx *gorm.DB) error {
    // Lógica post-creación (ej: notificaciones)
    return nil
}
```

### 6. Validaciones con Tags

```go
type Profile struct {
    Name     string `gorm:"type:varchar(255);not null" validate:"required,min=3,max=255"`
    Email    string `gorm:"type:varchar(255);unique;not null" validate:"required,email"`
    Age      int    `gorm:"not null" validate:"gte=0,lte=150"`
    Password string `gorm:"type:varchar(255);not null" validate:"required,min=8"`
}
```

### 7. Validación

Verificar que:

- [ ] El modelo sigue las convenciones de GORM v2
- [ ] Todos los campos tienen tags `json` para API
- [ ] Los índices están definidos correctamente (FKs, campos de búsqueda)
- [ ] Las relaciones tienen constraints apropiadas (CASCADE, SET NULL)
- [ ] Los campos tienen tipos de datos apropiados en MySQL
- [ ] Soft delete está configurado si es necesario (`gorm.DeletedAt`)
- [ ] La estructura de directorios corresponde a la arquitectura (MVC vs Clean)
- [ ] Los hooks no crean transacciones infinitas

### 8. Output

Presentar el modelo con:

- Estructura completa del modelo
- Comentarios explicando campos complejos
- Ejemplo de migración si es necesario
- Relaciones con otros modelos documentadas

## Recursos de la Skill

### Referencias (`references/`)

#### `references/gorm_conventions.md`

**Contenido**: Convenciones específicas del proyecto Reverence Hotels API para modelos GORM

**Cuándo consultar**: Al crear cualquier modelo nuevo para asegurar consistencia con el código existente

**Estructura**:

- Patrones de nomenclatura (campos, tablas, índices)
- Convenciones de arquitectura híbrida
- Configuración de multi-database
- Tags GORM más usados en el proyecto
- Errores comunes de migración GORM v1 → v2

**Búsqueda rápida**:

```bash
# Buscar convenciones de nombres
grep -i "nomenclature\|naming" references/gorm_conventions.md

# Buscar patrones de relación
grep -A 10 "relationship" references/gorm_conventions.md
```

---

#### `references/relationship_patterns.md`

**Contenido**: Patrones de relaciones entre modelos usados en el proyecto

**Cuándo consultar**: Al establecer relaciones entre modelos (Has One, Has Many, Many2Many)

**Estructura**:

- Ejemplos de cada tipo de relación
- Configuración de foreign keys y constraints
- Relaciones polimórficas
- Self-referencing relationships
- Joins preloading optimizados

**Búsqueda rápida**:

```bash
# Buscar tipo de relación específico
grep -i "many2many\|has many\|has one" references/relationship_patterns.md

# Buscar configuración de constraints
grep -A 5 "constraint" references/relationship_patterns.md
```

---

#### `references/migration_gotchas.md`

**Contenido**: Problemas conocidos y soluciones en migraciones GORM

**Cuándo consultar**: Si el modelo requiere migración o actualización de esquema

**Estructura**:

- Foreign key issues en GORM v1 → v2
- View recreation después de migración
- Problemas con multi-database
- Índices compuestos
- renaming columns y tablas

**Búsqueda rápida**:

```bash
# Buscar problema específico
grep -i "foreign key\|index\|view" references/migration_gotchas.md
```

---

#### `references/model_examples.md`

**Contenido**: Ejemplos reales de modelos del proyecto

**Cuándo consultar**: Para ver implementaciones de referencia

**Estructura**:

- `models/Profile/profile.go` - Modelo principal de perfiles
- `models/Invoices/issued_invoice.go` - Facturas emitidas (SII)
- `internal/backend/user/domain/entity.go` - Clean Architecture example

---

### Scripts (`scripts/`)

#### `scripts/generate_model.sh`

**Propósito**: Generar boilerplate para un nuevo modelo GORM

**Uso**:

```bash
bash scripts/generate_model.sh <ModelName> <module> [--clean-arch]
```

**Parámetros**:

- `ModelName`: Nombre del modelo en PascalCase (ej: UserProfile)
- `module`: Módulo destino (ej: Profile, Invoices, User)
- `--clean-arch`: Flag opcional para usar estructura Clean Architecture

**Output**: Archivo de modelo con boilerplate básico

**Ejemplo**:

```bash
# Generar modelo Legacy MVC
bash scripts/generate_model.sh Holiday Profile

# Generar modelo Clean Architecture
bash scripts/generate_model.sh AuditLog internal/backend/audit --clean-arch
```

---

#### `scripts/validate_model.go`

**Propósito**: Validar que un modelo cumple las convenciones del proyecto

**Uso**:

```bash
go run scripts/validate_model.go <path/to/model.go>
```

**Parámetros**:

- `path/to/model.go`: Ruta al archivo del modelo a validar

**Output**: Reporte de validación con errores y warnings

**Ejemplo**:

```bash
go run scripts/validate_model.go models/Profile/profile.go
```

---

#### `scripts/migrate_model.go`

**Propósito**: Ejecutar migración para un modelo específico

**Uso**:

```bash
go run scripts/migrate_model.go <ModelName> [--database <db>]
```

**Parámetros**:

- `ModelName`: Nombre del modelo a migrar
- `--database`: Base de datos (principal/sii/products) - default: principal

**Output**: SQL generado y resultado de la migración

**Ejemplo**:

```bash
# Migrar modelo a BD principal
go run scripts/migrate_model.go Profile

# Migrar modelo a BD SII
go run scripts/migrate_model.go IssuedInvoice --database sii
```

---

### Assets (`assets/`)

#### `assets/templates/model_template.go`

**Tipo**: Template de código Go

**Uso**: Boilerplate base para crear modelos nuevos

**Modificaciones**: Reemplazar placeholders como `{{ModelName}}`, `{{TableName}}`, `{{Fields}}`

---

#### `assets/templates/model_clean_arch_template.go`

**Tipo**: Template de código Go para Clean Architecture

**Uso**: Boilerplate para modelos en `internal/{module}/domain/`

**Modificaciones**: Reemplazar `{{ModuleName}}`, `{{EntityName}}`, `{{Fields}}`

---

#### `assets/boilerplate/gorm_model_snippets.txt`

**Tipo**: Fragmentos de código reutilizables

**Uso**: Copy-paste de patrones comunes (soft deletes, timestamps, relaciones)

**Contenido**:
- Timestamps con timezone
- Soft delete configuration
- Common relationship patterns
- Index definitions

## Ejemplos de Uso

### Ejemplo 1: Crear modelo para nueva entidad

**Input del usuario**:
> "Necesito un modelo GORM para gestionar documentos firmados digitalmente"

**Proceso**:

1. Identificar que es un módulo nuevo, probablemente Legacy MVC (`models/Signatures/`)
2. Consultar `references/gorm_conventions.md` para convenciones de nombres
3. Determinar campos necesarios: ID, UserID, DocumentType, Status, SignedAt, FileURL
4. Decidir arquitectura: Legacy MVC porque es módulo simple CRUD
5. Aplicar template de `assets/templates/model_template.go`

**Output esperado**:

```go
package Signatures

import (
    "time"
    "gorm.io/gorm"
)

type DocumentToSign struct {
    gorm.Model
    ID              uint           `gorm:"primarykey" json:"id"`
    UserID          uint           `gorm:"not null;index" json:"user_id"`
    User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    
    DocumentType    string         `gorm:"type:varchar(50);not null" json:"document_type"`
    DocumentName    string         `gorm:"type:varchar(255);not null" json:"document_name"`
    
    Status          string         `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, signed, rejected
    ExternalTxID    string         `gorm:"type:varchar(255);unique" json:"external_tx_id,omitempty"`
    
    SignedAt        *time.Time     `json:"signed_at,omitempty"`
    FileURL         string         `gorm:"type:varchar(500)" json:"file_url,omitempty"`
    
    // Porta Sigma integration
    SignatureData   string         `gorm:"type:text" json:"signature_data,omitempty"`
}

// TableName especifica el nombre de la tabla en MySQL
func (DocumentToSign) TableName() string {
    return "document_to_sign"
}

// BeforeCreate hook para generar ExternalTxID si no existe
func (d *DocumentToSign) BeforeCreate(tx *gorm.DB) error {
    if d.ExternalTxID == "" {
        d.ExternalTxID = generateTxID()
    }
    return nil
}
```

---

### Ejemplo 2: Añadir relación Many2Many existente

**Input del usuario**:
> "El modelo User necesita una relación muchos a muchos con Role para gestionar permisos"

**Proceso**:

1. Consultar `references/relationship_patterns.md` sección "Many to Many"
2. Revisar modelo User existente en `models/Auth/user.go`
3. Crear modelo Role si no existe
4. Establecer relación con tabla join `user_roles`

**Output esperado**:

```go
// En models/Auth/user.go - añadir campo:
type User struct {
    gorm.Model
    // ... campos existentes
    Roles []Role `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"roles,omitempty"`
}

// Crear nuevo archivo models/Auth/role.go:
package Auth

import (
    "gorm.io/gorm"
)

type Role struct {
    gorm.Model
    ID          uint   `gorm:"primarykey" json:"id"`
    Name        string `gorm:"type:varchar(50);not null;unique" json:"name"`
    Description string `gorm:"type:varchar(255)" json:"description"`
    IsActive    bool   `gorm:"default:true" json:"is_active"`
    
    Users []User `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

func (Role) TableName() string {
    return "roles"
}
```

---

### Ejemplo 3: Refactorizar modelo a Clean Architecture

**Input del usuario**:
> "Refactoriza el modelo Form de Legacy MVC a Clean Architecture"

**Proceso**:

1. Identificar modelo actual en `models/Forms/form.go`
2. Consultar `references/gorm_conventions.md` para estructura Clean Arch
3. Crear nueva estructura en `internal/backend/form/domain/entity.go`
4. Mover lógica de dominio a entity
5. Crear repository en `internal/backend/form/infrastructure/db/`
6. Actualizar referencias en controllers/services

**Output esperado**:

```go
// internal/backend/form/domain/entity.go
package domain

import (
    "time"
    "gorm.io/gorm"
)

// Form representa una entidad de formulario en el dominio
type Form struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    // Campos de identidad
    Name        string `gorm:"type:varchar(100);not null;unique" json:"name"`
    Title       string `gorm:"type:varchar(255);not null" json:"title"`
    Description string `gorm:"type:text" json:"description"`
    
    // Configuración de ruta API
    PathAPI     string `gorm:"type:varchar(500);not null" json:"path_api"`
    
    // Metadata
    Icon        string `gorm:"type:varchar(50)" json:"icon"`
    IsEnabled   bool   `gorm:"default:true;index" json:"is_enabled"`
    SortOrder   int    `gorm:"default:0" json:"sort_order"`
    
    // Relaciones
    Fields      []FormField `gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE" json:"fields,omitempty"`
    
    // Nivel asociado (para authorization)
    LevelID     uint `gorm:"index" json:"level_id,omitempty"`
}

// FormField representa un campo dentro de un formulario
type FormField struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    FormID     uint `gorm:"not null;index" json:"form_id"`
    
    Name       string `gorm:"type:varchar(100);not null" json:"name"`
    Label      string `gorm:"type:varchar(255);not null" json:"label"`
    FieldType  string `gorm:"type:varchar(50);not null" json:"field_type"` // text, number, date, select, etc.
    Required   bool   `gorm:"default:false" json:"required"`
    Options    string `gorm:"type:text" json:"options,omitempty"` // JSON para selects, checkboxes
    SortOrder  int    `gorm:"default:0" json:"sort_order"`
}
```

---

### Ejemplo 4: Añadir índices y optimizar modelo existente

**Input del usuario**:
> "El modelo Profile necesita índices para búsquedas frecuentes por email y DNI"

**Proceso**:

1. Leer modelo actual en `models/Profile/profile.go`
2. Identificar campos de búsqueda frecuentes: Email, DNI, UserID
3. Consultar `references/migration_gotchas.md` para patrones de índices
4. Añadir tags `index` y `uniqueIndex` según corresponda

**Output esperado**:

```go
type Profile struct {
    gorm.Model
    ID              uint   `gorm:"primarykey" json:"id"`
    
    // Identificación
    UserID          uint   `gorm:"not null;uniqueIndex:idx_user_id" json:"user_id"`
    User            User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
    
    // Datos personales
    Name            string `gorm:"type:varchar(255);not null" json:"name"`
    Surname         string `gorm:"type:varchar(255);not null" json:"surname"`
    
    // Identificadores únicos - CON ÍNDICES
    DNI             string `gorm:"type:varchar(20);unique;not null;index:idx_dni" json:"dni"`
    Email           string `gorm:"type:varchar(255);unique;not null;index:idx_email" json:"email"`
    
    // Contacto - ÍNDICE PARA BÚSQUEDAS
    Phone           string `gorm:"type:varchar(20);index:idx_phone" json:"phone"`
    
    // Employment - ÍNDICE COMPUESTO PARA FILTRADOS
    EmployeeID      string `gorm:"type:varchar(50);unique;index:idx_employee_id" json:"employee_id"`
    LevelID         uint   `gorm:"not null;index:idx_level_id" json:"level_id"`
    DepartmentID    uint   `gorm:"index:idx_department_id" json:"department_id,omitempty"`
    
    // Status - ÍNDICES PARA REPORTES
    IsActive        bool   `gorm:"default:true;index:idx_is_active" json:"is_active"`
    HireDate        *time.Time `json:"hire_date,omitempty"`
    
    // Índice compuesto para queries comunes: filtrar empleados activos por departamento
    // Esto optimiza: SELECT * FROM profiles WHERE department_id = X AND is_active = true
}

// Índices adicionales especificados en migración
// idx_profile_dept_active: (department_id, is_active)
// idx_profile_level_active: (level_id, is_active)
```

## Presentación de Resultados

Al completar la creación o modificación de un modelo:

1. **Resumir cambios**: "Modelo {ModelName} creado/modificado. Estructura: {resumen}"
2. **Formato de output**: Código Go completo con comentarios explicativos
3. **Incluir metadatos**:
   - Ubicación del archivo (ruta)
   - Arquitectura usada (Legacy MVC / Clean Architecture)
   - Base de datos destino (Principal/SII/Products)
   - Relaciones establecidas
   - Índices creados
4. **Notas de migración**: Si requiere migración, indicar pasos necesarios

**Ejemplo de resumen**:

```markdown
## Modelo DocumentToSign Creado

**Ubicación**: `models/Signatures/document_to_sign.go`
**Arquitectura**: Legacy MVC
**Base de datos**: Principal (sensesho_api)

**Estructura**:
- 10 campos incluyendo timestamps y soft delete
- Relación con User (foreignKey: UserID)
- Índices en UserID, ExternalTxID

**Migración requerida**:
```bash
go run scripts/migrate_model.go DocumentToSign
```

**Relaciones**:
- User (Belongs To)

**Índices**:
- idx_user_id en UserID
- unique index en ExternalTxID
```

## Troubleshooting

### Problema: Error "model not found" al migrar

**Síntoma**: `Error: 1005: Can't create table` o `foreign key constraint fails`

**Causa**: 
- Relación con tabla que aún no existe
- Foreign key mal configurado
- Orden incorrecto de migración

**Solución**:

1. Migrar primero modelos dependientes (ej: User antes que Profile)
2. Verificar que los tipos de datos de FK coinciden (uint con uint)
3. Consultar `references/migration_gotchas.md` sección "Foreign Keys"

```bash
# Migrar en orden correcto
go run scripts/migrate_model.go User
go run scripts/migrate_model.go Profile
```

---

### Problema: Índices duplicados o conflicto uniqueIndex

**Síntoma**: `Error 1061: Duplicate key name` o `Error 1062: Duplicate entry`

**Causa**: 
- Índice ya existe de migración anterior
- uniqueIndex en campo con datos duplicados existentes

**Solución**:

1. Eliminar índices conflictivos manualmente:
```sql
DROP INDEX idx_name ON table_name;
```

2. Limpiar datos duplicados antes de migrar:
```sql
DELETE t1 FROM table_name t1
INNER JOIN table_name t2
WHERE t1.id > t2.id AND t1.duplicate_field = t2.duplicate_field;
```

3. Re-ejecutar migración

---

### Problema: Relaciones no cargan con Preload

**Síntoma**: `Preload("User").Find(&profiles)` retorna estructuras vacías en relación

**Causa**: 
- Tag `gorm:"foreignKey"` mal configurado
- Nombre de campo no coincide con estructura
- Falta especificar `references`

**Solución**:

```go
// Incorrecto
type Profile struct {
    UserID uint `gorm:"index"`
    User   User `gorm:"foreignKey:UserID"`
}

// Correcto - especificar references si no es ID
type Profile struct {
    UserUUID string `gorm:"type:varchar(36);index"`
    User     User   `gorm:"foreignKey:UserUUID;references:UUID"`
}
```

Consultar `references/relationship_patterns.md` para más ejemplos.

---

### Problema: Soft delete no funciona correctamente

**Síntoma**: Registros eliminados aparecen en queries normales

**Causa**: 
- Falta campo `DeletedAt gorm.DeletedAt`
- Se está usando `Unscoped()` sin intención
- Driver MySQL no configurado correctamente

**Solución**:

1. Verificar que el modelo tiene DeletedAt:
```go
type Model struct {
    ID        uint           `gorm:"primarykey"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
    // ... más campos
}
```

2. No usar `Unscoped()` a menos que sea intencional:
```go
// Incorrecto - trae registros eliminados
db.Find(&profiles)

// Correcto - omite deleted
db.Find(&profiles)

// Solo para traer también eliminados
db.Unscoped().Find(&profiles)
```

---

### Problema: Hook crea recursión infinita

**Síntoma**: Stack overflow o timeout en operaciones de base de datos

**Causa**: Hook llama a `Save()` o `Update()` dentro de otro hook

**Solución**:

```go
// INCORRECTO - crea recursión
func (p *Profile) BeforeUpdate(tx *gorm.DB) error {
    p.UpdatedAt = time.Now()
    tx.Save(p) // ← Dispara BeforeUpdate nuevamente
    return nil
}

// CORRECTO - usa directamente el DB
func (p *Profile) BeforeUpdate(tx *gorm.DB) error {
    // No usar tx.Save/Update dentro de hooks
    p.UpdatedAt = time.Now()
    return nil
}
```

---

### Problema: Multi-database connection error

**Síntoma**: `Error 1046: No database selected` o `Error 1049: Unknown database`

**Causa**: 
- Modelo no especifica conexión correcta (db, dbSII, dbProducts)
- Migración ejecutada en database equivocado

**Solución**:

1. Verificar que la migración usa la conexión correcta:
```go
// En routes/echo.go o cmd/server/backend/
// Principal DB
db.AutoMigrate(&models.Profile{})

// SII DB
dbSII.AutoMigrate(&models.IssuedInvoice{})

// Products DB
dbProducts.AutoMigrate(&models.Article{})
```

2. Para queries manuales, especificar conexión:
```go
// Correcto
var profiles []models.Profile
database.db.Find(&profiles)

// Para SII
var invoices []models.IssuedInvoice
database.dbSII.Find(&invoices)
```

Consultar `references/gorm_conventions.md` sección "Multi-Database".

## Consideraciones Especiales

### Rendimiento

- **Índices compuestos**: Usar para queries que filtran por múltiples campos frecuentemente
  - Ejemplo: `(department_id, is_active)` optimiza búsquedas de empleados activos por departamento
- **Preload estratégico**: Usar `Preload()` solo para relaciones necesarias, evitar N+1 queries
- **Batch operations**: Usar `CreateInBatches()` para insertar múltiples registros (> 1000)

### Seguridad

- **GORM tags de seguridad**: Usar `omitempty` en campos sensibles para no exponer en JSON
  ```go
  Password string `gorm:"type:varchar(255);not null" json:"-"` // No serializa
  InternalNotes string `gorm:"type:text" json:"internal_notes,omitempty"`
  ```
- **Input sanitization**: Nunca confiar en input del usuario para nombres de tabla/columna
- **SQL Injection**: GORM previene automáticamente, pero validar raw queries si se usan

### Compatibilidad

- **GORM v2**: El proyecto migró de GORM v1 a v2. Revisar breaking changes en `references/migration_gotchas.md`
- **MySQL 8.0**: Asegurar compatibilidad de tipos (ej: `varchar` length máximo aumentó)
- **Timezones**: Usar `time.Time` con timezone UTC para consistencia
  ```go
  CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
  ```

### Migraciones

- **Order matters**: Migrar primero tablas padre, luego hijas (User → Profile → LevelPrivilege)
- **Foreign keys en GORM v2**: Requieren configuración explícita de constraints
  ```go
  FK uint `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
  ```
- **Views**: Recrear views después de cada migración (ver `migration/base.go`)

## Convenciones Específicas del Proyecto

### Nomenclatura de Tablas

```go
// GORM por defecto usa pluralización (User → users)
// Para overridar:
func (Profile) TableName() string {
    return "profiles" // Tabla en MySQL
}
```

### Campos de Auditoría

```go
type AuditModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
    
    CreatedBy uint `gorm:"index" json:"created_by,omitempty"`
    UpdatedBy uint `gorm:"index" json:"updated_by,omitempty"`
}
```

### Enums en MySQL

```go
// GORM no soporta enums nativamente, usar varchar con validación
type Status string

const (
    StatusPending   Status = "pending"
    StatusApproved  Status = "approved"
    StatusRejected  Status = "rejected"
)

type Document struct {
    Status Status `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
}
```

### JSON Fields

```go
// Para datos flexibles, usar tipo JSON de MySQL
type Metadata map[string]interface{}

type Config struct {
    ID       uint    `gorm:"primarykey" json:"id"`
    Metadata Metadata `gorm:"type:json" json:"metadata"`
}

// Uso
config.Metadata = map[string]interface{}{
    "key1": "value1",
    "key2": 123,
}
```

## Integración con Clean Architecture

### Domain Entity vs GORM Model

En `internal/{module}/domain/`:

```go
// Entidad pura de dominio - sin dependencias de GORM
type User struct {
    ID       UserID
    Email    Email
    Profile  Profile
}
```

En `infrastructure/db/` (persistence layer):

```go
// Modelo GORM para persistencia
type UserGORM struct {
    gorm.Model
    ID       uint   `gorm:"primarykey"`
    Email    string `gorm:"type:varchar(255);unique"`
    Profile  ProfileGORM `gorm:"foreignKey:UserID"`
}

// Mapper para convertir entre domain y GORM
func (g *UserGORM) ToDomain() *User {
    return &User{
        ID:      UserID(g.ID),
        Email:   Email(g.Email),
        Profile: g.Profile.ToDomain(),
    }
}
```

Consultar `examples/user_model_clean_arch.md` para implementación completa.

## Mejoras Futuras (Roadmap)

- [ ] Scripts para generar modelos desde schema de base de datos (reverse engineering)
- [ ] Validador automático de convenciones en pre-commit hooks
- [ ] Generador de diagramas ER desde modelos GORM
- [ ] Scripts para migración segura entre bases de datos (Principal ↔ SII ↔ Products)
- [ ] Integración con herramientas de validación de esquemas (golang-migrate)