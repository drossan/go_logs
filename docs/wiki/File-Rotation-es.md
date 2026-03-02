# Rotacion de Archivos

go_logs proporciona un `RotatingFileWriter` integrado para rotacion automatica de archivos de log por tamano. Esta implementacion no tiene dependencias externas y usa solo la biblioteca estandar de Go.

**[English](File-Rotation.md)** | **Espanol**

## Descripcion General

`RotatingFileWriter` rota automaticamente los archivos de log cuando alcanzan un tamano especificado, creando archivos de backup con sufijos incrementales.

### Comportamiento de Rotacion

Cuando ocurre una rotacion:
1. El archivo actual se cierra
2. Los backups existentes se renombran (`.3` -> `.4`, `.2` -> `.3`, `.1` -> `.2`)
3. El archivo actual se renombra a `.1`
4. Se crea un nuevo archivo vacio
5. Los backups antiguos mas alla de `maxBackups` se eliminan

### Ejemplo de Disposicion de Archivos

Con `maxBackups=3`:
```
app.log        # Archivo de log actual
app.log.1      # Backup mas reciente
app.log.2      # Segundo mas reciente
app.log.3      # Backup mas antiguo (tercera rotacion)
```

## Crear un RotatingFileWriter

### NewRotatingFileWriter

Crea un writer de archivo con rotacion basada en tamano.

```go
func NewRotatingFileWriter(filename string, maxSizeMB int, maxBackups int) (*RotatingFileWriter, error)
```

**Parametros:**
- `filename` - Ruta al archivo de log (ej: `/var/log/app.log`)
- `maxSizeMB` - Tamano maximo en megabytes antes de rotar
- `maxBackups` - Numero maximo de archivos de backup a mantener

**Ejemplo:**

```go
writer, err := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
if err != nil {
    panic(err)
}
defer writer.Close()

logger, _ := go_logs.New(
    go_logs.WithOutput(writer),
)
```

### Opcion WithRotatingFile

Opcion de conveniencia para crear un logger con rotacion de archivo:

```go
func WithRotatingFile(filename string, maxSizeMB int, maxBackups int) Option
```

**Ejemplo:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
defer logger.Sync()
```

## Configuracion

### Configuracion Basica

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile(
        "/var/log/miapp/app.log", // Ruta del archivo de log
        100,                       // Max 100 MB por archivo
        5,                         // Mantener 5 archivos de backup
    ),
)
```

### Con Formateador JSON

Para produccion, combinar con JSONFormatter:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithLevel(go_logs.InfoLevel),
)
```

### Con Multi-Salida

Loguear tanto a archivo como a consola:

```go
fileWriter, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)

logger, _ := go_logs.New(
    go_logs.WithMultiOutput(fileWriter, os.Stdout),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

## Variables de Entorno

Configurar rotacion via variables de entorno:

| Variable | Descripcion | Por Defecto |
|----------|-------------|-------------|
| `SAVE_LOG_FILE` | Habilitar logging a archivo | `0` |
| `LOG_FILE_NAME` | Nombre del archivo de log | `log.txt` |
| `LOG_FILE_PATH` | Directorio del archivo de log | directorio actual |
| `LOG_MAX_SIZE` | Tamano max en MB | `100` |
| `LOG_MAX_BACKUPS` | Max archivos de backup | `5` |

**Ejemplo:**

```bash
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_FILE_PATH=/var/log/miapp
export LOG_MAX_SIZE=100
export LOG_MAX_BACKUPS=5
```

```go
// El logger recoge las variables de entorno
logger, _ := go_logs.New()
```

## Mejores Practicas

### 1. Siempre Llamar Sync Antes de Salir

Asegurar que los datos en buffer se escriban:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("app.log", 100, 5),
)
defer logger.Sync() // Importante!
```

### 2. Usar Rutas Absolutas

Usa rutas absolutas para evitar confusion:

```go
// Bien
go_logs.WithRotatingFile("/var/log/miapp/app.log", 100, 5)

// Arriesgado - depende del directorio de trabajo
go_logs.WithRotatingFile("logs/app.log", 100, 5)
```

### 3. Crear Directorio de Logs

El writer crea el directorio automaticamente, pero asegurate de que la ruta padre exista:

```go
// Crear directorio de logs
os.MkdirAll("/var/log/miapp", 0755)

logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/miapp/app.log", 100, 5),
)
```

### 4. Usar JSON para Produccion

Combinar rotacion con formateador JSON:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### 5. Recomendaciones de Tamano

| Volumen de Logs | Tamano Max | Max Backups | Uso de Disco |
|-----------------|------------|-------------|--------------|
| Bajo (< 1 GB/dia) | 100 MB | 5 | 600 MB |
| Medio (1-10 GB/dia) | 100 MB | 20 | 2.1 GB |
| Alto (> 10 GB/dia) | 500 MB | 30 | 15.5 GB |

## Solucion de Problemas

### Permiso Denegado

```
Error: open /var/log/app.log: permission denied
```

**Solucion**: Asegurate de que la aplicacion tiene permisos de escritura en el directorio:

```bash
sudo mkdir -p /var/log/miapp
sudo chown $USER:$USER /var/log/miapp
```

### El Directorio No Existe

```
Error: open /var/log/miapp/app.log: no such file or directory
```

**Solucion**: El writer crea el directorio automaticamente, pero asegurate de que la ruta padre exista:

```go
os.MkdirAll("/var/log/miapp", 0755)
```

### La Rotacion No Ocurre

**Causas**:
- `maxSizeMB` es demasiado grande para tu volumen de logs
- Los logs no se estan escribiendo con suficiente frecuencia

**Solucion**: Reduce `maxSizeMB` para pruebas:

```go
writer, _ := go_logs.NewRotatingFileWriter("app.log", 1, 5) // 1 MB para pruebas
```

## Ver Tambien

- [Configuracion](Configuration-es.md) - Variables de entorno
- [Formateadores](Formatters-es.md) - Formateador JSON para produccion
- [Ejemplos](Examples-es.md) - Mas ejemplos de rotacion de archivos
