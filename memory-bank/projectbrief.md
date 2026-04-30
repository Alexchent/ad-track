# Project Brief

## Project Name

ad-track

## Purpose

`ad-track` is a Go-based advertising attribution callback service. Its current implementation focuses on integrating with the vivo Marketing Open API. The service receives ad click tracking data, stores click metadata by device identifier in Redis, and later matches activation/reporting requests against stored click data to upload conversion behavior back to vivo.

## Core Goals

- Receive and persist click tracking payloads from ad channels.
- Match attribution/report requests by OAID or IMEI.
- Upload activation conversion events to vivo for matched vivo traffic.
- Manage vivo OAuth tokens per advertiser in Redis.
- Provide operational endpoints for health checks and Prometheus metrics.
- Keep the service simple to deploy as a Gin HTTP server with YAML configuration.

## Primary Users

- Backend developers maintaining ad attribution integrations.
- Operations engineers deploying and monitoring the service.
- Marketing/ad-tech teams needing vivo conversion callback support.

## In Scope

- Gin HTTP API service.
- vivo authorization-code token exchange.
- vivo advertiser token persistence in Redis.
- Click data persistence in Redis with a 30-day TTL.
- Activation attribution callback to vivo.
- Structured request logging and basic observability endpoints.

## Out of Scope / Not Currently Implemented

- Full multi-channel attribution beyond vivo-specific behavior callback.
- Database persistence despite the presence of `MySQL.DSN` in config.
- Enabled request metrics middleware on all routes; `/metrics` is exposed, but `middleware.PrometheusMetrics()` is not currently registered.
- Automatic vivo token refresh; refresh-related code exists but is commented in `GetToken`.
- Complete production deployment manifests.

## Source Of Truth

- `README.md` contains the user-facing project documentation.
- `main.go`, `router.go`, `handler/`, `logic/`, `pkg/vivo/`, and `svc/` define the actual runtime behavior.