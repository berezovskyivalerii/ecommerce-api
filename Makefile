include .env
export 

psql-conn:
	@psql ${DB_URL}

migrations-up:
	@goose -dir sql/schema postgres ${DB_URL} up
	
migrations-down:
	@goose -dir sql/schema postgres ${DB_URL} down
	
run:
	@go run cmd/ecommerce/main.go
