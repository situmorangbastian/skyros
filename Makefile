.PHONY: service-up service-down test lint build migrate-up migrate-down tidy proto proto-tools

# Code generator versions. Keep these in sync with the headers of the
# committed *.pb.go files and with grpc-gateway in go.mod.
PROTOC_GEN_GO_VERSION           ?= v1.36.6
PROTOC_GEN_GO_GRPC_VERSION      ?= v1.5.1
PROTOC_GEN_GRPC_GATEWAY_VERSION ?= v2.33.0

service-up:
	@docker compose up -d

service-down:
	@docker compose down

build:
	@CGO_ENABLED=0 GOOS=linux go build ./...

test:
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

lint:
	@golangci-lint run ./...

tidy:
	@go mod tidy && go mod verify

proto-tools:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION)

proto: proto-tools
	@command -v buf >/dev/null || { echo "buf not found: brew install buf"; exit 1; }
	@cd proto && buf generate

migrate-up:
	@migrate -path ./userservice/migrations -database "$$USER_DATABASE_URL" up
	@migrate -path ./productservice/migrations -database "$$PRODUCT_DATABASE_URL" up
	@migrate -path ./orderservice/migrations -database "$$ORDER_DATABASE_URL" up

migrate-down:
	@migrate -path ./userservice/migrations -database "$$USER_DATABASE_URL" down
	@migrate -path ./productservice/migrations -database "$$PRODUCT_DATABASE_URL" down
	@migrate -path ./orderservice/migrations -database "$$ORDER_DATABASE_URL" down
