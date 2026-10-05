# Instalacion

Esta guia cubre la instalacion de go_logs y sus dependencias.

**[English](/installation)** | **Espanol**

## Requisitos

- **Go 1.27+** - Requerido para caracteristicas modernas de Go
- **Go Modules** - La biblioteca usa Go modules para gestion de dependencias

## Metodos de Instalacion

### Usando go get (Recomendado)

Instalar la ultima version:

```bash
go get github.com/drossan/go_logs/v3@latest
```

Instalar una version especifica:

```bash
go get github.com/drossan/go_logs/v3@latest
```

### Usando go.mod

Agregar a tu archivo `go.mod`:

```
require github.com/drossan/go_logs/v3 v3.1.0
```

Luego ejecutar:

```bash
go mod download
```

## Dependencias

go_logs tiene dependencias minimas:

### Dependencias Core (Requeridas)

- **fatih/color** - Soporte de colores ANSI para TextFormatter
  - Se instala automaticamente con `go get`

### Dependencias Opcionales

- **github.com/drossan/go_logs/slack/v3** - Notificador de Slack, en un módulo aparte (trae `slack-go/slack`)
  - Instalar con: `go get github.com/drossan/go_logs/slack/v3@latest`
  - Solo necesario para Slack: como notificador de `hooks.SlackHook` o, en la API v2, registrado con `go_logs.SetNotifier`
  - El módulo core `github.com/drossan/go_logs/v3` no depende de `slack-go/slack`

## Verificar Instalacion

Crea un archivo de prueba simple para verificar la instalacion:

```go
// test_install.go
package main

import (
    "fmt"
    "github.com/drossan/go_logs/v3"
)

func main() {
    logger, err := go_logs.New()
    if err != nil {
        panic(err)
    }

    logger.Info("go_logs instalado correctamente!")
    fmt.Println("Instalacion verificada!")
}
```

Ejecutar la prueba:

```bash
go run test_install.go
```

Salida esperada:

```
[2026/02/28 10:30:00] INFO go_logs instalado correctamente!
Instalacion verificada!
```

## Informacion de Version

Verificar la version instalada:

```go
package main

import (
    "fmt"
    "runtime/debug"
)

func main() {
    bi, ok := debug.ReadBuildInfo()
    if !ok {
        fmt.Println("No se pudo leer informacion de build")
        return
    }

    for _, dep := range bi.Deps {
        if dep.Path == "github.com/drossan/go_logs/v3" {
            fmt.Printf("Version de go_logs: %s\n", dep.Version)
        }
    }
}
```

## Historial de Versiones

| Version | Descripcion |
|---------|-------------|
| v3.x | API moderna con interfaz Logger, logging estructurado, hooks |
| v2.x | API legacy con funciones globales (todavia mantenida) |

Ambas versiones se mantienen y son 100% compatibles hacia atras.

## Actualizar

### De v2 a v3

El codigo v2 continua funcionando sin cambios:

```go
// Codigo v2 - todavia funciona en v3
import "github.com/drossan/go_logs/v3"

func main() {
    go_logs.Init()
    go_logs.InfoLog("Hola")  // Todavia funciona!
}
```

Para usar caracteristicas v3 junto con v2:

```go
import "github.com/drossan/go_logs/v3"

func main() {
    // API v2
    go_logs.InfoLog("Usando API v2")

    // API v3
    logger, _ := go_logs.New()
    logger.Info("Usando API v3", go_logs.String("version", "3.0"))
}
```

Ver [Migracion v2 a v3](/es/migration-v2-to-v3) para una guia completa de migracion.

## Solucion de Problemas

### Error: cannot find package

```
cannot find package "github.com/drossan/go_logs/v3" in any of:
    /usr/local/go/src/github.com/drossan/go_logs/v3 (from $GOROOT)
```

**Solucion**: Asegurate de que Go modules estan habilitados:

```bash
export GO111MODULE=on
go mod init tu-proyecto
go get github.com/drossan/go_logs/v3@latest
```

### Error: conflicto de version

```
go: github.com/drossan/go_logs/v3@v3.0.0: invalid version: unknown revision
```

**Solucion**: Actualiza tu go.mod y ejecuta `go mod tidy`:

```bash
go mod tidy
go get github.com/drossan/go_logs/v3@latest
```

### Error: fatih/color no encontrado

```
cannot find package "github.com/fatih/color"
```

**Solucion**: Ejecuta `go mod tidy` para instalar dependencias:

```bash
go mod tidy
```

### Problemas con Directorio Vendor

Si usas vendoring, actualiza el directorio vendor:

```bash
go mod vendor
```

## Estructura del Proyecto Despues de Instalar

Despues de la instalacion, tu estructura de proyecto deberia incluir:

```
tu-proyecto/
+-- go.mod                  # Contiene dependencia go_logs
+-- go.sum                  # Checksums de dependencias
+-- vendor/                 # (si usas vendoring)
|   +-- github.com/
|       +-- drossan/
|       |   +-- go_logs/    # Biblioteca principal
|       +-- fatih/
|           +-- color/      # Dependencia de color
+-- main.go
```

## Siguientes Pasos

Despues de la instalacion:

1. [Inicio Rapido](/es/getting-started) - Crea tu primer logger
2. [Referencia API](/es/api-reference) - Explora la API completa
3. [Ejemplos](/es/examples) - Ve ejemplos practicos

## Build Tags

go_logs no requiere ningun build tag. Los comandos de build estandar funcionan:

```bash
# Build
go build ./...

# Test
go test ./...

# Build con optimizaciones
go build -ldflags="-s -w" ./...
```

## Integracion con Docker

Cuando uses Docker, asegurate de que las dependencias esten disponibles:

```dockerfile
# Build stage
FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o myapp .

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/myapp .
CMD ["./myapp"]
```

## Integracion CI/CD

Para pipelines CI/CD, cachea los modulos Go:

```yaml
# Ejemplo GitHub Actions
steps:
  - uses: actions/checkout@v4

  - name: Set up Go
    uses: actions/setup-go@v5
    with:
      go-version: '1.27'
      cache: true

  - name: Download dependencies
    run: go mod download

  - name: Run tests
    run: go test -v ./...
```

## Obtener Ayuda

Si encuentras problemas:

1. Revisa [GitHub Issues](https://github.com/drossan/go_logs/issues)
2. Consulta la [Referencia API](/es/api-reference)
3. Revisa los [Ejemplos](/es/examples) para patrones de uso
