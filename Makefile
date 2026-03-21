.PHONY: service-up service-down test lint build migrate-up migrate-down tidy

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

migrate-up:
	@migrate -path ./userservice/migrations -database "$$USER_DATABASE_URL" up
	@migrate -path ./productservice/migrations -database "$$PRODUCT_DATABASE_URL" up
	@migrate -path ./orderservice/migrations -database "$$ORDER_DATABASE_URL" up

migrate-down:
	@migrate -path ./userservice/migrations -database "$$USER_DATABASE_URL" down
	@migrate -path ./productservice/migrations -database "$$PRODUCT_DATABASE_URL" down
	@migrate -path ./orderservice/migrations -database "$$ORDER_DATABASE_URL" down
