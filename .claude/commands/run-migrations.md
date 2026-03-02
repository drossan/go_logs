---
name: run-migrations
version: 1.0.0
author: platform-team
description: Ejecuta migraciones de base de datos en Reverence Hotels API con validación de esquema y verificación de integridad
usage: "run-migrations [--environment={local|pre|pro}] [--skip-seed={true|false}]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: [database-schema-validator, migration-coordinator]
---

# Comando: Run Migrations

## Objetivo

Ejecutar el proceso de migración de bases de datos para Reverence Hotels API, incluyendo:

- GORM AutoMigrate para todas las tablas
- Manejo de foreign keys (migración GORM v1→v2)
- Recreación de vistas (view_reduce_profiles)
- Seed data opcional (users, forms, levels, menu tree, SMTP config)
- Validación post-migración de integridad de datos

**Es un command ejecutable** que ejecuta acciones directas sobre las bases de datos sin generar plan previo.

## Contexto Requerido del Usuario

- [ ] Entorno de ejecución (local, pre, pro)
- [ ] Confirmación de backup de bases de datos (production)
- [ ] Skip seed data (true/false) - solo para ejecuciones específicas
- [ ] Validación de conexiones disponibles (3 bases de datos MySQL)

## Análisis Inicial (Obligatorio)

### Validaciones Pre-ejecución

El command debe evaluar:

- Estado de las 3 conexiones de base de datos (Principal, SII, Products)
- Espacio disponible en disco para migraciones
- Versión actual del esquema vs versión esperada
- Existencia de backups (entornos pre/pro)
- Permisos de escritura en bases de datos

```json
{
  "validation_passed": true,
  "risks": [
    "Modifica esquema de 3 bases de datos simultáneamente",
    "Requiere reinicio del servidor post-migración"
  ],
  "required_approvals": ["database-admin"],
  "estimated_complexity": "medium",
  "blocking_issues": [],
  "databases_status": {
    "principal": "connected",
    "sii": "connected",
    "products": "connected"
  }
}
```

## Selección de Agentes y Skills

El proceso de migración requiere un agente especializado en operaciones de bases de datos y migraciones de datos.

### Ejecución de Migraciones

```yaml
responsible: migration-specialist
accountable: go-orchestrator
consulted: [ multi-database, gorm-models ]
informed: [ technical-writer ]
```

**Rationale**:
- **migration-specialist**: Agente especializado en migraciones de datos y transiciones arquitectónicas
- **go-orchestrator**: Orquestador que valida y aprueba la ejecución técnica
- **multi-database skill**: Proporciona expertise en gestión de las 3 bases de datos MySQL del proyecto
- **gorm-models skill**: Proporciona conocimiento sobre modelos GORM y AutoMigrate

## Flujo de Trabajo Orquestado

### 1. Validación de Conexiones (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Verificar que las 3 bases de datos están accesibles

**Tareas**:

- Validar conexión a base de datos Principal (`sensesho_api`)
- Validar conexión a base de datos SII (`reverence_sii`)
- Validar conexión a base de datos Products (`economato`)
- Verificar permisos de ALTER, CREATE, DROP
- Comprobar espacio disponible en disco

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `multi-database`
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Las 3 conexiones están activas y responden
- [ ] Permisos de DDL confirmados
- [ ] Espacio en disco suficiente (>1GB libre)

---

### 2. Backup de Esquemas (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Crear snapshot del estado actual antes de migrar

**Tareas**:

- Exportar esquema actual de las 3 bases de datos
- Guardar timestamp de backup
- Almacenar en `.claude/backups/migration-{timestamp}/`

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `multi-database`
- **Dependencias**: Fase 1 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] 3 archivos de esquema generados
- [ ] Timestamp registrado en log de migraciones
- [ ] Backups almacenados en ubicación estandarizada

---

### 3. Ejecución de AutoMigrate (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Ejecutar GORM AutoMigrate en todas las entidades

**Tareas**:

- Ejecutar `db.AutoMigrate()` para todos los modelos GORM
- Ejecutar `dbSII.AutoMigrate()` para modelos de SII
- Ejecutar `dbProducts.AutoMigrate()` para modelos de economato
- Capturar y logear cualquier error o warning

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `gorm-models`, `multi-database`
- **Dependencias**: Fase 2 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] AutoMigrate ejecutado en las 3 bases de datos
- [ ] Logs de migración generados
- [ ] Errores críticos abortan el proceso

---

### 4. Manejo de Foreign Keys (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Resolver problemas de foreign keys de migración GORM v1→v2

**Tareas**:

- Eliminar foreign keys existentes problemáticas
- Recrear foreign keys con sintaxis GORM v2
- Validar referencias entre tablas

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `gorm-models`, `multi-database`
- **Dependencias**: Fase 3 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Foreign keys eliminadas y recreadas
- [ ] Referencias validadas
- [ ] Sin errores de integridad referencial

---

### 5. Recreación de Vistas (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Recrear vistas después de migración de esquema

**Tareas**:

- Eliminar vista `view_reduce_profiles` si existe
- Recrear vista con definición actualizada
- Validar que la vista retorna datos correctamente

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `multi-database`
- **Dependencias**: Fase 4 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Vista `view_reduce_profiles` recreada
- [ ] Query de prueba ejecutado exitosamente
- [ ] Sin errores de sintaxis SQL

---

### 6. Seed Data (Opcional) (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Poblar datos iniciales si se requiere

**Tareas**:

- Insertar usuarios iniciales (solo si no existen)
- Insertar forms del sistema
- Insertar niveles de usuario
- Insertar árbol de menú
- Insertar configuración SMTP

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `gorm-models`, `multi-database`
- **Dependencias**: Fase 5 completada
- **Validador**: go-orchestrator
- **Condición**: Solo si `--skip-seed=false`

**Criterios de Salida**:

- [ ] Seed data insertada sin duplicados
- [ ] Logs de cantidad de registros creados
- [ ] Validación de datos críticos (admin user, forms)

---

### 7. Validación Post-Migración (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Verificar integridad post-migración

**Tareas**:

- Ejecutar queries de validación en cada base de datos
- Verificar cantidad de registros críticos
- Validar que las vistas retornan datos
- Generar reporte de estado

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `multi-database`, `gorm-models`
- **Dependencias**: Fase 6 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Reporte de validación generado
- [ ] Todas las validaciones pasan
- [ ] Sin datos corruptos o perdidos

---

### 8. Generación de Reporte (migration-specialist | Validado por go-orchestrator)

**Objetivo**: Documentar resultados de la migración

**Tareas**:

- Generar reporte en Markdown con:
  - Timestamp de ejecución
  - Esquemas migrados
  - Tablas modificadas/creadas
  - Errores y warnings
  - Resultados de validaciones
- Guardar en `.claude/reports/migration-{timestamp}.md`

**Asignación**:

- **Agente**: migration-specialist
- **Skills**: `technical-writer`
- **Dependencias**: Fase 7 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Reporte generado en ubicación estandarizada
- [ ] Contiene todas las secciones requeridas
- [ ] Formato Markdown válido

## Uso de otros Commands y MCPs

```yaml
commands_invocados:
  - name: database-schema-validator
    trigger: post-fase-7
    output_required: validation-report.json

mcps_utilizados:
  - name: mysql-connection-checker
    config: .claude/mcp-configs/mysql-validator.json
    purpose: Validar conexiones y permisos

  - name: backup-creator
    config: .claude/mcp-configs/backup.json
    purpose: Crear snapshots de esquemas

contexto_compartido:
  location: .claude/context/migration-state.json
  format: JSON
  data:
    last_migration_timestamp: string
    last_migration_status: string
    databases_migrated: array
```

## Output y Artefactos

| Artefacto                 | Ubicación                                      | Formato    | Validador            | Obligatorio |
|---------------------------|------------------------------------------------|------------|----------------------|-------------|
| Reporte de migración      | `.claude/reports/migration-{timestamp}.md`     | Markdown   | `go-orchestrator`    | Sí          |
| Backup de esquemas        | `.claude/backups/migration-{timestamp}/`        | SQL        | -                    | Sí (pre/pro)|
| Log de ejecución          | `.claude/logs/run-migrations-{date}.log`       | Plain text | -                    | Sí          |
| Reporte de validación     | `.claude/reports/validation-{timestamp}.json`  | JSON       | `schema-validator`   | Sí          |
| Estado de migración       | `.claude/context/migration-state.json`         | JSON       | `migration-specialist` | Sí        |

## Rollback y Cancelación

Si el command falla o el usuario cancela durante la ejecución:

### Procedimiento de Rollback

1. **Detener migración en curso**: Abortar ejecución de AutoMigrate
2. **Restaurar desde backup**:
   - Identificar backup más reciente en `.claude/backups/`
   - Ejecutar scripts de restauración en las 3 bases de datos
   - Validar que el esquema se restauró correctamente
3. **Limpiar artefactos parciales**:
   - Eliminar reportes incompletos
   - Limpiar logs parciales
4. **Registrar rollback**:
   ```bash
   .claude/logs/rollback-migration-{timestamp}.log
   ```
5. **Notificar estado**: Actualizar `migration-state.json` con status "rolled_back"

### Estados Finales Posibles

- `completed`: Migración exitosa, todas las validaciones pasan
- `failed`: Error crítico durante AutoMigrate o seed
- `cancelled`: Cancelado por usuario durante ejecución
- `partial`: Migración completada con warnings no críticos

## Reglas Críticas

- **Validación obligatoria**: Ninguna migración sin verificar las 3 conexiones
- **Backups obligatorios**: En entornos pre/pro, siempre crear backup antes
- **Abort ante error crítico**: Cualquier error en AutoMigrate debe detener el proceso
- **Logs completos**: Cada fase debe registrar inicio, fin y resultado
- **No modificación de código**: Este command solo ejecuta migraciones, no modifica archivos Go
- **Restauración automática**: Si falla, intentar restaurar desde backup más reciente
- **Validación post-migración**: No considerar exitosa sin pasar todas las validaciones
- **Estado consistente**: Siempre dejar las bases de datos en estado válido

## Casos de Uso Especiales

### Migración en Producción

Para producción, el command debe:

1. Solicitar confirmación explícita del usuario
2. Verificar que existe backup reciente (<24h)
3. Ejecutar en modo detallado (verbose logging)
4. Generar reporte completo antes y después
5. Notificar a equipos dependientes

### Skip Seed Data

Si se especifica `--skip-seed=true`, omitir la fase 6 pero ejecutar todas las demás.

## Acción del Usuario

Para ejecutar las migraciones, especifica:

1. **Entorno**: `local`, `pre`, o `pro` (ej: `--environment=pro`)
2. **Skip Seed**: `true` si no deseas insertar datos de prueba (ej: `--skip-seed=true`)
3. **Confirmación**: Para producción, confirma que tienes backup reciente

**Ejemplos de uso válidos**:

> `run-migrations --environment=local`
> 
> "Ejecuta migraciones en entorno local con seed data incluido"

> `run-migrations --environment=pro --skip-seed=true`
> 
> "Ejecuta migraciones en producción sin seed data (requiere confirmación de backup)"

> `run-migrations --environment=pre`
> 
> "Ejecuta migraciones en pre-producción con seed data incluido"