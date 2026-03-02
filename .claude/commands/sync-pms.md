---
name: sync-pms
version: 1.0.0
author: platform-team
description: Comando para sincronizar datos del PMS (Property Management System) con Reverence Hotels API, incluyendo artículos, proveedores y establecimientos
usage: "sync-pms [--entity=articles|providers|establishments|all] [--force]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: [run-docker, check-auth]
---

# Comando: PMS Sync Orchestrator

## Objetivo

Orquestar la sincronización de datos entre el Property Management System (PMS) y Reverence Hotels API. Este comando:

- Ejecuta endpoints de sincronización protegidos por API Key
- Valida resultados de la sincronización
- Reporta estado detallado de la operación
- Maneja errores y reintentos

**No modifica código**, solo orquesta ejecución de endpoints existentes.

## Contexto Requerido del Usuario

- [ ] Entidad a sincronizar (articles, providers, establishments, all)
- [ ] Entorno de ejecución (local, docker, producción)
- [ ] API Key válida para PMS (configurada en `.env`)
- [ ] Forzar sincronización completa si es necesaria

## Análisis Inicial (Obligatorio)

### Validaciones Pre-ejecución

```json
{
  "validation_passed": true,
  "risks": [
    "Sobrescribe datos existentes en economato DB",
    "Puede afectar dependencias entre artículos y proveedores"
  ],
  "required_approvals": [],
  "estimated_complexity": "medium",
  "blocking_issues": []
}
```

### Pre-ejecución: Checklist Obligatorio

- [ ] ¿Existe API Key configurada? → Verificar variable de entorno
- [ ] ¿Está el servidor corriendo? → Iniciar si es necesario
- [ ] ¿Están disponibles los endpoints? → Validar rutas `/api/v1/pms-sync-*`
- [ ] ¿Es necesaria sincronización completa? → Evaluar flag `--force`

## Selección de Agentes y Skills

### Fase 1: Ejecución de Sincronización

```yaml
responsible: api-integration-expert
accountable: go-orchestrator
consulted: [ http-client, debug-master ]
informed: [ go-test-runner ]
```

**Justificación**: 
- `api-integration-expert`: Especializado en integraciones de APIs externas y servicios de terceros
- `go-orchestrator`: Coordina tareas de desarrollo para el proyecto Reverence Hotels API

### Fase 2: Validación de Resultados

```yaml
responsible: go-test-runner
accountable: go-reviewer
consulted: [ debug-master ]
informed: [ technical-writer ]
```

**Justificación**:
- `go-test-runner`: Especializado en testing y validación de calidad
- `go-reviewer`: Valida que la sincronización siga mejores prácticas del proyecto

## Flujo de Trabajo Orquestado

### 1. Verificación de Precondiciones (api-integration-expert | Validado por go-orchestrator)

**Objetivo**: Asegurar que el entorno está listo para la sincronización

**Tareas**:

- Verificar que el servidor está corriendo en puerto 1331
- Validar que la API Key está configurada correctamente
- Comprobar conectividad con los tres endpoints:
  - `POST /api/v1/pms-sync-articles`
  - `POST /api/v1/pms-sync-providers`
  - `POST /api/v1/pms-sync-establishments`
- Verificar estado de la base de datos `economato`

**Asignación**:

- **Agente**: api-integration-expert
- **Skills**: `http-client` (para hacer requests de prueba)
- **MCPs**: Ninguno
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Servidor responde en puerto 1331
- [ ] API Key configurada en headers
- [ ] Endpoints retornan 401 (no autorizado) o 200 (prueba pasada)
- [ ] Base de datos `economato` accesible

---

### 2. Ejecución de Sincronización (api-integration-expert | Validado por go-orchestrator)

**Objetivo**: Ejecutar los endpoints de sincronización según la entidad seleccionada

**Tareas**:

- **Para artículos**:
  - Ejecutar `POST /api/v1/pms-sync-articles` con header `X-API-KEY`
  - Capturar respuesta (status code, body, timing)
  - Validar estructura JSON de respuesta
  - Contar registros sincronizados

- **Para proveedores**:
  - Ejecutar `POST /api/v1/pms-sync-providers` con header `X-API-KEY`
  - Capturar respuesta
  - Validar estructura JSON
  - Contar registros sincronizados

- **Para establecimientos**:
  - Ejecutar `POST /api/v1/pms-sync-establishments` con header `X-API-KEY`
  - Capturar respuesta
  - Validar estructura JSON
  - Contar registros sincronizados

- **Para all**:
  - Ejecutar en orden: establishments → providers → articles (respetar dependencias)
  - Validar que cada paso fue exitoso antes de continuar

**Asignación**:

- **Agente**: api-integration-expert
- **Skills**: `http-client` (para ejecutar requests)
- **Dependencias**: Fase 1 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Endpoints ejecutados con status 200
- [ ] Respuestas JSON válidas
- [ ] Contador de registros capturado
- [ ] Tiempos de respuesta registrados

---

### 3. Validación de Resultados (go-test-runner | Validado por go-reviewer)

**Objetivo**: Verificar que la sincronización fue correcta y completa

**Tareas**:

- Validar datos en base de datos `economato`:
  - `articles`: Contar registros, verificar campos críticos
  - `providers`: Contar registros, verificar estado activo
  - `establishments`: Contar registros, verificar relaciones
- Ejecutar queries de validación:
  - Check de integridad referencial (foreign keys)
  - Verificación de duplicados
  - Validación de datos nulos o inconsistentes
- Comparar con registros previos (si existen)
- Generar reporte de diferencias

**Asignación**:

- **Agente**: go-test-runner
- **Skills**: `debug-master` (para identificar problemas en datos)
- **Dependencias**: Fase 2 completada exitosamente
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Cantidad de registros coincide con respuesta del API
- [ ] No hay duplicados
- [ ] Integridad referencial verificada
- [ ] Datos críticos no son nulos
- [ ] Reporte de validación generado

---

### 4. Reporte y Documentación (go-orchestrator | Validado por go-reviewer)

**Objetivo**: Generar reporte detallado de la sincronización

**Tareas**:

- Compilar métricas:
  - Tiempo total de ejecución
  - Tiempo por entidad
  - Registros sincronizados
  - Errores y warnings
- Generar reporte en formato Markdown:
  - Resumen ejecutivo
  - Detalle por entidad
  - Problemas detectados
  - Recomendaciones
- Guardar log de ejecución en `.claude/logs/sync-pms-{timestamp}.log`

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `technical-writer` (para generar documentación clara)
- **Dependencias**: Fase 3 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Reporte Markdown generado
- [ ] Log de ejecución guardado
- [ ] Métricas compiladas
- [ ] Recomendaciones documentadas

## Uso de otros Commands y MCPs

```yaml
commands_invocados:
  - name: check-auth
    trigger: pre-fase-1
    output_required: auth-status.json
    purpose: Validar configuración de API Key

  - name: run-docker
    trigger: opcional (si servidor no está corriendo)
    output_required: docker-container-id
    purpose: Iniciar servidor si es necesario

mcps_utilizados:
  - name: database-connector
    config: .claude/mcp-configs/economato-db.json
    purpose: Conectar a base de datos economato para validaciones

contexto_compartido:
  location: .claude/context/pms-sync-state.json
  format: JSON
  consumers: [ sync-pms, go-test-runner ]
  data:
    - last_sync_timestamp
    - entities_synced
    - records_count
```

## Output y Artefactos

| Artefacto              | Ubicación                                      | Formato    | Validador         | Obligatorio |
|------------------------|------------------------------------------------|------------|-------------------|-------------|
| Reporte de sincronización | `.claude/reports/sync-pms-{timestamp}.md`    | Markdown   | `go-reviewer`     | Sí          |
| Log de ejecución       | `.claude/logs/sync-pms-{timestamp}.log`        | Plain text | -                 | Sí          |
| Estado de validación   | `.claude/checklists/pms-validation-{id}.json`  | JSON       | `schema-validator`| Sí          |
| Métricas de rendimiento | `.claude/metrics/sync-pms-{timestamp}.json`   | JSON       | -                 | No          |

## Rollback y Cancelación

Si el command falla o el usuario cancela durante la ejecución:

### Procedimiento de Rollback

1. **Detener sincronización en curso**:
   - Cancelar requests HTTP pendientes
   - No intentar rollback automático en bases de datos (riesgo de pérdida de datos)

2. **Registrar estado parcial**:
   - Documentar qué entidades se sincronizaron correctamente
   - Registrar errores específicos
   - Guardar en `.claude/logs/partial-sync-{timestamp}.log`

3. **Recomendaciones al usuario**:
   - Si falló midway, sugerir re-ejecutar solo entidades fallidas
   - Si hay datos inconsistentes, sugerir restauración de backup
   - Documentar pasos manuales de recuperación

4. **Notificar estados**:
   - Marcar entidades fallidas en `.claude/context/pms-sync-state.json`
   - Enviar warning si hay riesgo de duplicados

### Estados Finales Posibles

- `completed`: Todas las entidades sincronizadas exitosamente
- `partial`: Algunas entidades sincronizadas, otras fallaron
- `failed`: Error crítico, no se completó ninguna sincronización
- `cancelled`: Usuario interrumpió la ejecución

## Reglas Críticas

- **No modificación de código**: Este command solo ejecuta endpoints existentes
- **Respeto de dependencias**: Establecimientos antes de proveedores antes de artículos
- **Validación obligatoria**: Siempre verificar datos después de sincronización
- **API Key segura**: Nunca exponer la API Key en logs o reportes
- **Manejo de errores**: No continuar si una entidad falla (a menos que sea explícito)
- **Transparencia**: Reportar siempre cantidad de registros y tiempos
- **Idempotencia**: Ejecutar múltiples veces no debe crear duplicados
- **Logging exhaustivo**: Registrar cada request y respuesta para debugging

---

## Acción del Usuario

Para ejecutar la sincronización PMS, especifica:

1. **Entidad**: ¿Qué deseas sincronizar?
   - `articles` (artículos del economato)
   - `providers` (proveedores)
   - `establishments` (establecimientos/hoteles)
   - `all` (todos en orden correcto)

2. **Entorno**: ¿Dónde se ejecutará?
   - `local` (servidor local con `go run init.go`)
   - `docker` (contenedor Docker)

3. **Forzar**: ¿Sincronización completa?
   - `--force` (forzar reemplazo completo de datos)

**Ejemplos de uso válidos**:

```bash
# Sincronizar solo artículos
sync-pms --entity=articles

# Sincronizar todo en orden
sync-pms --entity=all

# Forzar sincronización completa de proveedores
sync-pms --entity=providers --force

# Sincronizar en Docker
sync-pms --entity=all --env=docker
```

**Nota**: Asegúrate de que la API Key del PMS esté configurada en tu archivo `.env` como `PMS_API_KEY`.