# Golang Structure

Go REST API project with **custom distributed tracing** and **structured JSON logging** — designed as an alternative to OpenTelemetry that gives full control over log format, field masking, and cross-service correlation.

Integrated with **Grafana + Loki + Tempo + Alloy + Prometheus** observability stack.

## Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Register Use Case Flow](#register-use-case-flow)
- [Custom Distributed Tracing](#custom-distributed-tracing)
- [Structured JSON Log Format](#structured-json-log-format)
- [Application Logging (zap)](#application-logging-zap)
- [Data Masking](#data-masking)
- [Prometheus Metrics](#prometheus-metrics)
- [Database Migration](#database-migration)
- [Entity code generation (genentity)](#entity-code-generation-genentity)
- [Observability Stack](#observability-stack)
- [Monitoring Guide](#monitoring-guide)
- [Environment Variables](#environment-variables)
- [Makefile Commands](#makefile-commands)
- [Testing](#testing)

## Architecture

```
                           ┌────────────────┐
                           │    Grafana     │ :3001
                           │   Dashboards   │
                           └───────┬────────┘
                                   │ query
                   ┌───────────────┼────────────────┐
                   ▼               ▼                ▼
            ┌──────────┐    ┌────────────┐    ┌──────────┐
            │   Loki   │    │ Prometheus │    │  Tempo   │
            │  :3100   │    │   :9090    │    │  :3200   │
            └────▲─────┘    └─────▲──────┘    └──────────┘
                 │ push           │ scrape /metrics
            ┌────┴─────┐         │
            │  Alloy   │         │
            │ :12345   │         │
            └────▲─────┘         │
                 │ Docker logs   │
    ┌────────────┼───────────────┼────────────────┐
    ▼            ▼               ▼                ▼
┌────────┐ ┌────────────┐ ┌─────────────────┐ ┌──────────┐
│Gateway │ │Orchestrator│ │golang-structure │ │ Postgres │
│ :3000  │ │  :8081     │ │     :8080       │ │  :5432   │
└───┬────┘ └─────┬──────┘ └───────┬─────────┘ └────▲─────┘
    │            │                │                 │
    │            │                │          ┌──────┴──────┐
    │            │                │          │atlas-migrate│
    │            │                │          │ (schema)    │
    │            │                │          └─────────────┘
    │            │                │
    │ X-Trace-Id │  X-Trace-Id    │
    │ X-Source   │  X-Source      │
    └──►─────────└──►─────────────┘
```

In the checked-in `docker-compose.yaml`, the **golang-structure** service definition is **commented out** by default: use `make run` locally while Compose provides Postgres, Atlas migrations, and observability.

### Data Flow

| Signal | Path |
|--------|------|
| **Logs** | App stdout (JSON) → Alloy (Docker log scrape) → Loki → Grafana |
| **Metrics** | App `/metrics` ← Prometheus (scrape) → Grafana |
| **Traces** | Tempo (OTLP receiver, ready for future instrumentation) → Grafana |
| **Trace Correlation** | `traceId` in Loki logs → derived fields → search across services |

## Project Structure

```
.
├── cmd/                                    # Application entry points
│   ├── golangstructure/main.go             #   Main API service (:8080)
│   ├── gatewaydummie/main.go               #   Demo gateway (:3000)
│   ├── orchestratedummie/main.go           #   Demo orchestrator (:8081)
│   └── genentity/main.go                   #   CLI: Postgres introspection → GORM entity .go files
│
├── internal/golangstructure/               # Application-private code
│   ├── config.go                           #   Environment-based config struct
│   ├── domain/
│   │   ├── entity/                         #   GORM models (TblUsers, TblUserLogs; generated via genentity)
│   │   └── repository/                     #   Repository interfaces & implementations
│   ├── initial/                            #   Dependency injection / bootstrap
│   │   ├── initial.go                      #     App lifecycle (New, Run, Shutdown)
│   │   ├── config.go                       #     DB connection setup
│   │   ├── middleware.go                   #     Middleware registration
│   │   ├── router.go                       #     Route group wiring
│   │   ├── repository.go                   #     Repository construction
│   │   ├── sdk.go                          #     External SDK setup
│   │   └── use_case.go                     #     Use case construction
│   ├── middleware/
│   │   ├── healthz.go                      #   Liveness & readiness probes
│   │   ├── prometheus.go                   #   Prometheus metrics middleware
│   │   ├── logger.go                       #   Structured JSON logging middleware
│   │   └── cors.go                         #   CORS configuration
│   ├── router/
│   │   ├── auth.go                         #   /api/v1/auth routes
│   │   └── user.go                         #   /api/v1/users routes
│   └── usecase/                            #   Feature-based use cases
│       ├── auth/
│       │   ├── use_case.go                 #     Auth use case aggregate
│       │   └── register/                   #     POST /auth/register
│       └── user/
│           ├── use_case.go                 #     User use case aggregate
│           ├── createuser/                 #     POST /users
│           ├── getusers/                   #     GET /users
│           ├── updateuser/                 #     PUT /users/:user_id
│           └── deleteuser/                 #     DELETE /users/:user_id
│
├── pkg/                                    # Reusable shared libraries
│   ├── httpx/                              #   HTTP utilities
│   │   ├── logger.go                       #     Request logging middleware
│   │   ├── http.go                         #     HTTP client with header propagation
│   │   ├── error.go                        #     Error wrapping with stack traces
│   │   ├── model.go                        #     Log, Block, DBLog, AppLog structs
│   │   ├── const.go                        #     Headers, response codes
│   │   └── util.go                         #     DB/App log collectors, masking helpers
│   ├── logx/                               #   Application logger (zap-backed)
│   │   ├── logger.go                       #     Level-filtered logger that pushes to per-request collector
│   │   ├── writer.go                       #     stdout sink for fallback (no request context)
│   │   └── noop.go                         #     Noop logger for tests
│   ├── promx/                              #   Prometheus metrics middleware
│   │   └── promx.go                        #     HTTP metrics + /metrics endpoint
│   ├── db/sqlx/                            #   GORM setup + custom SQL logger
│   │   ├── sql.go                          #     Connection factory, query collector
│   │   ├── model.go                        #     Shared base DB model
│   │   └── const.go                        #     Driver constants
│   ├── db/gormx/                           #   Generic repository abstraction for GORM
│   │   ├── model.go                        #     Condition, relation, pagination query models
│   │   ├── repository.go                   #     BaseRepository (CRUD + transaction)
│   │   └── repository_test.go              #     Transaction and create behavior tests
│   ├── mask/                               #   Field masking engine
│   │   ├── mask.go                         #     JSON/map value masking
│   │   ├── sql.go                          #     SQL INSERT value masking
│   │   └── pattern.go                      #     Pattern parser ("4:2", "4:com")
│   ├── bcryptx/                            #   Password hashing
│   └── validatorx/                         #   Request validation
│
├── deployment/                             # Infrastructure configs
│   ├── golang-structure/Dockerfile         #   Multi-stage Go build
│   ├── atlas/                              #   Database migration (Atlas; used by docker compose atlas-migrate)
│   │   ├── atlas.hcl                       #     Atlas project config
│   │   └── migrations/                     #     SQL migration files
│   │       ├── 20260401000001_create_tbl_users.sql
│   │       ├── 20260503000001_create_tbl_user_logs.sql
│   │       └── atlas.sum
│   ├── liquibase/                          #   Optional Liquibase (local `make liquibase-up`)
│   │   ├── changelogs/                     #     master.yml + per-change YAML
│   │   ├── changesets/golangstructure/...  #     up/ down/ verify SQL per changeset
│   │   └── properties/dev.properties       #     JDBC URL + changelog path for CLI
│   ├── loki/loki-config.yaml               #   Loki log storage config
│   ├── tempo/tempo-config.yaml             #   Tempo trace storage config
│   ├── prometheus/prometheus.yaml          #   Prometheus scrape config
│   ├── alloy/config.alloy                  #   Grafana Alloy log collection pipeline
│   └── grafana/
│       ├── provisioning/
│       │   ├── datasources/datasources.yaml   # Loki + Prometheus + Tempo
│       │   └── dashboards/dashboards.yaml     # Dashboard auto-discovery
│       └── dashboards/
│           ├── golang-structure.json          # Service dashboard (metrics + logs)
│           └── centralized-logs.json          # Centralized log dashboard
│
├── docs/                                   # Documentation
│   └── apis/bruno/                         #   Bruno API collection (request samples)
├── docker-compose.yaml                     # Full stack definition
├── Makefile                                # Dev & deploy commands
├── .env.template                           # Environment variable template
├── go.mod
└── go.sum
```

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.26 |
| HTTP Framework | Fiber v3 |
| ORM | GORM (PostgreSQL, MySQL, SQLite, SQL Server) |
| Primary keys (users / user_logs) | UUID (`github.com/google/uuid`) in generated entities |
| Config | caarlos0/env + godotenv |
| Validation | go-playground/validator v10 |
| Password | bcrypt (golang.org/x/crypto) |
| Application Logging | Uber Zap (go.uber.org/zap) |
| DB Migration (Docker) | Atlas (`atlas-migrate` service) |
| DB Migration (local CLI, optional) | Liquibase + `make liquibase-up` |
| Metrics | Prometheus 3.3 + client_golang |
| Telemetry Collector | Grafana Alloy 1.8 |
| Log Storage | Grafana Loki 3.5 |
| Trace Storage | Grafana Tempo 2.7 |
| Dashboard | Grafana 11.6 |
| Database | PostgreSQL 18 |
| Container | Docker Compose |

## Getting Started

### Prerequisites

- Go 1.26+
- Docker & Docker Compose

### Run stack with Docker Compose

`docker-compose.yaml` currently starts **Postgres**, the **atlas-migrate** one-shot (applies `deployment/atlas/migrations`), and the **observability** services (Loki, Tempo, Prometheus, Alloy, Grafana). The **main API container is commented out** — run the Go service on the host for development.

```bash
make up
```

Use `docker compose logs -f` (or a specific service name) to follow logs.

### Develop the API locally

```bash
cp .env.template .env
# Edit .env: set SQLDB_USER / SQLDB_PASSWORD to match Postgres (e.g. admin / postgres from compose)

make up   # Postgres + migrations + observability

# Optional: apply Liquibase changesets instead of or after Atlas (requires Liquibase CLI installed)
# make liquibase-up

# Run the main API (listens on SERVICE_PORT, default 8080)
make run

# Optional: dummy services for the distributed tracing demo
make orchestrate-run   # terminal 2
make gateway-run       # terminal 3
```

### Access Points

| Service             | URL                    | Credentials   |
|---------------------|------------------------|---------------|
| Main API            | http://localhost:8080  | -             |
| Gateway (demo)      | http://localhost:3000  | -             |
| Orchestrator (demo) | http://localhost:8081  | -             |
| Grafana             | http://localhost:3001  | admin / admin |
| Prometheus          | http://localhost:9090  | -             |
| Loki API            | http://localhost:3100  | -             |
| Tempo API           | http://localhost:3200  | -             |
| Alloy UI            | http://localhost:12345 | -             |

## API Endpoints

| Method   | Path                     | Description                                         |
|----------|--------------------------|-----------------------------------------------------|
| `GET`    | `/livez`                 | Liveness probe (always 200)                         |
| `GET`    | `/readyz`                | Readiness probe (DB ping)                           |
| `POST`   | `/api/v1/auth/register`  | Register user and create `user_log` transactionally |
| `GET`    | `/api/v1/users/`         | List users                                          |
| `POST`   | `/api/v1/users/`         | Create user                                         |
| `PUT`    | `/api/v1/users/:user_id` | Update user (`user_id` = UUID string)               |
| `DELETE` | `/api/v1/users/:user_id` | Delete user (`user_id` = UUID string)                |
| `GET`    | `/metrics`               | Prometheus metrics endpoint                         |

## Register Use Case Flow

`POST /api/v1/auth/register` persists two records in a single transaction and emits structured app logs at every step:

1. Hash password with `bcryptx` (logs `debug: register flow started`)
2. Open transaction via `gormx.BaseRepository.Transaction(...)`
3. Create `tbl_users` row (`id` is a UUID from Postgres defaults) — logs `warn: create user conflict` on duplicate email, `error: create user failed` otherwise
4. Create `tbl_user_logs` row with action `register`, sharing the same `tx` via `WithTx(tx)` (logs `error: create user log failed` on failure)
5. Commit on success and log `info: register completed` (includes `userId` as string). Roll back everything otherwise

The flow lives in `internal/golangstructure/usecase/auth/register/service.go`. App logs are aggregated into the request's `appLogs` array via the injected `logx.Logger` — see [Application Logging](#application-logging-zap).

**Demo chain** (distributed tracing):

| Service              | Endpoint         | Forwards to                           |
|----------------------|------------------|---------------------------------------|
| Gateway `:3000`      | `POST /register` | Orchestrator `:8081`                  |
| Orchestrator `:8081` | `POST /register` | Main API `:8080/api/v1/auth/register` |

## Custom Distributed Tracing

This project implements distributed tracing **without OpenTelemetry**, using custom HTTP headers and structured JSON logs.

### Why Custom?

OpenTelemetry provides powerful tracing but limits control over log format and field-level customization. This custom approach gives:

- Full control over the JSON log structure
- Built-in sensitive data masking in logs
- Embedded DB query logs per request
- Cross-service trace correlation via `X-Trace-Id`
- Service chain visualization via `X-Source` header

### How It Works

#### 1. Trace ID Propagation (`X-Trace-Id`)

```
Gateway                  Orchestrator             golang-structure
   │                          │                         │
   │ generate UUID            │                         │
   │ X-Trace-Id: abc-123      │                         │
   ├─────────────────────────►│                         │
   │                          │ forward X-Trace-Id      │
   │                          ├────────────────────────►│
   │                          │                         │ use same traceId
   │                          │◄────────────────────────┤
   │◄─────────────────────────┤                         │
   │                          │                         │
   │ log: {traceId: abc-123}  │ log: {traceId: abc-123} │ log: {traceId: abc-123}
```

- First service generates a UUID `X-Trace-Id` if not present
- `httpx.Client()` forwards all incoming headers (including `X-Trace-Id`) to downstream services
- All services log with the **same `traceId`**, enabling cross-service log correlation

#### 2. Source Chain (`X-Source`)

```
Gateway log:
{
  "traceId": "abc-123",
  "current": { "service": "gateway-service", ... },
  "source": {                                          ← from Orchestrator
    "service": "orchestrate-service",
    "source": {                                        ← from golang-structure
      "service": "golang-structure-service",
      "dbLogs": [{ "sql": "INSERT INTO ...", ... }]
    }
  }
}
```

- Each service builds a `Block` with its request/response data and attaches it to `X-Source` response header
- Upstream services unmarshal `X-Source` into `logInfo.Source`, creating a **nested chain**
- `X-Internal-Call: true` signals inter-service calls; `X-Source` is stripped from external responses

#### 3. Key Headers

| Header            | Purpose                                                                |
|-------------------|------------------------------------------------------------------------|
| `X-Trace-Id`      | Correlation ID shared across all services in a request chain           |
| `X-Source`        | JSON-serialized `Block` from downstream, nested to form a call chain   |
| `X-Internal-Call` | Set to `"true"` on inter-service calls; controls `X-Source` visibility |

## Structured JSON Log Format

Every HTTP request produces a single JSON log line to stdout:

```json
{
  "traceId": "dafd8434-2c09-11f1-bcbe-6e6384c80408",
  "timestamp": "2026-03-30T14:27:06+07:00",
  "durationMs": "91",
  "current": {
    "service": "golang-structure-service",
    "method": "POST",
    "path": "localhost/api/v1/auth/register",
    "statusCode": "500",
    "code": "E50000",
    "errorMessage": "[httpx] : ERROR: relation \"tbl_users\" does not exist",
    "stackTraces": [
      "[register.handler] (...handler.go:51)",
      "[register.service] (...service.go:50)",
      "[user.repository] (...user.go:37)",
      "ERROR: relation \"tbl_users\" does not exist (SQLSTATE 42P01)"
    ],
    "dbLogs": [
      {
        "source": "repository/user.go:36",
        "sql": "INSERT INTO \"tbl_users\" ... VALUES ('test01',1,'com@******com','$2a$******u2')",
        "rows": 0,
        "durationMs": 21,
        "error": "ERROR: relation \"tbl_users\" does not exist (SQLSTATE 42P01)"
      }
    ],
    "appLogs": [
      {
        "timestamp": "2026-03-30T14:27:06+07:00",
        "level": "debug",
        "caller": "register/service.go:48",
        "message": "register flow started",
        "fields": { "email": "com@******com", "name": "test01" }
      },
      {
        "timestamp": "2026-03-30T14:27:06+07:00",
        "level": "error",
        "caller": "register/service.go:71",
        "message": "create user failed",
        "fields": { "error": "ERROR: relation \"tbl_users\" does not exist (SQLSTATE 42P01)" }
      }
    ],
    "request": {
      "headers": { "Content-Type": "application/json" },
      "body": { "name": "test01", "email": "com@******com", "password": "1234**78" }
    },
    "response": {
      "headers": { "Content-Type": "application/json; charset=utf-8" },
      "body": { "statusCode": 500, "code": "E50000", "data": null }
    }
  },
  "source": { }
}
```

### Log Fields

| Field         | Type  | Description                                 |
|--------------|--------|---------------------------------------------|
| `traceId`    | string | UUID correlating logs across services       |
| `timestamp`  | string | Request start time (RFC3339)                |
| `durationMs` | string | Total request processing time               |
| `current`    | Block  | This service's request data                 |
| `source`     | Block  | Downstream service chain (nested)           |

### Block Fields

| Field          | Type     | Description                               |
|----------------|----------|-------------------------------------------|
| `service`      | string   | Service name                              |
| `method`       | string   | HTTP method                               |
| `path`         | string   | Request path                              |
| `statusCode`   | string   | HTTP status code                          |
| `code`         | string   | Application response code (e.g. `E50000`) |
| `errorMessage` | string   | Root error message                        |
| `stackTraces`  | []string | Error stack trace chain                   |
| `dbLogs`       | []DBLog  | SQL queries executed during this request  |
| `appLogs`      | []AppLog | Application logs emitted via `logx.Logger`|
| `request`      | Body     | Masked request headers + body             |
| `response`     | Body     | Masked response headers + body            |

### DBLog Fields

| Field        | Type   | Description          |
|--------------|--------|----------------------|
| `source`     | string | File and line number |
| `sql`        | string | Masked SQL query     |
| `rows`       | int    | Rows affected        |
| `durationMs` | int    | Query execution time |
| `error`      | string | Query error (if any) |

### AppLog Fields

| Field       | Type           | Description                                                     |
|-------------|----------------|-----------------------------------------------------------------|
| `timestamp` | string         | When the log was emitted (RFC3339)                              |
| `level`     | string         | `debug` / `info` / `warn` / `error`                             |
| `caller`    | string         | Source file and line where the log was emitted                  |
| `message`   | string         | Log message                                                     |
| `fields`    | map[string]any | Structured fields from `zap.Field` (already masked, see below)  |

## Application Logging (zap)

Beyond the per-request log line, the project ships a custom application logger built on **Uber Zap** (`pkg/logx`). It is injected through the SDK and used inside use cases (e.g. `register.service`).

The logger has one defining property: **logs emitted while a request is being processed do not print to stdout — they are appended to `current.appLogs` of that request's JSON log line**, so a single request stays a single JSON entry.

### Behavior

| Where the log is emitted              | What happens                                                                                          |
|---------------------------------------|-------------------------------------------------------------------------------------------------------|
| Inside an HTTP request (with context) | Pushed to the per-request `AppLog` collector, masked with `MASK_PATTERN`, emitted in `current.appLogs`|
| Outside a request (startup, jobs)     | Falls back to JSON output via Zap to stdout                                                           |
| Below the configured `LOG_LEVEL`      | Dropped                                                                                               |

### Configuration

```bash
LOG_LEVEL=debug   # debug | info | warn | error
```

Default is `info`. Levels are inclusive (e.g. `warn` shows `warn` + `error`).

### Logger Interface

```go
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...zap.Field)
    Info(ctx context.Context, msg string, fields ...zap.Field)
    Warn(ctx context.Context, msg string, fields ...zap.Field)
    Error(ctx context.Context, msg string, fields ...zap.Field)
    With(fields ...zap.Field) Logger
    Sync() error
}
```

`logx.Noop()` is provided for unit tests so services can be constructed without a real logger.

### Example — Inject and Use

`internal/golangstructure/usecase/auth/register/service.go`:

```go
type service struct {
    // ...
    logger logx.Logger
}

func (s *service) Service(ctx context.Context, request *Request) error {
    s.logger.Debug(ctx, "register flow started",
        zap.String("email", request.Email),
        zap.String("name", request.Name),
    )
    // ...
    s.logger.Info(ctx, "register completed",
        zap.String("userId", user.ID.String()),
        zap.String("email", request.Email),
    )
    return nil
}
```

These calls show up under `current.appLogs` in the same JSON log entry as the request, with `email` already masked according to `MASK_PATTERN`.

### Wiring

```
.env (LOG_LEVEL)
        │
        ▼
initial.sdk ── logx.NewLogger(level)
        │
        ▼
auth.UseCase ── register.New(... , logger)
        │
        ▼
register.service.logger
```

## Data Masking

Sensitive data is automatically masked in logs using pattern-based rules configured via the `MASK_PATTERN` environment variable.

### Configuration

```bash
MASK_PATTERN={"password":"4:2","email":"4:com"}
```

### Pattern Syntax

| Pattern   | Input              | Output           | Description                                       |
|-----------|--------------------|------------------|---------------------------------------------------|
| `"4:2"`   | `12345678`         | `1234**78`       | Keep 4 prefix chars, 2 suffix chars               |
| `"4:com"` | `user@example.com` | `user@******com` | Keep 4 prefix chars, find "com" marker for suffix |

### What Gets Masked

- **Request/Response body** — JSON fields matching pattern keys
- **Headers** — Header values matching pattern keys
- **SQL queries** — Column values in `INSERT ... VALUES(...)` statements matching pattern keys
- **App logs** — `appLogs[].fields` keys matching pattern keys (recursive across nested maps/slices)

## Prometheus Metrics

The `pkg/promx` library provides a reusable Fiber middleware that exposes Prometheus metrics.

### Usage

```go
// One line: registers /metrics endpoint + records HTTP metrics
app.Use(middleware.Prometheus(app))
```

### Exposed Metrics

| Metric                          | Type      | Labels                                     | Description                        |
|---------------------------------|-----------|--------------------------------------------|------------------------------------|
| `http_requests_total`           | Counter   | `method`, `path`, `status_code`, `service` | Total HTTP request count           |
| `http_request_duration_seconds` | Histogram | `method`, `path`, `status_code`, `service` | Request latency (buckets: 5ms–10s) |
| `http_requests_in_flight`       | Gauge     | `service`                                  | Currently processing requests      |

### Go Runtime Metrics (auto-exposed)

The `prometheus/client_golang` library automatically exposes Go runtime metrics at `/metrics`:
 
| Metric                          | Description           |
|---------------------------------|-----------------------|
| `go_goroutines`                 | Number of goroutines  |
| `go_threads`                    | Number of OS threads  |
| `go_memstats_heap_alloc_bytes`  | Heap memory allocated |
| `go_memstats_stack_inuse_bytes` | Stack memory in use   |
| `go_gc_duration_seconds`        | GC pause duration     |

## Database Migration

The repo supports **two** migration paths:

| Path | When | Location |
|------|------|----------|
| **Atlas** | `docker compose up` runs the `atlas-migrate` service before apps would start | `deployment/atlas/migrations/` |
| **Liquibase** | Local CLI: `make liquibase-up` (requires [Liquibase](https://www.liquibase.org/) installed) | `deployment/liquibase/` |

Keep the database schema consistent with `internal/golangstructure/domain/entity/*.go` (regenerate with [`genentity`](#entity-code-generation-genentity) after schema changes). The Liquibase `up` SQL in this repo uses **UUID** primary keys for `tbl_users` and `tbl_user_logs`, matching the generated GORM models. The Atlas SQL under `deployment/atlas/migrations` is what Compose applies automatically — if you change PK types or columns, update **both** Atlas and Liquibase (or only the path you use) so they stay aligned.

### Atlas (Docker Compose)

```
deployment/atlas/
├── atlas.hcl                                    # Atlas project config
└── migrations/
    ├── 20260401000001_create_tbl_users.sql      # Initial users schema
    ├── 20260503000001_create_tbl_user_logs.sql  # User log schema (register flow)
    └── atlas.sum                                # Checksum integrity file
```

1. `docker compose up` starts **postgres** (with health check)
2. **atlas-migrate** runs `atlas migrate apply` against the database
3. Observability services start; run **`make run`** on the host to attach the API to the same DB

### Adding new Atlas migrations

```bash
# Create a new migration file (with Atlas installed locally)
atlas migrate new <migration_name> --dir "file://deployment/atlas/migrations"

# Write your SQL in the generated file, then update the checksum
atlas migrate hash --dir "file://deployment/atlas/migrations"
```

### Liquibase (optional, local)

```bash
# From repo root, after Postgres is up and .env / properties match your DB
make liquibase-up
```

Uses `deployment/liquibase/properties/dev.properties` and `changelogs/master.yml`. Each changeset can include **`verify/`** SQL files for post-deploy checks.

```
deployment/liquibase/
├── changelogs/
│   ├── master.yml
│   └── golangstructure/20260401000001.yaml
├── changesets/golangstructure/20260401000001/
│   ├── up/           # apply order
│   ├── down/         # rollback stubs
│   └── verify/       # verification scripts
└── properties/
    └── dev.properties
```

## Entity code generation (genentity)

`cmd/genentity` connects to **Postgres** using `SQLDB_*` from `.env` (loads `.env` from the current directory or repo root), introspects a schema (default `public`), and writes one `.go` file per table into the output directory.

```bash
# Makefile default output directory (package name = last segment, e.g. gen)
make gen-entity

# Match the checked-in layout (package entity, files next to hand-written code)
go run ./cmd/genentity -o internal/golangstructure/domain/entity
```

Flags (see `go run ./cmd/genentity -h`):

| Flag | Purpose |
|------|---------|
| `-o` / `-out` | **Required.** Output directory; Go **package name** is the last path segment (e.g. `gen` or `entity`) |
| `-schema` | Postgres schema to scan (default `public`) |
| `-ignore-file` | Optional path to ignore-tables list (default: embedded `ignore_tables.txt`) |

Tables listed in `cmd/genentity/ignore_tables.txt` are skipped. Only **`SQLDB_DRIVER=postgres`** is supported.

## Observability Stack

```
┌──────────────────────────────────────────────────────────────────┐
│                        Docker Compose                            │
│                                                                  │
│  ┌─────────────┐  stdout  ┌──────────┐  push   ┌──────┐          │
│  │ App Service │ ───────► │  Alloy   │ ──────► │ Loki │          │
│  │ (JSON logs) │          │  :12345  │         └──┬───┘          │ 
│  │             │          └──────────┘            │              │
│  │  /metrics ◄──── Prometheus (:9090)             │ query        │
│  │  (promx)    │           │                   ┌──▼─────┐        │
│  └─────────────┘           │   ┌────────┐      │Grafana │        │
│                            └──►│ TSDB   │─────►│ :3001  │        │
│  ┌─────────────┐               └────────┘      └──▲─────┘        │
│  │   Tempo     │ ◄────────────────────────────────┘ query        │
│  │  (traces)   │                                                 │ 
│  └─────────────┘                                                 │ 
└──────────────────────────────────────────────────────────────────┘
```

### Components

| Component      | Role | Port |
|----------------|---------------------------------------------------------------------------------------------------------------|-------|-
| **Alloy**      | Grafana's unified telemetry collector. Discovers Docker containers with label `logging=true`, scrapes stdout logs, parses JSON, extracts labels (`traceId`, `service`, `method`, `statusCode`, `code`), pushes to Loki. Built-in debug UI.          | 12345 |
| **Loki**       | Log aggregation and storage. Indexes logs by labels for fast querying via LogQL.                              | 3100  |
| **Prometheus** | Scrapes `/metrics` from Go app every 15s. Stores HTTP request metrics + Go runtime metrics. Query via PromQL. | 9090  |
| **Tempo**      | Distributed trace storage. Accepts OTLP gRPC/HTTP spans. Ready for future OTel instrumentation.               | 3200, 4317, 4318 |
| **Grafana**    | Visualization. Pre-configured datasources (Loki + Prometheus + Tempo) with `traceId` correlation.             | 3001  |

### Alloy Pipeline

Grafana Alloy is configured with a River-based pipeline (`deployment/alloy/config.alloy`):

1. **`discovery.docker`** — discover Docker containers with label `logging=true`
2. **`discovery.relabel`** — extract container name, compose service, job name as labels
3. **`loki.source.docker`** — tail container stdout/stderr logs
4. **`loki.process`** — JSON pipeline: parse fields, promote to Loki labels, parse timestamp
5. **`loki.write`** — push processed log entries to Loki

### Grafana Datasource Correlation

**Loki** — derived fields on `traceId`:
- **"Search TraceID in Logs"** — searches all services for the same `traceId`
- **"View in Tempo"** — jumps to Tempo trace view (when OTLP spans are available)

**Tempo** — traces-to-logs:
- Viewing a trace in Tempo → links back to Loki logs filtered by `traceId`

**Tempo** — traces-to-metrics:
- Links from Tempo to Prometheus metrics for the same service

## Monitoring Guide

### Dashboards

Open **Grafana** at http://localhost:3001 → **Dashboards** → **Golang Structure**

#### 1. Golang Structure - Service

Single-page view combining **metrics + logs** for the service:

| Section          | Panels                                                                                                 |
|------------------|--------------------------------------------------------------------------------------------------------|
| **HTTP Metrics** | Request Rate, Error Rate (4xx+5xx), Duration (avg), Duration (p50/p90/p99), In-Flight, Total Requests, 5xx Errors, Avg Latency |
| **Go Runtime** | Goroutines, Threads, Heap Alloc, Stack In Use, Goroutines Over Time, Memory Over Time, GC Pause Duration |
| **Service Logs** | All Logs (golang-structure), Error Logs Only                                                           |

#### 2. Centralized Logs

Cross-service log aggregation with a **service dropdown filter**:

| Section            | Panels                                                                                         |
|--------------------|------------------------------------------------------------------------------------------------|
| **Overview**       | Log Volume by Service (stacked bar), Total Requests, 5xx Errors, 4xx Errors, Avg Response Time |
| **Error Analysis** | Errors by Status Code, Errors by Service                                                       |
| **All Logs**       | Logs from all selected services with JSON parsing                                              |
| **Error Logs**     | Error logs only (4xx + 5xx)                                                                    |

### Explore (Manual Queries)

Go to **Explore** (compass icon) → select datasource:

#### Loki (LogQL)

```logql
# All logs
{compose_service=~".+"} | json

# Filter by service
{compose_service="golang-structure"} | json

# Errors only (5xx)
{compose_service=~".+", statusCode=~"5.."} | json

# Search by traceId (cross-service correlation)
{compose_service=~".+"} |= "YOUR-TRACE-ID" | json

# Slow requests (> 100ms)
{compose_service=~".+"} | json | durationMs > 100

# DB errors
{compose_service=~".+"} |= "dbLogs" |= "error" | json
```

#### Prometheus (PromQL)

```promql
# Request rate by endpoint
sum(rate(http_requests_total[5m])) by (method, path)

# P99 latency
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Error rate percentage
sum(rate(http_requests_total{status_code=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# Goroutines
go_goroutines{job="golang-structure"}

# Heap memory
go_memstats_heap_alloc_bytes{job="golang-structure"}
```

### Trace ID Correlation Workflow

1. Find a log entry in Grafana Explore or the Centralized Logs dashboard
2. Expand the log line details
3. Find the `traceId` field
4. Click **"Search TraceID in Logs"** → see all related logs across services
5. Trace the full request path: Gateway → Orchestrator → Main Service → DB

### Example

```bash
curl -X POST http://localhost:3000/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test","age":25,"email":"test@example.com","password":"12345678","confirm_password":"12345678"}'
```

Then in Grafana Explore:

```logql
{compose_service=~".+"} |= "<traceId-from-response-header>" | json
```

You will see **3 log entries** — one from each service — all correlated by the same `traceId`, with the `source` field showing the nested call chain.

## Environment Variables

| Variable                  | Default            | Description                                          |
|---------------------------|--------------------|------------------------------------------------------|
| `SERVICE_NAME`            | `golang-structure` | Service name in logs and metrics                     |
| `SERVICE_PORT`            | `8080`             | HTTP listen port                                     |
| `SERVICE_READ_TIMEOUT`    | `60`               | Fiber read timeout (seconds)                         |
| `MASK_PATTERN`            | `{}`               | JSON masking rules (also applied to `appLogs.fields`)|
| `LOG_LEVEL`               | `info`             | App logger threshold: `debug` / `info` / `warn` / `error` |
| `SQLDB_HOST`              | -                  | Database host                                        |
| `SQLDB_PORT`              | -                  | Database port                                        |
| `SQLDB_USER`              | -                  | Database user                                        |
| `SQLDB_PASSWORD`          | -                  | Database password                                    |
| `SQLDB_DB_NAME`           | -                  | Database name                                        |
| `SQLDB_SSL_MODE`          | -                  | SSL mode (disable/require)                           |
| `SQLDB_MAX_IDLE_CONNS`    | -                  | Max idle DB connections                              |
| `SQLDB_MAX_OPEN_CONNS`    | -                  | Max open DB connections                              |
| `SQLDB_CONN_MAX_LIFETIME` | -                  | DB connection max lifetime                           |
| `SQLDB_DRIVER`            | -                  | DB driver (postgres/mysql/sqlite/sqlserver)          |
| `BCRYPT_COST`             | `10`               | bcrypt hashing cost                                  |

## Makefile Commands

| Command           | Description |
|-------------------|-------------|
| `make help`       | List targets and short descriptions |
| `make run`        | Run main API: `go run cmd/golangstructure/main.go` |
| `make gateway-run`| Run gateway demo: `go run cmd/gatewaydummie/main.go` |
| `make orchestrate-run` | Run orchestrator demo: `go run cmd/orchestratedummie/main.go` |
| `make liquibase-up` | Run Liquibase `update` using `deployment/liquibase/properties/dev.properties` (Liquibase must be installed) |
| `make gen-entity` | Run codegen: `go run ./cmd/genentity -o internal/golangstructure/domain/entity/gen` (see [Entity code generation](#entity-code-generation-genentity); use `-o .../entity` for package `entity`) |
| `make up`         | `docker compose up -d` (Postgres, atlas-migrate, observability) |
| `make down`       | `docker compose down` |

## Testing

Unit tests cover:

| Path                                                                   | Coverage                                                                                  |
|------------------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| `internal/golangstructure/usecase/auth/register/service_test.go`       | Register service: success, duplicate user, user log error, transaction error, bcrypt error|
| `pkg/db/gormx/repository_test.go`                                      | Generic `BaseRepository` transaction and CRUD behavior                                    |
| `pkg/mask/mask_test.go`                                                | JSON/map masking engine                                                                   |
| `pkg/mask/sql_test.go`                                                 | SQL `INSERT ... VALUES(...)` masking                                                      |

Tests in the register suite use `mockery`-generated mocks (`internal/.../repository/mocks`) and `logx.Noop()` for the logger.

Run all tests:

```bash
go test ./...
```

## Bruno API Collection

Sample requests for every endpoint live under `docs/apis/bruno/`. Open the folder with [Bruno](https://www.usebruno.com/) to drive the local stack interactively without writing curl commands.

```
docs/apis/bruno/
├── golangstructure/
│   ├── auth/Register.bru
│   ├── health/{Livez,Readyz}.bru
│   └── users/{Get users, Create user, Update user, Delete user}.bru
├── orchestratedummie/Register.bru
└── gatewaydummie/Register.bru
```
