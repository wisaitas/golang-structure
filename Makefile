.PHONY: run orchestrate-run gateway-run \
       infra-up infra-down infra-logs \
       app-up app-down app-logs \
       up down logs

# ── Local Development ───────────────────────────────────────────────────
run:
	go run cmd/golangstructure/main.go

orchestrate-run:
	go run cmd/orchestratedummie/main.go

gateway-run:
	go run cmd/gatewaydummie/main.go

# ── Docker: Full Stack ──────────────────────────────────────────────────
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

# ── Docker: Infrastructure Only (DB + Observability) ────────────────────
infra-up:
	docker compose up -d postgres loki tempo alloy prometheus grafana

infra-down:
	docker compose down loki tempo alloy prometheus grafana

infra-logs:
	docker compose logs -f loki tempo alloy prometheus grafana

# ── Docker: App Services Only ──────────────────────────────────────────
app-up:
	docker compose up -d --build golang-structure gateway-dummie orchestrate-dummie

app-down:
	docker compose stop golang-structure gateway-dummie orchestrate-dummie

app-logs:
	docker compose logs -f golang-structure gateway-dummie orchestrate-dummie
