# Commands Disponibles

Este directorio contiene los comandos del proyecto Reverence Hotels API.

## Comandos Configurados

### plan-manage

**Descripción**: Meta-command para orquestar la creación, aprobación y ejecución de planes técnicos en el proyecto Reverence Hotels API

**Uso**: `plan-manage [create|approve|execute|list|cancel] [plan-id] [--force]`

---

### orchestrator

**Descripción**: Orchestration command for coordinating development tasks in Reverence Hotels API project. Analyzes requests and selects optimal agents and skills for execution.

**Uso**: `orchestrator [task-description] [--priority=normal] [--context=additional-info]`

---

### pre-flight

**Descripción**: Comando de pre-flight para validar el estado del proyecto antes de iniciar desarrollo, testing o despliegue en Reverence Hotels API

**Uso**: `pre-flight [--environment=local|pre|pro] [--check=deps|tests|db|all]`

---

### run-server

**Descripción**: Ejecuta el servidor de desarrollo de Reverence Hotels API con las opciones especificadas (migraciones, producción, etc.)

**Uso**: `run-server [--migrate] [--pro]`

---

### run-migrations

**Descripción**: Ejecuta migraciones de base de datos en Reverence Hotels API con validación de esquema y verificación de integridad

**Uso**: `run-migrations [--environment={local|pre|pro}] [--skip-seed={true|false}]`

---

### run-tests

**Descripción**: Ejecuta tests en el proyecto Reverence Hotels API con opciones de cobertura, filtrado por módulo y análisis de resultados

**Uso**: `run-tests [--coverage] [--module=path] [--verbose]`

---

### run-docker

**Descripción**: Comando para construir y ejecutar contenedores Docker de Reverence Hotels API con validación de configuración y verificación de estado

**Uso**: `run-docker [--migrate=yes] [--pro=yes] [--build] [--env=local|pre|pro]`

---

### check-auth

**Descripción**: Comando para verificar el sistema de autenticación y autorización en Reverence Hotels API, validando JWT, middleware, niveles de acceso y configuración de seguridad

**Uso**: `check-auth [--scope=full|jwt|middleware|authorization]`

---

### sync-pms

**Descripción**: Comando para sincronizar datos del PMS (Property Management System) con Reverence Hotels API, incluyendo artículos, proveedores y establecimientos

**Uso**: `sync-pms [--entity=articles|providers|establishments|all] [--force]`

---

### test-module

**Descripción**: Comando para test-module en Reverence Hotels API

**Uso**: `test-module [module-path] [--coverage] [--verbose]`

---

### clean-db

**Descripción**: Comando para clean-db en Reverence Hotels API (project). Limpia y resetea las bases de datos del proyecto, eliminando datos de desarrollo o pruebas según el entorno.

**Uso**: `clean-db [--environment=local|pre|pro] [--backup=yes|no] [--force]`

---

## Uso de los Comandos

Los comandos definen flujos de trabajo orquestados que utilizan uno o más agentes.

Para más información sobre cómo se crean estos comandos, consulta las guías en `.claude/embeds/command_guide.md`.
