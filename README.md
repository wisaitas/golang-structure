# Golang Structure

Go REST API project with **custom distributed tracing** and **structured JSON logging** — designed as an alternative to OpenTelemetry that gives full control over log format, field masking, and cross-service correlation.

Integrated with **Grafana + Loki + Tempo + Alloy + Prometheus** observability stack.

## Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Custom Distributed Tracing](#custom-distributed-tracing)
- [Structured JSON Log Format](#structured-json-log-format)
- [Data Masking](#data-masking)
- [Prometheus Metrics](#prometheus-metrics)
- [Observability Stack](#observability-stack)
- [Monitoring Guide](#monitoring-guide)
- [Environment Variables](#environment-variables)
- [Makefile Commands](#makefile-commands)

## Architecture

```
                           ┌────────────────┐
                           │    Grafana      │ :3001
                           │   Dashboards    │
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
┌────────┐ ┌───────────┐ ┌─────────────────┐ ┌──────────┐
│Gateway │ │Orchestrator│ │golang-structure │ │ Postgres │
│ :3000  │ │  :8081     │ │     :8080       │ │  :5432   │
└───┬────┘ └─────┬──────┘ └───────┬─────────┘ └──────────┘
    │            │                │
    │ X-Trace-Id │  X-Trace-Id   │
    │ X-Source   │  X-Source     │
    └──►─────────└──►─────────────┘
```

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
│   └── orchestratedummie/main.go           #   Demo orchestrator (:8081)
│
├── internal/golangstructure/               # Application-private code
│   ├── config.go                           #   Environment-based config struct
│   ├── domain/
│   │   ├── entity/                         #   GORM models (User, BaseEntity)
│   │   └── repository/                     #   Repository interfaces & implementations
│   ├── initial/                            #   Dependency injection / bootstrap
│   │   ├── initial.go                      #     App lifecycle (New, Run, Shutdown)
│   │   ├── config.go                       #     DB connection setup
│   │   ├── middleware.go                   #     Middleware registration
│   │   ├── router.go                       #     Route group wiring
│   │   ├── repository.go                  #     Repository construction
│   │   ├── sdk.go                          #     External SDK setup
│   │   └── use_case.go                    #     Use case construction
│   ├── middleware/
│   │   ├── prometheus.go                   #   Prometheus metrics middleware
│   │   ├── logger.go                       #   Structured JSON logging middleware
│   │   └── cors.go                         #   CORS configuration
│   ├── router/
│   │   ├── auth.go                         #   /api/v1/auth routes
│   │   └── user.go                         #   /api/v1/users routes
│   └── usecase/                            #   Feature-based use cases
│       ├── auth/register/                  #     POST /auth/register
│       └── user/{createuser,getusers,      #     CRUD /users
│                 updateuser,deleteuser}/
│
├── pkg/                                    # Reusable shared libraries
│   ├── httpx/                              #   HTTP utilities
│   │   ├── logger.go                       #     Request logging middleware
│   │   ├── http.go                         #     HTTP client with header propagation
│   │   ├── error.go                        #     Error wrapping with stack traces
│   │   ├── model.go                        #     Log, Block, DBLog structs
│   │   ├── const.go                        #     Headers, response codes
│   │   └── util.go                         #     DB log collector, masking helpers
│   ├── promx/                              #   Prometheus metrics middleware
│   │   └── promx.go                        #     HTTP metrics + /metrics endpoint
│   ├── db/sqlx/                            #   GORM setup + custom SQL logger
│   │   ├── sql.go                          #     Connection factory, query collector
│   │   ├── model.go                        #     BaseEntity (id, timestamps)
│   │   └── const.go                        #     Driver constants
│   ├── mask/                               #   Field masking engine
│   │   ├── mask.go                         #     JSON/map value masking
│   │   ├── sql.go                          #     SQL INSERT value masking
│   │   └── pattern.go                      #     Pattern parser ("4:2", "4:com")
│   ├── bcryptx/                            #   Password hashing
│   └── validatorx/                         #   Request validation
│
├── deployment/                             # Infrastructure configs
│   ├── golang-structure/Dockerfile         #   Multi-stage Go build
│   ├── loki/loki-config.yaml              #   Loki log storage config
│   ├── tempo/tempo-config.yaml            #   Tempo trace storage config
│   ├── prometheus/prometheus.yaml         #   Prometheus scrape config
│   ├── alloy/config.alloy                 #   Grafana Alloy log collection pipeline
│   └── grafana/
│       ├── provisioning/
│       │   ├── datasources/datasources.yaml  # Loki + Prometheus + Tempo
│       │   └── dashboards/dashboards.yaml    # Dashboard auto-discovery
│       └── dashboards/
│           ├── golang-structure.json          # Service dashboard (metrics + logs)
│           └── centralized-logs.json          # Centralized log dashboard
│
├── httptest/                               # REST client test files (.http)
├── docker-compose.yaml                     # Full stack definition
├── Makefile                                # Dev & deploy commands
├── .env.template                           # Environment variable template
├── go.mod
└── go.sum
```

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.26 |
| HTTP Framework | Fiber v3 |
| ORM | GORM (PostgreSQL, MySQL, SQLite, SQL Server) |
| Config | caarlos0/env + godotenv |
| Validation | go-playground/validator v10 |
| Password | bcrypt (golang.org/x/crypto) |
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

### Run Full Stack (Docker)

```bash
# Start everything: app + database + observability
make up

# View logs
make logs
```

### Run Infrastructure Only (develop locally)

```bash
# Start database + observability stack
make infra-up

# Copy and configure environment
cp .env.template .env
# Edit .env with your settings

# Run the main API locally
make run

# (Optional) Run dummy services for distributed tracing demo
make orchestrate-run   # terminal 2
make gateway-run       # terminal 3
```

### Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| Main API | http://localhost:8080 | - |
| Gateway (demo) | http://localhost:3000 | - |
| Orchestrator (demo) | http://localhost:8081 | - |
| Grafana | http://localhost:3001 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| Loki API | http://localhost:3100 | - |
| Tempo API | http://localhost:3200 | - |
| Alloy UI | http://localhost:12345 | - |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/auth/register` | Register a new user |
| `GET` | `/api/v1/users/` | List users |
| `POST` | `/api/v1/users/` | Create user |
| `PUT` | `/api/v1/users/:user_id` | Update user |
| `DELETE` | `/api/v1/users/:user_id` | Delete user |
| `GET` | `/metrics` | Prometheus metrics endpoint |

**Demo chain** (distributed tracing):

| Service | Endpoint | Forwards to |
|---------|----------|-------------|
| Gateway `:3000` | `POST /register` | Orchestrator `:8081` |
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

| Header | Purpose |
|--------|---------|
| `X-Trace-Id` | Correlation ID shared across all services in a request chain |
| `X-Source` | JSON-serialized `Block` from downstream, nested to form a call chain |
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

| Field | Type | Description |
|-------|------|-------------|
| `traceId` | string | UUID correlating logs across services |
| `timestamp` | string | Request start time (RFC3339) |
| `durationMs` | string | Total request processing time |
| `current` | Block | This service's request data |
| `source` | Block | Downstream service chain (nested) |

### Block Fields

| Field | Type | Description |
|-------|------|-------------|
| `service` | string | Service name |
| `method` | string | HTTP method |
| `path` | string | Request path |
| `statusCode` | string | HTTP status code |
| `code` | string | Application response code (e.g. `E50000`) |
| `errorMessage` | string | Root error message |
| `stackTraces` | []string | Error stack trace chain |
| `dbLogs` | []DBLog | SQL queries executed during this request |
| `request` | Body | Masked request headers + body |
| `response` | Body | Masked response headers + body |

### DBLog Fields

| Field | Type | Description |
|-------|------|-------------|
| `source` | string | File and line number |
| `sql` | string | Masked SQL query |
| `rows` | int | Rows affected |
| `durationMs` | int | Query execution time |
| `error` | string | Query error (if any) |

## Data Masking

Sensitive data is automatically masked in logs using pattern-based rules configured via the `MASK_PATTERN` environment variable.

### Configuration

```bash
MASK_PATTERN={"password":"4:2","email":"4:com"}
```

### Pattern Syntax

| Pattern | Input | Output | Description |
|---------|-------|--------|-------------|
| `"4:2"` | `12345678` | `1234**78` | Keep 4 prefix chars, 2 suffix chars |
| `"4:com"` | `user@example.com` | `user@******com` | Keep 4 prefix chars, find "com" marker for suffix |

### What Gets Masked

- **Request/Response body** — JSON fields matching pattern keys
- **Headers** — Header values matching pattern keys
- **SQL queries** — Column values in `INSERT ... VALUES(...)` statements matching pattern keys

## Prometheus Metrics

The `pkg/promx` library provides a reusable Fiber middleware that exposes Prometheus metrics.

### Usage

```go
// One line: registers /metrics endpoint + records HTTP metrics
app.Use(middleware.Prometheus(app))
```

### Exposed Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | `method`, `path`, `status_code`, `service` | Total HTTP request count |
| `http_request_duration_seconds` | Histogram | `method`, `path`, `status_code`, `service` | Request latency (buckets: 5ms–10s) |
| `http_requests_in_flight` | Gauge | `service` | Currently processing requests |

### Go Runtime Metrics (auto-exposed)

The `prometheus/client_golang` library automatically exposes Go runtime metrics at `/metrics`:

| Metric | Description |
|--------|-------------|
| `go_goroutines` | Number of goroutines |
| `go_threads` | Number of OS threads |
| `go_memstats_heap_alloc_bytes` | Heap memory allocated |
| `go_memstats_stack_inuse_bytes` | Stack memory in use |
| `go_gc_duration_seconds` | GC pause duration |

## Observability Stack

```
┌──────────────────────────────────────────────────────────────────┐
│                        Docker Compose                            │
│                                                                  │
│  ┌─────────────┐  stdout  ┌──────────┐  push   ┌──────┐        │
│  │ App Service  │ ───────► │  Alloy   │ ──────► │ Loki │        │
│  │ (JSON logs)  │          │  :12345  │         └──┬───┘        │
│  │              │          └──────────┘            │             │
│  │  /metrics ◄──── Prometheus (:9090)              │ query       │
│  │  (promx)     │          │                    ┌──▼─────┐      │
│  └─────────────┘           │   ┌────────┐      │Grafana │      │
│                            └──►│ TSDB   │─────►│ :3001  │      │
│  ┌─────────────┐               └────────┘      └──▲─────┘      │
│  │   Tempo     │ ◄────────────────────────────────┘ query       │
│  │  (traces)   │                                                │
│  └─────────────┘                                                │
└──────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Role | Port |
|-----------|------|------|
| **Alloy** | Grafana's unified telemetry collector. Discovers Docker containers with label `logging=true`, scrapes stdout logs, parses JSON, extracts labels (`traceId`, `service`, `method`, `statusCode`, `code`), pushes to Loki. Built-in debug UI. | 12345 |
| **Loki** | Log aggregation and storage. Indexes logs by labels for fast querying via LogQL. | 3100 |
| **Prometheus** | Scrapes `/metrics` from Go app every 15s. Stores HTTP request metrics + Go runtime metrics. Query via PromQL. | 9090 |
| **Tempo** | Distributed trace storage. Accepts OTLP gRPC/HTTP spans. Ready for future OTel instrumentation. | 3200, 4317, 4318 |
| **Grafana** | Visualization. Pre-configured datasources (Loki + Prometheus + Tempo) with `traceId` correlation. | 3001 |

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

| Section | Panels |
|---------|--------|
| **HTTP Metrics** | Request Rate, Error Rate (4xx+5xx), Duration (avg), Duration (p50/p90/p99), In-Flight, Total Requests, 5xx Errors, Avg Latency |
| **Go Runtime** | Goroutines, Threads, Heap Alloc, Stack In Use, Goroutines Over Time, Memory Over Time, GC Pause Duration |
| **Service Logs** | All Logs (golang-structure), Error Logs Only |

#### 2. Centralized Logs

Cross-service log aggregation with a **service dropdown filter**:

| Section | Panels |
|---------|--------|
| **Overview** | Log Volume by Service (stacked bar), Total Requests, 5xx Errors, 4xx Errors, Avg Response Time |
| **Error Analysis** | Errors by Status Code, Errors by Service |
| **All Logs** | Logs from all selected services with JSON parsing |
| **Error Logs** | Error logs only (4xx + 5xx) |

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

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `golang-structure` | Service name in logs and metrics |
| `SERVICE_PORT` | `8080` | HTTP listen port |
| `SERVICE_STAGE` | `dev` | Deployment stage |
| `MASK_PATTERN` | `{}` | JSON masking rules |
| `SQLDB_HOST` | - | Database host |
| `SQLDB_PORT` | - | Database port |
| `SQLDB_USER` | - | Database user |
| `SQLDB_PASSWORD` | - | Database password |
| `SQLDB_DB_NAME` | - | Database name |
| `SQLDB_SSL_MODE` | - | SSL mode (disable/require) |
| `SQLDB_MAX_IDLE_CONNS` | - | Max idle DB connections |
| `SQLDB_MAX_OPEN_CONNS` | - | Max open DB connections |
| `SQLDB_CONN_MAX_LIFETIME` | - | DB connection max lifetime |
| `SQLDB_DRIVER` | - | DB driver (postgres/mysql/sqlite/sqlserver) |
| `BCRYPT_COST` | `10` | bcrypt hashing cost |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run main API locally |
| `make gateway-run` | Run gateway demo locally |
| `make orchestrate-run` | Run orchestrator demo locally |
| `make up` | Docker: start full stack (build + up) |
| `make down` | Docker: stop full stack |
| `make logs` | Docker: follow all logs |
| `make infra-up` | Docker: start DB + observability only |
| `make infra-down` | Docker: stop observability stack |
| `make infra-logs` | Docker: follow observability logs |
| `make app-up` | Docker: build & start app services only |
| `make app-down` | Docker: stop app services |
| `make app-logs` | Docker: follow app service logs |
