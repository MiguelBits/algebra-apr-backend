# Variables
DOCKER_FILE ?= docker-compose.yml

DOCKER_COMPOSE = docker compose -f $(DOCKER_FILE)

# Targets
.PHONY: help
help:
	@echo Available targets:
	@echo start            Start containers
	@echo stop             Stop containers
	@echo restart          Restart containers
	@echo rebuild          Rebuild app container without others
	@echo rebuild-all      Rebuild all containers
	@echo migrate-up       Apply database migrations
	@echo migrate-down     Rollback last migration
	@echo migrate-and-run  Apply migrations and start application
	@echo logs             View container log
	@echo clean            Stop containers and remove volumes

.PHONY: start
start:
	$(DOCKER_COMPOSE) up -d

.PHONY: stop
stop:
	$(DOCKER_COMPOSE) down

.PHONY: restart
restart: stop start

c ?= app
.PHONY: rebuild
rebuild:
	$(DOCKER_COMPOSE) up --build -d --no-deps --force-recreate $(c)

.PHONY: rebuild-all
rebuild-all:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up -d --build --force-recreate

.PHONY: migrate-up
migrate-up:
	$(DOCKER_COMPOSE) run --rm app ./migrate up

.PHONY: migrate-down
migrate-down:
	$(DOCKER_COMPOSE) run --rm app ./migrate down

.PHONY: migrate-and-run
migrate-and-run:
	@echo "Stopping containers..."
	$(DOCKER_COMPOSE) down
	@echo "Starting database..."
	$(DOCKER_COMPOSE) up -d --build db
	@echo "Waiting for database to be ready..."
	@sleep 10
	@echo "Running migrations..."
	@if $(DOCKER_COMPOSE) run --rm --build app ./migrate up; then \
		echo "Migrations completed successfully!"; \
		echo "Starting application and nginx..."; \
		$(DOCKER_COMPOSE) up -d --build app nginx; \
		echo "Application and nginx started successfully!"; \
	else \
		echo "Migration failed! Application will not start."; \
		echo "Please check the migration logs and fix any issues."; \
		exit 1; \
	fi

.PHONY: logs
logs:
	$(DOCKER_COMPOSE) logs -f

.PHONY: clean
clean:
	$(DOCKER_COMPOSE) down -v
