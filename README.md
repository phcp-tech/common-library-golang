[![Go Reference](https://pkg.go.dev/badge/github.com/phcp-tech/common-library-golang.svg)](https://pkg.go.dev/github.com/phcp-tech/common-library-golang)


# phcp-library-golang

A shared Go library for the PHCP ecosystem, providing core infrastructure components for microservice development.

## Requirements

- Go 1.25+

## Installation

```bash
go get github.com/phcp-tech/common-library-golang
```

## Build and Test

```bash
go mod tidy
go build ./...
golangci-lint run ./...
go test ./...
```

### Run tests with coverage

```bash
# All packages
go test ./... -cover -timeout 60s

# Single package with per-function breakdown
go test ./env/... -cover -coverprofile=coverage.out -timeout 60s
go tool cover -func=coverage.out
```

### Subprocess coverage (Go 1.20+)

Some tests spawn a subprocess to exercise `os.Exit` paths. Pass `GOCOVERDIR`
to merge subprocess coverage into the parent report:

```bash
mkdir -p /tmp/cov
GOCOVERDIR=/tmp/cov go test ./env/... -cover -timeout 60s
go tool covdata func -i=/tmp/cov
```

## Module Overview

| Module | Import Path | Description |
|--------|-------------|-------------|
| env | `.../common-library-golang/env` | Configuration management |
| log | `.../common-library-golang/log` | Structured logging |
| ringuf | `.../common-library-golang/ringuf` | Ring buf |

---

## Configuration Management (env)

Built on [Koanf](https://github.com/knadh/koanf), supports TOML config files and environment variables. Priority: environment variables > config file.

```go
import "github.com/phcp-tech/common-library-golang/env"

cfg := env.GetConfig()
value := cfg.String("app.name")
```

---

## Logging (log)

Structured JSON logging based on `slog`, with file rotation and runtime log level adjustment.

```go
import "github.com/phcp-tech/common-library-golang/log"

logger := log.GetLogger()
logger.Info("server started", "port", 8080)
logger.Error("query failed", "err", err)
```

**Configuration** (via environment variables or config file):

| Key | Description | Default |
|-----|-------------|---------|
| `log.level` | Log level (debug/info/warn/error) | `info` |
| `log.writefile` | Write logs to file | `false` |
| `app.log.path` | Log file path | - |

