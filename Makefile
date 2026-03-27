.PHONY: run

run:
	go run cmd/golangstructure/main.go

orchestrate-run:
	go run cmd/orchestratedummie/main.go

gateway-run:
	go run cmd/gatewaydummie/main.go