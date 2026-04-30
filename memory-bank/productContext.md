# Product Context

## Why This Project Exists

Advertising platforms need conversion feedback to optimize campaign delivery and measure performance. `ad-track` bridges click tracking data and downstream activation/reporting events so vivo ad traffic can be attributed and reported back to vivo's marketing API.

## Problems It Solves

- Click tracking payloads arrive before user activation data and must be stored temporarily.
- Business attribution requests may only contain device identifiers, so the service needs to match OAID or IMEI to previously stored click data.
- vivo conversion upload requires platform-specific fields such as advertiser ID, package name, source ID, click ID, user ID type, and access token.
- vivo OAuth tokens are advertiser-scoped and need consistent storage and lookup.
- Operators need health and metrics endpoints for basic service monitoring.

## User-Facing Behavior

### Click Ingestion

A channel sends click tracking data to:

```http
POST /vivo/click?channel=vivo
```

The service stores each payload by `oaid` and/or `imei` in Redis with a 30-day TTL.

### Authorization

A vivo authorization code is submitted to:

```http
GET /vivo/auth?code=<authorization_code>
```

The service exchanges the code for token data, queries the advertiser UUID, and stores token JSON in Redis under:

```text
vivo_token_<clientId>_<advertiserId>
```

### Attribution Callback

The application/backend reports attribution data to:

```http
GET /report?oaid=<oaid>&imei=<imei>&user_id=<user_id>&package_name=<package_name>
```

The service finds prior click data, checks whether the channel is vivo, enriches the data with app user ID and package name, and uploads an `ACTIVATION` behavior event to vivo.

## Expected Experience

- API behavior should be predictable and easy to integrate with.
- Deployment should require only a YAML config file, Redis, and Go runtime/binary.
- Logs should contain trace IDs and request metadata for debugging.
- The service should fail clearly when required data is missing, such as empty device IDs, missing advertiser IDs, missing vivo tokens, or missing package-to-srcId mapping.

## Important Product Constraints

- `/report` requires at least one device identifier: `oaid` or `imei`.
- vivo attribution requires click data containing `advertiserId`.
- `package_name` in `/report` must match a key in `VIVO.APP`.
- vivo token must be authorized and stored before callbacks can succeed.
- Current production readiness depends on manually managing token expiry because auto-refresh is not enabled in runtime flow.