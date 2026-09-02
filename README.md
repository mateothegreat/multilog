# multilog

**Fan out log calls to multiple destinations simultaneously**: console, Elasticsearch, or your own custom logger, with per-logger level filtering, regex drop-filters, and structured data payloads.

Features:

- 🔀 **Multi-output fan-out**: one call logs to every registered destination.
- 🎚️ **Per-logger level filtering**: console at `DEBUG`, Elasticsearch at `WARN`, etc.
- ✂️ **Regex drop-filters**: silently drop messages by group or message pattern.
- 📦 **Structured data**: every call carries a `map[string]interface{}` payload.
- 🖥️ **Pretty console**: colorized text or JSON output, `TRACE` support, slog-compatible.
- 🔎 **Elasticsearch shipper**: indexes created on-the-fly with a default mapping.

![multilog console output](<docs/CleanShot 2024-07-04 at 19.28.48.png>)

![multilog documents in Kibana](<docs/CleanShot 2024-07-05 at 16.55.52.png>)

## Installation

```bash
go get -u github.com/mateothegreat/multilog
```

The Elasticsearch logger is a separate Go module; install it only if you need it:

```bash
go get -u github.com/mateothegreat/multilog/logger/elasticsearch
```

Requires Go 1.25+.

## Quick Start

```go
package main

import (
	"github.com/mateothegreat/multilog"
)

func init() {
	multilog.RegisterLogger(multilog.LogMethod("console"), multilog.NewConsoleLogger(&multilog.NewConsoleLoggerArgs{
		Level:  multilog.TRACE,
		Format: multilog.FormatText,
	}))
}

func main() {
	multilog.Info("my_app", "starting up", map[string]interface{}{
		"port": 8080,
		"env":  "dev",
	})
	multilog.Error("my_app", "something failed", map[string]interface{}{
		"attempt": 3,
		"err":     "connection refused",
	})
}
```

See it all in action:

```bash
go run ./examples/kitchensink
```

## Usage

### Log levels & methods

Package-level functions write to all registered loggers:

```go
func Trace(group string, message string, v map[string]interface{})
func Debug(group string, message string, v map[string]interface{})
func Info(group string, message string, v map[string]interface{})
func Warn(group string, message string, v map[string]interface{})
func Error(group string, message string, v map[string]interface{})
func Fatal(group string, message string, v map[string]interface{})
```

Levels are ordered: `TRACE(0)` < `DEBUG(1)` < `INFO(2)` < `WARN(3)` < `ERROR(4)` < `FATAL(5)`.

`Fatal` waits for all loggers to finish, sleeps 100ms as a best-effort flush, then calls `os.Exit(1)`.

### Registering loggers

Register a logger under a `LogMethod` name during `init()` (or before your first log call):

```go
err := multilog.RegisterLogger(multilog.LogMethod("console"), multilog.NewConsoleLogger(&multilog.NewConsoleLoggerArgs{
	Level:  multilog.DEBUG,
	Format: multilog.FormatJSON, // or multilog.FormatText
}))
```

`RegisterLogger` returns an error if that method name is already registered, and calls the logger's `Setup()` if it's non-nil. The console logger supports `FormatText` (colorized) and `FormatJSON`.

Alternatively, build a logger and register it manually with `multilog.NewLogger`:

```go
logger := multilog.NewLogger(multilog.LogMethod("my_logger"))
logger.Setup = func() { /* connect, warm up, etc. */ }
logger.Log = func(level multilog.LogLevel, group string, message string, v map[string]interface{}) {
	// ship it
}
```

### Drop filters

Each logger accepts `FilterDropPatterns []*string`: regex patterns matched against the **group or the message**. A match drops the log line for that logger only. `multilog.PtrString` is a helper for `*string` literals:

```go
multilog.NewConsoleLogger(&multilog.NewConsoleLoggerArgs{
	Format: multilog.FormatText,
	FilterDropPatterns: []*string{
		multilog.PtrString("block_this_group"),
		multilog.PtrString(".*drop.*"), // drop any message containing "drop"
	},
})
```

### Structured data

The third argument to every log call is a free-form `map[string]interface{}` (pass `nil` if unused). Console renders it inline; Elasticsearch indexes it under the `data` field.

```go
multilog.Warn("api", "slow request", map[string]interface{}{
	"path":     "/v1/users",
	"duration": "1.42s",
	"status":   200,
})
```

### Custom loggers

Anything can be a logger: implement `Setup` and `Log` on a `multilog.CustomLogger`:

```go
multilog.RegisterLogger(multilog.LogMethod("customerLogger1"), &multilog.CustomLogger{
	Log: func(level multilog.LogLevel, group string, message string, v map[string]interface{}) {
		log.Printf("logged via customerLogger1: %s: %s", group, message)
	},
})
```

### Elasticsearch logger

Ships documents shaped `{time, level, group, message, data}` to an index, creating the index on-the-fly if it's missing. `Config` is an alias for the [go-elasticsearch v8 Config](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/connecting.html); `DefaultMapping` provides a sane mapping (pass `""` to create the index without one).

```go
import (
	"crypto/tls"
	"net/http"

	"github.com/mateothegreat/multilog"
	elasticsearch "github.com/mateothegreat/multilog/logger/elasticsearch"
)

multilog.RegisterLogger(multilog.LogMethod("elasticsearch"), elasticsearch.NewElasticsearchLogger(&elasticsearch.NewElasticsearchLoggerArgs{
	Level: multilog.TRACE,
	Config: elasticsearch.Config{
		Addresses: []string{"https://localhost:9200"},
		Username:  "elastic",
		Password:  "elastic",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	},
	Index:   "logs-3",
	Mapping: elasticsearch.DefaultMapping, // or a custom mapping string
	FilterDropPatterns: []*string{
		multilog.PtrString(".*drop.*"),
	},
}))
```

### slog integration

The console logger is built on `log/slog`. Use `multilog.NewSlogLogger()` for a ready-made pretty `*slog.Logger`, or wire up `multilog.NewPrettyHandler` yourself for full control over output and options:

```go
logger := multilog.NewSlogLogger()
logger.Info("hello", "key", "value")
```

## Examples

| Example | Description |
|---|---|
| [examples/kitchensink](./examples/kitchensink) ([GitHub](https://github.com/mateothegreat/multilog/tree/main/examples/kitchensink)) | Console + two custom loggers, all levels |
| [examples/dropfilters](./examples/dropfilters) ([GitHub](https://github.com/mateothegreat/multilog/tree/main/examples/dropfilters)) | Regex drop filters + `Fatal` |
| [examples/elasticsearch](./examples/elasticsearch) ([GitHub](https://github.com/mateothegreat/multilog/tree/main/examples/elasticsearch)) | Console + Elasticsearch with `DefaultMapping` |

## Concurrency semantics

Logging is **concurrent but synchronous**: each call to `Trace`/`Debug`/`Info`/`Warn`/`Error`/`Fatal` fans out to all registered loggers in separate goroutines, then blocks until every logger has finished. It is safe to call from multiple goroutines, but a log call does not return until all loggers complete.

`Fatal()` waits for all loggers to complete, then sleeps 100ms as a best-effort flush for buffered backends before calling `os.Exit(1)`. This is **not** a delivery guarantee.

## Benchmarks

Benchmarks live in [`log_bench_test.go`](./log_bench_test.go) and run with:

```bash
go test -bench=. -benchmem -run=^$ .
```

Results from an actual run on Apple Silicon, Go 1.25, with `-benchtime=100x` (numbers vary by machine):

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| Info, no loggers | 77 | 16 | 1 |
| Info, 1 logger | 939 | 116 | 4 |
| Info, 2 loggers | 1305 | 217 | 6 |
| Info, 4 loggers | 2796 | 443 | 10 |
| Info, 8 loggers | 4358 | 837 | 18 |
| Info, concurrent (4 loggers, parallel goroutines) | 1550 | 574 | 10 |
| SnapshotLoggers (registry read + copy) | 130 | 64 | 1 |
| Info with structured data payload | 1783 | 784 | 8 |

### What this means in practice

- Every log call spins up one small background task (goroutine) per registered logger, then waits for all to finish. That costs roughly 55 bytes of memory and 2 allocations per extra logger. Cheap: even with 8 loggers a call completes in about 4 microseconds.
- Calling `Info` with zero loggers costs almost nothing (77 nanoseconds). It is safe to leave log calls in code paths that might have no loggers registered.
- Registry lookup (snapshot) is 130 nanoseconds and a single allocation. Reading the registered loggers is not a bottleneck.
- The structured data map you pass in is the biggest cost (about 784 bytes vs 116 bytes for an empty call) because Go boxes values into the map. Reuse payloads or keep them small on very hot paths.
- Concurrent logging from many goroutines works correctly and stays fast (about 1.5 microseconds per call) thanks to a read lock that only blocks writers (registration), not other loggers.
- Bottom line: log freely. The library's overhead is measured in microseconds and bytes, not milliseconds and megabytes. Your actual logger I/O (console printing, Elasticsearch network calls) will dominate total time.

Numbers come from a development laptop, so treat them as order-of-magnitude guidance; rerun on your own hardware with the command above.

## Testing

```bash
go test ./...
```

## Contributing

Issues and pull requests are welcome at [github.com/mateothegreat/multilog](https://github.com/mateothegreat/multilog). Keep changes minimal, match existing style, and make sure `go build ./...` and `go test ./...` pass.

## License

MIT. See [LICENSE](./LICENSE).
