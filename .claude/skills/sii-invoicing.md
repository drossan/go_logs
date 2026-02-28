---
name: sii-invoicing
description: Esta skill debe usarse cuando el usuario necesite trabajar con facturación electrónica española del SII (Suministro Inmediato de Información), incluyendo generación de XML, comunicación con la AEAT, gestión de certificados digitales y códigos de seguimiento CSV. Se activa con peticiones como "crear factura electrónica para el SII", "enviar facturas a Hacienda", "generar XML para AEAT", o "consultar estado de envío SII".
license: MIT
version: 1.0.0
author: Reverence Hotels Development Team
category: language
tags: [sii, aeat, electronic-invoicing, xml, go, echo-framework]---

# SII Invoicing Skill

Skill especializada en la implementación y gestión de facturación electrónica del SII (Suministro Inmediato de Información) de la Agencia Tributaria Española (AEAT). Proporciona conocimiento técnico sobre el módulo `services/Invoices/` del proyecto Reverence Hotels API, incluyendo generación de XML firmados, gestión de certificados digitales, comunicación con los endpoints de la AEAT y procesamiento de códigos de seguimiento CSV.

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:

- Se necesite implementar o modificar funcionalidad de facturación electrónica del SII
- Se requiera generar XML facturas emitidas, recibidas o intracomunitarias
- Se deba trabajar con certificados digitales (.pem) para firmar XMLs
- Sea necesario comunicarse con los endpoints de la AEAT (envío, consulta, anulación)
- Se precise gestionar códigos de seguimiento CSV para validar envíos
- Se necesite depurar errores en la comunicación con el SII
- Se requiera implementar cron jobs para envío automático de facturas

Triggers comunes:

- "Crear una nueva factura electrónica para el SII"
- "Enviar esta factura a Hacienda"
- "Generar XML firmado para factura emitida"
- "Consultar el estado de un envío SII"
- "Implementar validación de CSV"
- "El cron job de facturas está fallando"

## Workflow Principal

### 1. Análisis Inicial

Antes de proceder con cualquier tarea de facturación SII:

1. **Identificar el tipo de factura**:
   - Emitida (F1): Venta a clientes
   - Recibida (F2): Compras a proveedores
   - Intracomunitaria (F4): Operaciones UE
   - Facturas simplificadas (R1)

2. **Verificar configuración del entorno**:
   - Variables de entorno `SII_URL_*` configuradas correctamente
   - Certificados digitales (.pem) accesibles en `services/Invoices/certificates/`
   - Conexión a base de datos SII (`dbSII`) disponible

3. **Determinar el tipo de operación**:
   - Alta de nueva factura
   - Modificación de factura existente
   - Anulación de factura
   - Consulta de estado

### 2. Ejecución

#### Generación de XML Factura Emitida

```bash
# Service: services/Invoices/issued_invoices.go
# Método: GenerateIssuedInvoiceXML(invoice models.IssuedInvoice) ([]byte, error)
```

**Proceso**:

1. Obtener datos de la factura de `models.IssuedInvoice`
2. Generar estructura XML según esquema AEAT
3. Firmar XML con certificado digital
4. Validar XML contra schema XSD
5. Retornar XML firmado como `[]byte`

#### Envío a AEAT

```bash
# Service: services/Invoices/sii_communication.go
# Método: SendToSII(xmlData []byte, invoiceType string) (string, error)
```

**Proceso**:

1. Determinar endpoint según entorno (test/producción) y tipo de factura
2. Crear solicitud HTTP con headers SOAP
3. Adjuntar certificado digital (.pem)
4. Enviar petición POST al endpoint de la AEAT
5. Parsear respuesta XML para extraer CSV
6. Retornar código de seguimiento CSV

#### Validación de CSV

```bash
# Service: services/Invoices/sii_communication.go
# Método: ValidateCSV(csvCode string) (*models.SIIValidationResponse, error)
```

**Proceso**:

1. Construir SOAP request con código CSV
2. Enviar a endpoint de validación
3. Parsear respuesta XML del estado
4. Retornar estado: Aceptado, Aceptado con errores, Rechazado

#### Cron Job de Facturas

```bash
# Task: task/sii_invoices.go
# Schedule: Cada 5 minutos (entornos pre/pro)
```

**Proceso**:

1. Consultar facturas pendientes de envío (status = "pending")
2. Generar XML para cada factura
3. Enviar a AEAT en lote
4. Actualizar estado y CSV en base de datos
5. Registrar logs de auditoría

### 3. Validación

Verificar que:

- [ ] XML generado cumple con el schema XSD de la AEAT
- [ ] XML está correctamente firmado con certificado válido
- [ ] Endpoint AEAT utilizado corresponde al entorno correcto (test/pro)
- [ ] Código CSV recibido se almacenó correctamente
- [ ] Base de datos SII se actualizó con el resultado del envío
- [ ] Logs registran toda la operación para auditoría

### 4. Output

Presentar resultados:

- **Formato**: JSON estructurado con detalles de la operación
- **Incluir**:
  - Tipo de factura y ID
  - Código CSV (si aplica)
  - Estado del envío (enviado/pendiente/error)
  - Timestamp de la operación
- **Omitir**: Datos sensibles del certificado digital

## Recursos de la Skill

### Referencias (`references/`)

#### `references/sii_api_endpoints.md`

**Contenido**: Listado completo de endpoints de la AEAT para el SII, incluyendo URLs de test y producción para cada tipo de factura y operación.

**Cuándo consultar**: Al configurar endpoints de envío o cuando un endpoint devuelve errores 404/500.

**Estructura**:

- Endpoints de Facturas Emitidas (F1)
- Endpoints de Facturas Recibidas (F2)
- Endpoints de Facturas Intracomunitarias (F4)
- Endpoints de Consulta y Validación
- Diferencias entre entornos de Test y Producción

**Búsqueda rápida**:

```bash
# Buscar endpoint específico
grep -i "facturas emitidas" references/sii_api_endpoints.md

# Buscar URLs de producción
grep -A 5 "## Producción" references/sii_api_endpoints.md
```

---

#### `references/xml_schemas.md`

**Contenido**: Especificación detallada de los schemas XML de la AEAT para cada tipo de factura, incluyendo campos obligatorios, opcionales y formatos válidos.

**Cuándo consultar**: Al generar XMLs o depurar errores de validación de schema.

**Estructura**:

- Estructura general de un XML SII
- Campos obligatorios para Facturas Emitidas (F1)
- Campos obligatorios para Facturas Recibidas (F2)
- Tipos de IVA y desglose obligatorio
- Formatos de fecha (N8, N14, N19)
- Códigos de país y moneda

**Búsqueda rápida**:

```bash
# Buscar campo específico
grep -i "ImporteTotal" references/xml_schemas.md

# Buscar tipos de factura
grep -A 10 "## Tipo de Factura" references/xml_schemas.md
```

---

#### `references/certificate_management.md`

**Contenido**: Guía de gestión de certificados digitales (.pem) para firmar XMLs del SII, incluyendo renovación, conversión de formatos y validación.

**Cuándo consultar**: Al trabajar con certificados, errores de firma o al renovar certificados.

**Estructura**:

- Formatos de certificado soportados (.pfx, .p12, .pem)
- Conversión de formato usando OpenSSL
- Validación de certificado antes del uso
- Renovación de certificados caducados
- Seguridad y almacenamiento de claves privadas

**Búsqueda rápida**:

```bash
# Buscar comandos de conversión
grep -A 5 "openssl" references/certificate_management.md

# Buscar validación de certificado
grep -i "validate" references/certificate_management.md
```

---

#### `references/error_codes.md`

**Contenido**: Tabla de códigos de error que devuelve la AEAT, con causas comunes y soluciones para cada uno.

**Cuándo consultar**: Al depurar fallos en envíos al SII o respuestas con errores.

**Estructura**:

- Errores de autenticación (certificado inválido)
- Errores de validación de XML
- Errores de conexión SOAP
- Errores de negocio (periodo cerrado, duplicidad, etc.)
- Códigos de estado HTTP

**Búsqueda rápida**:

```bash
# Buscar error por código
grep -i "ERROR_1101" references/error_codes.md

# Buscar por categoría
grep -A 10 "## Errores de Validación" references/error_codes.md
```

---

#### `references/database_models.md`

**Contenido**: Modelos de base de datos GORM para el módulo de facturación SII, incluyendo tablas en `reverence_sii` y relaciones.

**Cuándo consultar**: Al trabajar con modelos de datos, crear queries o entender relaciones entre entidades.

**Estructura**:

- `models.IssuedInvoice` - Facturas emitidas
- `models.ReceivedInvoice` - Facturas recibidas
- `models.SIISendLog` - Logs de envíos
- `models.SIIValidationResponse` - Respuestas de validación
- Relaciones entre tablas

**Búsqueda rápida**:

```bash
# Buscar modelo específico
grep -A 20 "type IssuedInvoice" references/database_models.md

# Buscar campo de modelo
grep -i "CSVCode" references/database_models.md
```

---

### Scripts (`scripts/`)

#### `scripts/validate_xml.sh`

**Propósito**: Validar un archivo XML contra el schema XSD de la AEAT antes del envío.

**Uso**:

```bash
bash scripts/validate_xml.sh <archivo.xml> <tipo_factura>
```

**Parámetros**:

- `archivo.xml`: Archivo XML a validar (obligatorio)
- `tipo_factura`: Tipo de factura - F1/F2/F4 (obligatorio)

**Output**: Resultado de validación (válido/inválido) con detalles de errores si aplica.

**Ejemplo**:

```bash
bash scripts/validate_xml.sh factura_emitida.xml F1
```

---

#### `scripts/convert_certificate.sh`

**Propósito**: Convertir certificados digitales desde formato .pfx/.p12 a .pem para uso en la API.

**Uso**:

```bash
bash scripts/convert_certificate.sh <certificado.pfx> <password>
```

**Parámetros**:

- `certificado.pfx`: Archivo de certificado en formato PKCS#12 (obligatorio)
- `password`: Contraseña del certificado (obligatorio)

**Output**: Archivo .pem generado en `services/Invoices/certificates/`.

**Ejemplo**:

```bash
bash scripts/convert_certificate.sh certificado_empresa.pfx miPasswordSeguro123
```

---

#### `scripts/test_soap_connection.sh`

**Propósito**: Probar la conexión con los endpoints SOAP de la AEAT (test y producción).

**Uso**:

```bash
bash scripts/test_soap_connection.sh <entorno>
```

**Parámetros**:

- `entorno`: Entorno a probar - test|prod (obligatorio)

**Output**: Resultado de conexión exitosa/fallida con tiempo de respuesta.

**Ejemplo**:

```bash
bash scripts/test_soap_connection.sh test
```

---

### Assets (`assets/`)

#### `assets/xml_templates/issued_invoice_template.xml`

**Tipo**: Template XML para facturas emitidas

**Uso**: Template base para generar XMLs de facturas emitidas, con placeholders para datos dinámicos.

**Modificaciones**:

- Reemplazar `{{ID_FACTURA}}` con ID real de la factura
- Reemplazar `{{NIF_EMPRESA}}` con NIF de la empresa
- Reemplazar `{{FECHA_EXPEDICION}}` con fecha en formato N8
- Completar campos de IVA según corresponda

---

#### `assets/xml_templates/received_invoice_template.xml`

**Tipo**: Template XML para facturas recibidas

**Uso**: Template base para generar XMLs de facturas recibidas de proveedores.

**Modificaciones**: Similar a issued_invoice_template pero con campos específicos de factura recibida.

---

#### `assets/xsd_schemas/`

**Tipo**: Schemas XSD oficiales de la AEAT

**Uso**: Validar XMLs generados antes de enviar a producción.

**Contenido**:

- `sii_facturas_emitidas.xsd` - Schema para facturas emitidas
- `sii_facturas_recibidas.xsd` - Schema para facturas recibidas
- `sii_facturas_intracomunitarias.xsd` - Schema para operaciones UE
- `sii_consultas.xsd` - Schema para consultas y validaciones

## Ejemplos de Uso

### Ejemplo 1: Crear y enviar factura emitida al SII

**Input del usuario**:
> "Crear una nueva factura electrónica para el SII por los servicios del hotel del mes de enero"

**Proceso**:

1. Identificar que se necesita crear factura emitida (tipo F1)
2. Recopilar datos necesarios:
   - NIF de la empresa (emisor)
   - NIF del cliente (receptor)
   - Base imponible e IVA desglosado
   - Fecha de expedición
   - Tipo de factura (normal, simplificada, etc.)
3. Generar XML usando `services/Invoices/issued_invoices.go`
4. Firmar XML con certificado digital
5. Enviar a endpoint AEAT de facturas emitidas
6. Obtener y almacenar código CSV
7. Actualizar estado en base de datos

**Output esperado**:

```json
{
  "status": "success",
  "message": "Factura emitida creada y enviada al SII correctamente",
  "invoice_id": "F2025-000123",
  "csv_code": "P202500000001234567890",
  "sii_status": "Enviado",
  "timestamp": "2025-01-22T10:30:00Z",
  "aeat_response": {
    "codigo": "0",
    "descripcion": "Correcto",
    "csv": "P202500000001234567890"
  }
}
```

---

### Ejemplo 2: Consultar estado de envío SII

**Input del usuario**:
> "Consultar el estado de la factura con CSV P202500000001234567890"

**Proceso**:

1. Extraer código CSV del request
2. Construir SOAP request de consulta
3. Enviar a endpoint de validación de AEAT
4. Parsear respuesta XML
5. Retornar estado detallado

**Output esperado**:

```json
{
  "csv_code": "P202500000001234567890",
  "validation_status": "Aceptado",
  "registration_state": "Registrado en el SII",
  "timestamp": "2025-01-22T10:35:00Z",
  "details": {
    "invoice_id": "F2025-000123",
    "issue_date": "2025-01-15",
    "total_amount": 1250.00,
    "validation_timestamp": "2025-01-15T16:20:00Z"
  }
}
```

---

### Ejemplo 3: Depurar error de envío SII

**Input del usuario**:
> "El cron job de facturas está fallando con error de certificado inválido"

**Proceso**:

1. Revisar logs del cron job en `task/sii_invoices.go`
2. Identificar error específico (ej: "certificate expired")
3. Consultar `references/certificate_management.md`
4. Verificar fecha de caducidad del certificado actual
5. Ejecutar `scripts/test_soap_connection.sh test` para diagnosticar
6. Si certificado caducado, ejecutar `scripts/convert_certificate.sh` con nuevo certificado
7. Reiniciar cron job y verificar conexión

**Output esperado**:

```
Diagnóstico completado:
- Problema identificado: Certificado digital caducado el 2025-01-01
- Acción tomada: Convertido nuevo certificado desde .p12 a .pem
- Ubicación: services/Invoices/certificates/company_prod.pem
- Validación: Certificado válido hasta 2026-01-01
- Test de conexión: ✓ Exitoso (endpoint test de AEAT)
- Cron job reiniciado: ✓ Enviando facturas correctamente
```

---

### Ejemplo 4: Modificar factura ya enviada al SII

**Input del usuario**:
> "Modificar la factura F2025-000123 porque el IVA estaba mal calculado"

**Proceso**:

1. Recuperar factura existente de base de datos
2. Verificar que permite modificación (periodo no cerrado)
3. Calcular nuevo importe de IVA correcto
4. Generar XML de modificación con tipo de factura "F4" (Modificación)
4. Incluir campos rectificados y campos correctos
5. Enviar XML de modificación a AEAT
6. Obtener nuevo código CSV
7. Actualizar registro en base de datos con estado "Modificado"

**Output esperado**:

```json
{
  "status": "success",
  "message": "Factura modificada y enviada al SII",
  "invoice_id": "F2025-000123",
  "previous_csv": "P202500000001234567890",
  "new_csv": "P202500000001234567891",
  "modification_type": "Modificación de IVA",
  "changes": {
    "old_vat_amount": 230.00,
    "new_vat_amount": 262.50,
    "old_total": 1425.00,
    "new_total": 1457.50
  },
  "timestamp": "2025-01-22T11:00:00Z"
}
```

## Presentación de Resultados

Al completar cualquier operación de facturación SII:

1. **Resumir operación**: "{Tipo de operación} completada para {ID de factura}. Estado: {resultado}"
2. **Formato de output**: JSON estructurado con todos los detalles relevantes
3. **Incluir métricas**:
   - ID de factura y tipo
   - Código CSV (si aplica)
   - Estado del envío
   - Timestamp de operación
   - Detalles de respuesta AEAT
4. **Proteger datos sensibles**: Nunca incluir contenido de certificados o claves privadas

**Ejemplo de resumen completo**:

```json
{
  "operation": "Envío de Factura Emitida al SII",
  "invoice": {
    "id": "F2025-000123",
    "type": "F1",
    "issue_date": "2025-01-15",
    "receiver_nif": "B12345678",
    "total_amount": 1457.50,
    "vat_amount": 262.50
  },
  "sii_submission": {
    "csv_code": "P202500000001234567890",
    "status": "Enviado",
    "timestamp": "2025-01-22T10:30:00Z",
    "aeat_endpoint": "https://www2.agenciatributaria.es/es13/iiim"
  },
  "validation": {
    "xml_valid": true,
    "signature_valid": true,
    "certificate_valid": true,
    "xsd_compliant": true
  },
  "database": {
    "updated": true,
    "table": "issued_invoices",
    "new_status": "sent"
  },
  "audit": {
    "log_id": "LOG-2025-000456",
    "user": "system_cron",
    "duration_ms": 1234
  }
}
```

## Troubleshooting

### Problema: Error "Certificate Invalid" al enviar a AEAT

**Síntoma**: Respuesta de AEAT con error de autenticación del certificado

**Causa**: Certificado caducado, formato incorrecto, o ruta mal configurada

**Solución**:

1. Verificar fecha de caducidad del certificado:

```bash
openssl x509 -in services/Invoices/certificates/company_prod.pem -noout -dates
```

2. Si está caducado, obtener nuevo certificado y convertir:

```bash
bash scripts/convert_certificate.sh nuevo_certificado.pfx password123
```

3. Probar conexión:

```bash
bash scripts/test_soap_connection.sh prod
```

4. Si problema persiste, consultar `references/certificate_management.md`

---

### Problema: XML no pasa validación de schema

**Síntoma**: Error al validar XML contra XSD, o AEAT rechaza el XML

**Causa**: Campo mal formateado, tipo de dato incorrecto, o campo obligatorio faltante

**Solución**:

1. Validar XML manualmente:

```bash
bash scripts/validate_xml.sh factura_emitida.xml F1
```

2. Si hay error de validación, revisar:
   - Formato de fechas (debe ser N8, N14, o N19)
   - Códigos de país (ISO 3166-1 alpha-2)
   - Decimales con punto (no coma) como separador
   - Todos los campos obligatorios presentes

3. Consultar `references/xml_schemas.md` para verificar estructura correcta

---

### Problema: Cron job no envía facturas

**Síntoma**: Facturas marcadas como "pending" no se envían automáticamente

**Causa**: Variable de entorno `ENV` no configurada, o conexión a base de datos fallida

**Solución**:

1. Verificar variable de entorno:

```bash
echo $ENV
# Debe ser "pre" o "pro" para que el cron job se ejecute
```

2. Revisar logs del cron job:

```bash
tail -f /var/log/reverence-api/sii_cron.log
```

3. Verificar conexión a base de datos SII:

```bash
# En código Go, revisar que dbSII no sea nil
# Verificar que DATA_BASE_*_SII están en .env
```

4. Si el problema es de scheduling, verificar `task/sii_invoices.go`

---

### Problema: Código CSV no se recupera de respuesta AEAT

**Síntoma**: Envío parece exitoso pero CSV queda vacío en base de datos

**Causa**: Parser de respuesta XML no encuentra el campo CSV en la respuesta SOAP

**Solución**:

1. Revisar respuesta XML completa de AEAT (ver en logs)

2. Buscar patrón del CSV en la respuesta:

```bash
grep -o "CSV>[0-9]*<" logs/sii_response.xml
```

3. Si CSV existe pero parser no lo encuentra, actualizar regex de extracción en `services/Invoices/sii_communication.go`

4. Consultar `references/error_codes.md` para ver si la respuesta AEAT indica algún problema

---

### Problema: Error de conexión SOAP (500 Internal Server Error)

**Síntoma**: Endpoint AEAT devuelve HTTP 500

**Causa**: Endpoint caído, mantenimiento de AEAT, o request mal formado

**Solución**:

1. Verificar estado del endpoint AEAT:

```bash
bash scripts/test_soap_connection.sh prod
```

2. Si endpoint está caído, esperar y reintentar ( AEAT tiene ventanas de mantenimiento)

3. Si endpoint responde correctamente en test pero falla en producción, revisar:
   - Certificado de producción (diferente al de test)
   - URL correcta de producción
   - Headers SOAP requeridos para producción

4. Consultar estado de servicios AEAT en su web oficial

---

### Problema: Periodo ya cerrado en SII

**Síntoma**: Error al enviar factura: "El periodo de liquidación ya está cerrado"

**Causa**: Intentando enviar/modificar factura de un trimestre que ya fue cerrado en el SII

**Solución**:

1. Verificar que la fecha de la factura no pertenezca a un trimestre cerrado
2. Si es una modificación de factura de periodo cerrado, no se puede hacer
3. Para nuevas facturas de periodos cerrados, consultar normativa AEAT
4. En `references/error_codes.md` buscar "periodo cerrado" para más detalles

## Consideraciones Especiales

### Seguridad

- **Certificados digitales**: Nunca incluir contenido de certificados (.pem) en logs o mensajes de error
- **NIFs y datos fiscales**: Considerar datos sensibles, tratar conforme GDPR
- **Claves privadas**: Almacenar solo en variables de entorno o archivos con permisos restringidos (600)
- **Logs de auditoría**: Mantener registro completo de todas las operaciones SII para posibles inspecciones

### Rendimiento

- **Envío en lote**: El cron job envía facturas en lotes de hasta 100 por llamada para optimizar
- **Timeouts**: Configurar timeouts de 30 segundos para endpoints AEAT (pueden ser lentos)
- **Reintentos**: Implementar lógica de reintentos (3 intentos con backoff exponencial)
- **Validación previa**: Validar XMLs contra XSD antes de enviar para evitar rechazos de AEAT

### Entornos

- **Test (Pruebas)**:
  - URL: `https://www2.agenciatributaria.es/es13/iiim`
  - Certificado: Pruebas (no válido para producción)
  - No tiene validez legal, solo para desarrollo

- **Producción**:
  - URL: `https://www2.agenciatutributaria.es/es13/iiim` (diferente a test)
  - Certificado: Producción real emitido por autoridad certificadora
  - Los envíos tienen validez legal y obligan tributariamente

### Compatibilidad

- **Go version**: 1.21+ (requerido por librerías de criptografía)
- **Dependencias**:
  - `encoding/xml` para generación de XMLs
  - `crypto/tls` para comunicación HTTPS con certificados
  - `crypto/x509` para validación de certificados
- **Base de datos SII**: MySQL 5.7+ (reverence_sii)
- **Schema XSD**: Versión 1.1 de SuministroInmediato (AEAT 2024)

## Integración con el Proyecto

### Localización en la Arquitectura

El módulo de facturación SII se ubica en:

```
services/
└── Invoices/
    ├── issued_invoices.go          # Facturas emitidas
    ├── received_invoices.go        # Facturas recibidas
    ├── intracommunity_invoices.go  # Facturas intracomunitarias
    ├── sii_communication.go        # Comunicación con AEAT
    ├── xml_generator.go            # Generación de XMLs
    ├── xml_signer.go               # Firma de XMLs con certificado
    └── certificates/               # Directorio de certificados .pem
        ├── company_test.pem
        └── company_prod.pem

models/
└── Invoices/
    ├── issued_invoice.go           # Model GORM para facturas emitidas
    ├── received_invoice.go         # Model GORM para facturas recibidas
    └── sii_log.go                  # Logs de envíos SII

task/
└── sii_invoices.go                 # Cron job para envío automático

controllers/
└── Invoices/
    ├── issued_invoices_controller.go
    └── sii_controller.go           # Endpoints para operaciones SII

routes/
└── routes.go                       # Rutas API para facturación
```

### Endpoints API Relacionados

```
POST   /api/v1/issued-invoices          # Crear factura emitida + enviar SII
PUT    /api/v1/issued-invoices/:id      # Modificar factura existente
DELETE /api/v1/issued-invoices/:id      # Anular factura (si periodo abierto)
GET    /api/v1/issued-invoices/:id/csv  # Consultar estado por código CSV

POST   /api/v1/received-invoices        # Registrar factura recibida + enviar SII
GET    /api/v1/sii-validation/:csv      # Validar estado de envío en AEAT
GET    /api/v1/sii-stats                # Estadísticas de envíos SII
```

### Base de Datos SII

Tablas principales en `reverence_sii`:

```sql
issued_invoices           -- Facturas emitidas
received_invoices         -- Facturas recibidas
sii_send_logs             -- Historial de envíos
sii_validation_responses  -- Respuestas de validación CSV
```

### Variables de Entorno Requeridas

```bash
# Base de datos SII
DATA_BASE_USER_SII=user_sii
DATA_BASE_PASS_SII=password_sii
DATA_BASE_NAME_SII=reverence_sii
DATA_BASE_HOST_SII=db-sii.prod.internal

# Endpoints AEAT
SII_URL_TEST_ISSUED=https://www2.agenciatributaria.es/es13/iiim
SII_URL_PROD_ISSUED=https://www2.agenciatutributaria.es/es13/iiim
SII_URL_TEST_RECEIVED=https://www2.agenciatributaria.es/es13/iiim
SII_URL_PROD_RECEIVED=https://www2.agenciatutributaria.es/es13/iiim

# Certificados
SII_CERTIFICATE_PATH=/app/services/Invoices/certificates/company_prod.pem
SII_CERTIFICATE_PASSWORD=password_certificado
```

## Mejoras Futuras (Roadmap)

- [ ] **Batch send optimization**: Implementar envío en paralelo de múltiples facturas usando goroutines
- [ ] **Caching de validaciones**: Guardar en cache respuestas de validación CSV para evitar llamadas repetidas
- [ ] **Sistema de reintentos automático**: Cola de reintentos para envíos fallidos por errores transitorios
- [ ] **Soporte para facturas de asientos**: Implementar generación de XMLs para facturas de asientos (nuevo requerimiento AEAT 2025)
- [ ] **Dashboard de monitoring**: Endpoint con métricas de envíos, errores y latencias
- [ ] **Soporte para SIGNE**: Integración con sistema SIGNE de AEAT para facturas sin NIF
- [ ] **Validación en tiempo real**: Validar XMLs contra XSD antes de guardar en base de datos
- [ ] **Sistema de alertas**: Notificar Slack/email cuando el cron job falla o hay errores masivos

## Referencias Externas

### Documentación Oficial AEAT

- **Suministro Inmediato de Información**: https://www.agenciatributaria.es/AEAT.internet/Inicio/Proyectos_Fiscales/Suministro_Inmediato_Informacion/ 
- **Esquemas XML**: https://www.agenciatributaria.es/static_files/AEAT/Contenido_e_Informacion/Genericos/Modelos_y_Formularios/Suministro_Inmediato/FicherosSuministros/V_1_1/XSD/
- **Manual de uso**: PDF con instrucciones detalladas de implementación
- **Preguntas Frecuentes**: https://www.agenciatributaria.es/AEAT.internet/Inicio/Ayuda/Preguntas_Frecuentes/Proyectos_Fiscales/

### Herramientas de Desarrollo

- **Herramienta de validación XML AEAT**: Validador online de XMLs contra schemas oficiales
- **Herramienta de test SOAP**: Permite enviar peticiones SOAP de prueba a entornos de test
- **Portal de pruebas SII**: Entorno de pruebas gestionado por AEAT

---

**Versión de la skill**: 1.0.0  
**Última actualización**: 2025-01-22  
**Compatible con**: Reverence Hotels API v2.19.8+  
**Módulo principal**: services/Invoices/  