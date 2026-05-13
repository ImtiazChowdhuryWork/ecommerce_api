.PHONY: run build migrate seed tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

migrate:
	psql $$DATABASE_URL -f migrations/001_init.sql

seed:
	psql $$DATABASE_URL -f migrations/002_seed.sql

# Create the database (requires DATABASE_URL without the db name part)
createdb:
	createdb ecommerce

dropdb:
	dropdb ecommerce

.env:
	cp .env.example .env
