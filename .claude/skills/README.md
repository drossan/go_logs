# Skills Disponibles

Este directorio contiene las habilidades técnicas inyectables en los agentes del proyecto Reverence Hotels API.

## Skills por Categoría

### Language/

#### go-clean-architecture

**Descripción**: Esta skill debe usarse cuando el usuario necesite implementar o refactorizar módulos usando Clean Architecture/Hexagonal en Go. Se activa con peticiones como "crea un módulo con clean architecture", "refactoriza este controller a clean architecture", "implementa un nuevo feature usando hexagonal architecture", o "agrega un repositorio GORM siguiendo clean architecture".

**Propósito**: Provides clean-architecture-related expertise and capabilities

---

#### echo-routes

**Descripción**: Esta skill debe usarse cuando el usuario necesite crear, modificar, o diagnosticar rutas en aplicaciones Echo v4. Se activa con peticiones como "crea una nueva ruta en Echo", "añade un endpoint al API", "configura middleware en Echo", o "diagnosticar problemas de rutas".

**Propósito**: Provides echo-routes-related expertise and capabilities

---

#### gorm-models

**Descripción**: Esta skill debe usarse cuando el usuario necesite crear, modificar o refactorizar modelos GORM en el proyecto Reverence Hotels API. Se activa con peticiones como "crea un modelo GORM para", "añade un campo al modelo", "refactoriza el modelo Profile", o "necesito un nuevo modelo con relaciones".

**Propósito**: Provides gorm-models-related expertise and capabilities

---

#### jwt-auth

**Descripción**: Esta skill debe usarse cuando el usuario necesite implementar, configurar, o diagnosticar autenticación JWT en aplicaciones Go con Echo framework. Se activa con peticiones como "implementa autenticación JWT", "configura middleware JWT", "genera tokens JWT", o "diagnostica error de autenticación 401".

**Propósito**: Provides jwt-auth-related expertise and capabilities

---

#### multi-database

**Descripción**: Esta skill debe usarse cuando el usuario necesite trabajar con operaciones que involucren múltiples bases de datos MySQL en el proyecto Reverence Hotels API. Se activa con peticiones como "conecta a las tres bases de datos", "consulta la base de datos del SII", "sincroniza datos entre economato y la base principal", o "crea una query que use db y dbSII".

**Propósito**: Provides multi-database-related expertise and capabilities

---

#### sii-invoicing

**Descripción**: Esta skill debe usarse cuando el usuario necesite trabajar con facturación electrónica española del SII (Suministro Inmediato de Información), incluyendo generación de XML, comunicación con la AEAT, gestión de certificados digitales y códigos de seguimiento CSV. Se activa con peticiones como "crear factura electrónica para el SII", "enviar facturas a Hacienda", "generar XML para AEAT", o "consultar estado de envío SII".

**Propósito**: Provides sii-invoicing-related expertise and capabilities

---

### Base/

#### go-code-reviewer

**Descripción**: Esta skill debe usarse cuando el usuario necesite revisar código Go para identificar problemas de calidad, bugs potenciales, incumplimiento de mejores prácticas y desviaciones de los patrones arquitectónicos del proyecto. Se activa con peticiones como "revisa este código", "encuentra bugs en este handler", "evalúa la calidad de este servicio" o "revisa si sigue las convenciones del proyecto".

**Propósito**: Provides code-reviewer-related expertise and capabilities

---

#### debug-master

**Descripción**: Esta skill debe usarse cuando el usuario necesite investigar y resolver errores complejos, analizar stack traces, identificar race conditions, diagnosticar memory leaks, o proponer soluciones a bugs en código Go. Se activa con peticiones como "investiga este error", "¿por qué falla este código?", "tengo un panic en producción", o "ayuda a debuggear este problema".

**Propósito**: Provides debug-master-related expertise and capabilities

---

#### technical-writer

**Descripción**: Esta skill debe usarse cuando el usuario necesite crear, actualizar o mejorar documentación técnica, comentarios de código, guías de usuario o especificaciones. Se activa con peticiones como "documenta este código", "crea una guía de usuario", "mejora los comentarios de esta función", "genera documentación de API", o "escribe especificaciones técnicas".

**Propósito**: Provides technical-writer-related expertise and capabilities

---

## Cómo Funcionan las Skills

Las skills se inyectan en los agentes mediante el frontmatter YAML:

```yaml
---
skills:
  - go-expert
  - cobra-cli
  - http-client
---
```

