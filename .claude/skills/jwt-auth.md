---
name: jwt-auth
description: Esta skill debe usarse cuando el usuario necesite implementar, configurar, o diagnosticar autenticación JWT en aplicaciones Go con Echo framework. Se activa con peticiones como "implementa autenticación JWT", "configura middleware JWT", "genera tokens JWT", o "diagnostica error de autenticación 401".
license: Complete terms in LICENSE.txt
version: 1.0.0
author: reverence-hotels-team
category: language
tags: [jwt, authentication, go, echo-framework, security, middleware]---

# JWT Auth Skill

Skill especializada en autenticación JWT (JSON Web Tokens) para el proyecto Reverence Hotels API. Proporciona conocimiento técnico sobre implementación de JWT con Echo v4, middleware de autenticación, generación/validación de tokens, refresh tokens, y configuración de claims personalizados para autorización basada en niveles.

## Cuándo Usar Esta Skill

Esta skill debe usarse cuando:
- Se necesite implementar autenticación JWT en nuevos endpoints
- Se deba crear o validar middleware de autenticación
- Haya problemas con tokens expirados o inválidos
- Se requiera configurar claims personalizados (Level, UserID, Role)
- Necesite implementarse refresh token flow
- Se deba diagnosticar errores 401/403 en endpoints protegidos
- Se requiera verificar la configuración de JWT_SECRET y expiración

Triggers comunes:
- "Implementa autenticación JWT para este endpoint"
- "El token JWT no funciona, devuelve 401"
- "Genera un JWT token con claims personalizados"
- "Configura middleware JWT en Echo"
- "El refresh token expira muy rápido"
- "Añade autorización por nivel a este endpoint"

## Workflow Principal

### 1. Análisis Inicial

Antes de proceder:
1. Identificar el tipo de operación (generación/validación/refresh/diagnóstico)
2. Verificar que las variables de entorno JWT están configuradas
3. Determinar si se requiere autenticación básica o con claims personalizados
4. Revisar el middleware existente en `middleware/`

### 2. Ejecución

Según el tipo de operación:

#### Generación de Token
1. Revisar `services/Auth/` para implementación existente
2. Configurar claims estándar (sub, iat, exp) + custom (Level, UserID)
3. Usar `jwt-go` o `golang-jwt/jwt/v5` para firmar con JWT_SECRET
4. Establecer expiración apropiada (access: 15min, refresh: 7days)

#### Validación Middleware
1. Revisar `middleware/auth.go` o `middleware/jwt.go`
2. Extraer token del header `Authorization: Bearer <token>`
3. Validar firma y expiración
4. Extraer claims al contexto de Echo
5. Verificar autorización por nivel si es necesario

#### Diagnóstico de Errores
1. Verificar logs del middleware JWT
2. Validar que JWT_SECRET coincide entre generación y validación
3. Revisar expiración del token (claim `exp`)
4. Verificar formato del Authorization header
5. Consultar `references/jwt_errors.md` para errores comunes

### 3. Validación

Verificar que:
- [ ] El token se genera con firma válida
- [ ] El middleware extrae claims correctamente
- [ ] La autorización por nivel funciona (si aplica)
- [ ] El refresh token flow opera adecuadamente
- [ ] Los errores devuelven códigos HTTP correctos (401/403)

### 4. Output

Presentar resultados:
- Formato: Código implementado + explicación de cambios
- Incluir: Configuración de middleware, ejemplos de uso, pruebas
- Omitir: Logs excesivos de debugging

## Recursos de la Skill

### Scripts (`scripts/`)

#### `scripts/generate_jwt_token.go`

**Propósito**: Script standalone para generar tokens JWT de prueba

**Uso**:

```bash
go run scripts/generate_jwt_token.go --user-id 123 --level 5 --secret "your-secret"
```

**Parámetros**:
- `--user-id`: ID del usuario (claim sub)
- `--level`: Nivel de autorización (claim Level)
- `--secret`: Secreto JWT (default: leer de JWT_SECRET env var)
- `--duration`: Duración del token (default: 15m)

**Output**: Token JWT generado en stdout

**Ejemplo**:

```bash
go run scripts/generate_jwt_token.go --user-id 42 --level 3
# Output: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI0Mi...
```

---

#### `scripts/validate_jwt_token.go`

**Propósito**: Validar y decodificar tokens JWT para debugging

**Uso**:

```bash
go run scripts/validate_jwt_token.go --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Parámetros**:
- `--token`: Token JWT a validar (obligatorio)
- `--secret`: Secreto JWT (default: leer de JWT_SECRET)

**Output**: Claims decodificados + estado de validación

**Ejemplo**:

```bash
go run scripts/validate_jwt_token.go --token "eyJhbGci..."
# Output:
# ✓ Token válido
# Claims:
#   UserID: 42
#   Level: 3
#   Exp: 2025-01-21 10:30:00
```

---

### Referencias (`references/`)

#### `references/jwt_middleware_patterns.md`

**Contenido**: Patrones de implementación de middleware JWT en Echo v4

**Cuándo consultar**: Al implementar nuevo middleware o refactorizar autenticación existente

**Estructura**:

- Echo middleware basics
- JWT authentication pattern
- Authorization by level pattern
- Context data extraction
- Error handling best practices

**Búsqueda rápida**:

```bash
# Buscar patrón de autorización
grep -A 15 "Level-Based Authorization" references/jwt_middleware_patterns.md

# Buscar manejo de errores
grep -A 10 "Error Handling" references/jwt_middleware_patterns.md
```

---

#### `references/jwt_claims_specification.md`

**Contenido**: Especificación completa de claims usados en Reverence Hotels API

**Cuándo consultar**: Al añadir nuevos claims o modificar estructura de tokens

**Estructura**:

- Standard claims (sub, iat, exp, nbf)
- Custom claims (Level, UserID, Role, ProfileID)
- Claims para refresh tokens
- Validación de claims en middleware

**Búsqueda rápida**:

```bash
# Buscar claim específico
grep -i "Level" references/jwt_claims_specification.md

# Buscar refresh token claims
grep -A 10 "Refresh Token" references/jwt_claims_specification.md
```

---

#### `references/jwt_errors.md`

**Contenido**: Catálogo de errores comunes en JWT y sus soluciones

**Cuándo consultar**: Al diagnosticar errores 401/403 o tokens inválidos

**Búsqueda rápida**:

```bash
# Buscar error específico
grep -i "signature is invalid" references/jwt_errors.md

# Buscar por código de error
grep -A 5 "401" references/jwt_errors.md
```

---

#### `references/security_best_practices.md`

**Contenido**: Prácticas de seguridad para implementación de JWT

**Cuándo consultar**: Al auditar seguridad de autenticación o configurar JWT_SECRET

**Búsqueda rápida**:

```bash
# Buscar recomendaciones de secret
grep -A 10 "JWT_SECRET" references/security_best_practices.md

# Buscar expiración de tokens
grep -A 5 "Token Expiration" references/security_best_practices.md
```

---

### Assets (`assets/`)

#### `assets/templates/jwt_middleware.go`

**Tipo**: Template de código para middleware JWT en Echo

**Uso**: Copiar como base para nuevos middleware de autenticación

**Modificaciones**: Personalizar validación de claims según necesidades del endpoint

**Contenido principal**:
- Extracción de token del header
- Validación de firma y expiración
- Inyección de claims en contexto Echo
- Manejo de errores 401/403

---

#### `assets/templates/claims_struct.go`

**Tipo**: Template de estructura de claims personalizados

**Uso**: Base para definir claims estándar del proyecto

**Modificaciones**: Añadir campos custom según requerimientos del negocio

**Contenido principal**:
```go
type Claims struct {
    UserID    uint   `json:"sub"`
    Level     int    `json:"level"`
    Role      string `json:"role,omitempty"`
    ProfileID uint   `json:"profile_id,omitempty"`
    jwt.RegisteredClaims
}
```

---

## Ejemplos de Uso

### Ejemplo 1: Implementar autenticación JWT en nuevo endpoint

**Input del usuario**:
> "Añade autenticación JWT al endpoint POST /api/v1/employees"

**Proceso**:

1. Revisar `routes/routes.go` para ubicar la ruta del endpoint
2. Añadir middleware JWT a la ruta en `routes/echo.go`
3. Extraer claims del contexto en el controller
4. Validar nivel de autorización si es necesario
5. Testear con token generado usando `scripts/generate_jwt_token.go`

**Output esperado**:

```go
// En routes/echo.go
employeesGroup := e.Group("/api/v1/employees")
employeesGroup.Use(middleware.JWTAuth)
employeesGroup.POST("", controllers.Employee.Create)

// En controller
func Create(c echo.Context) error {
    claims := c.Get("claims").(*jwt.Claims)
    if claims.Level < 3 {
        return c.JSON(403, map[string]string{"error": "Insufficient level"})
    }
    // ... lógica del endpoint
}
```

---

### Ejemplo 2: Diagnosticar error 401 en endpoint protegido

**Input del usuario**:
> "El endpoint /api/v1/profiles devuelve 401 Unauthorized con token válido"

**Proceso**:

1. Validar el token usando `scripts/validate_jwt_token.go`
2. Revisar logs del middleware en `middleware/auth.go`
3. Verificar que JWT_SECRET coincide en generación y validación
4. Chequear formato del Authorization header (debe ser "Bearer <token>")
5. Consultar `references/jwt_errors.md` para error específico

**Output esperado**:

```
Diagnóstico: Token inválido por secreto incorrecto

Problema detectado:
- Token firmado con JWT_SECRET: "secret-A"
- Middleware validando con JWT_SECRET: "secret-B"
- Los secretos no coinciden

Solución:
1. Verificar variable de entorno JWT_SECRET
2. Asegurar que sea la misma en todos los servicios
3. Reiniciar aplicación tras corregir

Comando de validación:
go run scripts/validate_jwt_token.go --token "<token>" --secret "correct-secret"
```

---

### Ejemplo 3: Implementar refresh token flow

**Input del usuario**:
> "Implementa refresh tokens para renovar access tokens sin re-login"

**Proceso**:

1. Revisar `services/Auth/` para ver implementación actual
2. Crear endpoint POST /api/v1/auth/refresh
3. Validar refresh token (expiración más larga: 7 días)
4. Generar nuevo access token al validar refresh
5. Implementar rotación de refresh tokens (opcional, recomendado)
6. Consultar `references/jwt_middleware_patterns.md` sección "Refresh Flow"

**Output esperado**:

```go
// services/Auth/refresh_service.go
func RefreshAccessToken(refreshToken string) (string, error) {
    token, err := jwt.ParseWithClaims(refreshToken, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if err != nil || !token.Valid {
        return "", errors.New("invalid refresh token")
    }
    
    claims := token.Claims.(*RefreshClaims)
    
    // Generar nuevo access token
    accessClaims := AccessClaims{
        UserID: claims.UserID,
        Level: claims.Level,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    return generateToken(accessClaims)
}
```

---

## Presentación de Resultados

Al completar tareas de JWT:

1. **Resumir cambios**: "Implementada {feature} en {location}. Configuración: {summary}"
2. **Formato de output**: Código Go + explicación de integración
3. **Incluir detalles**:
   - Variables de entorno requeridas
   - Configuración de middleware
   - Ejemplos de requests con tokens
4. **Adjuntar tests**: Comandos para probar la implementación

**Ejemplo de resumen**:

```
Implementada autenticación JWT en endpoint /api/v1/employees

Cambios:
1. Añadido middleware JWT en routes/echo.go
2. Extraído claims en controller Employee.Create
3. Validado nivel mínimo (Level >= 3)

Configuración requerida:
- JWT_SECRET: (ya configurada)
- Expiración access token: 15 minutos
- Expiración refresh token: 7 días

Prueba:
curl -H "Authorization: Bearer <token>" \
  -X POST http://localhost:1331/api/v1/employees

Estado: ✓ Implementación completada
```

---

## Troubleshooting

### Problema: Token expira demasiado rápido

**Síntoma**: Error "token is expired" después de pocos minutos

**Causa**: Configuración de expiración muy corta en access token

**Solución**:

```go
// Ajustar expiración en services/Auth/token_generator.go
RegisteredClaims: jwt.RegisteredClaims{
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // Era 15m
    IssuedAt:  jwt.NewNumericDate(time.Now()),
}
```

Consultar `references/security_best_practices.md` para balance entre seguridad y UX.

---

### Problema: Error "signature is invalid"

**Síntoma**: Middleware rechaza tokens válidos con error de firma

**Causa**: JWT_SECRET diferente entre generación y validación

**Solución**:

1. Verificar variable de entorno JWT_SECRET
2. Asegurar que sea consistente en todos los entornos
3. No rotar el secreto sin invalidar tokens existentes

```bash
# Verificar secreto actual
echo $JWT_SECRET

# Validar token con secreto correcto
go run scripts/validate_jwt_token.go --token "<token>" --secret "$JWT_SECRET"
```

Consultar `references/jwt_errors.md` sección "Signature Errors".

---

### Problema: Claims no disponibles en contexto

**Síntoma**: `c.Get("claims")` devuelve nil en controller

**Causa**: Middleware no inyecta claims en contexto Echo

**Solución**:

```go
// En middleware/auth.go
func JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // ... validar token ...
        
        claims := token.Claims.(*Claims)
        
        // CRÍTICO: Inyectar claims en contexto
        c.Set("claims", claims)
        
        return next(c)
    }
}
```

Consultar `references/jwt_middleware_patterns.md` sección "Context Injection".

---

### Problema: Autorización por nivel no funciona

**Síntoma**: Usuario con Level insuficiente puede acceder a endpoint protegido

**Causa**: Validación de nivel no implementada en controller o middleware

**Solución**:

```go
// Opción 1: Validar en controller
func GetProfile(c echo.Context) error {
    claims := c.Get("claims").(*Claims)
    if claims.Level < requiredLevel {
        return c.JSON(403, map[string]string{
            "error": "Insufficient authorization level",
        })
    }
    // ... lógica del endpoint ...
}

// Opción 2: Middleware de autorización
func RequireLevel(level int) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            claims := c.Get("claims").(*Claims)
            if claims.Level < level {
                return c.JSON(403, map[string]string{
                    "error": fmt.Sprintf("Level %d required", level),
                })
            }
            return next(c)
        }
    }
}

// Uso:
profileGroup.Use(RequireLevel(3))
```

Consultar `references/jwt_middleware_patterns.md` sección "Level-Based Authorization".

---

### Problema: Refresh token reutilizable (security risk)

**Síntoma**: Mismo refresh token puede usarse múltiples veces

**Causa**: No hay rotación de refresh tokens o blacklist

**Solución**:

Implementar rotación de refresh tokens:

```go
type RefreshTokenUse struct {
    TokenID     string
    UserID      uint
    Used        bool
    ExpiresAt   time.Time
}

// Al usar refresh token:
1. Marcar token actual como usado
2. Generar nuevo refresh token
3. Guardar en BD
4. Si un refresh token ya usado se intenta reutilizar → revocar todos

// Consultar references/security_best_practices.md sección "Refresh Token Rotation"
```

---

### Problema: Permisos de ejecución en scripts

**Síntoma**: Error "permission denied" al ejecutar scripts en `scripts/`

**Solución**:

```bash
chmod +x scripts/*.go
chmod +x scripts/*.sh
```

---

## Consideraciones Especiales

### Rendimiento

- Validación JWT es rápida (~1-2ms por request)
- No cachear tokens (son de un solo uso)
- Evitar consultas a BD en middleware JWT (usar claims del token)
- Para endpoints muy frecuentes, considerar usar短期 access tokens (5min)

### Seguridad

- **CRÍTICO**: JWT_SECRET debe ser mínimo 32 caracteres, aleatorio, y rotarse periódicamente
- Nunca incluir información sensible en claims (contraseñas, datos de pago)
- Usar HTTPS siempre (tokens en headers pueden ser interceptados)
- Implementar blacklist para tokens revocados (logout forzado)
- Refresh tokens deben almacenarse en BD con expiración

**Consultar `references/security_best_practices.md` antes de producción.**

### Compatibilidad

- Requisitos: Go 1.18+, Echo v4, golang-jwt/jwt/v5
- Tokens son compatibles backward si claims no cambian
- Migración desde JWT v4 a v5 requiere actualizar imports

---

## Especificación Técnica del Proyecto

### Arquitectura de Autenticación en Reverence Hotels API

#### Middleware JWT Existente

Ubicación: `middleware/auth.go` o `middleware/jwt.go`

```go
const (
    JWTSecretKey = "JWT_SECRET"
    ContextKey   = "claims"
)

type Claims struct {
    UserID uint   `json:"sub"`
    Level  int    `json:"level"`
    jwt.RegisteredClaims
}

func JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Extracción del header
        authHeader := c.Request().Header.Get("Authorization")
        if authHeader == "" {
            return c.JSON(401, map[string]string{"error": "Missing authorization header"})
        }
        
        // Formato: "Bearer <token>"
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Validación
        token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv(JWTSecretKey)), nil
        })
        
        if err != nil || !token.Valid {
            return c.JSON(401, map[string]string{"error": "Invalid token"})
        }
        
        // Extracción de claims
        claims := token.Claims.(*Claims)
        
        // Inyección en contexto
        c.Set(ContextKey, claims)
        
        return next(c)
    }
}
```

#### Generación de Tokens

Ubicación: `services/Auth/token_service.go`

```go
func GenerateAccessToken(user *models.User) (string, error) {
    secret := os.Getenv("JWT_SECRET")
    
    claims := &Claims{
        UserID: user.ID,
        Level:  user.Level,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "reverence-hotels-api",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(user *models.User) (string, error) {
    // Similar pero con expiración de 7 días
    claims := &RefreshClaims{
        UserID: user.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

#### Sistema de Autorización por Niveles

El proyecto usa un sistema granular de autorización:

```
User → Level → LevelPrivileges → Form → PathAPI
                         ↓
                   Read/Write privileges
```

**Integración con JWT**:
- El claim `Level` en el JWT indica el nivel del usuario
- El middleware de autorización verifica si `Level` tiene privilegios para el `PathAPI` solicitado
- GET requests requieren privilegio `Read`
- POST/PUT/DELETE requieren privilegio `Write`

**Ejemplo de middleware de autorización**:

```go
func AuthorizeByPath(path string, requiredPermission string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            claims := c.Get("claims").(*Claims)
            
            // Consultar LevelPrivileges en BD
            hasPrivilege := db.Where("level_id = ? AND permission = ?", claims.Level, requiredPermission).
                             First(&LevelPrivilege{})
            
            if !hasPrivilege {
                return c.JSON(403, map[string]string{"error": "Forbidden"})
            }
            
            return next(c)
        }
    }
}
```

---

## Mejoras Futuras (Roadmap)

- [ ] Implementar blacklist de tokens en Redis para logout forzado
- [ ] Añadir soporte para múltiples dispositivos (tokens concurrentes)
- [ ] Implementar fingerprinting de tokens para detectar robos
- [ ] Rotación automática de JWT_SECRET con período de gracia
- [ ] Métricas de uso de tokens para detectar patrones sospechosos
- [ ] Soporte para scopes OAuth2 además de levels
- [ ] Implementar token binding a IP/User-Agent