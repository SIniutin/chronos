SHELL := /bin/sh

BACKEND_DIR := backend
FRONTEND_DIR := frontend

COMPOSE := docker compose -f $(BACKEND_DIR)/docker-compose.yml --project-directory $(BACKEND_DIR)
DB_URL ?= postgres://postgres:postgres@localhost:5432/history-db?sslmode=disable
API_BASE_URL ?= http://localhost:8080
ANDROID_API_BASE_URL ?= http://10.0.2.2:8080
WEB_PORT ?= 3000
GOCACHE ?= /tmp/go-build

.PHONY: help manual manual-reset manual-android manual-web backend-up backend-down backend-reset-db backend-logs backend-db backend-migrate backend-seed backend-restart frontend-bootstrap frontend-pub frontend-run frontend-run-android frontend-run-web test-backend build-backend

help:
	@printf '%s\n' 'Manual test targets:'
	@printf '%s\n' '  make manual              Start backend in Docker, run migrations, then run Flutter on Linux desktop'
	@printf '%s\n' '  make manual-reset        Drop local Docker Postgres volume, then run manual'
	@printf '%s\n' '  make manual-android      Same, but Flutter uses http://10.0.2.2:8080 for Android emulator'
	@printf '%s\n' '  make manual-web          Same, but starts Flutter web-server on WEB_PORT=3000'
	@printf '%s\n' '  make backend-up          Start backend app + Postgres in Docker'
	@printf '%s\n' '  make backend-db          Start only Postgres'
	@printf '%s\n' '  make backend-migrate     Apply Goose migrations to local Docker Postgres'
	@printf '%s\n' '  make backend-seed        Seed local Docker Postgres with structured history content'
	@printf '%s\n' '  make backend-down        Stop backend containers'
	@printf '%s\n' '  make backend-reset-db    Stop containers and delete the local Docker Postgres volume'
	@printf '%s\n' '  make frontend-run        Run Flutter app with API_BASE_URL'
	@printf '%s\n' '  make frontend-bootstrap  Generate missing Flutter platform folders'

manual: backend-db backend-migrate backend-up frontend-run

manual-reset: backend-reset-db manual

manual-android: backend-db backend-migrate backend-up frontend-run-android

manual-web: backend-db backend-migrate backend-up frontend-run-web

backend-db:
	$(COMPOSE) up -d db
	$(COMPOSE) exec -T db sh -c 'until pg_isready -U postgres -d history-db; do sleep 1; done'

backend-up:
	$(COMPOSE) up -d --build app

backend-restart:
	$(COMPOSE) up -d --build app

backend-migrate:
	cd $(BACKEND_DIR) && DATABASE_URL="$(DB_URL)" GOCACHE=$(GOCACHE) go run ./cmd/history-migrate

backend-seed:
	cd $(BACKEND_DIR) && DATABASE_URL="$(DB_URL)" GOCACHE=$(GOCACHE) go run ./cmd/history-seed

backend-logs:
	$(COMPOSE) logs -f app

backend-down:
	$(COMPOSE) down

backend-reset-db:
	$(COMPOSE) down -v

frontend-pub:
	cd $(FRONTEND_DIR) && flutter pub get

frontend-bootstrap:
	@if [ ! -d "$(FRONTEND_DIR)/linux" ] || [ ! -d "$(FRONTEND_DIR)/android" ] || [ ! -d "$(FRONTEND_DIR)/web" ]; then \
		cd $(FRONTEND_DIR) && flutter create --platforms=linux,android,web .; \
	else \
		printf '%s\n' 'Flutter platform folders already exist'; \
	fi

frontend-run: frontend-bootstrap frontend-pub
	cd $(FRONTEND_DIR) && flutter run -d linux --dart-define=API_BASE_URL=$(API_BASE_URL)

frontend-run-android: frontend-bootstrap frontend-pub
	cd $(FRONTEND_DIR) && flutter run -d android --dart-define=API_BASE_URL=$(ANDROID_API_BASE_URL)

frontend-run-web: frontend-bootstrap frontend-pub
	cd $(FRONTEND_DIR) && flutter run -d web-server --web-hostname 0.0.0.0 --web-port $(WEB_PORT) --dart-define=API_BASE_URL=$(API_BASE_URL)

test-backend:
	cd $(BACKEND_DIR) && GOCACHE=$(GOCACHE) go test ./...

build-backend:
	cd $(BACKEND_DIR) && GOCACHE=$(GOCACHE) go build -o /tmp/history-app-build ./cmd/history-app
