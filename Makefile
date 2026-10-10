-include .env
export 

.PHONY: psql-conn migrations-up migrations-down run

MIGRATION_DB_URL ?= $(if $(LOCAL_DB_URL),$(LOCAL_DB_URL),$(DB_URL))

psql-conn:
	@psql "$(MIGRATION_DB_URL)"

migrations-up:
	@goose -dir sql/schema postgres "$(MIGRATION_DB_URL)" up

migrations-down:
	@goose -dir sql/schema postgres "$(MIGRATION_DB_URL)" down

run:
	@go run cmd/ecommerce/main.go
