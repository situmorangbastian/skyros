.PHONY: service-up
service-up:
	@docker compose up -d

.PHONY: service-down
service-down:
	@docker compose down

.PHONY: migrate-up
migrate-up:
	@migrate -database "mysql://root:password@tcp(127.0.0.1:3306)/userservice" \
	-path=userservice/internal/mysql/migrations up
	@migrate -database "mysql://root:password@tcp(127.0.0.1:3306)/productservice" \
	-path=productservice/internal/mysql/migrations up
	@migrate -database "mysql://root:password@tcp(127.0.0.1:3306)/orderservice" \
	-path=orderservice/internal/mysql/migrations up
