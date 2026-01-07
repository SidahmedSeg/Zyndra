.PHONY: run test migrate-up migrate-down install-deps build docker-build docker-up docker-down docker-logs

# Development
run:
	go run cmd/server/main.go

test:
	go test ./...

migrate-up:
	migrate -path migrations/postgres -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations/postgres -database "$$DATABASE_URL" down

migrate-version:
	migrate -path migrations/postgres -database "$$DATABASE_URL" version

install-deps:
	go get github.com/go-chi/chi/v5
	go get github.com/go-chi/chi/v5/middleware
	go get github.com/jackc/pgx/v5
	go get github.com/golang-migrate/migrate/v4
	go get github.com/kelseyhightower/envconfig
	go get github.com/joho/godotenv
	go get github.com/google/uuid

# Build
build:
	go build -o bin/click-deploy ./cmd/server

# Docker commands
docker-build:
	docker build -t click-deploy:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

docker-restart:
	docker-compose restart api

# Production helpers
setup-env:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example. Please update with your configuration."; \
	else \
		echo ".env file already exists."; \
	fi

# Run migrations in Docker
docker-migrate-up:
	docker-compose exec api migrate -path migrations/postgres -database "$$DATABASE_URL" up

docker-migrate-down:
	docker-compose exec api migrate -path migrations/postgres -database "$$DATABASE_URL" down
