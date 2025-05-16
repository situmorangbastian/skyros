.PHONY: service-up
service-up:
	@docker compose up -d

.PHONY: service-down
service-down:
	@docker compose down
