---
outline: [2, 3]
---

# Changelog

Release notes for `github.com/drossan/go_logs`. Versions follow [Semantic Versioning](https://semver.org/); the Slack module `github.com/drossan/go_logs/slack/v3` is tagged separately as `slack/vX.Y.Z`.

## v3.1.0 (unreleased)

First release that can actually be installed as v3: `go get github.com/drossan/go_logs/v3@v3.1.0`.

::: warning v3.0.x was never installable under /v3
The `v3.0.0`–`v3.0.4` tags declare `module github.com/drossan/go_logs` (no `/v3` suffix) in their `go.mod`, so the Go proxy cannot serve them as `github.com/drossan/go_logs/v3`. They stay published but invalid for `/v3`; use `v3.1.0` or later.
:::

### 🚨 Breaking changes

- The module path is now `github.com/drossan/go_logs/v3`. `async`, `hooks`, `http`, `otel` and `signal` are packages of that single module (`github.com/drossan/go_logs/v3/async`, …) instead of separate modules.
- Slack is out of the core: the notifier lives in the separate module `github.com/drossan/go_logs/slack/v3` (`slack.NewNotifier`, `slack.NewNotifierFromEnv`) and the `adapters` package is gone. The v2 API sends notifications only after `go_logs.SetNotifier(n)`; see the [migration guide](/migration-v2-to-v3) and [Configuration](/configuration).

### 🚀 Enhancements

- `go_logs.SetNotifier(domain.Notifier)` registers the notifier used by the v2 API (`nil` disables it).
- `go_logs.Flusher` interface: writers are flushed after every entry. `RotatingFileWriter`, `EnhancedRotatingFileWriter`, `MultiWriter` and `SamplingWriter` implement it.
- `Logger.Sync()` now flushes and syncs the output, ignoring the `EINVAL`/`ENOTTY`/`EBADF` errors returned by terminals and pipes.
- The log level is shared by a logger and all its children: `SetLevel` on any of them affects the whole tree.

### 🩹 Fixes

- `Init()` (and the v2 auto-initialization) never terminates the process: empty variables fall back to their defaults, invalid values fall back and print a warning to stderr, and an unopenable log file disables file output with a warning. With no environment the default level is `info`, so `InfoLog("x")` prints to the console.
- `With()` no longer shares field memory between sibling loggers (a data race that could log another request's fields).
- No `fsync` per log entry when writing to files (it limited file logging to a few hundred messages per second).
- `async`: `Sync()` on a child waits for its entries, `Close()` is idempotent on the root and a no-op on children, and `Fatal` on a child includes its fields.
- `GetCaller` reports the package name without the `/vN` suffix.
- The `hooks` package builds on Windows (the syslog hook is excluded there).

### 📖 Documentation

- This documentation site (English and Spanish), with corrected code samples.
- `LICENSE` (MIT) added.

### 🤖 CI

- Workflow running `gofmt`, `go vet`, `go test -race` (root module and `slack/`) and a Windows cross-build on every push and pull request.

## v3.0.0 – v3.0.4 (2026-03-02)

`v3.0.1`–`v3.0.4` only changed the release workflow. `v3.0.0` introduced the v3 API; the features below were previously listed as "v3.1"–"v3.5" but all shipped in `v3.0.0`.

### 🚀 Enhancements

- `Logger` interface with typed structured fields, child loggers (`With()`) and context propagation (`trace_id`, `span_id`).
- `TextFormatter` and `JSONFormatter`, extensible hooks, sensitive-data redaction and a dependency-free `RotatingFileWriter`.
- Caller info and stack traces (`WithCaller`, `WithStackTrace`, `WithCallerLevel`).
- `MultiWriter`, time-based rotation (daily/hourly), gzip compression and `MaxAge` retention (`EnhancedRotatingFileWriter`).
- Sampling (`SamplingWriter`), global logger (`SetDefault`), testing helpers (`CaptureBuffer`, `MockLogger`).
- OpenTelemetry exporter (`otel`) and syslog hooks, local and remote.
- Metrics, async logging (`async`), dynamic level over HTTP (`http`) and SIGHUP rotation (`signal`).
- The v2 global API keeps working.

## v1.0.0 – v1.2.5 (2023-12-30 – 2024-05-14)

The legacy global API (`InfoLog`, `ErrorLog`, …), referred to as "v2" in this documentation.

- `v1.0.0`: initial release with basic logging.
- `v1.2.0`: centralized logging and Slack notifications.
- `v1.2.1`–`v1.2.5`: fixes to notification settings and log file management.
