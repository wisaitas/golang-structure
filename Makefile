.PHONY: help run orchestrate-run gateway-run liquibase-up up down

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  run - Run the main API locally"
	@echo "  orchestrate-run - Run the orchestrator demo locally"
	@echo "  gateway-run - Run the gateway demo locally"
	@echo "  liquibase must be installed"
	@echo "  liquibase-up - Run the Liquibase migrations"
	@echo "  up - Start the full stack"
	@echo "  down - Stop the full stack"

run:
	go run cmd/golangstructure/main.go

orchestrate-run:
	go run cmd/orchestratedummie/main.go

gateway-run:
	go run cmd/gatewaydummie/main.go

liquibase-up:
	cd deployment/liquibase && liquibase --defaults-file=properties/dev.properties update

up:
	docker compose up -d

down:
	docker compose down