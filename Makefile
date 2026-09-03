include .env
export

.PHONY: run db-up db-down migrate-up migrate-down

run:
	@go run cmd/app/main.go

db-up:
	@docker compose up -d

db-down:
	@docker compose down

migrate-up:
	@migrate -path migrations -database "$(PG_URL)" -verbose up

migrate-down:
	@migrate -path migrations -database "$(PG_URL)" -verbose down