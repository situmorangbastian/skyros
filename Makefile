GITHUB_TOKEN=

# Docker
.PHONY: mysql-up
mysql-up:
	@docker-compose up -d skyros.mysql

.PHONY: mysql-down
mysql-down:
	@docker stop skyros.mysql.database

.PHONY: service-up
service-up:
	@docker-compose up -d skyros.userservice
	@docker-compose up -d skyros.productservice

.PHONY: service-down
service-down:
	@docker stop skyros.userservice.svc
	@docker stop skyros.productservice.svc
	@docker rm skyros.userservice.svc
	@docker rm skyros.productservice.svc

# Database Migration
.PHONY: migrate-prepare
migrate-prepare:
	@GO111MODULE=off go get -tags 'mysql' -u github.com/golang-migrate/migrate/cmd/migrate

.PHONY: service-migrate-up
service-migrate-up:
	@migrate -database "mysql://root:password@tcp(127.0.0.1:3306)/userservice" \
	-path=userservice/internal/mysql/migrations up
	@migrate -database "mysql://root:password@tcp(127.0.0.1:3306)/productservice" \
	-path=productservice/internal/mysql/migrations up

# Build Docker Services
.PHONY: service-docker
service-docker:
	@docker build --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) -f Dockerfile-userservice -t skyros-user-service:latest .
	@docker build --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) -f Dockerfile-productservice -t skyros-product-service:latest .
