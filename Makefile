-include .env
export

LOCAL_DB_URL ?= postgres://postgres:postgres@localhost:5433/ecommerce?sslmode=disable
MIGRATIONS_DIR ?= ./sql/schema

.PHONY: psql-conn migrations-up migrations-down run

psql-conn:
	@psql "$(LOCAL_DB_URL)"

migrations-up:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(LOCAL_DB_URL)" up

migrations-down:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(LOCAL_DB_URL)" down

run:
	@go run cmd/ecommerce/main.go
