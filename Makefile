.PHONY: help
.DEFAULT_GOAL := help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

gen-proto: ## generate Go server interface from .proto
	protoc -I pb/ pb/events.proto --go_out=plugins=grpc:pb

run-server: ## Run the Go server
	go run server/main.go

run-client: ## Run the Node client
	node client/main.js