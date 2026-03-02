---
name: run-server
version: 1.0.0
author: Reverence Hotels Development Team
description: Ejecuta el servidor de desarrollo de Reverence Hotels API con las opciones especificadas (migraciones, producción, etc.)
usage: "run-server [--migrate] [--pro]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: []
---

# Comando: Run Server

## Objetivo

Ejecutar el servidor de desarrollo de Reverence Hotels API con las opciones especificadas por el usuario, validando que el entorno esté correctamente configurado y verificando que el servidor inicie correctamente.

**Este comando NO modifica código**, solo ejecuta y valida el arranque del servidor.

## Contexto Requerido del Usuario

- [ ] **Opción de migración**: ¿Ejecutar migraciones de base de datos? (`--migrate`)
- [ ] **Modo de ejecución**: ¿Desarrollo o producción? (`--pro`)
- [ ] **Puerto personalizado** (opcional): ¿Usar puerto diferente al default (1331)?
- [ ] **Variables de entorno**: ¿Archivo `.env` específico?

## Análisis Inicial (Obligatorio)

Antes de iniciar el servidor, el comando debe evaluar:

- **Estado de las bases de datos**: Verificar que las 3 conexiones MySQL están configuradas
- **Disponibilidad del puerto**: Confirmar que el puerto 1331 (o el especificado) está libre
- **Dependencias de Go**: Validar que todas las dependencias están instaladas
- **Archivos de configuración**: Verificar existencia de `.env` y archivos de certificados
- **Estado de migraciones**: Identificar si es necesario ejecutar migraciones
- **Conexiones externas**: Validar configuración de SII, Porta Sigma, Slack si aplica

### Pre-ejecución: Checklist Obligatorio

El comando debe verificar:

- [ ] ¿Existe el archivo `init.go`?
- [ ] ¿Las variables de entorno `DATA_BASE_*` están configuradas?
- [ ] ¿El puerto especificado está libre?
- [ ] ¿Existen los archivos de certificados `.pem` (si se va a usar SII)?
- [ ] ¿Go está instalado y versión es compatible?
- [ ] ¿Todas las dependencias están disponibles (`go mod` está actualizado)?

**Output esperado**: JSON de validación antes de continuar.

```json
{
  "validation_passed": true,
  "environment": "local",
  "port_available": 1331,
  "databases_configured": ["principal", "sii", "products"],
  "migrations_required": false,
  "certificates_found": ["sii_cert.pem"],
  "risks": [],
  "blocking_issues": []
}
```

## Selección de Agentes y Skills

El comando utiliza un único agente especializado en orquestación de tareas de desarrollo Go:

### Fase 1: Ejecución del Servidor

```yaml
responsible: go-orchestrator
accountable: go-orchestrator
consulted: [ echo-routes, multi-database ]
informed: []
```

**Justificación de selección**:
- **go-orchestrator**: Su descripción indica "Orchestration Agent especializado en coordinación de tareas de desarrollo para el proyecto Reverence Hotels API. Aplica razonamiento estructurado sin conocimiento técnico hardcodeado." Esto lo hace ideal para ejecutar el servidor e interpretar resultados de arranque.
- **echo-routes**: Proporciona "echo-routes-related expertise" para validar que las rutas de Echo se registren correctamente.
- **multi-database**: Proporciona "multi-database-related expertise" para verificar la conexión correcta a las 3 bases de datos MySQL.

## Flujo de Trabajo Orquestado

### 1. Validación del Entorno (go-orchestrator | Auto-validado)

**Objetivo**: Verificar que el entorno esté correctamente configurado antes de iniciar el servidor.

**Tareas**:

- Verificar existencia de archivo `init.go`
- Validar que las variables de entorno críticas estén definidas (`DATA_BASE_*`, `ENV`, `API_URL`)
- Comprobar disponibilidad del puerto (default: 1331)
- Verificar que los archivos de certificados `.pem` existen (si aplica)
- Validar que Go y las dependencias están disponibles
- Verificar estado de los 3 archivos de configuración de base de datos

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `multi-database` (para validar configuración de las 3 bases de datos)
- **Validador**: go-orchestrator (auto-validación)

**Criterios de Salida**:

- [ ] Todas las validaciones pasan sin errores críticos
- [ ] JSON de validación generado con `validation_passed: true`
- [ ] warnings reportados si hay problemas no bloqueantes

---

### 2. Ejecución del Servidor (go-orchestrator | Auto-validado)

**Objetivo**: Iniciar el servidor de Reverence Hotels API con las opciones especificadas.

**Tareas**:

- Construir comando de ejecución según parámetros:
  - Sin flags: `go run init.go`
  - Con migraciones: `go run init.go --migrate=yes`
  - Producción: `go run init.go --migrate=yes --pro=yes`
- Ejecutar el servidor en background
- Capturar stdout y stderr
- Monitorear mensajes de inicio (Echo setup, rutas registradas, conexiones DB)
- Esperar mensaje "Server started successfully" o equivalente
- Reportar estado de las conexiones a las 3 bases de datos
- Reportar estado de migraciones si se ejecutaron
- Reportar rutas registradas (total ~302 endpoints)
- Verificar que el servidor escucha en el puerto correcto

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: 
  - `echo-routes` (para validar que las rutas Echo se registraron correctamente)
  - `multi-database` (para verificar las 3 conexiones MySQL)
- **Dependencias**: Fase 1 completada
- **Validador**: go-orchestrator (interpreta logs de inicio)

**Criterios de Salida**:

- [ ] Servidor iniciado sin errores críticos
- [ ] Puerto escuchando correctamente
- [ ] Conexiones a las 3 bases de datos establecidas
- [ ] Rutas Echo registradas (verificar en logs)
- [ ] Migraciones ejecutadas exitosamente (si aplica)
- [ ] Server escuchando en puerto 1331 (o especificado)

---

### 3. Verificación de Salud del Servidor (go-orchestrator | Auto-validado)

**Objetivo**: Confirmar que el servidor está respondiendo correctamente.

**Tareas**:

- Esperar 3-5 segundos después del inicio
- Realizar request de health check si existe endpoint `/health` o similar
- Verificar que el servidor responde en el puerto configurado
- Revisar logs en busca de warnings o errores post-inicio
- Reportar métricas básicas:
  - Estado del servidor (running/failed)
  - Puertos escuchando
  - Bases de datos conectadas
  - Rutas registradas
  - Entorno (local/pre/pro)
  - PID del proceso

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `echo-routes` (para validar que el servidor Echo responde)
- **Dependencias**: Fase 2 completada exitosamente
- **Validador**: go-orchestrator (verifica respuesta HTTP)

**Criterios de Salida**:

- [ ] Servidor responde a requests HTTP
- [ ] No hay errores críticos en logs
- [ ] Métricas de estado reportadas al usuario
- [ ] URL de acceso confirmada (http://localhost:1331)

## Uso de otros Commands y MCPs

Este comando **no invoca otros commands** ya que su propósito es autocontenido.

```yaml
commands_invocados: []

mcps_utilizados: []

contexto_compartido:
  location: .claude/context/server-state.json
  format: JSON
  data:
    pid: process_id
    port: 1331
    start_time: timestamp
    environment: local|pre|pro
    databases: [principal, sii, products]
```

## Output y Artefactos

| Artefacto         | Ubicación                               | Formato    | Validador        | Obligatorio |
|-------------------|-----------------------------------------|------------|------------------|-------------|
| Reporte de estado | `.claude/reports/server-state-{ts}.md`  | Markdown   | -                | Sí          |
| JSON de validación| `.claude/logs/server-validation-{ts}.json`| JSON     | schema-validator | Sí          |
| Log de ejecución  | `.claude/logs/run-server-{date}.log`    | Plain text | -                | Sí          |

**Formato del reporte de estado**:

```markdown
# Estado del Servidor - Reverence Hotels API

## Información General

- **Estado**: ✅ Running
- **PID**: 12345
- **Puerto**: 1331
- **Entorno**: local
- **URL**: http://localhost:1331
- **Tiempo de inicio**: 2.3s

## Bases de Datos

- **Principal** (sensesho_api): ✅ Conectada
- **SII** (reverence_sii): ✅ Conectada
- **Products** (economato): ✅ Conectada

## Migraciones

- **Ejecutadas**: ✅ Sí (si --migrate=yes)
- **Tablas creadas**: 45
- **Views recreadas**: 1 (view_reduce_profiles)

## Rutas

- **Total endpoints**: 302
- **API Restricted**: 245
- **API Public**: 12
- **API Key Protected**: 8
- **Intranet**: 37

## Servicios Externos

- **SII (AEAT)**: ⚠️ No configurado en local
- **Porta Sigma**: ⚠️ No configurado en local
- **Slack**: ⚠️ Token no configurado

## Logs de Inicio

\`\`\`
2025-01-20T10:30:15Z INFO Initializing Echo framework
2025-01-20T10:30:16Z INFO Connecting to databases...
2025-01-20T10:30:17Z INFO Database principal connected
2025-01-20T10:30:17Z INFO Database sii connected
2025-01-20T10:30:17Z INFO Database products connected
2025-01-20T10:30:17Z INFO Registering 302 routes...
2025-01-20T10:30:18Z INFO Server started successfully on port 1331
\`\`\`

## Próximos Pasos

1. Accede a la API en: http://localhost:1331
2. Revisa los endpoints disponibles en la documentación
3. Para detener el servidor, usa Ctrl+C o el comando `kill-server`
```

## Rollback y Cancelación

Si el comando falla o el usuario cancela durante la ejecución:

### Procedimiento de Rollback

1. **Detener el servidor**: Si está corriendo en background, enviar `SIGTERM` al proceso
2. **Liberar el puerto**: Verificar que el puerto 1331 esté libre
3. **Eliminar artefactos parciales**:
   - Borrar logs incompletos en `.claude/logs/`
   - Limpiar reportes parciales en `.claude/reports/`
4. **Restaurar estado previo**: No hay cambios en código, pero cerrar conexiones a DB si se abrieron
5. **Registrar cancelación**:
   ```
   .claude/logs/cancelled-run-server-{timestamp}.log
   ```
6. **Reportar estado final**: Indicar que el servidor no está corriendo

### Estados Finales Posibles

- `running`: Servidor iniciado correctamente y respondiendo
- `failed`: Error al iniciar (puerto ocupado, DB no disponible, etc.)
- `cancelled`: Cancelado por usuario durante validación o inicio
- `timeout`: El servidor no inició en el tiempo esperado (>30s)

## Reglas Críticas

- **No modificación de código**: Este comando solo ejecuta el servidor existente
- **Selección RACI obligatoria**: Agente go-orchestrator seleccionado explícitamente
- **Validación previa obligatoria**: Ningún comando `go run` sin verificar el entorno primero
- **Ejecución en background**: El servidor debe correr en background para permitir monitoreo
- **Captura de logs**: Todo stdout/stderr debe capturarse para diagnóstico
- **Verificación de salud**: No considerar "exitoso" si el servidor no responde HTTP
- **Gestión de puerto**: Fallar temprano si el puerto está ocupado
- **Manejo de bases de datos**: Verificar las 3 conexiones antes de considerar exitoso
- **Timeout razonable**: Si el servidor no inicia en 30 segundos, considerar fallo
- **No dejar procesos zombie**: Si falla, asegurar que ningún proceso quede corriendo

## Acción del Usuario

Para ejecutar el servidor, especifica las opciones deseadas:

1. **Opciones de ejecución**:
   - `--migrate`: Ejecutar migraciones de base de datos (tablas, seed data, views)
   - `--pro`: Modo producción (habilita todos los cron jobs)

2. **Ejemplos de uso**:

   ```bash
   # Desarrollo simple (sin migraciones)
   run-server

   # Desarrollo con migraciones
   run-server --migrate

   # Producción completa
   run-server --migrate --pro
   ```

3. **Información del entorno**:
   - ¿Cuál es el entorno? (local/pre/pro)
   - ¿Qué archivo `.env` debo usar? (default: `.env.local`)
   - ¿Necesito un puerto diferente al 1331?

**Ejemplo de solicitud válida**:
> "Inicia el servidor en modo desarrollo con migraciones. Usa el archivo .env.local. Necesito verificar que las 3 bases de datos conecten correctamente."