include .envrc

MIGRATIONS_PATH := ./cmd/migrate/migrations

.PHONY: migration migrate-up migrate-down migrate-version \
        migrate-force gen-docs run build test

migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" up $(filter-out $@,$(MAKECMDGOALS))

migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" down $(or $(filter-out $@,$(MAKECMDGOALS)),1)

migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" version

migrate-force:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" force $(filter-out $@,$(MAKECMDGOALS))

gen-docs:
	@swag fmt -d cmd,internal
	@swag init -g api/main.go -d cmd,internal

run:
	@go run ./cmd/api

build:
	@go build -o bin/api ./cmd/api

test:
	@go test -v -race ./...

%:
	@: