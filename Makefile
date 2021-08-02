SOURCES := $(shell find . -name '*.go' -type f -not -path './vendor/*'  -not -path '*/mocks/*')

IMAGE_NAME = skyros

# Dependency Management
.PHONY: vendor
vendor: go.mod go.sum
	@GO111MODULE=on go get ./...

# Testing
.PHONY: unittest
unittest: vendor
	GO111MODULE=on go test -short -covermode=atomic ./...

.PHONY: test
test: vendor
	GO111MODULE=on go test -covermode=atomic ./...

# Build
.PHONY: docker
docker: vendor $(SOURCES)
	@docker build -t $(IMAGE_NAME):latest .

.PHONY: run
run:
	@docker-compose up -d

.PHONY: stop
stop:
	@docker-compose down

# Database Migration
.PHONY: migrate-prepare
migrate-prepare:
	@GO111MODULE=off go get -tags 'mysql' -u github.com/golang-migrate/migrate/cmd/migrate

.PHONY: migrate-up
migrate-up:
	@migrate -database "mysql://root:root@tcp(127.0.0.1:3306)/skyros" \
	-path=internal/mysql/migrations up


# Docker
.PHONY: mysql-up
mysql-up:
	@docker-compose up -d mysql

.PHONY: mysql-down
mysql-down:
	@docker stop skyros.mysql
