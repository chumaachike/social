
include .envrc

MIGRATIONS_PATH := ./cmd/migrate/migrations

.PHONY: migration migrate-up migrate-down migrate-version

migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" up

migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" down $(filter-out $@,$(MAKECMDGOALS))

migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_ADDR)" version

# Prevent Make from treating migration names or counts as targets.
%:
	@:
