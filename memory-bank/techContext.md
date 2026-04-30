# Tech Context

## Runtime

- Language: Go
- Go version from `go.mod`: `1.25.8`
- Service type: HTTP API server
- Default config file: `conf.yaml`
- Config flag: `-f`

## Main Dependencies

Direct dependencies from `go.mod`:

- `github.com/gin-gonic/gin`: HTTP routing and handlers.
- `github.com/gin-contrib/requestid`: request ID middleware.
- `github.com/go-redis/redis/v8`: Redis client.
- `github.com/prometheus/client_golang`: Prometheus metrics and `/metrics` handler.
- `github.com/zeromicro/go-zero`: YAML config loading through `conf.MustLoad`.
- `github.com/fvbock/endless`: graceful restart/shutdown HTTP server.
- `go.uber.org/zap`: used in some logging field calls.
- `gopkg.in/natefinch/lumberjack.v2`: log file rotation.

## External Services

### Redis

Required for:

- Click data storage.
- vivo advertiser token storage.

Click data key format:

```text
click:<device_id>
```

vivo token key format:

```text
vivo_token_<clientId>_<advertiserId>
```

### vivo Marketing API

Default host:

```text
https://marketing-api.vivo.com.cn
```

Used endpoints:

- OAuth token exchange.
- Refresh token endpoint, implemented but not active in runtime flow.
- Advertiser account query.
- Behavior upload callback.
- Summary query helper.

## Configuration Shape

The runtime config is represented by `config.Config`:

```yaml
Port: ":8080"
Env: "dev"
CachePrefix: "ad-track"

Redis:
  Addr: "127.0.0.1:6379"
  Password: ""
  Db: 0

MySQL:
  DSN: ""

Log:
  Filename: "./logs/ad-track.log"
  Encoding: "json"
  Level: "info"
  MaxSize: 100
  MaxAge: 7
  Compress: true

VIVO:
  Host: "https://marketing-api.vivo.com.cn"
  ClientId: "your-client-id"
  ClientSecret: "your-client-secret"
  APP:
    "com.example.app": "your-vivo-src-id"
```

## Development Commands

Install/update dependencies:

```bash
go mod tidy
```

Run locally:

```bash
go run . -f conf.yaml
```

Build binary:

```bash
go build -o ad-track .
```

Run tests:

```bash
go test ./...
```

Run built binary:

```bash
./ad-track -f conf.yaml
```

## Operational Commands

Graceful restart:

```bash
kill -HUP <pid>
```

Graceful shutdown:

```bash
kill -TERM <pid>
```

Health check:

```bash
curl http://127.0.0.1:8080/health
```

Metrics:

```bash
curl http://127.0.0.1:8080/metrics
```

## Logging

- Logger setup: `pkg/logger/logger.go`.
- Uses `slog` globally.
- Outputs to both log file and stdout.
- Supports JSON and console text handlers.
- Uses `lumberjack` for file rotation.
- Request trace ID key: `traceID`.
- Optional user ID context key: `user_id`.

## Technical Constraints

- `conf.yaml` is required at runtime unless another file is passed with `-f`.
- Redis must be reachable before click/token flows can work.
- vivo callback requires external network access to vivo Marketing API.
- Token refresh is not currently active in `GetToken`.
- Metrics middleware exists but is not wired into router setup.
- `go.sum` should remain synchronized with `go.mod`.

## Testing Notes

- Existing test file: `pkg/vivo/vivo_summary_test.go`.
- No full integration test suite currently documents Redis + vivo behavior.
- Tests involving vivo API should avoid real credentials unless explicitly configured as integration tests.