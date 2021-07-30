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
