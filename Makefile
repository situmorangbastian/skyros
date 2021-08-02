SOURCES := $(shell find . -name '*.go' -type f -not -path './vendor/*'  -not -path '*/mocks/*')

IMAGE_NAME = skyros

# Dependency Management
.PHONY: vendor
vendor: go.mod go.sum
	@GO111MODULE=on go get ./...

# Linter
.PHONY: lint-prepare
lint-prepare:
	@echo "Installing golangci-lint"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0

.PHONY: lint
lint:
	golangci-lint run ./...

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
