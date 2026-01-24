APP_NAME=api
CMD_DIR=cmd/api
TMP_DIR=tmp
LOG_DIR=logs
DOCS_DIR=internal/app/docs

AIR_BIN=$(shell go env GOPATH)/bin/air
SWAG_BIN=$(shell go env GOPATH)/bin/swag

.PHONY: dev build run clean test lint docs docker-build docker-up docker-down docker-logs logs-up logs-down

## 📚 Generar documentación Swagger
docs:
	@echo "📚 Generating API documentation..."
	@if ! command -v $(SWAG_BIN) > /dev/null; then \
		echo "📦 Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	$(SWAG_BIN) init -g $(CMD_DIR)/main.go -o $(DOCS_DIR) --parseDependency --parseInternal
	@./scripts/fix-swagger.sh

## 🔥 Desarrollo con hot reload
dev: docs
	@echo "🔥 Starting development server with hot reload..."
	@if ! command -v $(AIR_BIN) > /dev/null; then \
		echo "📦 Installing Air..."; \
		go install github.com/cosmtrek/air@v1.49.0; \
	fi
	@mkdir -p $(TMP_DIR) $(LOG_DIR)
	$(AIR_BIN) -c .air.toml 2>&1 | stdbuf -oL tee $(LOG_DIR)/app.log

## 🏗️ Build manual
build: docs
	@echo "🏗️ Building binary..."
	go build -o $(TMP_DIR)/$(APP_NAME) ./$(CMD_DIR)

## ▶️ Run sin hot reload
run: docs build
	@echo "▶️ Running binary..."
	./$(TMP_DIR)/$(APP_NAME)

## 🧹 Limpieza
clean:
	@echo "🧹 Cleaning artifacts..."
	rm -rf $(TMP_DIR) $(LOG_DIR)

## 🧪 Tests
test:
	go test ./... -race -count=1

## 🔍 Lint (requiere golangci-lint)
lint:
	golangci-lint run

## 🐳 Docker build
docker-build: docs
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME) .

## 🚀 Docker up (producción con Traefik)
docker-up:
	@echo "🚀 Starting containers..."
	@touch acme.json && chmod 600 acme.json
	docker compose up -d

## 🛑 Docker down
docker-down:
	@echo "🛑 Stopping containers..."
	docker compose down

## 📋 Docker logs
docker-logs:
	docker compose logs -f api

## 🔄 Docker restart
docker-restart: docker-down docker-up

## 📊 Observability up (Loki + Grafana)
logs-up:
	@echo "📊 Starting Loki + Grafana..."
	@mkdir -p $(LOG_DIR)
	docker compose -f docker-compose.observability.yml up -d
	@echo "✅ Grafana: http://localhost:3000 (admin/admin)"

## 📊 Observability down
logs-down:
	@echo "📊 Stopping Loki + Grafana..."
	docker compose -f docker-compose.observability.yml down

## 📊 Observability logs
logs-view:
	docker compose -f docker-compose.observability.yml logs -f

## 🌱 Database seed (super admins)
seed:
	@echo "🌱 Running database seeds..."
	go run ./cmd/seed
