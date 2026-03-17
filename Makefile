.PHONY: help install run build test migrate migrate-up migrate-down docker-up docker-down clean

help:
	@echo "Available commands:"
	@echo "  make install      - Install dependencies"
	@echo "  make run          - Run the server"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run tests"
	@echo "  make migrate      - Run all migrations"
	@echo "  make migrate-up   - Run next migration"
	@echo "  make migrate-down - Rollback last migration"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make docker-restart - Rebuild and restart backend"
	@echo "  make docker-build - Rebuild all services"
	@echo "  make docker-logs  - View live logs"
	@echo "  make db-shell     - Open Postgres shell"
	@echo "  make check        - Run fmt, lint, and test"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run static analysis"
	@echo "  make clean        - Clean build artifacts and docker"

install:
	go mod download
	go mod tidy

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v -cover ./...

# migrate1:
# 	@echo "Running migrations..."
# 	@for file in migrations/*.up.sql; do \
# 		echo "Applying $$file..."; \
# 		PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -f $$file; \
# 	done

migrate:
	@echo "Running migrations..."
	@docker exec -i file-storage-postgres psql -U postgres -d file_storage < migrations/001_initial_schema.up.sql
	@docker exec -i file-storage-postgres psql -U postgres -d file_storage < migrations/002_sync_tokens.up.sql
	@docker exec -i file-storage-postgres psql -U postgres -d file_storage < migrations/003_add_upload_to_field.up.sql
	@echo "✅ Migrations completed"

migrate-up:
	@echo "Run specific migration manually or use a migration tool like golang-migrate"

migrate-down:
	@echo "Run rollback migration manually or use a migration tool like golang-migrate"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-restart:
	docker compose up -d --build backend
	docker image prune -f

docker-prune:
	docker system prune -af --volumes

docker-build:
	docker compose build

docker-logs:
	docker compose logs -f --tail 50

db-shell:
	docker exec -it file-storage-postgres psql -U postgres -d file_storage

fmt:
	go fmt ./...

lint:
	go vet ./...

check: fmt lint test

clean:
	rm -rf bin/
	go clean
	docker system prune -f

# Admin commands (SSH-only in production)
create-admin:
	go run cmd/admin/create_admin.go

list-admins:
	go run cmd/admin/list_admins.go

reset-admin-password:
	go run cmd/admin/reset_admin_password.go

# Sync token management (SSH-only in production)
create-sync-token:
	go run cmd/admin/create_sync_token.go

list-sync-tokens:
	go run cmd/admin/list_sync_tokens.go

rotate-sync-token:
	go run cmd/admin/rotate_sync_token.go

revoke-sync-token:
	go run cmd/admin/revoke_sync_token.go

sync-token-stats:
	go run cmd/admin/sync_token_stats.go

# Development shortcuts
dev: docker-up migrate run

fresh: docker-down docker-up migrate run