COMPOSE_FILE ?= docker-compose.yml

.PHONY: build
build:
	docker compose -f $(COMPOSE_FILE) build

.PHONY: up
up:
	docker compose -f $(COMPOSE_FILE) up --build

.PHONY: down
down:
	docker compose -f $(COMPOSE_FILE) down

.PHONY: clean
clean:
	docker compose -f $(COMPOSE_FILE) down -v

.PHONY: unit-test
unit-test:
	go generate ./...
	go test ./internal/... -count=1 -cover

.PHONY: e2e-test
e2e-test:
	docker compose -f docker-compose-e2e.yml down -v
	docker compose -f docker-compose-e2e.yml up --build --force-recreate e2e-tests
