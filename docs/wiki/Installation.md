# Installation

This guide covers installing go_logs and its dependencies.

## Requirements

- **Go 1.21+** - Required for modern Go features
- **Go Modules** - The library uses Go modules for dependency management

## Installation Methods

### Using go get (Recommended)

Install the latest version:

```bash
go get github.com/drossan/go_logs@latest
```

Install a specific version:

```bash
go get github.com/drossan/go_logs@v3.0.0
```

### Using go.mod

Add to your `go.mod` file:

```
require github.com/drossan/go_logs v3.0.0
```

Then run:

```bash
go mod download
```

## Dependencies

go_logs has minimal dependencies:

### Core Dependencies (Required)

- **fatih/color** - ANSI color support for TextFormatter
  - Automatically installed with `go get`

### Optional Dependencies

- **slack-go/slack** - Required only for Slack notifications
  - Install with: `go get github.com/slack-go/slack`
  - Only needed if using `SlackHook` or v2 Slack notifications

## Verifying Installation

Create a simple test file to verify the installation:

```go
// test_install.go
package main

import (
    "fmt"
    "github.com/drossan/go_logs"
)

func main() {
    logger, err := go_logs.New()
    if err != nil {
        panic(err)
    }

    logger.Info("go_logs installed successfully!")
    fmt.Println("Installation verified!")
}
```

Run the test:

```bash
go run test_install.go
```

Expected output:

```
[2026/02/28 10:30:00] INFO go_logs installed successfully!
Installation verified!
```

## Version Information

Check the installed version:

```go
package main

import (
    "fmt"
    "runtime/debug"
)

func main() {
    bi, ok := debug.ReadBuildInfo()
    if !ok {
        fmt.Println("Could not read build info")
        return
    }

    for _, dep := range bi.Deps {
        if dep.Path == "github.com/drossan/go_logs" {
            fmt.Printf("go_logs version: %s\n", dep.Version)
        }
    }
}
```

## Version History

| Version | Description |
|---------|-------------|
| v3.x | Modern API with Logger interface, structured logging, hooks |
| v2.x | Legacy API with global functions (still maintained) |

Both versions are maintained and 100% backward compatible.

## Upgrading

### From v2 to v3

v2 code continues to work without changes:

```go
// v2 code - still works in v3
import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()
    go_logs.InfoLog("Hello")  // Still works!
}
```

To use v3 features alongside v2:

```go
import "github.com/drossan/go_logs"

func main() {
    // v2 API
    go_logs.InfoLog("Using v2 API")

    // v3 API
    logger, _ := go_logs.New()
    logger.Info("Using v3 API", go_logs.String("version", "3.0"))
}
```

See [Migration v2 to v3](Migration-v2-to-v3.md) for a complete migration guide.

## Troubleshooting

### Error: cannot find package

```
cannot find package "github.com/drossan/go_logs" in any of:
    /usr/local/go/src/github.com/drossan/go_logs (from $GOROOT)
```

**Solution**: Ensure Go modules are enabled:

```bash
export GO111MODULE=on
go mod init your-project
go get github.com/drossan/go_logs@latest
```

### Error: version conflict

```
go: github.com/drossan/go_logs@v3.0.0: invalid version: unknown revision
```

**Solution**: Update your go.mod and run `go mod tidy`:

```bash
go mod tidy
go get github.com/drossan/go_logs@latest
```

### Error: fatih/color not found

```
cannot find package "github.com/fatih/color"
```

**Solution**: Run `go mod tidy` to install dependencies:

```bash
go mod tidy
```

### Vendor Directory Issues

If using vendoring, update the vendor directory:

```bash
go mod vendor
```

## Project Structure After Installation

After installation, your project structure should include:

```
your-project/
+-- go.mod                  # Contains go_logs dependency
+-- go.sum                  # Dependency checksums
+-- vendor/                 # (if using vendoring)
|   +-- github.com/
|       +-- drossan/
|       |   +-- go_logs/    # Main library
|       +-- fatih/
|           +-- color/      # Color dependency
+-- main.go
```

## Next Steps

After installation:

1. [Getting Started](Getting-Started.md) - Create your first logger
2. [API Reference](API-Reference.md) - Explore the full API
3. [Examples](Examples.md) - See practical examples

## Build Tags

go_logs does not require any build tags. Standard build commands work:

```bash
# Build
go build ./...

# Test
go test ./...

# Build with optimizations
go build -ldflags="-s -w" ./...
```

## Docker Integration

When using Docker, ensure dependencies are available:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
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

## CI/CD Integration

For CI/CD pipelines, cache Go modules:

```yaml
# GitHub Actions example
steps:
  - uses: actions/checkout@v4

  - name: Set up Go
    uses: actions/setup-go@v5
    with:
      go-version: '1.21'
      cache: true

  - name: Download dependencies
    run: go mod download

  - name: Run tests
    run: go test -v ./...
```

## Getting Help

If you encounter issues:

1. Check [GitHub Issues](https://github.com/drossan/go_logs/issues)
2. Review the [API Reference](API-Reference.md)
3. Check the [Examples](Examples.md) for usage patterns
