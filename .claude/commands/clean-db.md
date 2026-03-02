---
name: clean-db
version: 1.0.0
author: platform-team
description: Comando para clean-db en Reverence Hotels API (project). Limpia y resetea las bases de datos del proyecto, eliminando datos de desarrollo o pruebas según el entorno.
usage: "clean-db [--environment=local|pre|pro] [--backup=yes|no] [--force]"
type: executable
writes_code: false
creates_plan: false
requires_approval: true
dependencies: []
---

# Comando: Clean Database

## Objetivo

Limpiar y resetear las bases de datos del proyecto Reverence Hotels API de manera segura, eliminando datos temporales, de pruebas o de desarrollo según el entorno especificado.

**Este comando NO genera código**, solo ejecuta operaciones de limpieza sobre las bases de datos existentes.

**Advertencia**: Este comando es destructivo y puede resultar en pérdida de datos si se ejecuta en el entorno incorrecto. Requiere confirmación explícita para entornos de producción.

## Contexto Requerido del Usuario

- [ ] Entorno objetivo (local, pre, pro)
- [ ] Tipo de limpieza (all, test-data, temp-tables, logs)
- [ ] ¿Crear backup antes de limpiar? (yes/no)
- [ ] Confirmación explícita para entornos pre/pro
- [ ] Tablas o entidades específicas a excluir (opcional)

## Análisis Inicial (Obligatorio)

Antes de cualquier acción, el command debe evaluar:

- Entorno actual (ENV variable)
- Bases de datos afectadas (3 instancias MySQL)
- Tipo de datos a eliminar
- Riesgo de pérdida de datos críticos
- Necesidad de backup previo
- Conexiones activas que podrían bloquear operaciones
- **Agente y skill óptimos para operaciones de base de datos**
- Espacio disponible para backups

### Pre-ejecución: Checklist Obligatorio

El command debe verificar:

- [ ] ¿El entorno es producción? → Requiere confirmación explícita + --force
- [ ] ¿Existe suficiente espacio en disco para backup? → Fallar si no
- [ ] ¿Las conexiones a las 3 bases de datos están activas? → Verificar conectividad
- [ ] ¿El usuario tiene privilegios DROP/DELETE? → Verificar permisos
- [ ] ¿Hay operaciones en curso (migraciones, syncs)? → Abortar y advertir

**Output esperado**: JSON de validación antes de continuar.

```json
{
  "validation_passed": true,
  "environment": "local",
  "databases": ["sensesho_api", "reverence_sii", "economato"],
  "risks": ["Pérdida de datos no recuperable sin backup", "Bloqueo de operaciones concurrentes"],
  "required_approvals": ["user-confirmation"],
  "estimated_impact": "all-data",
  "backup_required": true,
  "blocking_issues": []
}
```

## Selección de Agentes y Skills (Framework RACI)

El command debe **elegir explícitamente** los agentes y skills más adecuados utilizando el modelo RACI:

### Fase 1: Validación y Backup

```yaml
responsible: go-orchestrator
accountable: planning-agent
consulted: [ multi-database ]
informed: [ technical-writer ]
```

**Justificación**:
- **go-orchestrator**: Coordinación de tareas de desarrollo del proyecto, aplica razonamiento estructurado
- **planning-agent**: Validación de planificación y estimación de riesgos
- **multi-database**: Proporciona expertise para trabajar con las 3 bases de datos MySQL del proyecto

### Fase 2: Ejecución de Limpieza

```yaml
responsible: go-orchestrator
accountable: go-reviewer
consulted: [ multi-database ]
informed: [ go-test-runner ]
```

**Justificación**:
- **go-orchestrator**: Ejecuta las operaciones de limpieza de forma estructurada
- **go-reviewer**: Valida que las operaciones sean correctas y seguras
- **multi-database**: Maneja las operaciones en las 3 instancias MySQL
- **go-test-runner**: Verifica que el sistema quede en estado funcional post-limpieza

## Flujo de Trabajo Orquestado

### 1. Validación de Precondiciones (go-orchestrator | Validado por planning-agent)

**Objetivo**: Verificar que es seguro proceder con la limpieza

**Tareas**:

- Verificar entorno actual (local/pre/pro) mediante variable ENV
- Validar conectividad con las 3 bases de datos (db, dbSII, dbProducts)
- Verificar permisos necesarios (DROP, DELETE, TRUNCATE)
- Comprobar espacio disponible para backups (si aplica)
- Detectar operaciones activas que puedan interferir
- Validar que no hay migraciones en curso

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `multi-database`
- **Validador**: planning-agent

**Criterios de Salida**:

- [ ] Entorno identificado y validado
- [ ] Conexiones a las 3 BDs verificadas
- [ ] Permisos confirmados
- [ ] Espacio suficiente disponible (si backup requerido)
- [ ] No hay operaciones bloqueantes

---

### 2. Backup de Seguridad (go-orchestrator | Validado por planning-agent)

**Objetivo**: Crear backup de las bases de datos antes de la limpieza

**Tareas**:

- Crear backup de `sensesho_api` (base de datos principal)
- Crear backup de `reverence_sii` (base de datos SII)
- Crear backup de `economato` (base de datos productos)
- Comprimir backups y guardar con timestamp
- Verificar integridad de backups creados
- Registrar ubicación de backups en log

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `multi-database`
- **Dependencias**: Fase 1 completada
- **Validador**: planning-agent

**Criterios de Salida**:

- [ ] Backups de las 3 BDs creados exitosamente
- [ ] Backups comprimidos y almacenados
- [ ] Integridad de backups verificada
- [ ] Ubicación registrada en `.claude/logs/`

---

### 3. Ejecución de Limpieza (go-orchestrator | Validado por go-reviewer)

**Objetivo**: Limpiar las bases de datos según el tipo especificado

**Tareas**:

- **Para `sensesho_api`**:
  - Eliminar datos de pruebas (si test-data)
  - Limpiar tablas temporales (si temp-tables)
  - Resetear secuencias de auto-increment
  - Eliminar logs antiguos (si logs)
  
- **Para `reverence_sii`**:
  - Limpiar facturas de prueba
  - Eliminar registros temporales de comunicación AEAT
  
- **Para `economato`**:
  - Limpiar artículos y proveedores de prueba
  - Eliminar pedidos temporales

- **Opcional: Full Reset** (si --all):
  - TRUNCATE de todas las tablas (excepto seed data)
  - Recrear seed data básica (users, forms, levels)

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `multi-database`
- **Dependencias**: Fase 2 completada (backup exitoso)
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Datos eliminados según tipo de limpieza especificado
- [ ] Seed data recreada (si aplica)
- [ ] Contadores de auto-increment reseteados
- [ ] No errores de foreign key constraints

---

### 4. Verificación Post-Limpieza (go-orchestrator | Validado por go-test-runner)

**Objetivo**: Asegurar que el sistema queda en estado funcional

**Tareas**:

- Verificar conectividad con las 3 BDs
- Validar que tablas críticas existen y tienen estructura correcta
- Comprobar que seed data básica está presente
- Ejecutar tests deSmoke para validar API funcional
- Verificar que no hay datos huérfanos (foreign key violations)
- Generar reporte de estado post-limpieza

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `multi-database`
- **Dependencias**: Fase 3 completada
- **Validador**: go-test-runner

**Criterios de Salida**:

- [ ] Conexiones a BDs funcionales
- [ ] Estructura de tablas intacta
- [ ] Seed data verificada
- [ ] Tests de humo pasando
- [ ] Reporte de estado generado

---

## Uso de otros Commands y MCPs

```yaml
commands_invocados:
  - name: pre-flight
    trigger: pre-ejecución
    purpose: Validar estado del proyecto antes de limpieza

mcps_utilizados:
  - name: mysql-connector
    purpose: Ejecutar comandos SQL en las 3 instancias

contexto_compartido:
  location: .claude/context/db-state.json
  format: JSON
  data:
    - environment
    - databases_status
    - backup_location
    - cleanup_type
```

## Output y Artefactos

| Artefacto              | Ubicación                                    | Formato    | Validador          | Obligatorio |
|------------------------|----------------------------------------------|------------|--------------------|-------------|
| Reporte de limpieza    | `.claude/reports/clean-db-{timestamp}.md`    | Markdown   | -                  | Sí          |
| Log de ejecución       | `.claude/logs/clean-db-{date}.log`           | Plain text | -                  | Sí          |
| Backup de BDs          | `.backups/db-clean-{timestamp}/`             | SQL.gz     | -                  | Sí*          |
| Estado post-limpieza   | `.claude/context/db-state.json`              | JSON       | `schema-validator` | Sí          |
| Checklist de validación | `.claude/checklists/clean-db.json`          | JSON       | `schema-validator` | Sí          |

*Solo obligatorio si se especifica --backup=yes

## Rollback y Cancelación

Si el command falla o el usuario cancela durante la ejecución:

### Procedimiento de Rollback

1. **Detener ejecución inmediata**: Cancelar operaciones SQL en curso
2. **Restaurar desde backup** (si ya se creó):
   - Detener servidor si está corriendo
   - Restaurar `sensesho_api` desde backup
   - Restaurar `reverence_sii` desde backup
   - Restaurar `economato` desde backup
   - Reiniciar servidor
3. **Eliminar artefactos parciales**:
   - Borrar reportes incompletos en `.claude/reports/`
   - Limpiar archivos temporales
4. **Registrar cancelación**:
   ```
   .claude/logs/cancelled-clean-db-{timestamp}.log
   ```
5. **Notificar estado**: Reportar resultado de rollback al usuario

### Estados Finales Posibles

- `completed`: Limpieza exitosa, sistema funcional
- `failed`: Error durante limpieza (rollback ejecutado)
- `cancelled`: Cancelado por usuario (rollback ejecutado si aplicable)
- `partial`: Limpieza parcial (algunas BDs limpiadas)

## Reglas Críticas

- **No modificación de código**: Este comando NO modifica archivos de código Go
- **Backup obligatorio**: Nunca limpiar sin backup previo (excepto entorno local explícito)
- **Confirmación en producción**: Entornos pre/pro requieren --force + confirmación explícita
- **Validación de entorno**: Verificar ENV antes de cualquier operación
- **Manejo de 3 BDs**: Operar correctamente sobre db, dbSII y dbProducts
- **Verificación post-limpieza**: Sistema debe quedar en estado funcional
- **Registro obligatorio**: Todo proceso debe quedar en logs
- **Restauración de seed data**: Si es full reset, recrear datos básicos
- **Protección de datos críticos**: Nunca eliminar usuarios admin o configuración base

## Tipos de Limpieza Soportados

### `all` (Full Reset)
Elimina todos los datos y recrea seed data básica.
- **Riesgo**: ALTO
- **Backup**: Obligatorio
- **Uso**: Solo desarrollo local

### `test-data`
Elimina solo datos de pruebas, preserva configuración.
- **Riesgo**: Medio
- **Backup**: Recomendado
- **Uso**: Desarrollo y testing

### `temp-tables`
Limpia tablas temporales y cache.
- **Riesgo**: Bajo
- **Backup**: Opcional
- **Uso**: Mantenimiento rutinario

### `logs`
Elimina logs antiguos de base de datos.
- **Riesgo**: Muy bajo
- **Backup**: No requerido
- **Uso**: Liberación de espacio

## Ejemplos de Uso

### Limpieza completa en desarrollo
```bash
clean-db --environment=local --type=all --backup=yes
```
**Output**: Backup creado + todas las BDs reseteadas + seed data recreada

### Limpieza de datos de prueba
```bash
clean-db --environment=pre --type=test-data --backup=yes
```
**Output**: Backup creado + datos de prueba eliminados + configuración preservada

### Limpieza de logs en producción
```bash
clean-db --environment=pro --type=logs --backup=no --force
```
**Output**: Logs antiguos eliminados (requiere confirmación explícita)

## Seguridad y Protección

### Protecciones Implementadas

1. **Verificación de entorno**: Valida ENV antes de ejecutar
2. **Confirmación en producción**: Requiere --force para pre/pro
3. **Backup automático**: Crea backup antes de cualquier operación destructiva
4. **Validación de permisos**: Verifica privilegios antes de ejecutar SQL
5. **Detección de bloqueos**: Aborta si hay operaciones en curso
6. **Rollback automático**: Restaura backup si falla la limpieza
7. **Logging completo**: Registro de todas las operaciones

### Datos Nunca Eliminados

Incluso con `--type=all`, estos datos se preservan:
- Usuario admin inicial
- Formularios del sistema
- Niveles de acceso básicos
- Configuración de SMTP (si existe)
- Esquema de tablas (DROP TABLE nunca se usa)

## Métricas de Ejecución

El command reporta las siguientes métricas:

- Tiempo total de ejecución
- Tamaño de backups creados
- Filas eliminadas por BD
- Espacio liberado
- Estado final de cada BD
- Errores o warnings

## Acción del Usuario

Para ejecutar la limpieza de bases de datos, especifica:

1. **Entorno**: local, pre o pro
2. **Tipo de limpieza**: all, test-data, temp-tables, logs
3. **Backup**: yes o no
4. **Confirmación**: Para entornos pre/pro, confirma explícitamente

**Ejemplo de solicitud válida**:
> "Limpia los datos de prueba del entorno de pre-producción, crea backup antes de proceder. Tipo: test-data, backup: yes"

**Advertencia**: Este comando es destructivo. Asegúrate de tener un backup reciente antes de ejecutar en entornos pre/pro.