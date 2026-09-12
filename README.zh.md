# PHCP: common-library-golang

<p align="center">
  <a href="./README.md">English</a> | <a href="./README.zh.md">简体中文</a>
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/phcp-tech/common-library-golang.svg)](https://pkg.go.dev/github.com/phcp-tech/common-library-golang)
[![LICENSE](https://img.shields.io/badge/license-Apache--2.0-green.svg)](https://github.com/phcp-tech/common-library-golang/blob/main/LICENSE)
![CI](https://github.com/phcp-tech/common-library-golang/actions/workflows/deploy-build-test.yml/badge.svg)
[![codecov](https://codecov.io/gh/phcp-tech/common-library-golang/branch/main/graph/badge.svg)](https://app.codecov.io/gh/phcp-tech/common-library-golang)

![common-library-golang](docs/common-library-golang-banner.png)

Common-library-golang 是 PHCP 生态系统内使用的一组功能组件集合，为微服务开发提供了大量组件，内置了诸如环境配置、日志、数据库等开箱即用的功能。为了让它能够被更广泛地使用，现已基于 Apache 许可证开源。

## 环境要求

- Go 1.25+

## 安装

```bash
go get github.com/phcp-tech/common-library-golang
```

## 构建与测试

```bash
go mod tidy
go vet ./...
go build ./...
```

### 快速模式测试 — 用于 CI 的快速验证

部分测试会启动一个真实的 TCP 监听器（例如 `httpserver` 的集成测试）。传入
`-short` 参数可以跳过这些测试，保持测试运行的轻量化：

```bash
go test ./... -short
```

### 完整测试运行 — 包含集成测试

```bash
go test ./...
```

### 运行覆盖率测试

```bash
# 全部包
go test ./... -cover -timeout 60s

```

## 组件包

| 分组 | 包 | 导入路径 | 描述 |
|-------|---------|-------------|-------------|
| 基础 | [`env`](#env--配置管理) | `.../common-library-golang/env` | TOML 配置 + 环境变量加载器 |
| 基础 | [`log`](#log--结构化日志) | `.../common-library-golang/log` | 带文件轮转和环形缓冲区的结构化 JSON 日志器 |
| 基础 | [`health`](#health--可组合健康检查) | `.../common-library-golang/health` | 面向 `/health` 端点的可组合健康检查聚合器 |
| 基础 | [`version`](#version--应用版本元数据) | `.../common-library-golang/version` | 面向 `/version` 端点的应用版本与构建元数据 |
| 基础 | [`shutdown`](#shutdown--优雅关闭) | `.../common-library-golang/shutdown` | 阻塞直到收到 OS 信号或程序化触发，随后继续执行清理 |
| 基础 | [`ringbuf`](#ringbuf--环形缓冲区) | `.../common-library-golang/ringbuf` | 无锁环形缓冲区（SPSC 和 MPSC） |
| 基础 | [`maps`](#maps--线程安全的并发映射) | `.../common-library-golang/maps` | 支持可插拔替换策略的线程安全泛型并发映射 |
| 基础 | [`cache`](#cache--进程内缓存-otter) | `.../common-library-golang/cache` | 由 [Otter](https://github.com/maypok86/otter) 支持的带 TTL 的进程内键值缓存 |
| 基础 | [`cgroup`](#cgroup--linux-资源限制) | `.../common-library-golang/cgroup` | 从 cgroup v2 读取 CPU 和内存资源限制（仅限 Linux；其他平台返回 0） |
| 基础 | [`metrics`](#metrics--运行时指标快照) | `.../common-library-golang/metrics` | 运行时与系统指标快照（CPU、内存、goroutine、cgroup 限制、运行时长） |
| Bootstrap | [`bootstrap`](#bootstrap--应用生命周期编排器) | `.../common-library-golang/bootstrap` | 顺序 Init + LIFO Close 编排器；第 1 次 `Add()` = env，第 2 次 `Add()` = log（约定） |
| Bootstrap | [`env/component`](#bootstrap-组件包) | `.../common-library-golang/env/component` | `env` 包的 `IComponent` 适配器 |
| Bootstrap | [`log/component`](#bootstrap-组件包) | `.../common-library-golang/log/component` | `log` 包的 `IComponent` 适配器 |
| Bootstrap | [`auth/component`](#bootstrap-组件包) | `.../common-library-golang/auth/component` | `auth`（casbin）包的 `IComponent` 适配器 |
| Bootstrap | [`token/component`](#bootstrap-组件包) | `.../common-library-golang/token/component` | `token`（JWT）包的 `IComponent` 适配器；若 jwt.issuer 或 jwt.access.secretcode 未设置或仍为占位符，或 jwt.refresh.secretcode 被显式声明却仍保持占位符值，则 Init 会失败 |
| Bootstrap | [`gin/component`](#bootstrap-组件包) | `.../common-library-golang/gin/component` | `gin` 引擎的 `IComponent` 适配器 |
| Bootstrap | [`redis/component`](#bootstrap-组件包) | `.../common-library-golang/redis/component` | `redis` 包的 `IComponent` 适配器 |
| Bootstrap | [`dbgorm/clickhouse/component`](#bootstrap-组件包) | `.../common-library-golang/dbgorm/clickhouse/component` | GORM ClickHouse 连接的 `IComponent` 适配器（Init 时执行 ping） |
| Bootstrap | [`dbgorm/mysql/component`](#bootstrap-组件包) | `.../common-library-golang/dbgorm/mysql/component` | GORM MySQL 连接的 `IComponent` 适配器（Init 时执行 ping） |
| Bootstrap | [`dbgorm/postgres/component`](#bootstrap-组件包) | `.../common-library-golang/dbgorm/postgres/component` | GORM PostgreSQL 连接的 `IComponent` 适配器（Init 时执行 ping） |
| Bootstrap | [`dbgorm/sqlite/component`](#bootstrap-组件包) | `.../common-library-golang/dbgorm/sqlite/component` | GORM SQLite 连接的 `IComponent` 适配器 |
| Bootstrap | [`dbsqlc/mysql/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlc/mysql/component` | MySQL 连接的 `IComponent` 适配器（sql.Open，惰性连接） |
| Bootstrap | [`dbsqlc/postgres/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlc/postgres/component` | PostgreSQL 连接池的 `IComponent` 适配器 |
| Bootstrap | [`dbsqlc/sqlite/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlc/sqlite/component` | SQLite 连接的 `IComponent` 适配器 |
| Bootstrap | [`dbsqlx/clickhouse/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlx/clickhouse/component` | sqlx ClickHouse 连接的 `IComponent` 适配器（Init 时执行 ping） |
| Bootstrap | [`dbsqlx/clickhouse-native/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlx/clickhouse-native/component` | 原生驱动 ClickHouse 连接的 `IComponent` 适配器（惰性打开） |
| Bootstrap | [`dbsqlx/mysql/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlx/mysql/component` | sqlx MySQL 连接的 `IComponent` 适配器（Init 时执行 ping） |
| Bootstrap | [`dbsqlx/postgres/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlx/postgres/component` | sqlx PostgreSQL 连接的 `IComponent` 适配器（Init 时执行 ping） |
| Bootstrap | [`dbsqlx/sqlite/component`](#bootstrap-组件包) | `.../common-library-golang/dbsqlx/sqlite/component` | sqlx SQLite 连接的 `IComponent` 适配器 |
| Bootstrap | [`httpserver/component`](#bootstrap-组件包) | `.../common-library-golang/httpserver/component` | HTTP 服务器的 `IComponent` 适配器；`ComponentWithRunner` 用于注入自定义 runner |
| Bootstrap | [`httpserver/componentwithlambda`](#bootstrap-组件包) | `.../common-library-golang/httpserver/componentwithlambda` | 支持 AWS Lambda 的 `IComponent` 适配器；通过 `app.runmode` 选择 runner |
| 数据库 | [`redis`](#redis--redis-客户端) | `.../common-library-golang/redis` | 带连接池和键扫描工具的 Redis 客户端（单机与集群模式） |
| 数据库 | [`dbgorm/clickhouse`](#dbgormclickhouse--gorm-clickhouse) | `.../common-library-golang/dbgorm/clickhouse` | 基于 GORM 的 ClickHouse —— Open 时立即 ping，共享进程级 `*gorm.DB` |
| 数据库 | [`dbgorm/mysql`](#dbgormmysql--gorm-mysql) | `.../common-library-golang/dbgorm/mysql` | 基于 GORM 的 MySQL —— Open 时立即 ping，共享进程级 `*gorm.DB` |
| 数据库 | [`dbgorm/postgres`](#dbgormpostgres--gorm-postgresql) | `.../common-library-golang/dbgorm/postgres` | 基于 GORM 的 PostgreSQL —— Open 时立即 ping，共享进程级 `*gorm.DB` |
| 数据库 | [`dbgorm/sqlite`](#dbgormsqlite--gorm-sqlite) | `.../common-library-golang/dbgorm/sqlite` | 基于 GORM 的 SQLite —— 嵌入式，无需网络，Open 时自动创建文件 |
| 数据库 | [`dbgorm/sqlite/vfs`](#dbgormsqlitevfs--通过-vfs-嵌入的-gorm-sqlite) | `.../common-library-golang/dbgorm/sqlite/vfs` | 基于 GORM、通过嵌入式文件系统的 SQLite —— 二进制内嵌的只读数据库 |
| 数据库 | [`dbsqlc/mysql`](#dbsqlcmysql--mysql-连接) | `.../common-library-golang/dbsqlc/mysql` | 通过标准 `database/sql` 和 go-sql-driver 实现的 MySQL 连接 |
| 数据库 | [`dbsqlc/postgres`](#dbsqlcpostgres--postgresql-连接池) | `.../common-library-golang/dbsqlc/postgres` | 通过 pgx/v5 实现的 PostgreSQL 连接池 |
| 数据库 | [`dbsqlc/sqlite`](#dbsqlcsqlite--sqlite-连接) | `.../common-library-golang/dbsqlc/sqlite` | 通过纯 Go 的 modernc 驱动实现的 SQLite 连接 |
| 数据库 | [`dbsqlc/sqlite/vfs`](#dbsqlcsqlitevfs--通过-vfs-嵌入的-sqlite) | `.../common-library-golang/dbsqlc/sqlite/vfs` | 通过 VFS 基于嵌入式文件系统的 SQLite（二进制内嵌数据库） |
| 数据库 | [`dbsqlx/clickhouse`](#dbsqlxclickhouse--sqlx-clickhouse) | `.../common-library-golang/dbsqlx/clickhouse` | 基于 sqlx 的 ClickHouse（clickhouse-go/v2 的 database/sql 驱动）—— Open 时立即 ping，共享进程级 `*sqlx.DB` |
| 数据库 | [`dbsqlx/clickhouse-native`](#dbsqlxclickhouse-native--clickhouse-原生驱动) | `.../common-library-golang/dbsqlx/clickhouse-native` | 通过 clickhouse-go/v2 实现的 ClickHouse 原生 TCP 客户端 —— 返回 `driver.Conn`，而非 `*sqlx.DB` |
| 数据库 | [`dbsqlx/mysql`](#dbsqlxmysql--sqlx-mysql) | `.../common-library-golang/dbsqlx/mysql` | 基于 sqlx 的 MySQL —— Open 时立即 ping，共享进程级 `*sqlx.DB` |
| 数据库 | [`dbsqlx/postgres`](#dbsqlxpostgres--sqlx-postgresql) | `.../common-library-golang/dbsqlx/postgres` | 基于 sqlx（pgx stdlib 驱动）的 PostgreSQL —— Open 时立即 ping 并执行 `SHOW search_path` |
| 数据库 | [`dbsqlx/sqlite`](#dbsqlxsqlite--sqlx-sqlite) | `.../common-library-golang/dbsqlx/sqlite` | 基于 sqlx 的 SQLite —— 嵌入式，无需网络，Open 时自动创建文件 |
| 数据库 | [`dbsqlx/sqlite/vfs`](#dbsqlxsqlitevfs--通过-vfs-嵌入的-sqlx-sqlite) | `.../common-library-golang/dbsqlx/sqlite/vfs` | 基于 sqlx、通过嵌入式文件系统的 SQLite —— 二进制内嵌的只读数据库 |
| 网络 | [`network`](#network--网络工具) | `.../common-library-golang/network` | 网络工具辅助函数 |
| 网络 | [`auth`](#auth--casbin-rbac-鉴权) | `.../common-library-golang/auth` | 面向 Gin 的 Casbin RBAC 鉴权中间件 |
| 网络 | [`token`](#token--jwt-认证) | `.../common-library-golang/token` | JWT 访问/刷新令牌的创建、解析及 Gin 中间件 |
| 网络 | [`gin`](#gin--gin-引擎工厂) | `.../common-library-golang/gin` | 带 slog 请求日志和 CORS 支持的 Gin 引擎工厂 |
| 网络 | [`gin/pprof`](#ginpprof--性能分析端点) | `.../common-library-golang/gin/pprof` | 面向 Gin 的可选 pprof 性能分析端点（需显式启用） |
| 网络 | [`httpclient`](#httpclient--resty-http-客户端) | `.../common-library-golang/httpclient` | 基于 Resty 的 HTTP 客户端，支持重试、JWT 鉴权和 JSON 辅助函数 |
| 网络 | [`httpclient/retryable`](#httpclientretryable--可重试的-http-客户端) | `.../common-library-golang/httpclient/retryable` | 封装 hashicorp/go-retryablehttp、返回标准 *http.Client 的客户端 |
| 网络 | [`httpserver`](#httpserver--生产级-httphttps-服务器) | `.../common-library-golang/httpserver` | 具备超时控制与优雅关闭的生产级 HTTP/HTTPS 服务器 |
| 网络 | [`httpserver/lambda`](#httpserverlambda--aws-lambda-适配器) | `.../common-library-golang/httpserver/lambda` | 实现 httpserver.IRunner 的 AWS Lambda 适配器（需显式启用） |

---

## env — 配置管理

加载一个 TOML 配置文件，并在其上合并 OS 环境变量。
基于 [koanf](https://github.com/knadh/koanf) 构建。实现了单例模式：
只有第一次 `InitEnv` 调用会真正生效。

```go
import "github.com/phcp-tech/common-library-golang/env"

// Call once at startup before any Env() usage.
if err := env.InitEnv("config.toml"); err != nil {
    log.Fatal(err)
}

host := env.Env().String("server.host")
port := env.Env().Int("server.port")
```

对于单二进制部署场景，配置文件可以在编译期内嵌：

```go
//go:embed config.toml
var configFS embed.FS

env.InitEnv("config.toml", &configFS)
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/env#pkg-examples)。

---

## log — 结构化日志

通过 `log/slog` 实现的结构化 JSON 日志，使用 UTC 时间戳。
文件写入完全**异步**：调用方将格式化后的日志条目推入内部的 `RingMPSC` 缓冲区后立即返回；由专用的消费者 goroutine 执行实际的 I/O。这使得该包适用于**高吞吐、对延迟敏感的场景**，在这类场景中阻塞在磁盘 I/O 上是不可接受的。

**`InitLog` 必须在应用启动时调用一次，且要在任何日志函数之前调用。**
省略参数则以 INFO 级别输出到标准输出；传入 `Config` 可自定义行为。

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

可用的日志函数：`Debug` / `Info` / `Warn` / `Error` 及其 `f`（格式化）和 `With`（结构化键值对）变体。日志级别可通过 `SetLevel` 在运行时修改。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/log#pkg-examples)。

---

## health — 可组合健康检查

面向 `/health` HTTP 端点的、基于接口的健康检查聚合器。
每个基础设施包都提供一个 `Checker`，由其自行维护组件名称和可达性状态；`Check` 会运行所有 checker 并返回它们的
汇总结果。

```go
import (
    "github.com/phcp-tech/common-library-golang/health"
    db    "github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
    cache "github.com/phcp-tech/common-library-golang/redis"
)

// Compose any number of checkers in a /health handler.
router.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, health.Check(
        c.Request.Context(),
        db.HealthChecker(),
        cache.HealthChecker(),
    ))
})
// JSON response: [{"name":"postgres","status":1},{"name":"redis","status":1}]
```

| 常量 | 值 | 含义 |
|---|---|---|
| `health.StatusHealthy` | `1` | 组件可达 |
| `health.StatusUnhealthy` | `0` | 组件不可达或尚未初始化 |

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/health#pkg-examples)。

---

## version — 应用版本元数据

为 `/version` HTTP 端点返回应用版本与构建元数据。
从 `env.Env()` 读取 `app.name`、`app.version` 和 `app.env.value`，并
根据内嵌的 Go 构建信息填充 `GoVersion` / `BuildInfo`。
要求在应用启动时已调用过 `env.InitEnv`。

```go
import "github.com/phcp-tech/common-library-golang/version"

router.GET("/version", func(c *gin.Context) {
    c.JSON(http.StatusOK, version.Get())
})
// JSON response:
// {
//   "name": "my-service",
//   "version": "1.2.3",
//   "environment": "production",
//   "goVersion": "go1.26.1",
//   "buildInfo": "v1.2.3"
// }
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/version#pkg-examples)。

---

## shutdown — 优雅关闭

用于应用关闭协调的两个基础组件：

- **`Wait`** 阻塞调用方所在的 goroutine，直到收到 OS 信号（`SIGINT`、`SIGTERM`、`SIGHUP`、`SIGQUIT`）或调用了 `Trigger`。返回后，调用方执行清理并退出。
- **`Trigger`** 从任意 goroutine 以编程方式解除 `Wait` 的阻塞 —— 适用于 `/shutdown` HTTP 端点或指标失败处理器。可安全地多次调用。

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

> **注意：** 当进程被 OS 强制终止时（例如 Windows 上 IDE 的停止按钮会调用
> `TerminateProcess` —— 不会发送任何信号），`Trigger` 不会产生任何效果。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/shutdown#pkg-examples)。

---

## ringbuf — 环形缓冲区

面向生产者-消费者流水线的高性能固定容量环形缓冲区。

| 类型 | 使用场景 | 线程安全性 |
|------|----------|---------------|
| `RingSPSC` | 单生产者、单消费者 | 无锁（仅使用原子操作） |
| `RingMPSC` | 多生产者、单消费者 | 生产者一侧使用互斥锁 |

两种类型都支持可选的 `ProcessFunc`，用于自动启动一个消费者 goroutine；
也支持手动调用 `Pop` / `TryPop` 以实现由调用方自行管理的消费方式。
`Push` 在缓冲区已满时会阻塞（背压）；`TryPush` 会立即返回 false。

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

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/ringbuf#pkg-examples)。

### 性能

基准测试运行方式：`go test -bench=. -benchtime=3s -benchmem`。
所有操作都产生**零堆内存分配**。

**测试环境**

| | |
|---|---|
| CPU | 11th Gen Intel® Core™ i7-11850H @ 2.50 GHz（8 核 / 16 线程） |
| 内存 | 32 GB |
| 操作系统 | Windows 11 Enterprise |
| Go | 1.26.2 windows/amd64 |

**测试结果**

| 基准测试 | ns/op | 吞吐量 | 分配次数 |
|-----------|------:|----------:|-------:|
| `SPSC Push`（1 生产者-1 消费者，阻塞） | 76.65 | ~13.0 M ops/s | 0 |
| `SPSC TryPush`（1 生产者-1 消费者，非阻塞） | 87.91 | ~11.4 M ops/s | 0 |
| `SPSC ProducerConsumer`（配合 ProcessFunc 的端到端流程） | 83.72 | ~11.9 M ops/s | 0 |
| `SPSC Push string`（贴近真实日志负载） | 91.06 | ~11.0 M ops/s | 0 |
| `MPSC Push`（1 个生产者） | 151.1 | ~6.6 M ops/s | 0 |
| `MPSC Push`（4 个并发生产者） | 192.2 | ~5.2 M ops/s | 0 |
| `MPSC Push`（8 个并发生产者） | 200.7 | ~5.0 M ops/s | 0 |
| `MPSC ProducerConsumer`（4 生产者端到端） | 197.5 | ~5.1 M ops/s | 0 |
| `Go channel`（带缓冲 4096，作为参考） | 126.3 | ~7.9 M ops/s | 0 |

**关键观察**

- 在相同的单生产者/单消费者模式下，SPSC 比带缓冲的 Go channel **快约 1.6 倍**。
- MPSC 通过互斥锁换取多生产者安全性；在 8 个并发生产者的情况下，其吞吐量仍保持在 **500 万 ops/s** 以上。
- 随着生产者并发数增加，吞吐量能够优雅地扩展：从 1 个生产者增加到 8 个，每个条目的延迟仅增加约 33%。
- 两种类型都保持**每次操作零内存分配**，使它们适合作为 `log` 包的异步 I/O 支撑机制。

---

## maps — 线程安全的并发映射

基于 [orcaman/concurrent-map](https://github.com/orcaman/concurrent-map) 构建、支持
可插拔替换策略的线程安全泛型并发映射。

提供了两种实现：

| 类型 | 键 | 值 | 策略 |
|------|-----|-------|----------|
| `CMap` | `string`（固定） | `int64`（固定） | 数值更大者获胜（内置） |
| `CMapGen[K, V]` | 任意 `comparable` | 任意类型 | 未配置时：总是覆盖写入。可通过 `SetDefaultCompare` 或 `SetDefaultStrategy` 配置 |

内置的策略类型：`NumericGreaterStrategy[V cmp.Ordered]`（真实的数值/有序比较 —— 要求 `V` 满足 `cmp.Ordered`）、`AlwaysReplaceStrategy`、`TimestampStrategy`。在调用 `SetDefaultCompare`/`SetDefaultStrategy` 之前，`CMapGen.Replace` 的行为等同于 `ReplaceAlways` —— 因为 `V` 是无约束的，所以无法为任意类型自动接入某种策略。

```go
import "github.com/phcp-tech/common-library-golang/maps"

// Generic map — configure a default compare function.
m := maps.NewCMapGen[string, int64]()
m.SetDefaultCompare(func(old, new int64) bool { return new > old })

m.Set("EURUSD", 10500)
m.Replace("EURUSD", 10600)       // stored: 10600 > 10500
m.ReplaceAlways("EURUSD", 9000)  // always stored
m.ReplaceIfNotExists("GBPUSD", 12800) // stored only if key absent

// Read-modify-write via callback (not atomic for concurrent writes to the same key — see below).
m.UpsertWithCallback("USDJPY", 15100, func(exists bool, old, new int64) int64 {
    if !exists || new > old { return new }
    return old
})

// Fixed string→int64 map (no configuration needed).
cm := maps.NewCMap()
cm.Set("tick", 10000)
cm.Replace("tick", 10050) // stored: 10050 > 10000
```

`Replace`/`ReplaceWithCompare`/`ReplaceWithStrategy`/`UpsertWithCallback` 都是通过两次分别加锁的操作来完成"读旧值、写新值"的，而不是单一的原子操作 —— 对*不同*键的并发写入是安全的，即便对*同一个*键并发写入也是安全的（不会损坏数据或 panic），但在竞争之后最终留下的值，是最后一次执行的 `Set` 所写入的值，未必是比较/策略/回调逻辑判定应当获胜的那个值。若需要对单个键实现真正原子的读-改-写，请直接调用底层的 `cmap.ConcurrentMap.Upsert` —— 它会在整个回调执行期间持有分片锁。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/maps#pkg-examples)。

### 性能

基准测试运行方式：`go test -bench=. -benchtime=3s -benchmem`。

**测试环境** —— 与 ringbuf 基准测试相同的机器：
Intel® Core™ i7-11850H @ 2.50 GHz · 8 核 / 16 线程 · 32 GB 内存 · Go 1.26.2 / Windows 11

**CMap**（固定 `string→int64`，内置的数值更大者获胜策略）

| 基准测试 | ns/op | 吞吐量 | B/op | 分配次数 |
|-----------|------:|----------:|-----:|-------:|
| `Set` | 27.51 | ~36.3 M ops/s | 0 | 0 |
| `Get` | 15.70 | ~63.7 M ops/s | 0 | 0 |
| `Replace` | 43.99 | ~22.7 M ops/s | 0 | 0 |
| `Set`（并行，16 个 goroutine） | 70.82 | ~14.1 M ops/s | 0 | 0 |
| `Replace`（并行，16 个 goroutine） | 151.5 | ~6.6 M ops/s | 0 | 0 |

**CMapGen**（泛型 `[K comparable, V any]`）

| 基准测试 | ns/op | 吞吐量 | B/op | 分配次数 | 备注 |
|-----------|------:|----------:|-----:|-------:|------|
| `Set` | 31.75 | ~31.5 M ops/s | 0 | 0 | |
| `Get` | 18.20 | ~55.0 M ops/s | 0 | 0 | |
| `Replace`（`NumericGreaterStrategy`，显式配置） | 47.63 | ~21.0 M ops/s | 0 | 0 | |
| `ReplaceWithCompare` | 46.32 | ~21.6 M ops/s | 0 | 0 | |
| `ReplaceAlways` | 30.05 | ~33.3 M ops/s | 0 | 0 | |
| `UpsertWithCallback` | 49.29 | ~20.3 M ops/s | 0 | 0 | |
| `Set`（并行，16 个 goroutine） | 71.17 | ~14.1 M ops/s | 0 | 0 | |
| `Replace`（并行，16 个 goroutine） | 167.9 | ~6.0 M ops/s | 0 | 0 | |
| 混合读写（并行） | 25.96 | ~38.5 M ops/s | 0 | 0 | |

**`sync.Map`**（标准库参考基准）

| 基准测试 | ns/op | 吞吐量 | B/op | 分配次数 |
|-----------|------:|----------:|-----:|-------:|
| `Store` | 60.90 | ~16.4 M ops/s | 56 | 1 |
| `Load` | 10.48 | ~95.4 M ops/s | 0 | 0 |
| `Store`（并行，16 个 goroutine） | 110.0 | ~9.1 M ops/s | 56 | 1 |

**关键观察**

- `CMap.Set` 比 `sync.Map.Store` **快约 2.2 倍**，且不产生任何内存分配。
- 无论是由显式配置的 `NumericGreaterStrategy`/`SetDefaultStrategy`、`SetDefaultCompare` 驱动，还是未做任何配置（此时行为等同于 `ReplaceAlways`，同样是零分配），`CMapGen.Replace` 都是零分配的 —— 不存在会退化为字符串格式化的默认路径。
- 当不需要条件逻辑时，`CMapGen.ReplaceAlways` 是最快的写入路径（约 3330 万 ops/s）。
- 在 16 个 goroutine 并行写入竞争下，由于单一键的热点竞争，`CMap.Replace` 的吞吐量会下降到约 660 万 ops/s；将写入分散到多个键上可以恢复吞吐量。

---

## cache — 进程内缓存 (Otter)

一个支持 TTL 的进程内键值缓存，定义在 `ICache` 接口之后，
使调用方永远不必依赖具体实现。目前唯一的
实现 `OtterCache` 基于
[maypok86/otter](https://github.com/maypok86/otter)（一种 W-TinyLFU 准入
缓存）。这个包没有 `Default()`/`Component()`，只提供普通的
构造函数，与 `httpclient`/`maps`/`ringbuf` 一致。

```go
import "github.com/phcp-tech/common-library-golang/cache"

c := cache.NewOtterCache() // capacity 10,000 entries, default TTL 1 hour

// Config is optional — customize capacity and/or the default TTL:
c2 := cache.NewOtterCache(cache.Config{MaxSize: 500, DefaultTTL: 10 * time.Minute})

// DefaultTTL: cache.NoExpiry builds a cache that never expires entries —
// this is a whole-cache setting: on a NoExpiry cache every Set's expire
// argument is ignored (see below), there's no per-key "except this one".
permanent := cache.NewOtterCache(cache.Config{DefaultTTL: cache.NoExpiry})

_ = c.Set("session-token", "abc123", 5*time.Minute) // custom per-entry TTL
_ = c.Set("counter", 1, 0)                          // 0 (or negative) uses the cache's default TTL

val, ok := c.Get("session-token")

_ = c.Update("counter", 2) // replaces the value, leaves the TTL untouched
_ = c.Delete("session-token")

c.Size()   // number of entries currently held (estimated)
c.Keys()   // []interface{} snapshot of all keys
c.Values() // []interface{} snapshot of all values
_ = c.Clear()
```

**`Set` 中的 `expire <= 0` 并不意味着"永不过期"** —— 它意味着"使用
缓存的默认 TTL"（除非 `Config.DefaultTTL` 覆盖，否则为 1 小时）。这
与某些其他缓存 API 中零值 TTL 的含义恰恰相反，如果你正在从那类 API
移植代码，请务必仔细核对这一点。对于需要条目真正永不过期的缓存，
请改用 `Config.DefaultTTL: NoExpiry` 来构造 —— 不要试图通过在单次
`Set` 上传入一个任意大的 `expire` 时长来模拟"永久"。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/cache#pkg-examples)。

### 性能

基准测试运行方式：`go test -bench=. -benchtime=2s -benchmem`。

**测试环境**：Intel® Core™ i7-11850H @ 2.50 GHz · 8 核 / 16 线程 · 32 GB 内存 · Go 1.26.2 / Windows 11

| 基准测试 | ns/op | 吞吐量 | B/op | 分配次数 |
|-----------|------:|----------:|-----:|-------:|
| `Set` | 389.4 | ~2.6 M ops/s | 120 | 4 |
| `Set`（自定义 TTL） | 1175 | ~0.85 M ops/s | 118 | 4 |
| `Get`（命中） | 638.8 | ~1.6 M ops/s | 16 | 2 |
| `Get`（未命中） | 176.6 | ~5.7 M ops/s | 24 | 2 |
| 混合（80% 读 / 20% 写） | 239.7 | ~4.2 M ops/s | 34 | 2 |
| 混合，并行（16 个 goroutine） | 85.43 | ~11.7 M ops/s | 33 | 2 |
| `Delete` | 110.0 | ~9.1 M ops/s | 15 | 1 |
| `Set`（1 KB 值） | 448.7 | ~2.2 M ops/s（2282 MB/s） | 123 | 4 |
| `Get`（1 KB 值，命中） | 181.8 | ~5.5 M ops/s（5633 MB/s） | 16 | 1 |
| `Get`（Zipf 热点键分布） | 619.7 | ~1.6 M ops/s | 10 | 1 |

**关键观察**

- 缓存**未命中比命中快约 3.6 倍**（176.6 ns 对比 638.8 ns）—— 命中会执行 Otter 的 `afterRead` 记账（记录命中、更新 TinyLFU 准入/淘汰策略状态），而未命中只记录一个计数器后就返回；这是准入缓存设计本身固有的特性，而不是缺陷。
- 带自定义 TTL 的 `Set`（1175 ns）比普通 `Set`（389.4 ns）**耗时约 3 倍** —— 自定义 TTL 需要额外一次调用（`Set` 之后再调用 `SetExpiresAfter`），而不是一次写入完成。
- 16 个 goroutine 并行的混合读写吞吐量（约 1170 万 ops/s）**高于**、而非低于单 goroutine 的 `Get`/`Set` 数值 —— 与上文 `maps.CMap` 的单键热点情形不同，这个基准测试将键分散在一个含 10,000 个键的空间中，因此竞争始终较低，这些数字体现的是真实的并行扩展性，而不是热点造成的假象。

---

## cgroup — Linux 资源限制

从 **cgroup v2** 统一层级结构（`/sys/fs/cgroup`）读取 CPU 和内存资源限制。
面向容器化工作负载（Kubernetes、Docker）设计，即进程运行在配置了 CPU 和内存约束的
cgroup 之中。

在**非 Linux 平台**（macOS、Windows）上，所有函数都返回 `(0, nil)` —— 没有可用的
cgroup 文件系统，也不会产生错误。

| 函数 | cgroup v2 文件 | 返回值 |
|----------|---------------|---------|
| `CPULimitMilli()` | `cpu.max` | 以毫核为单位的 CPU 限制；0 表示不限制 |
| `CPURequestMilli()` | `cpu.weight` | 通过 weight→shares→mCPU 换算得到的、以毫核为单位的 CPU 请求量 |
| `MemoryLimitBytes()` | `memory.max` | 以字节为单位的内存限制；0 表示不限制 |
| `MemoryRequestBytes()` | `memory.low` | 以字节为单位的内存软限制；0 表示未设置 |

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
参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/cgroup#pkg-examples)。

---

## metrics — 运行时指标快照

以扁平的 `[]NameValue` 切片形式收集关键运行时与系统指标的一次性
快照，适用于 `/metrics` 或 `/health` HTTP 端点。

每次调用都会通过 `gopsutil` **在 1 秒内采样 CPU 使用率**，因此它
适合用于周期性轮询（例如每 10-30 秒一次），而不适合用于热路径。

| 指标名称 | 描述 |
|-------------|-------------|
| `cpuPercent` | 进程在所有核心上的 CPU 使用率（%），在 1 秒内采样 |
| `memorySize` | 常驻内存集大小（RSS），单位 MiB |
| `threads` | OS 线程数 |
| `goroutines` | 存活的 goroutine 数量 |
| `gomaxprocs` | `runtime.GOMAXPROCS(0)` |
| `numCPU` | `runtime.NumCPU()` |
| `cpuRequest` | cgroup v2 的 CPU 请求量，单位毫核（未设置或非 Linux 时为 0） |
| `cpuLimit` | cgroup v2 的 CPU 限制，单位毫核（不限制或非 Linux 时为 0） |
| `memoryRequest` | cgroup v2 的内存软限制，单位字节（未设置或非 Linux 时为 0） |
| `memoryLimit` | cgroup v2 的内存限制，单位字节（不限制或非 Linux 时为 0） |
| `age` | 以 `Xd Xh Xm Xs` 格式表示的进程运行时长 |

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

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/metrics#pkg-examples)。

---

## bootstrap — 应用生命周期编排器

`bootstrap` 负责编排应用启动时的顺序初始化以及 LIFO（后进先出）清理。

> **组件注册顺序约定**
>
> **第一次** `Add()` 调用必须注册 **env** 组件。
> **第二次** `Add()` 调用必须注册 **log** 组件。
> 之后的所有调用（`Add`、`AddParallel`、`PreReady`）没有顺序限制。
>
> 之所以有这个约定，是因为：
> - 每个组件的 `Init()` 都会通过 `env.Env()` 读取配置，所以 env 必须最先初始化；
> - Go 的 `slog` 包从程序启动起就有一个可用的默认实例（写入到 stderr），因此
>   任何阶段（包括 `log.Init()` 之前）的 `Init()` 失败都会被 `slog` 捕获；
> - `env.Close()` 是一个空操作，因此 LIFO 顺序自然会让 `log.Close()` 成为最后一个有意义的关闭操作。

```go
import (
    "github.com/phcp-tech/common-library-golang/bootstrap"
    envComp  "github.com/phcp-tech/common-library-golang/env/component"
    logComp  "github.com/phcp-tech/common-library-golang/log/component"
    dbComp   "github.com/phcp-tech/common-library-golang/dbsqlc/postgres/component"
    ginComp  "github.com/phcp-tech/common-library-golang/gin/component"
    httpComp "github.com/phcp-tech/common-library-golang/httpserver/component"
)

func main() {
    var router *gin.Engine
    bootstrap.New().
        Add(envComp.Component("config/app.toml", &configFS)). // 1st — MUST be env
        Add(logComp.Component()).                              // 2nd — MUST be log
        AddParallel(dbComp.Component()).
        PreReady(migrate).
        PreReady(initServices).
        Add(ginComp.Component(func(r *gin.Engine) { router = r; mount(r) })).
        Add(httpComp.Component(func() http.Handler { return router })).
        PostReady(func() { slog.Info("server ready") }).
        Run()
}
```

### API

| 方法 | 签名 | 描述 |
|--------|-----------|-------------|
| `New` | `() *App` | 创建一个空的编排器 |
| `Add` | `(cs ...IComponent) *App` | 顺序阶段：按注册顺序执行 Init，按 LIFO 顺序执行 Close。**第 1 次调用 = env，第 2 次调用 = log** |
| `AddParallel` | `(cs ...IComponent) *App` | 并发阶段：所有 Init 调用并行执行；必须全部成功才会进入下一步 |
| `PreReady` | `(fn func() error) *App` | 在启动流程中内联的一次性设置函数；非 nil 错误会中止启动；没有 Close，不会纳入 LIFO 追踪 |
| `PostReady` | `(fn func()) *App` | 所有步骤成功后调用的通知钩子；多次调用会按注册顺序累积 |
| `Run` | `()` | 执行所有步骤，等待 SIGINT/SIGTERM，然后按 LIFO 顺序关闭 |

### IComponent 接口

```go
type IComponent interface {
    Name() string   // displayed in log messages
    Init() error    // non-nil error aborts startup and exits with code 1
    Close()         // called during shutdown in LIFO order; must never panic
}
```

当存在 Close 逻辑但不需要一个完整结构体时，使用 `bootstrap.Func` 将一对函数包装为 `IComponent`：

```go
bootstrap.Func("worker", startWorker, stopWorker)
```

### 启动与关闭流程

```
New()
  → Add(env).Init   ← slog (default handler writes to stderr before log.Init)
  → Add(log).Init   ← slog
  → steps (in registration order)
        ├── stepPhase    → Init, added to closeStack
        └── stepPreReady → fn(), not added to closeStack
  → PostReady callbacks
  → wait for signal
        ↓
  ← single closeAll (LIFO)
        custom components (reverse order)
        → log.Close   (last meaningful close — env.Close is a no-op)
        → env.Close   (no-op)
```

### PreReady

`PreReady` 将一个函数注册进启动流程，与 `Add`/`AddParallel` 共享同一个有序列表。非 nil 错误会像组件 `Init()` 失败一样中止启动。`PreReady` 步骤没有 Close，也不参与 LIFO 关闭流程。

```go
.AddParallel(dbComp.Component()).
PreReady(migrate).        // runs after DB is ready
PreReady(initServices).   // runs after migrate
Add(ginComp.Component(mount)).
```

| | `bootstrap.Func(..., nil)` | `PreReady(fn)` |
|---|---|---|
| 出现在步骤列表中 | ✓ 作为 `stepPhase` | ✓ 作为 `stepPreReady` |
| 参与 LIFO closeStack | ✓（Close 是空操作） | ✗ |
| 语义 | "一个没有 Close 的组件" | "启动流程中的一次性代码" |

### PostReady

`PostReady` 可以被多次调用。每次调用都会**追加**一个回调；所有回调都会在每个步骤都成功之后、进程阻塞等待 OS 信号之前按注册顺序运行。适用于不影响请求处理正确性的动作：

```go
.PostReady(func() { slog.Info("server ready", "addr", ":8080") }).
PostReady(func() { discovery.Register(serviceID) })
```

放在 `main` 中 `Run()` 之后的代码，或者放在 `Run()` 之前的 `defer` 中的代码，会在完整的关闭流程执行完毕后运行，适合用于关闭后的收尾工作。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/bootstrap#pkg-examples)。

---

### Bootstrap 组件包

每个基础包都附带一个配套的 `component/` 子包，用于将其适配为 `IComponent` 接口。组件包只在 `Init()` 内部通过 `env.Env()` 读取所有配置 —— 绝不在构造函数中读取 —— 因此在调用 `bootstrap.New()` 之前构造它们是安全的。

| 组件包 | Init 时读取的 Env 键 |
|-------------------|-----------------------|
| `env/component` | _（无 —— 它自己读取配置文件）_ |
| `log/component` | `log.level`、`log.file.path`、`log.file.max.size.mb`、`log.file.max.backups`、`log.file.max.age.days`、`log.file.compress` |
| `auth/component` | _（model 和 policy 作为构造函数参数传入）_ |
| `token/component` | `jwt.issuer`（必填 —— 在所有服务中必须一致；若为空或为 `"TOBE_REPLACED"` 则 Init 失败）、`jwt.access.secretcode`（必填 —— 若为空或为 `"TOBE_REPLACED"` 则 Init 失败）、`jwt.refresh.secretcode`（可选，若未设置则忽略；若被显式留为 `"TOBE_REPLACED"` 则 Init 失败） |
| `gin/component` | `app.env.value`、`cors.allow.origins.prod`、`cors.allow.origins.dev` |
| `redis/component` | `redis.clusters`、`redis.database`、`redis.password` |
| `dbgorm/clickhouse/component` | `db.host`、`db.port`、`db.name`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbgorm/mysql/component` | `db.host`、`db.port`、`db.name`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbgorm/postgres/component` | `db.host`、`db.port`、`db.name`、`db.schema`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbgorm/sqlite/component` | `db.sqlite.path` |
| `dbsqlc/mysql/component` | `db.host`、`db.port`、`db.name`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbsqlc/postgres/component` | `db.host`、`db.port`、`db.name`、`db.schema`、`db.username`、`db.password`、`db.pool.*` |
| `dbsqlc/sqlite/component` | `db.sqlite.path` |
| `dbsqlx/clickhouse/component` | `db.host`、`db.port`、`db.name`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbsqlx/clickhouse-native/component` | `db.host`、`db.port`、`db.name`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime` |
| `dbsqlx/mysql/component` | `db.host`、`db.port`、`db.name`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbsqlx/postgres/component` | `db.host`、`db.port`、`db.name`、`db.schema`、`db.username`、`db.password`、`db.max.open.conns`、`db.max.idle.conns`、`db.conn.max.lifetime`、`db.conn.max.idletime` |
| `dbsqlx/sqlite/component` | `db.sqlite.path` |
| `httpserver/component` | `http.server.port` |
| `httpserver/componentwithlambda` | `app.runmode`（`"aws_lambda"` → Lambda runner）、`http.server.port`（其他模式） |

---

## redis — Redis 客户端

同时支持**单机**和**集群**模式的线程安全 Redis 客户端，
基于 [go-redis/v9](https://github.com/redis/go-redis) 实现。
连接池设置（`PoolSize`、`MinIdleConns`）可通过 `Config` 配置，
带有合理的默认值（100 / 5）。调用方在组合根处从 `env.Env()` 读取这些值 ——
本包本身不依赖 env。

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

通过 `InitDefault` / `Default` 实现单例模式。

### 健康检查

`HealthChecker()` 返回一个 [`health.Checker`](#health--可组合健康检查)，用于 ping 默认客户端。
当尚未初始化任何客户端，或 ping 失败时，会报告 `StatusUnhealthy`。

```go
import (
    "github.com/phcp-tech/common-library-golang/health"
    "github.com/phcp-tech/common-library-golang/redis"
)

results := health.Check(c.Request.Context(), redis.HealthChecker())
// → []health.Result{{Name: "redis", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/redis#pkg-examples)。


---

## dbgorm/clickhouse — GORM ClickHouse

基于 [GORM](https://gorm.io) 的 ClickHouse 适配器。GORM 的驱动会在 **`Open` 期间 ping
服务器** —— 当数据库不可达时，`InitDefault` 会立即返回非 nil 错误，
导致 bootstrap 中止启动。
默认实例是一个进程级的 `*gorm.DB`，通过 `dbgorm.Default()` /
`dbgorm.SetDefault()` 共享。

> **注意：** ClickHouse 会自动禁用 `PrepareStmt`，因为该驱动会
> 把 `Prepare()` 映射为批量 INSERT；对于常规 SELECT 查询它必须关闭。
> 这一点已在 `dbgorm.Open` 内部透明处理。

```go
import (
    dbgorm     "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/dbgorm/clickhouse"
)

err := clickhouse.InitDefault(&clickhouse.Config{
    Host:     "localhost",
    Port:     "9440",   // native TCP+TLS; use 8123/8443 for HTTP
    Database: "mydb",
    Username: "user",
    Password: "pass",
})
if err != nil {
    log.Fatal(err)
}

db := dbgorm.Default() // *gorm.DB, ready for GORM operations
```

### 健康检查

使用根包中的 [`dbgorm.HealthChecker()`](#health--可组合健康检查) ——
该 checker 与具体方言无关，可配合任何通过 `InitDefault` 初始化的驱动使用。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), dbgorm.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbgorm/clickhouse#pkg-examples)。

---

## dbgorm/mysql — GORM MySQL

基于 [GORM](https://gorm.io) 的 MySQL 适配器。与 `dbsqlc/mysql` 不同，GORM 的驱动
会在 **`Open` 期间 ping 服务器** —— 当数据库不可达时，`InitDefault` 会立即
返回非 nil 错误，导致 bootstrap 中止启动。
默认实例是一个进程级的 `*gorm.DB`，通过 `dbgorm.Default()` /
`dbgorm.SetDefault()` 共享（没有 `sync.Once` —— 可在运行时被替换）。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/dbgorm/mysql"
)

// Open and store as the process-wide default.
err := mysql.InitDefault(&mysql.Config{
    Host:     "localhost",
    Port:     "3306",
    Database: "mydb",
    Username: "user",
    Password: "pass",
})
if err != nil {
    log.Fatal(err) // server was unreachable — GORM pinged on Open
}

db := dbgorm.Default() // *gorm.DB, ready for GORM operations
```

对于需要多个数据库的场景，请直接使用 `NewMySQL`，并将返回的
`*gorm.DB` 存放在显式字段中，而不是共享的默认实例中。

### 健康检查

使用根包中的 [`dbgorm.HealthChecker()`](#health--可组合健康检查) ——
该 checker 与具体方言无关，可配合任何通过 `InitDefault` 初始化的驱动使用。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), dbgorm.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbgorm/mysql#pkg-examples)。

---

## dbgorm/postgres — GORM PostgreSQL

基于 [GORM](https://gorm.io) 的 PostgreSQL 适配器。GORM 的驱动会在 **`Open` 期间 ping
服务器并执行 `SHOW search_path`** —— 当数据库不可达时，`InitDefault` 会立即
返回非 nil 错误，导致 bootstrap 中止启动。
支持可选的 `SearchPath` 以实现 schema 隔离。
默认实例是一个进程级的 `*gorm.DB`，通过 `dbgorm.Default()` /
`dbgorm.SetDefault()` 共享。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/dbgorm/postgres"
)

err := postgres.InitDefault(&postgres.Config{
    Host:       "localhost",
    Port:       "5432",
    Database:   "mydb",
    Username:   "user",
    Password:   "pass",
    SearchPath: "myschema", // optional
})
if err != nil {
    log.Fatal(err) // server was unreachable
}

db := dbgorm.Default() // *gorm.DB, ready for GORM operations
```

### 健康检查

使用根包中的 [`dbgorm.HealthChecker()`](#health--可组合健康检查) ——
该 checker 与具体方言无关，可配合任何通过 `InitDefault` 初始化的驱动使用。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), dbgorm.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbgorm/postgres#pkg-examples)。

---

## dbgorm/sqlite — GORM SQLite

基于 [glebarez/sqlite](https://github.com/glebarez/sqlite)（纯 Go，无需 CGO）的
[GORM](https://gorm.io) SQLite 适配器。
SQLite 是一个**嵌入式**数据库 —— 不涉及任何网络连接。
`InitDefault` 会打开（并自动创建）文件；只有在路径为空
或文件无法创建时才会失败。为防止并发写入时出现
`database is locked` 错误，`MaxOpenConns` 被固定为 1。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    "github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
)

// File-based database.
err := sqlite.InitDefault(&sqlite.Config{
    Path: "file:app.db?cache=shared",
})
if err != nil {
    log.Fatal(err)
}

db := dbgorm.Default() // *gorm.DB, ready for GORM operations

// In-memory database (tests, ephemeral data).
db, _ := sqlite.NewSQLite(&sqlite.Config{
    Path: "file::memory:?cache=shared",
})
```

> **没有 HealthChecker：** SQLite 是嵌入式的 —— 如果它出了问题，应用本身也已经出了问题。
> 为一个嵌入式数据库提供健康检查端点没有任何运维价值。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbgorm/sqlite#pkg-examples)。

---

## dbgorm/sqlite/vfs — 通过 VFS 嵌入的 GORM SQLite

使用 [modernc.org/sqlite/vfs](https://pkg.go.dev/modernc.org/sqlite/vfs) 和 Go 的
`embed.FS`，打开一个**内嵌在 Go 二进制文件中**的 SQLite 数据库。
被嵌入的文件必须位于所提供 `embed.FS` 内的 `config/sqlite.db` 路径下。

**只有在需要将 SQLite 数据库作为二进制文件的一部分进行分发时，才导入这个子包。**
对于常规的基于文件的数据库，请改用 `dbgorm/sqlite`。

`InitDefault` 使用 `sync.Once` —— 第一次调用生效；后续调用会被静默忽略。

```go
import (
    dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
    sqlvfs "github.com/phcp-tech/common-library-golang/dbgorm/sqlite/vfs"
)

//go:embed config/sqlite.db
var sqliteFS embed.FS

// Open directly (no singleton).
db, err := sqlvfs.New(&sqliteFS)

// Or store as the process-wide default (sync.Once).
if err := sqlvfs.InitDefault(&sqliteFS); err != nil {
    log.Fatal(err)
}
db := dbgorm.Default() // *gorm.DB, ready for GORM operations
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbgorm/sqlite/vfs#pkg-examples)。

---

## dbsqlc/mysql — MySQL 连接

供 [sqlc](https://sqlc.dev/)（`sql_package: "database/sql"`）使用的标准 `database/sql` 连接，
基于 [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) 实现。
连接池设置（`MaxOpenConns`、`MaxIdleConns`、`ConnMaxLifetime`、`ConnMaxIdletime`）
可通过 `Config` 配置，并带有合理的默认值。
通过 `InitDefault` / `Default` 实现单例模式。

> **注意：** `sql.Open` 是**惰性**的 —— `InitDefault` 期间不会建立任何连接。
> `Init()` 总是返回 nil。如果需要在启动时进行主动的连通性检查，请在
> `PreReady` 步骤中显式调用 `PingContext`。

```go
import "github.com/phcp-tech/common-library-golang/dbsqlc/mysql"

// Singleton mode: call once at startup.
err := mysql.InitDefault(&mysql.Config{
    Host:     "localhost",
    Port:     "3306",
    Database: "mydb",
    Username: "user",
    Password: "pass",
})
if err != nil {
    log.Fatal(err)
}

db := mysql.Default() // *sql.DB, pass to sqlc Queries
```

对于需要多个连接的场景，请直接使用 `NewMySQL`，而不是单例模式。

### 健康检查

`HealthChecker()` 返回一个 [`health.Checker`](#health--可组合健康检查)，用于 ping 默认连接。
当尚未初始化任何连接，或 ping 失败时，会报告 `StatusUnhealthy`。

```go
import (
    "github.com/phcp-tech/common-library-golang/health"
    "github.com/phcp-tech/common-library-golang/dbsqlc/mysql"
)

results := health.Check(c.Request.Context(), mysql.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/mysql#pkg-examples)。

---

## dbsqlc/postgres — PostgreSQL 连接池

供 [sqlc](https://sqlc.dev/)（`sql_package: "pgx/v5"`）使用的 pgx/v5 连接池。
连接池的创建是**惰性**的：`NewPostgres` 会立即返回而不建立任何
连接，因此启动时不需要一个存活的服务器。
通过 `InitDefault` / `Default` 实现单例模式。

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

对于需要多个连接池的场景，请直接使用 `NewPostgres`，而不是单例模式。

### 健康检查

`HealthChecker()` 返回一个 [`health.Checker`](#health--可组合健康检查)，用于 ping 默认连接池。
当尚未初始化任何连接池，或 ping 失败时，会报告 `StatusUnhealthy`。

```go
import (
    "github.com/phcp-tech/common-library-golang/health"
    db "github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
)

results := health.Check(c.Request.Context(), db.HealthChecker())
// → []health.Result{{Name: "postgres", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/postgres#pkg-examples)。

---

## dbsqlc/sqlite — SQLite 连接

纯 Go 的 SQLite 驱动（[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)，无需 CGO）。
供 [sqlc](https://sqlc.dev/)（`sql_package: "database/sql"`）使用。
通过 `InitDefault` / `Default` 实现单例模式。

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

SQLite 一次只允许一个写入者。当未使用 WAL 模式时，`NewSQLite` 会自动调用
`SetMaxOpenConns(1)` 以防止出现 `database is locked` 错误。

> **没有 HealthChecker：** SQLite 是嵌入式的 —— 如果它出了问题，应用本身也已经出了问题。
> 为一个嵌入式数据库提供健康检查端点没有任何运维价值。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/sqlite#pkg-examples)。

---

## dbsqlc/sqlite/vfs — 通过 VFS 嵌入的 SQLite

使用 [modernc.org/sqlite/vfs](https://pkg.go.dev/modernc.org/sqlite/vfs) 和 Go 的
`embed.FS`，打开一个**内嵌在 Go 二进制文件中**的 SQLite 数据库。数据库文件
必须位于内嵌文件系统中的 `config/sqlite.db` 路径下。

**只有在数据库作为二进制文件的一部分进行分发时，才导入这个子包。**
对于常规的基于文件的数据库，请改用 `dbsqlc/sqlite`。

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

对于需要多个 VFS 连接的场景，请直接使用 `New`，而不是单例模式。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlc/sqlite/vfs#pkg-examples)。

---

## dbsqlx/clickhouse — sqlx ClickHouse

基于 [github.com/vinovest/sqlx](https://github.com/vinovest/sqlx) 构建的 ClickHouse 适配器，
通过 [clickhouse-go/v2](https://github.com/ClickHouse/clickhouse-go) 的 `database/sql`
驱动（注册为 `"clickhouse"`）实现 —— 与 `gorm.io/driver/clickhouse` 内部
使用的驱动模式相同。`dbsqlx.Open` 会**主动 ping 服务器** ——
当数据库不可达时，`InitDefault` 会立即返回非 nil 错误，
导致 bootstrap 中止启动。
默认实例是一个进程级的 `*sqlx.DB`，通过 `dbsqlx.Default()` /
`dbsqlx.SetDefault()` 共享。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/dbsqlx/clickhouse"
)

err := clickhouse.InitDefault(&clickhouse.Config{
    Host:     "localhost",
    Port:     "9440", // native TCP+TLS
    Database: "mydb",
    Username: "user",
    Password: "pass",
})
if err != nil {
    log.Fatal(err) // server was unreachable — Open pinged eagerly
}

db := dbsqlx.Default() // *sqlx.DB, ready for Get/Select/Exec
```

### 健康检查

使用根包中的 [`dbsqlx.HealthChecker()`](#health--可组合健康检查) ——
该 checker 与具体方言无关，可配合任何通过 `InitDefault` 初始化的驱动使用。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), dbsqlx.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlx/clickhouse#pkg-examples)。

---

## dbsqlx/clickhouse-native — ClickHouse 原生驱动

基于 [clickhouse-go/v2](https://github.com/ClickHouse/clickhouse-go) 的
原生 TCP ClickHouse 客户端。
返回 `driver.Conn` —— **而非** `*sqlx.DB` —— 因此它不会与父级
`dbsqlx` 包中的 `dbsqlx.Open`/`Default`/`Exec`/`Transact` 集成。请将
这个包用于直接、高性能的原生协议访问；如果希望使用共享的
`*sqlx.DB` 便利方法，请改用 [`dbsqlx/clickhouse`](#dbsqlxclickhouse--sqlx-clickhouse)。

> **惰性打开：** `clickhousenative.NewClickHouse` 只会配置连接；TCP
> 拨号发生在第一次操作（Ping/Query 等）时。因此 `InitDefault` 总是
> 返回 nil。如果需要主动检查，请在 `PreReady` 步骤中使用 `conn.Ping(ctx)`。

```go
import "github.com/phcp-tech/common-library-golang/dbsqlx/clickhouse-native"

// Singleton mode — call once at startup.
err := clickhousenative.InitDefault(&clickhousenative.Config{
    Host:     "localhost",
    Port:     "9440", // native TCP+TLS
    Database: "mydb",
    Username: "user",
    Password: "pass",
})
if err != nil {
    log.Fatal(err)
}

conn := clickhousenative.Default() // driver.Conn, ready to use
```

### 健康检查

`HealthChecker()` 返回一个 [`health.Checker`](#health--可组合健康检查)，
它会对默认客户端调用 `conn.Ping(ctx)`。
当尚未初始化任何客户端，或 ping 失败时，会报告 `StatusUnhealthy`。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx/clickhouse-native"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), clickhousenative.HealthChecker())
// → []health.Result{{Name: "clickhouse-native", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlx/clickhouse-native#pkg-examples)。

---

## dbsqlx/mysql — sqlx MySQL

基于 [github.com/vinovest/sqlx](https://github.com/vinovest/sqlx)
（一个带反射结构体扫描和 Go 泛型的 `database/sql` 扩展）构建的 MySQL 适配器，
通过 `go-sql-driver/mysql` 驱动实现。与 `dbsqlc/mysql` 不同，`dbsqlx.Open` 会**主动
ping 服务器** —— 当数据库不可达时，`InitDefault` 会立即返回非 nil 错误，
导致 bootstrap 中止启动。
默认实例是一个进程级的 `*sqlx.DB`，通过 `dbsqlx.Default()` /
`dbsqlx.SetDefault()` 共享（没有 `sync.Once` —— 可在运行时被替换）。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/dbsqlx/mysql"
)

// Open and store as the process-wide default.
err := mysql.InitDefault(&mysql.Config{
    Host:     "localhost",
    Port:     "3306",
    Database: "mydb",
    Username: "user",
    Password: "pass",
})
if err != nil {
    log.Fatal(err) // server was unreachable — Open pinged eagerly
}

db := dbsqlx.Default() // *sqlx.DB, ready for Get/Select/Exec
```

对于需要多个数据库的场景，请直接使用 `NewMySQL`，并将返回的
`*sqlx.DB` 存放在显式字段中，而不是共享的默认实例中。

### 健康检查

使用根包中的 [`dbsqlx.HealthChecker()`](#health--可组合健康检查) ——
该 checker 与具体方言无关，可配合任何通过 `InitDefault` 初始化的驱动使用。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), dbsqlx.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlx/mysql#pkg-examples)。

---

## dbsqlx/postgres — sqlx PostgreSQL

基于 [github.com/vinovest/sqlx](https://github.com/vinovest/sqlx) 构建的 PostgreSQL 适配器，
通过 `pgx/v5/stdlib` 驱动实现（在 `database/sql` 中注册为 `"pgx"`）。
`dbsqlx.Open` 会**主动 ping 服务器并执行 `SHOW search_path`** ——
当数据库不可达时，`InitDefault` 会立即返回非 nil 错误。
支持可选的 `SearchPath` 以实现 schema 隔离。
默认实例是一个进程级的 `*sqlx.DB`，通过 `dbsqlx.Default()` /
`dbsqlx.SetDefault()` 共享。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/dbsqlx/postgres"
)

err := postgres.InitDefault(&postgres.Config{
    Host:       "localhost",
    Port:       "5432",
    Database:   "mydb",
    Username:   "user",
    Password:   "pass",
    SearchPath: "myschema", // optional
})
if err != nil {
    log.Fatal(err) // server was unreachable
}

db := dbsqlx.Default() // *sqlx.DB, ready for Get/Select/Exec
```

### 健康检查

使用根包中的 [`dbsqlx.HealthChecker()`](#health--可组合健康检查) ——
该 checker 与具体方言无关，可配合任何通过 `InitDefault` 初始化的驱动使用。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/health"
)

results := health.Check(c.Request.Context(), dbsqlx.HealthChecker())
// → []health.Result{{Name: "database", Status: health.StatusHealthy}}
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlx/postgres#pkg-examples)。

---

## dbsqlx/sqlite — sqlx SQLite

基于 [github.com/vinovest/sqlx](https://github.com/vinovest/sqlx) 构建的 SQLite 适配器，
通过纯 Go 的 [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) 驱动实现
（无需 CGO）。SQLite 是一个**嵌入式**数据库 —— 不涉及任何网络连接。
`InitDefault` 会打开（并自动创建）文件及其父目录；只有在路径为空
或文件无法创建时才会失败。每个连接都会应用标准的
PRAGMA 设置（WAL、外键、忙等待超时、32 MB 缓存）。
`MaxOpenConns` 被固定为 4 / `MaxIdleConns` 被固定为 2 —— 因为 SQLite 一次只
允许一个写入者，即便在 WAL 模式下，更大的连接池也不会带来任何收益。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    "github.com/phcp-tech/common-library-golang/dbsqlx/sqlite"
)

// File-based database.
err := sqlite.InitDefault(&sqlite.Config{
    Path: "file:app.db?cache=shared",
})
if err != nil {
    log.Fatal(err)
}

db := dbsqlx.Default() // *sqlx.DB, ready for Get/Select/Exec

// In-memory database (tests, ephemeral data).
db, _ = sqlite.NewSQLite(&sqlite.Config{
    Path: "file::memory:?cache=shared",
})
```

> **没有 HealthChecker：** SQLite 是嵌入式的 —— 如果它出了问题，应用本身也已经出了问题。
> 为一个嵌入式数据库提供健康检查端点没有任何运维价值。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlx/sqlite#pkg-examples)。

---

## dbsqlx/sqlite/vfs — 通过 VFS 嵌入的 sqlx SQLite

使用 [modernc.org/sqlite/vfs](https://pkg.go.dev/modernc.org/sqlite/vfs) 和 Go 的
`embed.FS`，打开一个**内嵌在 Go 二进制文件中**的 SQLite 数据库。
被嵌入的文件必须位于所提供 `embed.FS` 内的 `config/sqlite.db` 路径下。

**只有在需要将 SQLite 数据库作为二进制文件的一部分进行分发时，才导入这个子包。**
对于常规的基于文件的数据库，请改用 `dbsqlx/sqlite`。

`InitDefault` 使用 `sync.Once` —— 第一次调用生效；后续调用会被静默忽略。

```go
import (
    "github.com/phcp-tech/common-library-golang/dbsqlx"
    sqlvfs "github.com/phcp-tech/common-library-golang/dbsqlx/sqlite/vfs"
)

//go:embed config/sqlite.db
var sqliteFS embed.FS

// Open directly (no singleton).
db, err := sqlvfs.New(&sqliteFS)

// Or store as the process-wide default (sync.Once).
if err := sqlvfs.InitDefault(&sqliteFS); err != nil {
    log.Fatal(err)
}
db = dbsqlx.Default() // *sqlx.DB, ready for Get/Select/Exec
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/dbsqlx/sqlite/vfs#pkg-examples)。

---

## network — 网络工具

用于请求 IP 提取、本地网卡检查，以及 IPv4 ↔ uint32 转换的辅助函数。

### IP 地址辅助函数

```go
import "github.com/phcp-tech/common-library-golang/network"

// Real client IP — inspects X-Forwarded-For → X-Real-IP → RemoteAddr.
// Returns the first address in X-Forwarded-For when multiple proxies are chained.
// "::1" is normalised to "127.0.0.1".
ip := network.GetRemoteIp(req)

// All non-loopback IPv4 addresses on the local machine.
addrs := network.GetLocalIpAddress()
```

### IPv4 ↔ uint32 转换

支持两种字节序：

| 函数 | 字节序 | 典型用途 |
|----------|-----------|-------------|
| `Ip2IntWithBigEndian` | 大端序（网络字节序） | 标准 TCP/IP、数据库 |
| `Int2IpWithBigEndian` | 大端序 | 同上 |
| `Ip2IntWithLittleEndian` | 小端序 | MT4 / MT5 交易平台 |
| `Int2IpWithLittleEndian` | 小端序 | 同上 |

```go
// Big-endian (network byte order): "1.2.3.4" → 0x01020304
n := network.Ip2IntWithBigEndian("1.2.3.4")   // → 16909060
ip := network.Int2IpWithBigEndian(0x01020304)  // → "1.2.3.4"

// Little-endian (MT4/MT5 format): "1.2.3.4" → 0x04030201
n := network.Ip2IntWithLittleEndian("1.2.3.4")   // → 67305985
ip := network.Int2IpWithLittleEndian(0x04030201)  // → "1.2.3.4"

// Both return 0 / "" for empty, invalid, or IPv6 input.
```

### 地址校验

```go
// IsValidAddr checks if the string is a valid IP address (with or without port)
// or a resolvable hostname.
network.IsValidAddr("192.168.1.1:8080") // true
network.IsValidAddr("192.168.1.1")      // true
network.IsValidAddr("example.com:443")  // true (DNS lookup)
network.IsValidAddr("999.999.999.999")  // false
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/network#pkg-examples)。

---

## auth — Casbin RBAC 鉴权

基于 [Casbin](https://casbin.org/) 的基于角色的访问控制（RBAC）Gin 中间件。
按照**方法二语义**执行策略检查：用户角色列表中的每一个角色都必须
通过策略检查 —— 只要有任何一个角色被拒绝，请求就会以 HTTP 403 Forbidden 被拒绝。

`InitCasbin` 必须在应用启动时调用一次，且要在注册 `Authorize` 之前。
Model 和 policy 既可以从内存字符串（`fs=true`）加载，也可以从
磁盘上的文件（`fs=false`）加载。

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

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/auth#pkg-examples)。

---

## token — JWT 认证

基于 HS256 的 JWT 访问令牌和刷新令牌的创建、解析，以及一个 Gin bearer-token 中间件。
基于 [golang-jwt/jwt](https://github.com/golang-jwt/jwt) 构建。

**`InitToken` 必须在应用启动时调用一次**，且要在任何 token 函数之前。
密钥和 issuer 通常在 `env.InitEnv()` 之后从 `env.Env()` 中读取。
建议通过 [`token/component`](#bootstrap-组件包) 将其接入 `bootstrap` 链路 ——
如果 `jwt.access.secretcode` 为空或仍是占位符值，它的 `Init()` 会快速失败，
而不是让每一个 token 都被一个非密钥字符串静默地签名/校验。

```go
import "github.com/phcp-tech/common-library-golang/token"

// Initialise once at startup (composition root), after env.InitEnv().
token.InitToken(
    env.Env().String("jwt.issuer"),             // e.g. "phcp"
    env.Env().String("jwt.access.secretcode"),  // access token signing key
    env.Env().String("jwt.refresh.secretcode"), // refresh token signing key
)

// Create a short-lived access token (valid for 1 hour).
tok, err := token.CreateToken(userId, username, orgId, productId, roles, time.Hour)

// Validate and parse an access token from an incoming request.
user, err := token.ParseToken(tok)
fmt.Println(user.Username, user.UserId, user.OrgId, user.ProductId, user.Roles)

// Create a long-lived refresh token (no roles embedded).
refresh, err := token.CreateRefreshToken(userId, username, orgId, productId, 24*time.Hour)

// Register as a Gin middleware; stores LoginUser in context under key "userInfo".
r := gin.New()
r.Use(token.Authenticate())
```

如果 `Authorization: Bearer <token>` 请求头缺失、格式错误，或携带了无效/过期的
令牌，`Authenticate` 会以 HTTP 401 中止请求。

`CreateToken`/`ParseToken`/`CreateRefreshToken`/`ParseRefreshToken` 都会拒绝
未初始化或使用占位符密钥的情形 —— 无论是从未调用过 `InitToken`（accessSecret/
refreshSecret 保持零值 `""`），还是调用时传入了 `token.PlaceholderSecret`
（即 `"TOBE_REPLACED"`，本工作区 `config/app.toml`
文件中提供的脚手架占位值）—— 都会返回错误，而不是让每一个 token 都被一个
众所周知的、非密钥的字符串静默地签名/校验（HMAC 接受任意密钥，包括空值
或常量值，因此这种情况本身不会失败）。`token.IsUsableSecret` 是这一检查
对外导出的入口 —— `token/component` 在 bootstrap 的 `Init()` 阶段调用的正是
同一个函数来校验配置，因此这两层防护对"何为已配置"的判定永远不会不一致。

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/token#pkg-examples)。

---

## gin — Gin 引擎工厂

预配置好的 [Gin](https://github.com/gin-gonic/gin) 引擎，内置基于 slog 的结构化
请求日志（`slog-gin`）和可选的 CORS 支持。始终运行在 `ReleaseMode` 下。

`InitGin` 是唯一的入口点。传入一组允许的来源即可启用 CORS；
包含 `*` 的条目会被当作单层通配符模式处理
（例如 `https://*.example.com` 能匹配 `https://api.example.com`，但不能匹配 `https://a.b.example.com`）。

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

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/gin#pkg-examples)。

---

## gin/pprof — 性能分析端点

通过 [gin-contrib/pprof](https://github.com/gin-contrib/pprof) 在 Gin 引擎上注册
Go 运行时 pprof 端点。

**仅在需要性能分析的服务中导入这个子包。**
导入它是一种显式的选择性启用 —— 不会影响只导入了 `gin` 的服务。

挂载了两个路由组：

| 路径 | 用途 |
|------|-------------|
| `/debug/pprof/*` | 标准 Go pprof 路径，用于直接访问 |
| `<path>/admin/pprof/*` | 对 API 网关友好的别名 |

```go
import (
    libgin   "github.com/phcp-tech/common-library-golang/gin"
    ginpprof "github.com/phcp-tech/common-library-golang/gin/pprof"
)

router := libgin.InitGin(nil)
ginpprof.Mount(router, "/api/v1") // mounts /debug/pprof/* and /api/v1/admin/pprof/*
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/gin/pprof#pkg-examples)。

---

## httpclient — Resty HTTP 客户端

基于 [go-resty/resty](https://github.com/go-resty/resty) 构建的功能丰富的 HTTP 客户端，
预配置了重试、JWT bearer-token 支持，以及自动 JSON 处理。
适用于结构化的服务间调用。

**默认情况下，`RetryMax` 只重试传输层的失败**（连接
被拒绝、超时、DNS 失败等）—— 这是 resty 自身的默认重试
条件，它只看往返请求返回的 `error`。一个格式良好的 HTTP 500/503/429
响应并不是一个 Go error，因此仅靠 `RetryMax`
永远不会重试这些情况。将 `RetryOnServerErrors` 设为 `true`，即可对 429 和 5xx 响应
也进行重试。

```go
import "github.com/phcp-tech/common-library-golang/httpclient"

// Default settings: Timeout=10s, RetryMax=3 (transport-level failures only).
cli := httpclient.NewHttpClient()

// Custom settings — zero-value fields fall back to defaults.
cli := httpclient.NewHttpClient(httpclient.Config{
    Timeout:             15 * time.Second,
    RetryMax:            5,
    InsecureSkipVerify:  true, // only for internal self-signed certs
    RetryOnServerErrors: true, // also retry on HTTP 429/5xx responses
})

resp, err := cli.Get(url, jwtToken, nil)
resp, err := cli.Post(url, jwtToken, body)
resp, err := cli.Put(url, jwtToken, body)
resp, err := cli.Delete(url, jwtToken, body)

// Access the underlying resty.Client for advanced use.
restyClient := cli.Client()
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpclient#pkg-examples)。

---

## httpclient/retryable — 可重试的 HTTP 客户端

封装 [hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp)，
并将其以标准的 `*http.Client` 形式暴露出来，支持可配置的重试和超时。
当现有代码已经在直接使用 `net/http`，而你需要在不切换到 resty 的情况下
加入重试行为时，使用这个子包。

与上文的 `httpclient` 不同，这个包的默认重试策略除了传输层失败之外，
已经会对 HTTP 429 和大多数 5xx 响应进行重试 ——
这里不需要 `RetryOnServerErrors` 的等价项。

**只在需要时才导入这个子包** —— 它会独立于父级 `httpclient` 包
引入 hashicorp 的库。

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

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpclient/retryable#pkg-examples)。

---

## httpserver — 生产级 HTTP/HTTPS 服务器

基于 `net/http` 构建、带有安全默认配置的生产级 HTTP/HTTPS 服务器。
它直接封装 `http.Server`（而非 `gin.Run()`），以提供：

- **可配置的超时** —— `ReadTimeout`、`WriteTimeout`、`IdleTimeout`、`ReadHeaderTimeout`。
  零值的 `WriteTimeout` 会回退到默认的 60 秒，而不是"无限制" —— 如需无限制（例如大文件下载），请传入 `httpserver.NoWriteTimeout`。
- **TLS 1.2+ 及强加密套件** —— 当设置了 `CrtFile` 和 `KeyFile` 时启用
- **优雅关闭** —— `Shutdown(ctx)` 会在停止之前排空正在处理中的请求

`IRunner` 接口统一了 HTTP 和 Lambda 两种模式，使组合根可以自行
选择后端，而应用其余部分的代码保持不变。

```go
import "github.com/phcp-tech/common-library-golang/httpserver"

// Plain HTTP with package defaults.
runner := httpserver.NewHttpServer(httpserver.Config{Port: "8080"})

// HTTPS with custom timeouts.
runner := httpserver.NewHttpServer(httpserver.Config{
    Port:         "8443",
    CrtFile:      "/etc/ssl/server.crt",
    KeyFile:      "/etc/ssl/server.key",
    ReadTimeout:  15 * time.Second,
    WriteTimeout: httpserver.NoWriteTimeout, // unlimited (required for file downloads) — a literal 0 falls back to the 60s default instead
})

// Start in a goroutine; block until OS signal, then shut down gracefully.
go func() { _ = runner.Start(ginRouter) }()
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = runner.Shutdown(ctx)
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpserver#pkg-examples)。

---

## httpserver/lambda — AWS Lambda 适配器

通过 [aws-lambda-go-api-proxy](https://github.com/awslabs/aws-lambda-go-api-proxy)
实现的 `httpserver.IRunner`，适配 AWS Lambda。
它将 `APIGatewayProxyRequest` 事件桥接到任意 `http.Handler`（包括
`*gin.Engine`），无需修改 handler。

**仅在 Lambda 部署场景中导入这个子包** —— 它会引入 AWS
Lambda SDK。以普通 HTTP 服务器形式运行的服务不受影响。

```go
import (
    "github.com/phcp-tech/common-library-golang/httpserver"
    lambdarunner "github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

// Composition root selects the runner based on the deployment environment.
var runner httpserver.IRunner
if isLambda {
    runner = lambdarunner.NewHttpServer()
} else {
    runner = httpserver.NewHttpServer(httpserver.Config{Port: port})
}
_ = runner.Start(ginRouter) // same call regardless of mode
```

参见[完整示例](https://pkg.go.dev/github.com/phcp-tech/common-library-golang/httpserver/lambda#pkg-examples)。
