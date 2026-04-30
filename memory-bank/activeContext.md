# Active Context

## Current Focus

The current task is documentation and project context recovery. A comprehensive `README.md` has been generated, and the Memory Bank is being initialized so future sessions can understand the project without re-exploring the codebase from scratch.

## Recent Changes

- Expanded `README.md` from a minimal title into a full project guide.
- Created `memory-bank/projectbrief.md`.
- Created `memory-bank/productContext.md`.
- Created `memory-bank/systemPatterns.md`.
- Created `memory-bank/techContext.md`.

## Current Project Understanding

`ad-track` is a Go + Gin service for vivo ad attribution callbacks. It stores incoming click data in Redis by OAID/IMEI and later uses report requests to upload vivo activation events after matching stored click data.

Core runtime flow:

```text
/vivo/click -> Redis click data
/vivo/auth  -> Redis vivo advertiser token
/report     -> Redis click lookup -> vivo behavior upload
```

## Active Decisions

- `README.md` is the primary user-facing project documentation.
- `memory-bank/` is the long-term context store for future Codee sessions.
- Memory Bank files are written in English for concise technical continuity.
- The documentation should reflect the actual code behavior, including known gaps, not only ideal behavior.

## Important Patterns To Preserve

- Handler factories accept `*svc.ServiceContext` and return `gin.HandlerFunc`.
- Redis key format for clicks is `click:<device_id>`.
- Redis key format for vivo tokens is `vivo_token_<clientId>_<advertiserId>`.
- Attribution behavior is currently vivo-focused and reports `ACTIVATION`.
- Device identifier type is inferred by field and 32-character MD5 length.
- `VIVO.APP` maps package names to vivo `srcId`.

## Known Risks / Follow-Up Items

- `pkg/vivo.GetToken` returns placeholder `"xxxxxsss"` when Redis token lookup fails, which may cause confusing vivo callback failures.
- Automatic token refresh is implemented but commented out in `GetToken`.
- `middleware.PrometheusMetrics()` is not registered despite metric definitions.
- `middleware.CORS` and rate-limit middleware are not registered.
- `logic.NewClick` creates a fresh Redis client instead of reusing `ServiceContext`.
- No `conf.yaml` example file exists in the repository aside from the README snippet.
- Integration tests for Redis and vivo callback flows are not present.

## Last Validation

After generating `README.md`, a Python check confirmed required sections were present:

```text
README.md lines: 401
Missing sections: none
```

## Next Steps

- Finish creating Memory Bank core files.
- Verify all required Memory Bank files exist.
- Future implementation work should update `activeContext.md` and `progress.md` when changes alter project behavior.