# PHCP: common-library-golang

[![Go Reference](https://pkg.go.dev/badge/github.com/phcp-tech/common-library-golang.svg)](https://pkg.go.dev/github.com/phcp-tech/common-library-golang)
[![LICENSE](https://img.shields.io/badge/license-Apache--2.0-green.svg)](https://github.com/phcp-tech/common-library-golang/blob/main/LICENSE)
![CI](https://github.com/phcp-tech/common-library-golang/actions/workflows/deploy-build-test.yml/badge.svg)
[![codecov](https://codecov.io/gh/phcp-tech/common-library-golang/branch/main/graph/badge.svg)](https://codecov.io/gh/phcp-tech/common-library-golang)

Common-library-golang is a collection of functional components used within the PHCP ecosystem, providing numerous components for microservice development, it provides some out-of-the-box functions, such as environment, log, database, etc. To make it more widely available, it is now open-sourced under the Apache license.

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
```

### Short mode test — fast CI verification

Some tests start a real TCP listener (e.g. `httpserver` integration tests). Pass
`-short` to skip them and keep the test run lightweight:

```bash
go test ./... -short
```

### Full test run — including integration tests

```bash
go test ./...
```

### Run test with coverage

```bash
# All packages
go test ./... -cover -timeout 60s

```

## Packages

| Group | Package | Import path | Description |
|-------|---------|-------------|-------------|
| Basic | [`env`](#env--configuration-management) | `.../common-library-golang/env` | TOML config + environment variable loader |
| Basic | [`log`](#log--structured-logging) | `.../common-library-golang/log` | Structured JSON logger with file rotation and ringbuf |
| Basic | [`shutdown`](#shutdown--graceful-shutdown) | `.../common-library-golang/shutdown` | Block until OS signal or programmatic trigger, then continue for cleanup |
| Basic | [`ringbuf`](#ringbuf--ring-buffers) | `.../common-library-golang/ringbuf` | Lock-free ring buffers (SPSC and MPSC) |
| Basic | [`maps`](#maps--thread-safe-concurrent-maps) | `.../common-library-golang/maps` | Thread-safe generic concurrent maps with pluggable replacement strategies |
| Basic | [`cgroup`](#cgroup--linux-resource-limits) | `.../common-library-golang/cgroup` | Read CPU and memory resource limits from cgroup v2 (Linux only; returns 0 on other platforms) |
| Basic | [`metrics`](#metrics--runtime-metrics-snapshot) | `.../common-library-golang/metrics` | Runtime and system metrics snapshot (CPU, memory, goroutines, cgroup limits, uptime) |
| Database | [`redis`](#redis--redis-client) | `.../common-library-golang/redis` | Redis client (standalone and cluster) with connection pool and key scan utilities |
| Database | [`dbsqlc/postgres`](#dbsqlcpostgres--postgresql-connection-pool) | `.../common-library-golang/dbsqlc/postgres` | PostgreSQL connection pool via pgx/v5 |
| Database | [`dbsqlc/sqlite`](#dbsqlcsqlite--sqlite-connection) | `.../common-library-golang/dbsqlc/sqlite` | SQLite connection via pure-Go modernc driver |
| Database | [`dbsqlc/sqlite/vfs`](#dbsqlcsqlitevfs--embedded-sqlite-via-vfs) | `.../common-library-golang/dbsqlc/sqlite/vfs` | SQLite over embedded FS via VFS (binary-embedded databases) |
| Network | [`auth`](#auth--casbin-rbac-authorisation) | `.../common-library-golang/auth` | Casbin RBAC authorisation middleware for Gin |
| Network | [`token`](#token--jwt-authentication) | `.../common-library-golang/token` | JWT access/refresh token creation, parsing, and Gin middleware |
| Network | [`gin`](#gin--gin-engine-factory) | `.../common-library-golang/gin` | Gin engine factory with slog request logging and CORS |
| Network | [`gin/pprof`](#ginpprof--profiling-endpoints) | `.../common-library-golang/gin/pprof` | Optional pprof profiling endpoints for Gin (explicit opt-in) |
| Network | [`httpclient`](#httpclient--resty-http-client) | `.../common-library-golang/httpclient` | Resty-based HTTP client with retry, JWT auth, and JSON helpers |
| Network | [`httpclient/retryable`](#httpclientretryable--retryable-http-client) | `.../common-library-golang/httpclient/retryable` | hashicorp/go-retryablehttp wrapper returning standard *http.Client |
| Network | [`httpserver`](#httpserver--production-httphttps-server) | `.../common-library-golang/httpserver` | Production HTTP/HTTPS server with timeouts and graceful shutdown |
| Network | [`httpserver/lambda`](#httpserverlambda--aws-lambda-adapter) | `.../common-library-golang/httpserver/lambda` | AWS Lambda adapter implementing httpserver.Runner (explicit opt-in) |

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
defer log.Close() // flush async buffer and close file on shutdown
```

Available log functions: `Debug` / `Info` / `Warn` / `Error` and their `f` (format) and `With` (structured key-value) variants. Log level can be changed at runtime with `SetLevel`.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/log#pkg-examples).

---

## shutdown — Graceful Shutdown

Two primitives for application shutdown coordination:

- **`Wait`** blocks the calling goroutine until an OS signal (`SIGINT`, `SIGTERM`, `SIGHUP`, `SIGQUIT`) or `Trigger` is called. After it returns, the caller performs cleanup and exits.
- **`Trigger`** unblocks `Wait` programmatically from any goroutine — useful for a `/shutdown` HTTP endpoint or a metrics failure handler. Safe to call multiple times.

```go
import "github.com/phcp-tech/common-library-golang/shutdown"

// In main: start services, then block until shutdown.
shutdown.Wait()

// Cleanup runs here (or via defer before Wait).
runner.Shutdown(ctx)
```

```go
// From a /shutdown HTTP endpoint or metrics failure handler.
shutdown.Trigger()
```

> **Note:** `Trigger` has no effect when the process is forcibly terminated by the OS
> (e.g. IDE stop button on Windows calls `TerminateProcess` — no signal is delivered).

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/shutdown#pkg-examples).

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

### Performance

Benchmarks run with `go test -bench=. -benchtime=3s -benchmem`.
All operations produce **zero heap allocations**.

**Test environment**

| | |
|---|---|
| CPU | 11th Gen Intel® Core™ i7-11850H @ 2.50 GHz (8 cores / 16 threads) |
| RAM | 32 GB |
| OS | Windows 11 Enterprise |
| Go | 1.26.2 windows/amd64 |

**Results**

| Benchmark | ns/op | Throughput | Allocs |
|-----------|------:|----------:|-------:|
| `SPSC Push` (1P-1C, blocking) | 76.65 | ~13.0 M ops/s | 0 |
| `SPSC TryPush` (1P-1C, non-blocking) | 87.91 | ~11.4 M ops/s | 0 |
| `SPSC ProducerConsumer` (end-to-end with ProcessFunc) | 83.72 | ~11.9 M ops/s | 0 |
| `SPSC Push string` (realistic log payload) | 91.06 | ~11.0 M ops/s | 0 |
| `MPSC Push` (1 producer) | 151.1 | ~6.6 M ops/s | 0 |
| `MPSC Push` (4 producers concurrent) | 192.2 | ~5.2 M ops/s | 0 |
| `MPSC Push` (8 producers concurrent) | 200.7 | ~5.0 M ops/s | 0 |
| `MPSC ProducerConsumer` (4P end-to-end) | 197.5 | ~5.1 M ops/s | 0 |
| `Go channel` (buffered 4096, reference) | 126.3 | ~7.9 M ops/s | 0 |

**Key observations**

- SPSC is **~1.6× faster** than a buffered Go channel for the same single-producer / single-consumer pattern.
- MPSC trades some throughput for multi-producer safety via a mutex; with 8 concurrent producers it remains above **5 M ops/s**.
- Throughput scales gracefully under producer contention: going from 1 to 8 producers adds only ~33% latency per item.
- Both types maintain **zero allocations per operation**, making them suitable as the async I/O backbone for the `log` package.

---

## maps — Thread-Safe Concurrent Maps

Thread-safe generic concurrent maps backed by
[orcaman/concurrent-map](https://github.com/orcaman/concurrent-map) with
pluggable replacement strategies.

Two implementations are provided:

| Type | Key | Value | Strategy |
|------|-----|-------|----------|
| `CMap` | `string` (fixed) | `int64` (fixed) | Greater-value wins (built-in) |
| `CMapGen[K, V]` | any `comparable` | any | Configurable via `SetDefaultCompare` or `SetDefaultStrategy` |

Built-in strategy types: `NumericGreaterStrategy`, `AlwaysReplaceStrategy`, `TimestampStrategy`.

```go
import "github.com/phcp-tech/common-library-golang/maps"

// Generic map — configure a default compare function.
m := maps.NewCMapGen[string, int64]()
m.SetDefaultCompare(func(old, new int64) bool { return new > old })

m.Set("EURUSD", 10500)
m.Replace("EURUSD", 10600)       // stored: 10600 > 10500
m.ReplaceAlways("EURUSD", 9000)  // always stored
m.ReplaceIfNotExists("GBPUSD", 12800) // stored only if key absent

// Atomic read-modify-write.
m.UpsertWithCallback("USDJPY", 15100, func(exists bool, old, new int64) int64 {
    if !exists || new > old { return new }
    return old
})

// Fixed string→int64 map (no configuration needed).
cm := maps.NewCMap()
cm.Set("tick", 10000)
cm.Replace("tick", 10050) // stored: 10050 > 10000
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/maps#pkg-examples).

### Performance

Benchmarks run with `go test -bench=. -benchtime=3s -benchmem`.

**Test environment** — same machine as ringbuf benchmarks:
Intel® Core™ i7-11850H @ 2.50 GHz · 8 cores / 16 threads · 32 GB RAM · Go 1.26.2 / Windows 11

**CMap** (fixed `string→int64`, built-in greater-value strategy)

| Benchmark | ns/op | Throughput | B/op | Allocs |
|-----------|------:|----------:|-----:|-------:|
| `Set` | 33.01 | ~30.3 M ops/s | 0 | 0 |
| `Get` | 15.68 | ~63.8 M ops/s | 0 | 0 |
| `Replace` | 43.42 | ~23.0 M ops/s | 0 | 0 |
| `Set` (parallel, 16 goroutines) | 90.67 | ~11.0 M ops/s | 0 | 0 |
| `Replace` (parallel, 16 goroutines) | 434.8 | ~2.3 M ops/s | 0 | 0 |

**CMapGen** (generic `[K comparable, V any]`)

| Benchmark | ns/op | Throughput | B/op | Allocs | Note |
|-----------|------:|----------:|-----:|-------:|------|
| `Set` | 94.20 | ~10.6 M ops/s | 0 | 0 | |
| `Get` | 60.54 | ~16.5 M ops/s | 0 | 0 | |
| `Replace` (default strategy) | 137.0 | ~7.3 M ops/s | 32 | 3 | ⚠ fmt.Sprint fallback allocates |
| `ReplaceWithCompare` | 52.94 | ~18.9 M ops/s | 0 | 0 | recommended |
| `ReplaceAlways` | 29.78 | ~33.6 M ops/s | 0 | 0 | |
| `UpsertWithCallback` | 58.68 | ~17.0 M ops/s | 0 | 0 | |
| `Set` (parallel, 16 goroutines) | 74.32 | ~13.5 M ops/s | 0 | 0 | |
| `Replace` (parallel, 16 goroutines) | 81.61 | ~12.3 M ops/s | 32 | 3 | ⚠ see above |
| Mixed read/write (parallel) | 26.08 | ~38.3 M ops/s | 10 | 1 | |

**`sync.Map`** (stdlib reference)

| Benchmark | ns/op | Throughput | B/op | Allocs |
|-----------|------:|----------:|-----:|-------:|
| `Store` | 60.96 | ~16.4 M ops/s | 56 | 1 |
| `Load` | 10.66 | ~93.8 M ops/s | 0 | 0 |
| `Store` (parallel, 16 goroutines) | 103.7 | ~9.6 M ops/s | 56 | 1 |

**Key observations**

- `CMap.Set` is **~1.8× faster** than `sync.Map.Store` and allocates nothing.
- `CMapGen.Replace` with a typed compare function (`ReplaceWithCompare` / `SetDefaultCompare`) is **zero-allocation**; the no-argument `Replace()` fallback uses `fmt.Sprint` internally and produces 3 allocations — always call `SetDefaultCompare` or `SetDefaultStrategy` at construction time.
- `CMapGen.ReplaceAlways` is the fastest write path (~33.6 M ops/s) when no conditional logic is needed.
- Under 16-goroutine parallel write contention `CMap.Replace` degrades to ~2.3 M ops/s due to single-key hot-spot; spreading writes across many keys restores throughput.

---

## cgroup — Linux Resource Limits

Reads CPU and memory resource limits from the **cgroup v2** unified hierarchy
(`/sys/fs/cgroup`). Intended for containerised workloads (Kubernetes, Docker)
where the process runs inside a cgroup with configured CPU and memory constraints.

On **non-Linux platforms** (macOS, Windows) all functions return `(0, nil)` — no
cgroup filesystem is available and no error is raised.

| Function | cgroup v2 file | Returns |
|----------|---------------|---------|
| `CPULimitMilli()` | `cpu.max` | CPU limit in millicores; 0 = unlimited |
| `CPURequestMilli()` | `cpu.weight` | CPU request in millicores via weight→shares→mCPU |
| `MemoryLimitBytes()` | `memory.max` | Memory limit in bytes; 0 = unlimited |
| `MemoryRequestBytes()` | `memory.low` | Memory soft limit in bytes; 0 = not set |

```go
import "github.com/phcp-tech/common-library-golang/cgroup"

// Read CPU limit and cap GOMAXPROCS accordingly.
if milli, err := cgroup.CPULimitMilli(); err == nil && milli > 0 {
    cpus := milli / 1000
    if cpus < 1 {
        cpus = 1
    }
    runtime.GOMAXPROCS(cpus)
}

// Size an in-process cache as 25 % of the container memory limit.
if limitBytes, err := cgroup.MemoryLimitBytes(); err == nil && limitBytes > 0 {
    cacheBytes := limitBytes / 4
    cache.SetMaxSize(cacheBytes)
}
```
See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/cgroup#pkg-examples).

---

## metrics — Runtime Metrics Snapshot

Collects a one-shot snapshot of key runtime and system metrics as a flat
`[]NameValue` slice, suitable for a `/metrics` or `/health` HTTP endpoint.

Each call samples **CPU usage over one second** via `gopsutil`, so it is
intended for periodic polling (e.g. every 10–30 s), not hot-path use.

| Metric name | Description |
|-------------|-------------|
| `cpuPercent` | Process CPU usage (%) across all cores, sampled over 1 s |
| `memorySize` | Resident set size (RSS) in MiB |
| `threads` | OS thread count |
| `goroutines` | Live goroutine count |
| `gomaxprocs` | `runtime.GOMAXPROCS(0)` |
| `numCPU` | `runtime.NumCPU()` |
| `cpuRequest` | cgroup v2 CPU request in millicores (0 if not set or non-Linux) |
| `cpuLimit` | cgroup v2 CPU limit in millicores (0 if unlimited or non-Linux) |
| `memoryRequest` | cgroup v2 memory soft limit in bytes (0 if not set or non-Linux) |
| `memoryLimit` | cgroup v2 memory limit in bytes (0 if unlimited or non-Linux) |
| `age` | Process uptime formatted as `Xd Xh Xm Xs` |

```go
import "github.com/phcp-tech/common-library-golang/metrics"

// Periodic metrics poll — call from a background goroutine.
snapshot := metrics.GetMetrics()

// Look up a specific entry.
for _, nv := range snapshot {
    if nv.Name == "goroutines" {
        slog.Info("runtime", "goroutines", nv.Value)
    }
}

// Serialize to JSON for a /health endpoint.
b, _ := json.Marshal(snapshot)
w.Write(b)
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/metrics#pkg-examples).

---

## redis — Redis Client

Thread-safe Redis client supporting both **standalone** and **cluster** modes,
backed by [go-redis/v9](https://github.com/redis/go-redis).
Connection pool settings (`PoolSize`, `MinIdleConns`) are configurable via `Config`
with sensible defaults (100 / 5). The caller reads values from `env.Env()` at the
composition root — this package has no dependency on env.

```go
import "github.com/phcp-tech/common-library-golang/redis"

// Standalone mode (single address).
cli := redis.NewRedisClient(&redis.Config{
    Clusters: []string{env.Env().String("redis.addr")},
    Password: env.Env().String("redis.password"),
    DB:       env.Env().Int("redis.db"),
})
defer cli.Close()

// Cluster mode — more than one address enables cluster automatically.
cli := redis.NewRedisClient(&redis.Config{
    Clusters:     env.Env().Strings("redis.clusters"),
    PoolSize:     200,
    MinIdleConns: 10,
})

// Basic operations.
cli.Set(ctx, "key", "value", 5*time.Minute)
val, err := cli.Get(ctx, "key")
cli.Del(ctx, "key")
cli.Unlink(ctx, "key") // async deletion — preferred for large keys

// Remove all keys starting with a prefix.
cli.CleanCache(ctx, "user:42:")

// Count matching keys across all nodes.
n, err := cli.GetKeysCount(ctx, "session:*")
```

Implements the singleton pattern via `InitDefault` / `Default`.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/redis#pkg-examples).

---

## dbsqlc/postgres — PostgreSQL Connection Pool

pgx/v5 connection pool for use with [sqlc](https://sqlc.dev/) (`sql_package: "pgx/v5"`).
Pool creation is **lazy**: `NewPostgres` returns immediately without establishing any
connections, so no live server is required at startup.
Implements the singleton pattern via `InitDefault` / `Default`.

```go
import "github.com/phcp-tech/common-library-golang/dbsqlc/postgres"

// Singleton mode: call once at startup.
err := postgres.InitDefault(&postgres.Config{
    Host:            "localhost",
    Port:            "5432",
    Database:        "mydb",
    Username:        "user",
    Password:        "pass",
    MaxOpenConns:    100,
    MaxIdleConns:    25,
    ConnMaxLifetime: 60, // minutes
    ConnMaxIdletime: 10, // minutes
    SearchPath:      "myschema", // optional
})
if err != nil {
    log.Fatal(err)
}

pool := postgres.Default() // *pgxpool.Pool, pass to sqlc Queries
```

For cases that require multiple pools, use `NewPostgres` directly instead of the singleton.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/postgres#pkg-examples).

---

## dbsqlc/sqlite — SQLite Connection

Pure-Go SQLite driver ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite), no CGO required).
For use with [sqlc](https://sqlc.dev/) (`sql_package: "database/sql"`).
Implements the singleton pattern via `InitDefault` / `Default`.

```go
import "github.com/phcp-tech/common-library-golang/dbsqlc/sqlite"

// In-memory database (tests and short-lived operations).
db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})

// File-based database with WAL mode and foreign key enforcement.
err := sqlite.InitDefault(&sqlite.Config{
    Path: "file:app.db?_journal_mode=WAL&_foreign_keys=on",
})
if err != nil {
    log.Fatal(err)
}

db := sqlite.Default() // *sql.DB, pass to sqlc Queries
```

SQLite allows only one writer at a time. `NewSQLite` calls `SetMaxOpenConns(1)` automatically
to prevent `database is locked` errors when WAL mode is not in use.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/sqlite#pkg-examples).

---

## dbsqlc/sqlite/vfs — Embedded SQLite via VFS

Opens a SQLite database that is **embedded inside the Go binary** using
[modernc.org/sqlite/vfs](https://pkg.go.dev/modernc.org/sqlite/vfs) and Go's
`embed.FS`. The database file must reside at `config/sqlite.db` inside the embedded
filesystem.

**Import this sub-package only when the database is distributed as part of the binary.**
For regular file-based databases use `dbsqlc/sqlite` instead.

```go
import sqlitevfs "github.com/phcp-tech/common-library-golang/dbsqlc/sqlite/vfs"

//go:embed config/sqlite.db
var sqliteFS embed.FS

// Singleton mode: call once at startup.
if err := sqlitevfs.InitDefault(&sqliteFS); err != nil {
    log.Fatal(err)
}

db := sqlitevfs.Default() // *sql.DB, pass to sqlc-generated Queries
```

For cases that require multiple VFS connections, use `New` directly instead of the singleton.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/sqlite/vfs#pkg-examples).

---

## auth — Casbin RBAC Authorisation

Role-based access control (RBAC) Gin middleware backed by
[Casbin](https://casbin.org/). Enforces policy using **method-2 semantics**:
every role in the user's role list must pass the policy check — if any role
is denied, the request is rejected with HTTP 403 Forbidden.

`InitCasbin` must be called once at application startup before `Authorize` is
registered. Model and policy can be loaded from in-memory strings (`fs=true`)
or from files on disk (`fs=false`).

```go
import "github.com/phcp-tech/common-library-golang/auth"

// Load model and policy from in-memory strings (e.g. embedded at build time).
if err := auth.InitCasbin(true, modelString, policyString); err != nil {
    log.Fatal(err)
}

// Load model and policy from files.
if err := auth.InitCasbin(false, "model.conf", "policy.csv"); err != nil {
    log.Fatal(err)
}

// Register as a Gin middleware after token.Authenticate.
// It reads roles from the "userInfo" context key set by token.Authenticate.
r := gin.New()
r.Use(token.Authenticate())
r.Use(auth.Authorize())
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/auth#pkg-examples).

---

## token — JWT Authentication

HS256 JWT access and refresh token creation, parsing, and a Gin bearer-token middleware.
Built on [golang-jwt/jwt](https://github.com/golang-jwt/jwt).

**`InitToken` must be called once at application startup** before any token function.
The secrets and issuer are typically read from `env.Env()` after `env.InitEnv()`.

```go
import "github.com/phcp-tech/common-library-golang/token"

// Initialise once at startup (composition root), after env.InitEnv().
token.InitToken(
    env.Env().String("jwt.issuer"),             // e.g. "phcp"
    env.Env().String("jwt.access.secretcode"),  // access token signing key
    env.Env().String("jwt.refresh.secretcode"), // refresh token signing key
)

// Create a short-lived access token (valid for 1 hour).
tok, err := token.CreateToken(userId, username, productId, roles, time.Hour)

// Validate and parse an access token from an incoming request.
user, err := token.ParseToken(tok)
fmt.Println(user.Username, user.UserId, user.Roles)

// Create a long-lived refresh token (no roles embedded).
refresh, err := token.CreateRefreshToken(userId, username, productId, 24*time.Hour)

// Register as a Gin middleware; stores LoginUser in context under key "userInfo".
r := gin.New()
r.Use(token.Authenticate())
```

`Authenticate` aborts with HTTP 401 if the `Authorization: Bearer <token>` header is
missing, malformed, or carries an invalid/expired token.

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/token#pkg-examples).

---

## gin — Gin Engine Factory

Pre-configured [Gin](https://github.com/gin-gonic/gin) engine with structured slog-based
request logging (`slog-gin`) and optional CORS support. Always runs in `ReleaseMode`.

`InitGin` is the single entry point. Pass a slice of allowed origins to enable CORS;
entries containing `*` are treated as single-level wildcard patterns
(e.g. `https://*.example.com` matches `https://api.example.com` but not `https://a.b.example.com`).

```go
import libgin "github.com/phcp-tech/common-library-golang/gin"

// No CORS.
router := libgin.InitGin(nil)

// Exact origins and wildcard patterns can be mixed.
router := libgin.InitGin([]string{
    "https://app.example.com",   // exact
    "https://*.example.com",     // wildcard: root + any single subdomain
})
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/gin#pkg-examples).

---

## gin/pprof — Profiling Endpoints

Registers Go runtime pprof endpoints on a Gin engine via
[gin-contrib/pprof](https://github.com/gin-contrib/pprof).

**Import this sub-package only in services that require profiling.**
Importing it is an explicit opt-in — it does not affect services that only import `gin`.

Two route groups are mounted:

| Path | Purpose |
|------|---------|
| `/debug/pprof/*` | Standard Go pprof path for direct access |
| `<path>/admin/pprof/*` | API-gateway-friendly alias |

```go
import (
    libgin   "github.com/phcp-tech/common-library-golang/gin"
    ginpprof "github.com/phcp-tech/common-library-golang/gin/pprof"
)

router := libgin.InitGin(nil)
ginpprof.Mount(router, "/api/v1") // mounts /debug/pprof/* and /api/v1/admin/pprof/*
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/gin/pprof#pkg-examples).

---

## httpclient — Resty HTTP Client

Feature-rich HTTP client built on [go-resty/resty](https://github.com/go-resty/resty)
with pre-configured retry, JWT bearer-token support, and automatic JSON handling.
Suitable for structured service-to-service calls.

```go
import "github.com/phcp-tech/common-library-golang/httpclient"

// Default settings: Timeout=10s, RetryMax=3.
cli := httpclient.NewHttpClient()

// Custom settings — zero-value fields fall back to defaults.
cli := httpclient.NewHttpClient(httpclient.Config{
    Timeout:            15 * time.Second,
    RetryMax:           5,
    InsecureSkipVerify: true, // only for internal self-signed certs
})

resp, err := cli.Get(url, jwtToken, nil)
resp, err := cli.Post(url, jwtToken, body)
resp, err := cli.Put(url, jwtToken, body)
resp, err := cli.Delete(url, jwtToken, body)

// Access the underlying resty.Client for advanced use.
restyClient := cli.Client()
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpclient#pkg-examples).

---

## httpclient/retryable — Retryable HTTP Client

Wraps [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp)
and exposes it as a standard `*http.Client` with configurable retry and timeout.
Use this sub-package when existing code already uses `net/http` directly and you
need to add retry behaviour without switching to resty.

**Import this sub-package only when needed** — it pulls in the hashicorp library
independently of the parent `httpclient` package.

```go
import "github.com/phcp-tech/common-library-golang/httpclient/retryable"

cli := retryable.NewHttpClient(retryable.Config{
    Timeout:  15 * time.Second,
    RetryMax: 5,
})

// Obtain a standard *http.Client with retry built in.
stdClient := cli.Client().StandardClient()
resp, err := stdClient.Do(req)
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpclient/retryable#pkg-examples).

---

## httpserver — Production HTTP/HTTPS Server

A production-ready HTTP/HTTPS server built on `net/http` with secure defaults.
It wraps `http.Server` directly (not `gin.Run()`) to provide:

- **Configurable timeouts** — `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout`
- **TLS 1.2+ with strong cipher suites** — activated when `CrtFile` and `KeyFile` are set
- **Graceful shutdown** — `Shutdown(ctx)` drains in-flight requests before stopping

The `Runner` interface unifies HTTP and Lambda modes so the composition root
selects the backend and the rest of the application code is unchanged.

```go
import "github.com/phcp-tech/common-library-golang/httpserver"

// Plain HTTP with package defaults.
runner := httpserver.NewHttpServer(httpserver.Config{Port: "8080"})

// HTTPS with custom timeouts.
runner := httpserver.NewHttpServer(httpserver.Config{
    Port:        "8443",
    CrtFile:     "/etc/ssl/server.crt",
    KeyFile:     "/etc/ssl/server.key",
    ReadTimeout: 15 * time.Second,
    WriteTimeout: 0, // 0 = unlimited (required for file downloads)
})

// Start in a goroutine; block until OS signal, then shut down gracefully.
go func() { _ = runner.Start(ginRouter) }()
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = runner.Shutdown(ctx)
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpserver#pkg-examples).

---

## httpserver/lambda — AWS Lambda Adapter

Implements `httpserver.Runner` for AWS Lambda via
[aws-lambda-go-api-proxy](https://github.com/awslabs/aws-lambda-go-api-proxy).
It bridges `APIGatewayProxyRequest` events to any `http.Handler` (including
`*gin.Engine`) with no handler changes required.

**Import this sub-package only for Lambda deployments** — it pulls in the AWS
Lambda SDK. Services that run as plain HTTP servers are unaffected.

```go
import (
    "github.com/phcp-tech/common-library-golang/httpserver"
    lambdarunner "github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

// Composition root selects the runner based on the deployment environment.
var runner httpserver.Runner
if isLambda {
    runner = lambdarunner.NewHttpServer()
} else {
    runner = httpserver.NewHttpServer(httpserver.Config{Port: port})
}
_ = runner.Start(ginRouter) // same call regardless of mode
```

See [full examples](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpserver/lambda#pkg-examples).
