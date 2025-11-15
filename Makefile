# Survival Game Makefile
# Manages backend (Go) and frontend (TypeScript + Vite) development

.PHONY: help install dev start stop backend frontend backend-build frontend-build build test clean

# Default target
.DEFAULT_GOAL := help

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Configuration
BACKEND_PORT := 3033
FRONTEND_PORT := 5173

##@ General

help: ## Display this help message
	@echo "$(BLUE)Survival Game - Development Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(YELLOW)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

dev: ## Start both backend and frontend in parallel (recommended)
	@echo "$(GREEN)Starting development servers...$(NC)"
	@echo "$(BLUE)Backend:$(NC)  http://localhost:$(BACKEND_PORT)"
	@echo "$(BLUE)Frontend:$(NC) http://localhost:$(FRONTEND_PORT)"
	@$(MAKE) -j2 backend frontend

start: dev ## Alias for 'make dev'

backend: ## Start backend Go WebSocket server (port 3033)
	@echo "$(GREEN)Starting backend server on port $(BACKEND_PORT)...$(NC)"
	@cd . && go run main.go

frontend: ## Start frontend Vite dev server (port 5173)
	@echo "$(GREEN)Starting frontend dev server on port $(FRONTEND_PORT)...$(NC)"
	@cd frontend && npm run dev

stop: ## Stop all running servers (requires manual Ctrl+C)
	@echo "$(YELLOW)Press Ctrl+C to stop the servers$(NC)"

##@ Installation

install: install-backend install-frontend ## Install all dependencies (backend + frontend)

install-backend: ## Install backend Go dependencies
	@echo "$(GREEN)Installing backend dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Backend dependencies installed$(NC)"

install-frontend: ## Install frontend npm dependencies
	@echo "$(GREEN)Installing frontend dependencies...$(NC)"
	@cd frontend && npm install
	@echo "$(GREEN)Frontend dependencies installed$(NC)"

##@ Building

build: backend-build frontend-build ## Build both backend and frontend for production

backend-build: ## Build backend Go binary
	@echo "$(GREEN)Building backend...$(NC)"
	@go build -o bin/survival main.go
	@echo "$(GREEN)Backend built: bin/survival$(NC)"

frontend-build: ## Build frontend for production
	@echo "$(GREEN)Building frontend...$(NC)"
	@cd frontend && npm run build
	@echo "$(GREEN)Frontend built: frontend/dist/$(NC)"

##@ Testing

test: test-backend ## Run all tests

test-backend: ## Run backend Go tests
	@echo "$(GREEN)Running backend tests...$(NC)"
	@go test ./... -v

##@ Cleaning

clean: clean-backend clean-frontend ## Clean all build artifacts

clean-backend: ## Clean backend build artifacts
	@echo "$(YELLOW)Cleaning backend build artifacts...$(NC)"
	@rm -rf bin/
	@echo "$(GREEN)Backend cleaned$(NC)"

clean-frontend: ## Clean frontend build artifacts
	@echo "$(YELLOW)Cleaning frontend build artifacts...$(NC)"
	@cd frontend && rm -rf dist/ node_modules/.vite
	@echo "$(GREEN)Frontend cleaned$(NC)"

##@ Production

run-production: ## Run production build (backend binary + serve frontend)
	@echo "$(GREEN)Starting production server...$(NC)"
	@./bin/survival &
	@cd frontend && npm run preview

##@ Utilities

check: ## Check if required tools are installed
	@echo "$(BLUE)Checking required tools...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Error: Go is not installed$(NC)"; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "$(RED)Error: Node.js is not installed$(NC)"; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "$(RED)Error: npm is not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Go:$(NC)    $$(go version)"
	@echo "$(GREEN)✓ Node:$(NC)  $$(node --version)"
	@echo "$(GREEN)✓ npm:$(NC)   $$(npm --version)"
	@echo "$(GREEN)All required tools are installed$(NC)"

deps-update: ## Update all dependencies
	@echo "$(YELLOW)Updating Go dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(YELLOW)Updating npm dependencies...$(NC)"
	@cd frontend && npm update
	@echo "$(GREEN)Dependencies updated$(NC)"

logs-backend: ## Show backend logs (if running in background)
	@echo "$(BLUE)Backend logs:$(NC)"
	@tail -f logs/backend.log 2>/dev/null || echo "$(YELLOW)No log file found. Run backend with logging enabled.$(NC)"

##@ Docker (Future)

docker-build: ## Build Docker image (not implemented yet)
	@echo "$(YELLOW)Docker support coming soon...$(NC)"

docker-up: ## Start services with Docker Compose (not implemented yet)
	@echo "$(YELLOW)Docker support coming soon...$(NC)"
