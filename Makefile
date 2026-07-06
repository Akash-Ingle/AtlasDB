.PHONY: all build test lint clean dev dev-down migrate seed help

# ── Variables ───────────────────────────────────────────────────────
GO         := go
GOFLAGS    := -race
BINARY_API := bin/api-server
BINARY_PROC:= bin/processor

# ── Build ───────────────────────────────────────────────────────────
all: build

build: ## Build all binaries
	$(GO) build -o $(BINARY_API)  ./cmd/api-server
	$(GO) build -o $(BINARY_PROC) ./cmd/processor

build-api: ## Build API server only
	$(GO) build -o $(BINARY_API) ./cmd/api-server

build-processor: ## Build processor only
	$(GO) build -o $(BINARY_PROC) ./cmd/processor

# ── Test ────────────────────────────────────────────────────────────
test: ## Run all tests with race detector
	$(GO) test $(GOFLAGS) ./...

test-v: ## Run tests verbose
	$(GO) test $(GOFLAGS) -v ./...

test-cover: ## Run tests with coverage
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# ── Lint ────────────────────────────────────────────────────────────
lint: ## Run linter
	golangci-lint run ./...

# ── Docker ──────────────────────────────────────────────────────────
dev: ## Start full dev environment
	docker compose up --build -d

dev-down: ## Stop dev environment
	docker compose down

dev-logs: ## Tail logs
	docker compose logs -f

dev-ps: ## Show running containers
	docker compose ps

infra: ## Start only infrastructure (postgres, redis, prometheus, grafana)
	docker compose up -d postgres redis prometheus grafana

# ── Database ────────────────────────────────────────────────────────
migrate: ## Run database migrations
	$(GO) run ./cmd/api-server migrate

psql: ## Connect to postgres
	docker compose exec postgres psql -U atlas -d atlasdb

# ── Seed ────────────────────────────────────────────────────────────
seed: ## Generate sample events
	$(GO) run ./scripts/generate-events.go

# ── Clean ───────────────────────────────────────────────────────────
clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

# ── Help ────────────────────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
