---
name: run-docker
version: 1.0.0
author: platform-team
description: Comando para construir y ejecutar contenedores Docker de Reverence Hotels API con validación de configuración y verificación de estado
usage: "run-docker [--migrate=yes] [--pro=yes] [--build] [--env=local|pre|pro]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: [pre-flight, docker-cli]
---

# Comando: Run Docker

## Objetivo

Orquestar la construcción y ejecución de contenedores Docker para Reverence Hotels API, asegurando que:

1. La configuración de Docker es válida
2. Las variables de entorno están correctamente definidas
3. El contenedor se construye exitosamente
4. La aplicación se inicia correctamente dentro del contenedor
5. Las migraciones se ejecutan si se solicitan

**Output**: Reporte de estado del contenedor y logs de validación.

## Contexto Requerido del Usuario

- [ ] Entorno de ejecución (local, pre-production, production)
- [ ] ¿Ejecutar migraciones de base de datos? (--migrate=yes/no)
- [ ] ¿Modo producción? (--pro=yes/no)
- [ ] ¿Forzar rebuild de imagen? (--build)
- [ ] Archivo de entorno a utilizar (.env, .env.production, .env.pre)

## Análisis Inicial (Obligatorio)

### Validaciones Pre-ejecución

El command debe evaluar:

- **Estado de Docker**: Docker daemon está ejecutándose
- **Puerto disponible**: El puerto 1331 no está en uso
- **Archivos de configuración**: Dockerfile y .env correspondiente existen
- **Recursos del sistema**: CPU y memoria suficientes
- **Red Docker**: Red de Docker disponible para contenedores

### Pre-ejecución: Checklist Obligatorio

```json
{
  "validation_passed": true,
  "risks": [
    "Puerto 1331 ya en uso por otro proceso",
    "Archivo .env.production no encontrado",
    "Imagen base de Go puede necesitar actualización"
  ],
  "required_approvals": [],
  "estimated_complexity": "low",
  "blocking_issues": [],
  "docker_daemon_status": "running",
  "port_1331_available": true
}
```

## Selección de Agentes y Skills

### Fase 1: Validación de Pre-condiciones

```yaml
responsible: go-orchestrator
accountable: go-orchestrator
consulted: [ pre-flight ]
informed: []
```

**Justificación**: 
- `go-orchestrator` está especializado en coordinación de tareas de desarrollo para el proyecto Reverence Hotels API
- Su descripción indica "aplica razonamiento estructurado sin conocimiento técnico hardcodeado"
- Ideal para orquestar la validación y ejecución de Docker

### Fase 2: Construcción de Imagen Docker

```yaml
responsible: go-orchestrator
accountable: go-orchestrator
consulted: []
informed: []
```

### Fase 3: Ejecución y Validación

```yaml
responsible: go-orchestrator
accountable: go-orchestrator
consulted: [ debug-master ]
informed: []
```

## Flujo de Trabajo Orquestado

### 1. Validación de Pre-condiciones (go-orchestrator)

**Objetivo**: Verificar que el entorno está listo para ejecutar Docker

**Tareas**:

- Verificar que Docker daemon está ejecutándose (`docker info`)
- Validar que el puerto 1331 está disponible
- Comprobar existencia de Dockerfile en la raíz del proyecto
- Verificar archivo de entorno correspondiente (.env, .env.pre, .env.production)
- Validar formato del archivo de entorno
- Verificar espacio en disco disponible (mínimo 2GB)

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: None requerido para validaciones básicas
- **MCPs**: None
- **Validador**: go-orchestrator (self-validating)

**Criterios de Salida**:

- [ ] Docker daemon responde correctamente
- [ ] Puerto 1331 está libre
- [ ] Dockerfile existe y es válido
- [ ] Archivo .env correspondiente existe
- [ ] Espacio en disco suficiente

---

### 2. Construcción de Imagen Docker (go-orchestrator)

**Objetivo**: Construir la imagen Docker de Reverence Hotels API

**Tareas**:

- Leer versión actual desde package.json o variable de entorno
- Ejecutar `docker build` con tags apropiados (ej: reverence-hotels-api:v2.19.8)
- Validar que la construcción finalice sin errores
- Verificar tamaño de imagen generada (debe ser < 1GB)
- Etiquetar imagen como `latest` si es exitoso

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: None requerido
- **Dependencias**: Fase 1 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Imagen Docker construida exitosamente
- [ ] Imagen etiquetada con versión correcta
- [ ] Tamaño de imagen dentro de límites aceptables
- [ ] No hay warnings críticos en el build

---

### 3. Ejecución de Contenedor (go-orchestrator)

**Objetivo**: Iniciar el contenedor con la configuración apropiada

**Tareas**:

- Detener y eliminar contenedor anterior si existe (docker rm -f reverence-api)
- Crear nuevo contenedor con configuración:
  - Nombre: reverence-api
  - Puertos: 1331:1331
  - Env file: según entorno especificado
  - Flags: --migrate=[yes/no], --pro=[yes/no]
- Ejecutar contenedor en modo detached (-d)
- Obtener container ID

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: None requerido
- **Dependencias**: Fase 2 completada
- **Validador**: go-orchestrator

**Criterios de Salida**:

- [ ] Contenedor creado y ejecutándose
- [ ] Puerto 1331 mapeado correctamente
- [ ] Variables de entorno inyectadas correctamente

---

### 4. Validación de Estado (go-orchestrator | Validado por debug-master)

**Objetivo**: Verificar que la aplicación está funcionando correctamente dentro del contenedor

**Tareas**:

- Verificar estado del contenedor (docker ps)
- Obtener logs iniciales del contenedor (docker logs reverence-api)
- Validar que el servidor escucha en puerto 1331
- Verificar que las 3 conexiones a bases de datos se establecieron
- Comprobar que no hay errores de startup en los logs
- Si --migrate=yes, verificar que migraciones se ejecutaron
- Ejecutar health check si está disponible

**Asignación**:

- **Agente**: go-orchestrator
- **Skills**: `debug-master` (para diagnosticar problemas si los hay)
- **MCPs**: None
- **Validador**: Self-validation con skill debug-master como soporte

**Criterios de Salida**:

- [ ] Contenedor está en estado "running"
- [ ] Server inició correctamente (log: "Server started on port 1331")
- [ ] Conexiones a BD establecidas (3 conexiones)
- [ ] Migraciones ejecutadas si se solicitaron
- [ ] No hay errores críticos en logs

---

## Uso de otros Commands y MCPs

### Commands Invocados

```yaml
commands_invocados:
  - name: pre-flight
    trigger: pre-ejecución (fase 1)
    output_required: validación de entorno completa
    purpose: Validar estado del proyecto antes de ejecutar Docker
```

### MCPs Utilizados

No se requieren MCPs específicos para este command. Todo se ejecuta mediante comandos de Docker CLI y Bash.

---

## Output y Artefactos

| Artefacto              | Ubicación                                   | Formato    | Validador        | Obligatorio |
|------------------------|---------------------------------------------|------------|------------------|-------------|
| Reporte de ejecución   | `.claude/reports/docker-run-{date}.md`      | Markdown   | go-orchestrator  | Sí          |
| Logs del contenedor    | `.claude/logs/docker-logs-{date}.log`       | Plain text | -                | Sí          |
| JSON de estado         | `.claude/reports/docker-state-{date}.json`  | JSON       | schema-validator | Sí          |
| Docker inspect         | `.claude/reports/docker-inspect-{date}.json`| JSON       | -                | No          |

### Formato del Reporte de Ejecución

```markdown
# Docker Run Report - {timestamp}

## Configuración
- **Entorno**: {local|pre|pro}
- **Migraciones**: {yes|no}
- **Modo Producción**: {yes|no}
- **Imagen**: reverence-hotels-api:{version}

## Validaciones
- [x] Docker daemon: running
- [x] Puerto 1331: available
- [x] Dockerfile: valid
- [x] .env file: found

## Construcción
- **Status**: {success|failed}
- **Imagen ID**: {sha256}
- **Tamaño**: {MB}
- **Tiempo**: {seconds}

## Ejecución
- **Container ID**: {id}
- **Status**: {running|exited}
- **Puertos**: 0.0.0.0:1331->1331/tcp

## Validación de Aplicación
- [x] Server iniciado
- [x] BD Principal conectada
- [x] BD SII conectada
- [x] BD Products conectada
- [x] Migraciones ejecutadas (si aplica)

## Logs de Startup
{primeras 20 líneas de logs}

## Próximos Pasos
- Ver logs en vivo: `docker logs -f reverence-api`
- Acceder API: http://localhost:1331
- Detener: `docker stop reverence-api`
```

---

## Rollback y Cancelación

### Si el usuario cancela durante la construcción:

1. **Detener build**: `docker build --no-cache` se detiene con Ctrl+C
2. **Eliminar imágenes parciales**: `docker image prune -f`
3. **Registrar cancelación**:
   ```
   .claude/logs/cancelled-docker-build-{timestamp}.log
   ```

### Si el contenedor falla al iniciar:

1. **Obtener logs de error**: `docker logs reverence-api`
2. **Analizar con skill debug-master**: Investigar causa del fallo
3. **Limpiar contenedor fallido**: `docker rm -f reverence-api`
4. **Generar reporte de fallo** en `.claude/reports/docker-failed-{date}.md`

### Si hay problemas post-deployment:

1. **Detener contenedor**: `docker stop reverence-api`
2. **Verificar rollback necesario** (¿volver a versión anterior?)
3. **Limpiar recursos**: `docker system prune -f`
4. **Notificar al usuario** con diagnóstico completo

### Estados Finales Posibles

- `success`: Contenedor corriendo correctamente, API accesible
- `failed_build`: Error construyendo imagen
- `failed_start`: Contenedor no inició o se crasheó
- `validation_failed`: Pre-condiciones no cumplidas
- `cancelled`: Usuario canceló ejecución

---

## Reglas Críticas

- **No modificación de código**: Este command no crea ni modifica archivos de código
- **Orquestación pura**: Solo ejecuta comandos de Docker CLI y Bash
- **Validaciones obligatorias**: No construir imagen si pre-condiciones fallan
- **Logs completos**: Siempre capturar output de comandos Docker
- **Limpieza automática**: Eliminar contenedores anteriores antes de crear nuevos
- **Reporte detallado**: Generar artefactos en `.claude/reports/` siempre
- **Non-destructive**: Nunca eliminar imágenes sin confirmación explícita
- **Idempotencia**: Ejecutar múltiples veces debe dejar el sistema en estado válido

---

## Diagnóstico de Problemas Comunes

### Puerto 1331 en uso

```bash
# Identificar proceso
lsof -i :1331

# Opción 1: Matar proceso
kill -9 {PID}

# Opción 2: Usar puerto diferente
docker run -p 1332:1331 ...
```

### Docker daemon no running

```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

### Build falla por falta de memoria

```bash
# Aumentar memoria de Docker (Docker Desktop > Settings > Resources)
# O limpiar caché de Docker
docker system prune -a -f
```

### Contenedor inicia pero API no responde

1. Verificar logs: `docker logs reverence-api`
2. Verificar variables de entorno: `docker exec reverence-api env`
3. Entrar al contenedor: `docker exec -it reverence-api sh`
4. Verificar proceso: `docker exec reverence-api ps aux`

---

## Ejemplos de Uso

### Caso 1: Desarrollo Local

```bash
/run-docker --env=local --migrate=no
```

**Resultado esperado**:
- Imagen construida (o cacheada si existe)
- Contenedor iniciado en modo desarrollo
- Sin migraciones
- API accesible en http://localhost:1331

### Caso 2: Pre-producción con Migraciones

```bash
/run-docker --env=pre --migrate=yes --build
```

**Resultado esperado**:
- Imagen reconstruida forzadamente (--build)
- Contenedor con .env.pre
- Migraciones ejecutadas al startup
- Logs de migraciones capturados

### Caso 3: Producción

```bash
/run-docker --env=pro --migrate=yes --pro=yes
```

**Resultado esperado**:
- Contenedor en modo producción (--pro=yes)
- Migraciones ejecutadas
- Todas las 3 BD conectadas
- Cron jobs activados
- Health checks ejecutándose

---

## Acción del Usuario

Para ejecutar el contenedor Docker de Reverence Hotels API, proporciona:

1. **Entorno**: ¿local, pre-production o production? (default: local)
2. **Migraciones**: ¿Ejecutar migraciones de BD? (yes/no, default: no)
3. **Modo Producción**: ¿Iniciar en modo producción? (yes/no, default: no)
4. **Rebuild**: ¿Forzar reconstrucción de imagen? (yes/no, default: no)
5. **Archivo de entorno**: ¿Qué archivo .env usar? (default: .env para local)

**Ejemplo de solicitud válida**:
> "Ejecuta Docker en modo pre-producción con migraciones y rebuild de imagen."

**Uso rápido**:
```bash
/run-docker --env=pre --migrate=yes --build