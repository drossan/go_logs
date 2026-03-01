# Migracion de v2 a v3

Esta guia te ayuda a migrar de go_logs v2 (funciones globales legacy) a v3 (interfaz Logger moderna). La buena noticia es que el codigo v2 continua funcionando - puedes migrar gradualmente o no hacerlo.

**[English](Migration-v2-to-v3.md)** | **Espanol**

## Diferencias Clave

| Caracteristica | v2 | v3 |
|----------------|-----|-----|
| Estilo de API | Funciones globales | Interfaz Logger |
| Campos Estructurados | No | Si (campos tipados) |
| Child Loggers | No | Si (`With()`) |
| Soporte de Contexto | No | Si (`LogCtx()`) |
| Multiples Instancias | No | Si |
| Hooks | No | Si |
| Inyeccion de Dependencias | No | Si (basada en interfaces) |
| Configuracion | Solo entorno | Opciones + Entorno |
| Inicializacion | `Init()` requerido | Auto-inicializa |

## Compatibilidad hacia Atras

**El codigo v2 funciona sin cambios en v3:**

```go
// Codigo v2 - todavia funciona en v3!
import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()
    go_logs.InfoLog("Hola desde v2")
    go_logs.ErrorLog("Algo salio mal")
}
```

Puedes usar APIs v2 y v3 juntas:

```go
import "github.com/drossan/go_logs"

func main() {
    // API v2
    go_logs.InfoLog("Usando API v2")

    // API v3
    logger, _ := go_logs.New()
    logger.Info("Usando API v3", go_logs.String("version", "3.0"))
}
```

## Estrategias de Migracion

### Estrategia 1: Sin Migracion

Continua usando la API v2. Esta completamente soportada y mantenida.

```go
// Sin cambios necesarios
go_logs.Init()
go_logs.InfoLog("Aplicacion iniciada")
```

**Pros:**
- Sin cambios de codigo
- Riesgo cero
- Funciona inmediatamente

**Contras:**
- Sin logging estructurado
- Sin child loggers
- Sin propagacion de contexto
- Sin hooks

### Estrategia 2: Migracion Gradual

Migra incrementalmente, empezando con codigo nuevo:

```go
// Codigo antiguo - mantener como esta
func funcionLegacy() {
    go_logs.InfoLog("Funcion legacy")
}

// Codigo nuevo - usar v3
func nuevaFuncion() {
    logger, _ := go_logs.New()
    logger.Info("Nueva funcion",
        go_logs.String("feature", "estructurado"),
    )
}
```

### Estrategia 3: Migracion Completa

Migra completamente a la API v3 para todas las nuevas features.

## Tabla de Equivalencia API

### Inicializacion

| v2 | v3 |
|----|-----|
| `go_logs.Init()` | `logger, _ := go_logs.New()` |
| `go_logs.Close()` | `logger.Sync()` |

### Funciones de Logging

| v2 | v3 |
|----|-----|
| `go_logs.TraceLog(msg)` | `logger.Trace(msg)` |
| `go_logs.DebugLog(msg)` | `logger.Debug(msg)` |
| `go_logs.InfoLog(msg)` | `logger.Info(msg)` |
| `go_logs.SuccessLog(msg)` | `logger.Log(go_logs.SuccessLevel, msg)` |
| `go_logs.WarningLog(msg)` | `logger.Warn(msg)` |
| `go_logs.ErrorLog(msg)` | `logger.Error(msg)` |
| `go_logs.FatalLog(msg)` | `logger.Fatal(msg)` |

### Logging Formateado

| v2 | v3 |
|----|-----|
| `go_logs.Infof(format, args...)` | `logger.Info(fmt.Sprintf(format, args...))` |
| `go_logs.Errorf(format, args...)` | `logger.Error(fmt.Sprintf(format, args...))` |

### Configuracion

| v2 (Entorno) | v3 (Programatico) |
|--------------|-------------------|
| `LOG_LEVEL=debug` | `go_logs.WithLevel(go_logs.DebugLevel)` |
| `LOG_FORMAT=json` | `go_logs.WithFormatter(go_logs.NewJSONFormatter())` |
| `SAVE_LOG_FILE=1` | `go_logs.WithRotatingFile(...)` |

## Ejemplos de Migracion

### Logger Basico

**v2:**
```go
package main

import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()
    go_logs.InfoLog("Aplicacion iniciada")
    go_logs.ErrorLog("Algo salio mal")
}
```

**v3:**
```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithOutput(os.Stdout),
    )
    logger.Info("Aplicacion iniciada")
    logger.Error("Algo salio mal")
}
```

### Con Campos Estructurados

**v2 (concatenacion):**
```go
go_logs.Infof("Usuario %s conectado desde %s", username, ip)
```

**v3 (estructurado):**
```go
logger.Info("Usuario conectado",
    go_logs.String("username", username),
    go_logs.String("ip", ip),
)
```

### Con Salida a Archivo

**v2:**
```go
// Establecer variables de entorno
os.Setenv("SAVE_LOG_FILE", "1")
os.Setenv("LOG_FILE_NAME", "app.log")
os.Setenv("LOG_FILE_PATH", "/var/log")

go_logs.Init()
go_logs.InfoLog("Aplicacion iniciada")
```

**v3:**
```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
)
defer logger.Sync()

logger.Info("Aplicacion iniciada")
```

### Con Manejo de Errores

**v2:**
```go
err := hacerAlgo()
if err != nil {
    go_logs.ErrorLog("Operacion fallo: " + err.Error())
}
```

**v3:**
```go
err := hacerAlgo()
if err != nil {
    logger.Error("Operacion fallo",
        go_logs.Err(err),
        go_logs.String("operation", "hacerAlgo"),
    )
}
```

### Con Contexto

**v2 (sin soporte de contexto):**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Sin forma de propagar trace ID
    go_logs.InfoLog("Peticion recibida")
}
```

**v3:**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := go_logs.WithTraceID(r.Context(), r.Header.Get("X-Trace-ID"))

    logger.LogCtx(ctx, go_logs.InfoLevel, "Peticion recibida")
}
```

### Con Child Loggers

**v2 (repeticion):**
```go
func procesarPeticion(requestID string) {
    go_logs.InfoLog("Peticion " + requestID + " iniciada")
    go_logs.InfoLog("Peticion " + requestID + " procesando")
    go_logs.InfoLog("Peticion " + requestID + " completada")
}
```

**v3:**
```go
func procesarPeticion(requestID string) {
    reqLogger := logger.With(go_logs.String("request_id", requestID))

    reqLogger.Info("Peticion iniciada")
    reqLogger.Info("Procesando")
    reqLogger.Info("Peticion completada")
}
```

## Checklist de Migracion

- [ ] Identificar todas las llamadas de logging v2
- [ ] Crear instancia de logger v3
- [ ] Reemplazar llamadas simples de logging
- [ ] Agregar campos estructurados donde sea apropiado
- [ ] Crear child loggers para contexto repetido
- [ ] Agregar propagacion de contexto para handlers HTTP
- [ ] Actualizar tests para usar v3
- [ ] Remover llamadas `go_logs.Init()` (ya no necesarias para v3)
- [ ] Actualizar documentacion

## Ver Tambien

- [Referencia API](API-Reference-es.md) - API v3 completa
- [Inicio Rapido](Getting-Started-es.md) - Guia de inicio rapido
- [Ejemplos](Examples-es.md) - Mas ejemplos v3
