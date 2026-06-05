# PHCP:common-library-golang

[![Go Reference](https://pkg.go.dev/badge/github.com/phcp-tech/common-library-golang.svg)](https://pkg.go.dev/github.com/phcp-tech/common-library-golang)
[![LICENSE](https://img.shields.io/badge/license-Apache--2.0-green.svg)](https://github.com/phcp-tech/common-library-golang/blob/main/LICENSE)


Common-library-golang is a collection of functional components used within the PHCP ecosystem, providing numerous components for microservice development. To make it more widely available, it is now open-sourced under the Apache license.

## Requirements

- Go 1.25+

## Installation

```bash
go get github.com/phcp-tech/common-library-golang
```

## Build and Test

```bash
go mod tidy
go vet ./...
go build ./...
go test ./...
```

### Run tests with coverage

```bash
# All packages
go test ./... -cover -timeout 60s

```

## Packages

| Package | Import path | Description |
|---------|-------------|-------------|
| `env` | `.../common-library-golang/env` | TOML config + environment variable loader |
| `log` | `.../common-library-golang/log` | Structured JSON logger with file rotation and ringbuf |
| `ringbuf` | `.../common-library-golang/ringbuf` | Lock-free ring buffers (SPSC and MPSC) |

---

## env — Configuration Management

Loads a TOML config file and merges OS environment variables on top.
Built on [koanf](https://github.com/knadh/koanf). Implements the singleton pattern:
only the first `InitEnv` call takes effect.

```go
import "github.com/phcp-tech/common-library-golang/env"

// Call once at startup before any Env() usage.
if err := env.InitEnv("config.toml"); err != nil {
    log.Fatal(err)
}

host := env.Env().String("server.host")
port := env.Env().Int("server.port")
```

For single-binary deployments the config file can be embedded at compile time:

```go
//go:embed config.toml
var configFS embed.FS

env.InitEnv("config.toml", &configFS)
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/env#pkg-examples).

---

## log — Structured Logging

Structured JSON logging via `log/slog` with UTC timestamps.
File writes are fully **asynchronous**: the caller returns immediately after pushing the formatted entry into an internal `RingMPSC` buffer; a dedicated consumer goroutine performs the actual I/O. This makes the package suitable for **high-throughput, latency-sensitive scenarios** where blocking on disk I/O is unacceptable.

**`InitLog` must be called once at application startup before any log function.**
Omit the argument for stdout output at INFO level; pass a `Config` to customise.

```go
import "github.com/phcp-tech/common-library-golang/log"

// Stdout mode at default INFO level.
log.InitLog()
log.Info("application started")
log.Infof("listening on port %d", 8080)
log.InfoWith("request", "method", "GET", "path", "/api/v1", "status", 200)

// Stdout mode with custom level.
log.InitLog(&log.Config{Level: "debug"})

// File mode: set FilePath to enable rotating file logging.
log.InitLog(&log.Config{
    Level:      "info",
    FilePath:   "/var/log/app.log",
    MaxSizeMB:  100,
    MaxBackups: 7,
    MaxAgeDays: 30,
    Compress:   true,
})
defer log.CloseLogFile() // flush async buffer and close file on shutdown
```

Available log functions: `Debug` / `Info` / `Warn` / `Error` and their `f` (format) and `With` (structured key-value) variants. Log level can be changed at runtime with `SetLevel`.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/log#pkg-examples).

---

## ringbuf — Ring Buffers

High-performance fixed-capacity ring buffers for producer-consumer pipelines.

| Type | Use case | Thread safety |
|------|----------|---------------|
| `RingSPSC` | Single producer, single consumer | Lock-free (atomic only) |
| `RingMPSC` | Multiple producers, single consumer | Mutex on producer side |

Both types support an optional `ProcessFunc` that starts a consumer goroutine
automatically, or manual `Pop` / `TryPop` for caller-managed consumption.
`Push` blocks when the buffer is full (backpressure); `TryPush` returns false
immediately.

```go
import "github.com/phcp-tech/common-library-golang/ringbuf"

// SPSC — single producer, automatic consumer goroutine
rb := ringbuf.NewRingSPSC(ringbuf.RingSPSCConfig[string]{
    Capacity:    1024,
    ProcessFunc: func(s string) { fmt.Println(s) },
})
rb.Push("hello")
rb.Close() // drain and wait for consumer to finish

// MPSC — multiple producers, automatic consumer goroutine
rb := ringbuf.NewRingMPSC(ringbuf.RingMPSCConfig[[]byte]{
    Capacity:    4096,
    ProcessFunc: func(b []byte) { os.Stdout.Write(b) },
})
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/ringbuf#pkg-examples).
