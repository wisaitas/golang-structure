# Golang Structure

Go REST API project with **custom distributed tracing** and **structured JSON logging** — designed as an alternative to OpenTelemetry that gives full control over log format, field masking, and cross-service correlation.

Integrated with **Grafana + Loki + Tempo** observability stack for log aggregation, search, and visualization.

## Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Custom Distributed Tracing](#custom-distributed-tracing)
- [Structured JSON Log Format](#structured-json-log-format)
- [Data Masking](#data-masking)
- [Observability Stack](#observability-stack)
- [Monitoring Guide](#monitoring-guide)
- [Makefile Commands](#makefile-commands)

## Architecture

```
                         ┌──────────────┐
                         │   Grafana    │ :3001
                         │  Dashboard   │
                         └──────┬───────┘
                                │ query
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
              ┌──────────┐ ┌────────┐ ┌───────────┐
              │   Loki   │ │ Tempo  │ │  Promtail │
              │   :3100  │ │ :3200  │ │  (agent)  │
              └──────────┘ └────────┘ └─────┬─────┘
                    ▲                       │ scrapes Docker logs
                    │ push                  │
                    └───────────────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                  ▼
      ┌──────────────┐ ┌───────────────┐ ┌─────────────────┐
      │   Gateway    │ │  Orchestrator │ │ golang-structure│
      │   :3000      │ │   :8081       │ │     :8080       │
      └──────┬───────┘ └──────┬────────┘ └───────┬─────────┘
             │                │                  │
             │  X-Trace-Id    │  X-Trace-Id      │
             │  X-Source      │  X-Source        │
             │  X-Internal    │  X-Internal      ▼
             └─────►──────────└──────►────── ┌──────────┐
                                             │ Postgres │
                                             │  :5432   │
                                             └──────────┘
```

**Request Flow:** Gateway → Orchestrator → golang-structure → PostgreSQL

Each service outputs a **structured JSON log line** to stdout per request. Promtail scrapes Docker container logs and pushes them to Loki. Grafana queries Loki and correlates logs by `traceId`.

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
│   ├── promtail/promtail-config.yaml      #   Log scraping pipeline
│   └── grafana/
│       ├── provisioning/
│       │   ├── datasources/datasources.yaml  # Loki + Tempo auto-provisioned
│       │   └── dashboards/dashboards.yaml    # Dashboard auto-discovery
│       └── dashboards/
│           └── golang-structure.json          # Pre-built dashboard
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
| Log Storage | Grafana Loki 3.5 |
| Trace Storage | Grafana Tempo 2.7 |
| Log Agent | Promtail 3.5 |
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
| Loki API | http://localhost:3100 | - |
| Tempo API | http://localhost:3200 | - |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/auth/register` | Register a new user |
| `GET` | `/api/v1/users/` | List users |
| `POST` | `/api/v1/users/` | Create user |
| `PUT` | `/api/v1/users/:user_id` | Update user |
| `DELETE` | `/api/v1/users/:user_id` | Delete user |

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
      "headers": { "Content-Type": "application/json", "..." : "..." },
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

## Observability Stack

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Compose                          │
│                                                             │
│  ┌─────────────┐    stdout    ┌──────────┐    push   ┌──────┐
│  │ App Service  │ ──────────► │ Promtail │ ────────► │ Loki │
│  │ (JSON logs)  │             └──────────┘           └──┬───┘
│  └─────────────┘                                        │
│                                                         │ query
│  ┌─────────────┐                                     ┌──▼─────┐
│  │   Tempo     │ ◄──────────────────────────────────►│Grafana │
│  │ (traces)    │         query                       │ :3001  │
│  └─────────────┘                                     └────────┘
└─────────────────────────────────────────────────────────────┘
```

| Component | Role | Port |
|-----------|------|------|
| **Promtail** | Scrapes Docker container logs via Docker socket, parses JSON, extracts labels (`traceId`, `service`, `method`, `statusCode`, `code`), pushes to Loki | - |
| **Loki** | Log aggregation and storage. Indexes logs by labels for fast querying via LogQL | 3100 |
| **Tempo** | Distributed trace storage. Accepts OTLP gRPC/HTTP spans. Ready for future OTel instrumentation | 3200, 4317, 4318 |
| **Grafana** | Visualization dashboard. Pre-configured datasources (Loki + Tempo) with `traceId` correlation derived fields | 3001 |

### Promtail Pipeline

Promtail is configured to:
1. Discover Docker containers with label `logging=promtail`
2. Parse JSON log lines from stdout
3. Extract fields: `traceId`, `service`, `method`, `statusCode`, `code`
4. Set extracted fields as Loki **labels** for fast filtering
5. Parse `timestamp` from the log's RFC3339 timestamp field

### Grafana Datasource Correlation

Loki datasource is configured with **derived fields**:
- Clicking `traceId` in a log → **"Search TraceID in Logs"** → searches all services for the same `traceId`
- Clicking `traceId` in a log → **"View in Tempo"** → jumps to Tempo trace view (when OTLP spans are available)

Tempo datasource is configured with **traces-to-logs**:
- Viewing a trace in Tempo → links back to Loki logs filtered by the same `traceId`

## Monitoring Guide

### 1. Pre-built Dashboard

Navigate to **Grafana** (http://localhost:3001) → **Dashboards** → **Golang Structure** → **Golang Structure - Observability**

Dashboard panels:
| Panel | Description |
|-------|-------------|
| Log Volume | Request count per service over time |
| Total Requests | Total request count |
| Error Requests (5xx) | Server error count |
| Client Errors (4xx) | Client error count |
| Avg Response Time | Average response duration (ms) |
| Errors by Status Code | Error distribution by HTTP status |
| Errors by Service | Error distribution by service |
| Response Time by Service | Average latency per service |
| P95 Response Time | 95th percentile latency per service |
| DB Query Duration | Average database query time |
| DB Errors | Database error count |
| All Logs | Full log explorer with JSON parsing |
| Error Logs Only | Filtered error log explorer |

### 2. Log Exploration (Grafana Explore)

Go to **Explore** (compass icon) → Select **Loki** datasource

#### View all logs
```logql
{compose_service=~".+"} | json
```

#### Filter by service
```logql
{compose_service="golang-structure"} | json
```

#### Filter errors only (5xx)
```logql
{compose_service=~".+", statusCode=~"5.."} | json
```

#### Filter client errors (4xx)
```logql
{compose_service=~".+", statusCode=~"4.."} | json
```

#### Search by Trace ID (cross-service correlation)
```logql
{compose_service=~".+"} |= "YOUR-TRACE-ID-HERE" | json
```

This shows logs from **all services** in the request chain (gateway → orchestrator → main service) — the core of distributed tracing.

#### Find slow requests (> 100ms)
```logql
{compose_service=~".+"} | json | durationMs > 100
```

#### Find requests with DB errors
```logql
{compose_service=~".+"} |= "dbLogs" |= "error" | json
```

#### Filter by HTTP method
```logql
{compose_service=~".+", method="POST"} | json
```

### 3. Trace ID Correlation Workflow

1. Find a log entry in Grafana Explore
2. Expand the log line details
3. Find the `traceId` field
4. Click **"Search TraceID in Logs"** → see all related logs across services
5. Trace the full request path: Gateway → Orchestrator → Main Service → DB

### 4. Distributed Trace Example

Send a request through the gateway:
```bash
curl -X POST http://localhost:3000/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test","age":25,"email":"test@example.com","password":"12345678","confirm_password":"12345678"}'
```

Then in Grafana Explore, search for the `traceId` from the response header:
```logql
{compose_service=~".+"} |= "<traceId-from-response>" | json
```

You will see **3 log entries** — one from each service — all correlated by the same `traceId`, with the `source` field showing the nested call chain.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `golang-structure` | Service name in logs |
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
