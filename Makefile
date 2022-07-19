# Docker
.PHONY: mysql-up
mysql-up:
	@docker-compose up -d skyros.mysql

.PHONY: mysql-down
mysql-down:
	@docker stop skyros.mysql.database

.PHONY: userservice-up
userservice-up:
	@docker-compose up -d skyros.userservice

# Database Migration
.PHONY: migrate-prepare
migrate-prepare:
	@GO111MODULE=off go get -tags 'mysql' -u github.com/golang-migrate/migrate/cmd/migrate

.PHONY: userservice-migrate-up
userservice-migrate-up:
	@migrate -database "mysql://root:password@tcp(127.0.0.1:3306)/userservice" \
	-path=userservice/internal/mysql/migrations up

# Build Docker Services
.PHONY: service-docker
service-docker:
	@docker build -f Dockerfile-userservice -t skyros-user-service:latest .
