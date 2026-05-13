.PHONY: help run orchestrate-run gateway-run liquibase-up gen-entity up down \
	atlas-hash-dev atlas-hash-uat atlas-hash-all atlas-up-dev atlas-up-uat \
	atlas-diff-dev-uat atlas-export-uat-from-dev

# Atlas URLs ให้ตรง docker-compose: postgres (dev) :5432 / postgres2 (uat) :5433
ATLAS_URL_DEV ?= postgres://admin:postgres@127.0.0.1:5432/golang-structure-db?sslmode=disable
ATLAS_URL_UAT ?= postgres://admin:postgres@127.0.0.1:5433/golang-structure-db-uat?sslmode=disable

# schema diff: SQL จากสถานะ --from ไปหา --to (ค่าเริ่มต้น UAT -> dev ให้เห็นว่าต้องปรับ UAT อย่างไรให้ใกล้ dev)
ATLAS_SCHEMA_DIFF_FROM ?= $(ATLAS_URL_UAT)
ATLAS_SCHEMA_DIFF_TO   ?= $(ATLAS_URL_DEV)

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  run - Run the main API locally"
	@echo "  orchestrate-run - Run the orchestrator demo locally"
	@echo "  gateway-run - Run the gateway demo locally"
	@echo "  liquibase must be installed"
	@echo "  liquibase-up - Run the Liquibase migrations"
	@echo "  atlas-hash-dev / atlas-hash-uat - Regenerate atlas.sum for that env dir"
	@echo "  atlas-hash-all - Hash both dev and uat migration dirs"
	@echo "  atlas-up-dev / atlas-up-uat - Apply Atlas migrations (run hash after editing SQL)"
	@echo "  atlas-diff-dev-uat - แสดง SQL ทาง stdout (ยังไม่สร้างไฟล์)"
	@echo "  atlas-export-uat-from-dev - บันทึก diff UAT->dev ลง deployment/atlas/migrations/uat/*.sql แล้ว hash + up uat เอง"
	@echo "  gen-entity - Generate domain models from live Postgres into entity/gen"
	@echo "  up - Start the full stack"
	@echo "  down - Stop the full stack"

run:
	go run cmd/golangstructure/main.go

orchestrate-run:
	go run cmd/orchestratedummie/main.go

gateway-run:
	go run cmd/gatewaydummie/main.go

atlas-hash-dev:
	atlas migrate hash --dir "file://deployment/atlas/migrations/dev"

atlas-hash-uat:
	atlas migrate hash --dir "file://deployment/atlas/migrations/uat"

atlas-hash-all: atlas-hash-dev
	@if [ -d deployment/atlas/migrations/uat ]; then $(MAKE) atlas-hash-uat; else echo "Skip atlas-hash-uat (no migrations/uat yet)"; fi

atlas-up-dev:
	atlas migrate apply --dir "file://deployment/atlas/migrations/dev" --url "$(ATLAS_URL_DEV)"

atlas-up-uat:
	atlas migrate apply --dir "file://deployment/atlas/migrations/uat" --url "$(ATLAS_URL_UAT)"

atlas-diff-dev-uat:
	atlas schema diff --from "$(ATLAS_SCHEMA_DIFF_FROM)" --to "$(ATLAS_SCHEMA_DIFF_TO)"

atlas-export-uat-from-dev:
	@mkdir -p deployment/atlas/migrations/uat
	@out=deployment/atlas/migrations/uat/$$(date +%Y%m%d%H%M%S)_from_dev_schema.sql; \
	atlas schema diff --from "$(ATLAS_SCHEMA_DIFF_FROM)" --to "$(ATLAS_SCHEMA_DIFF_TO)" > "$$out"; \
	echo "Wrote $$out"; \
	echo "แก้ไฟล์นี้: ลบ CREATE SCHEMA/TABLE ของ atlas_schema_revisions (ถ้ามี) แล้วค่อย hash + up"; \
	echo "Next: edit file -> make atlas-hash-uat && make atlas-up-uat"

liquibase-up:
	cd deployment/liquibase && liquibase --defaults-file=properties/dev.properties update

gen-entity:
	go run ./cmd/genentity -o internal/golangstructure/domain/entity

up:
	docker compose up -d

down:
	docker compose down