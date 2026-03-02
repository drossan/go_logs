---
name: check-auth
version: 1.0.0
author: reverence-hotels-team
description: Comando para verificar el sistema de autenticación y autorización en Reverence Hotels API, validando JWT, middleware, niveles de acceso y configuración de seguridad
usage: "check-auth [--scope=full|jwt|middleware|authorization]"
type: executable
writes_code: false
creates_plan: false
requires_approval: false
dependencies: []
---

# Comando: Check Auth

## Objetivo

Verificar el estado y configuración del sistema de autenticación y autorización de Reverence Hotels API, proporcionando un diagnóstico completo sobre:

- Configuración de JWT (secret, expiration, algorithms)
- Middleware de autenticación en Echo
- Sistema de autorización basado en niveles (Level)
- Rutas protegidas y su configuración
- Integración con forms y privilegios

**Output esperado**: Reporte detallado de estado y lista de problemas detectados (si existen).

**Este comando es de tipo ejecutable**: NO genera plan, NO modifica código, solo analiza y reporta.

## Contexto Requerido del Usuario

- [ ] Alcance del chequeo (completo o específico)
- [ ] Entorno objetivo (local, pre-production, production)
- [ ] Rutas o módulos específicos a verificar (opcional)

## Análisis Inicial (Obligatorio)

Antes de iniciar el análisis, el comando debe evaluar:

- Estado actual del entorno (local/pre/pro)
- Archivos de configuración presentes (.env, config files)
- Estructura de middlewares disponible
- Modelos de autenticación/autorización implementados

### Pre-ejecución: Checklist Obligatorio

- [ ] ¿El proyecto Go está correctamente inicializado? → Verificar go.mod
- [ ] ¿Existen los archivos de configuración de JWT? → Buscar middleware y config
- [ ] ¿La estructura de middlewares está presente? → Verificar directorio middleware/
- [ ] ¿Los modelos de Level/Form están disponibles? → Revisar models/
- [ ] ¿El contexto del usuario es suficiente? → Solicitar aclaraciones si es necesario

**Output esperado**: JSON de validación antes de continuar.

```json
{
  "validation_passed": true,
  "environment": "local",
  "jwt_config_found": true,
  "middleware_structure": "present",
  "models_available": ["Level", "Form", "User"],
  "blocking_issues": []
}
```

## Selección de Agentes y Skills

### Fase 1: Análisis de Configuración JWT y Middleware

```yaml
responsible: go-debugger
accountable: go-reviewer
consulted: [ jwt-auth, echo-routes, debug-master ]
informed: [ go-orchestrator ]
```

**Justificación RACI**:
- **go-debugger** como Responsible porque su descripción menciona "investigating and resolving complex bugs, analyzing error messages, identifying race conditions" - perfecto para diagnosticar configuración
- **go-reviewer** como Accountable porque valida "calidad de código, mejores prácticas y convenciones del proyecto"
- **Skills**:
  - `jwt-auth`: proporciona "expertise and capabilities" para autenticación JWT
  - `echo-routes`: proporciona "expertise and capabilities" para rutas y middleware en Echo v4
  - `debug-master`: proporciona "expertise and capabilities" para técnicas de debugging

### Fase 2: Verificación de Sistema de Autorización

```yaml
responsible: go-reviewer
accountable: software-architect-tdd-ddd
consulted: [ go-code-reviewer, multi-database ]
informed: [ planning-agent ]
```

**Justificación RACI**:
- **go-reviewer** como Responsible porque es "Senior Go Code Reviewer especializado en razonamiento sobre calidad de código"
- **software-architect-tdd-ddd** como Accountable porque valida arquitectura y patrones de diseño
- **Skills**:
  - `go-code-reviewer`: proporciona "expertise and capabilities" para revisar calidad de código
  - `multi-database`: proporciona "expertise and capabilities" para verificar consultas a múltiples bases de datos (Principal para auth)

## Flujo de Trabajo Orquestado

### 1. Validación de Configuración JWT (go-debugger | Validado por go-reviewer)

**Objetivo**: Verificar que la configuración de JWT sea correcta y segura

**Tareas**:

- Revisar variables de entorno JWT_SECRET, JWT_EXPIRATION
- Validar que el secret tenga longitud mínima de 32 caracteres
- Verificar algoritmo de firma (RS256 recomendado)
- Comprobar configuración de claims (exp, iat, nbf)
- Analizar middleware de JWT en Echo

**Asignación**:

- **Agente**: go-debugger
- **Skills**: `jwt-auth`, `echo-routes`, `debug-master`
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Configuración JWT validada
- [ ] Secret seguro verificado
- [ ] Middleware de Echo correctamente configurado
- [ ] Reporte de problemas de seguridad generados

---

### 2. Análisis de Middleware de Autorización (go-reviewer | Validado por software-architect-tdd-ddd)

**Objetivo**: Verificar que el sistema de autorización basado en niveles funcione correctamente

**Tareas**:

- Revisar middleware de autorización en `middleware/`
- Verificar tabla de Level y LevelPrivileges
- Analizar mapeo entre Forms y PathAPI
- Comprobar lógica de validación Read/Write
- Validar que las rutas protegidas estén correctamente configuradas

**Asignación**:

- **Agente**: go-reviewer
- **Skills**: `go-code-reviewer`, `multi-database`
- **Dependencias**: Fase 1 completada
- **Validador**: software-architect-tdd-ddd

**Criterios de Salida**:

- [ ] Sistema de autorización analizado
- [ ] Mapeo de rutas a privilegios verificado
- [ ] Lógica Read/Write validada
- [ ] Reporte de incidencias generado

---

### 3. Verificación de Integración con Forms (go-debugger | Validado por go-reviewer)

**Objetivo**: Validar la integración entre autenticación y el sistema de forms dinámicos

**Tareas**:

- Revisar modelo Form y su relación con PathAPI
- Verificar que todos los endpoints estén mapeados a forms
- Comprobar rutas con múltiples paths separados por |
- Validar consistencia entre routes.go y forms en BD
- Analizar edge cases (rutas dinámicas, parámetros)

**Asignación**:

- **Agente**: go-debugger
- **Skills**: `gorm-models`, `debug-master`
- **Dependencias**: Fase 2 completada
- **Validador**: go-reviewer

**Criterios de Salida**:

- [ ] Integración Form-PathAPI verificada
- [ ] Rutas mapeadas correctamente
- [ ] Inconsistencias documentadas
- [ ] Reporte final generado

## Uso de otros Commands y MCPs

Este comando es autónomo y no invoca otros commands ni MCPs externos.

Todo el análisis se realiza mediante:
- Lectura de archivos de configuración
- Revisión de código fuente (middleware, models, routes)
- Análisis estático de estructuras de datos

## Output y Artefactos

| Artefacto                | Ubicación                                  | Formato    | Obligatorio |
|--------------------------|--------------------------------------------|------------|-------------|
| Reporte de diagnóstico   | `.claude/reports/auth-check-{date}.md`     | Markdown   | Sí          |
| Checklist de validación  | `.claude/checklists/auth-validation.json`  | JSON       | Sí          |
| Log de ejecución         | `.claude/logs/check-auth-{date}.log`       | Plain text | Sí          |

### Estructura del Reporte

```markdown
# Check Auth - Reporte de Diagnóstico

## Resumen Ejecutivo
- Estado General: ✅ Pass | ⚠️ Warnings | ❌ Fail
- Entorno: local/pre/pro
- Fecha: YYYY-MM-DD HH:MM:SS

## 1. Configuración JWT
- Secret: ✅ Seguro (256+ bits) | ⚠️ Débil (<256 bits) | ❌ No configurado
- Algoritmo: RS256 | HS256 | Otro
- Expiración: {horas} horas
- Problemas detectados:
  - [ ] Problema 1
  - [ ] Problema 2

## 2. Middleware de Autorización
- Rutas protegidas: {count}
- Niveles configurados: {count}
- Forms mapeados: {count}
- Problemas detectados:
  - [ ] Problema 1
  - [ ] Problema 2

## 3. Integración Forms-PathAPI
- Consistencia: ✅ 100% | ⚠️ Parcial | ❌ Inconsistente
- Rutas sin mapear: {count}
- Problemas detectados:
  - [ ] Problema 1

## Recomendaciones
1. Recomendación prioritaria 1
2. Recomendación 2
3. Recomendación 3
```

## Rollback y Cancelación

Como este es un comando de tipo **executable** (sin impacto en código), no requiere rollback.

Si el usuario cancela durante la ejecución:

1. **Detener análisis en curso**
2. **Guardar reporte parcial** en `.claude/reports/auth-check-partial-{timestamp}.md`
3. **Registrar cancelación** en `.claude/logs/cancelled-check-auth-{timestamp}.log`
4. **Notificar** que el análisis está incompleto

## Reglas Críticas

- **Solo lectura**: Este comando NO puede modificar ningún archivo
- **Sin generación de planes**: Si se detectan problemas que requieren cambios, solo reportarlos
- **Sin creación de código**: Si se requiere implementar fixes, derivar a otro comando o plan
- **Análisis completo**: Verificar JWT, middleware, autorización y forms
- **Reporte obligatorio**: Siempre generar un reporte en `.claude/reports/`
- **Validación previa**: Ejecutar checklist antes de iniciar análisis
- **Idempotencia**: Ejecutar múltiples veces debe producir el mismo resultado
- **Entorno específico**: Adaptar análisis según ENV (local/pre/pro)

---

## Acción del Usuario

Ejecuta el check de autenticación proporcionando:

1. **Alcance**: 
   - `--scope=full`: Análisis completo (JWT + middleware + autorización + forms)
   - `--scope=jwt`: Solo configuración JWT
   - `--scope=middleware`: Solo middleware de autorización
   - `--scope=authorization`: Solo sistema de niveles y privilegios

2. **Entorno**: Especifica si estás verificando local, pre-production o production

**Ejemplo de solicitud válida**:
> "Ejecuta check-auth con scope full para el entorno local. Quiero verificar que todo el sistema de autenticación y autorización esté correctamente configurado."