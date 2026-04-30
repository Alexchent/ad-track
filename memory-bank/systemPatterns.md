# System Patterns

## Architecture Overview

`ad-track` is a single-process HTTP service built with Gin. Runtime dependencies are initialized in `main.go` and `svc/servicecontext.go`, then injected into handlers through `ServiceContext`.

```text
HTTP request
  -> Gin router
  -> middleware.RequestLogger
  -> handler
  -> logic layer
  -> Redis / vivo API client
```

## Key Components

### Entry Point

- `main.go`
  - Loads YAML config via `go-zero/core/conf`.
  - Initializes global `slog` logger through `pkg/logger`.
  - Creates `svc.ServiceContext`.
  - Creates a Gin engine.
  - Registers request ID, request logging, recovery middleware, and routes.
  - Starts an `endless` HTTP server for graceful restart/shutdown.

### Routing

- `router.go`
  - `GET /health`: health check.
  - `GET /metrics`: Prometheus endpoint.
  - `GET /vivo/auth`: vivo authorization-code token exchange.
  - `POST /vivo/click`: vivo click ingestion.
  - `GET /report`: attribution report and vivo activation callback.

### Service Context

- `svc/servicecontext.go`
  - Holds loaded `config.Config`.
  - Creates Redis client for vivo token storage.
  - Creates `pkg/vivo.AdService`.

### Handler Layer

- `handler/vivo.go`
  - `GetAuthorizationCode`: exchanges vivo auth code and stores advertiser token.
  - `ProcessVIVOClick`: receives click tracking arrays and delegates storage.
- `handler/click.go`
  - Saves click payloads by OAID and IMEI when present.
- `handler/attribute.go`
  - Finds click data by OAID/IMEI and triggers channel-specific attribution callback.

### Logic Layer

- `logic/click.go`
  - Persists click data as Redis Hash values.
  - Converts arbitrary JSON values to Redis-safe string values.
  - Uses `click:<device_id>` key format.
  - Applies a fixed 30-day TTL.
- `logic/vivo.go`
  - Implements vivo attribution callback behavior.
  - Converts device identifiers to vivo `userIdType`.
  - Resolves package name to `srcId` from `VIVO.APP`.
  - Reads advertiser token and uploads activation behavior.

### vivo API Client

- `pkg/vivo/*`
  - Encapsulates vivo Marketing API endpoint construction.
  - Exchanges authorization codes for token data.
  - Queries advertiser identity.
  - Stores and retrieves advertiser-scoped tokens from Redis.
  - Uploads behavior conversion data.
  - Supports summary query helper APIs.

## Important Data Flows

### Click Ingestion Flow

```text
POST /vivo/click
  -> bind []map[string]interface{}
  -> attach channel query value to each item
  -> SaveData by oaid if present
  -> SaveData by imei if present
  -> Redis HSET click:<device_id>
  -> Redis EXPIRE 30 days
```

### Attribution Flow

```text
GET /report
  -> validate oaid or imei
  -> Redis HGETALL click:<oaid>
  -> if not found, Redis HGETALL click:<imei>
  -> check channel contains "vivo"
  -> add app_uid from user_id
  -> add pkgName from package_name
  -> create VivoApi
  -> upload ACTIVATION event
```

### vivo Token Flow

```text
GET /vivo/auth?code=...
  -> GetAccessToken(code)
  -> queryAdvertiser(access_token)
  -> save token under vivo_token_<clientId>_<advertiserId>
```

## Design Patterns

- Dependency context pattern: `ServiceContext` carries app config and shared clients.
- Handler factory pattern: handlers are functions that accept `*svc.ServiceContext` and return `gin.HandlerFunc`.
- Interface-based attribution: `logic.Attribute` defines `Active(data map[string]string) error`, allowing channel-specific implementations.
- Cache-as-join pattern: Redis stores click data temporarily so later activation reports can be joined by device ID.
- Platform client package: `pkg/vivo` isolates external vivo API request/response details.
- Structured logging middleware: request metadata and trace ID are injected into request context.

## Critical Implementation Details

- `handler.AttributeReport` only invokes vivo callback when stored click `channel` contains `vivo` case-insensitively.
- `logic.VivoApi.callbackVivoBehavior` requires:
  - non-empty OAID or IMEI
  - non-empty `advertiserId`
  - configured `VIVO.APP[pkgName]`
  - non-empty stored access token
- Click ID lookup accepts both `clickId` and `ClickId`.
- Device ID type is inferred by length:
  - 32-char OAID -> `OAID_MD5`
  - other OAID -> `OAID`
  - 32-char IMEI -> `IMEI_MD5`
  - other IMEI -> `IMEI`
- `pkg/vivo.GetToken` currently returns placeholder `"xxxxxsss"` if Redis lookup fails after retries, which can mask missing token state and should be treated carefully.

## Observability Patterns

- `/health` returns plain `ok`.
- `/metrics` exposes Prometheus handler.
- Metrics middleware is implemented but not registered in `main.go` or `router.go`.
- Request logging uses request ID as `traceId`.

## Known Architectural Gaps

- Redis clients are created in both `svc.NewServiceContext` and `logic.NewClick`, so click logic creates a new client instead of reusing the service context client.
- MySQL config exists but has no runtime usage.
- Rate limit and CORS middleware files exist but are not currently registered.
- Automatic token refresh code exists but is not active.