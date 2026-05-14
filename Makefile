.PHONY: help start stop db api app install build

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  start    Start all services (db, api, app)"
	@echo "  stop     Stop Docker services (db)"
	@echo "  db       Start the database (Docker)"
	@echo "  api      Start the API server"
	@echo "  app      Start the frontend dev server"
	@echo "  install  Install frontend dependencies"
	@echo "  build    Build api and app for production"

start: db
	@echo ">>> Starting API and App (Ctrl+C to stop)..."
	@(cd api && make start) & (cd app && npm install && npm run dev); wait

stop:
	docker compose stop

db:
	docker compose up -d db
	@echo ">>> Waiting for database to be ready..."
	@docker compose exec db sh -c 'until pg_isready -U hackathon -d hackathon; do sleep 1; done'
	@echo ">>> Database is ready."

api: db
	cd api && make start

app:
	cd app && npm install && npm run dev

install:
	cd app && npm install

build:
	cd api && make build
	cd app && npm run build
