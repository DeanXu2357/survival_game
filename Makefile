# Survival Game Makefile
# Manages backend (Go) and frontend (TypeScript + Vite) development

.PHONY: help install dev dev-bg start stop backend frontend backend-build frontend-build build test clean logs logs-backend logs-frontend status

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

dev: ## Start both servers in separate terminals (RECOMMENDED - see below)
	@echo "$(YELLOW)⚠️  Recommended approach: Use two separate terminals$(NC)"
	@echo ""
	@echo "$(BLUE)Terminal 1:$(NC)"
	@echo "  $$ make backend"
	@echo ""
	@echo "$(BLUE)Terminal 2:$(NC)"
	@echo "  $$ make frontend"
	@echo ""
	@echo "$(YELLOW)Or use 'make dev-bg' to run both in background$(NC)"

dev-bg: ## Start both servers in background
	@echo "$(GREEN)Starting development servers in background...$(NC)"
	@mkdir -p logs
	@echo "$(BLUE)Starting backend on port $(BACKEND_PORT)...$(NC)"
	@nohup go run main.go > logs/backend.log 2>&1 & echo $$! > .backend.pid
	@sleep 1
	@echo "$(BLUE)Starting frontend on port $(FRONTEND_PORT)...$(NC)"
	@(cd frontend && nohup npm run dev > ../logs/frontend.log 2>&1 & echo $$! > ../.frontend.pid)
	@sleep 2
	@if [ -f .backend.pid ] && [ -f .frontend.pid ]; then \
		echo "$(GREEN)✓ Backend started  (PID: $$(cat .backend.pid))  - http://localhost:$(BACKEND_PORT)$(NC)"; \
		echo "$(GREEN)✓ Frontend started (PID: $$(cat .frontend.pid)) - http://localhost:$(FRONTEND_PORT)$(NC)"; \
	else \
		echo "$(RED)✗ Failed to start servers. Check logs for details.$(NC)"; \
		exit 1; \
	fi
	@echo ""
	@echo "$(YELLOW)View logs:$(NC)"
	@echo "  Backend:  tail -f logs/backend.log"
	@echo "  Frontend: tail -f logs/frontend.log"
	@echo "  Both:     make logs"
	@echo ""
	@echo "$(YELLOW)Check status:$(NC)"
	@echo "  make status"
	@echo ""
	@echo "$(YELLOW)Stop servers:$(NC)"
	@echo "  make stop"

start: dev ## Alias for 'make dev'

backend: ## Start backend Go WebSocket server (port 3033)
	@echo "$(GREEN)Starting backend server on port $(BACKEND_PORT)...$(NC)"
	@go run main.go

frontend: ## Start frontend Vite dev server (port 5173)
	@echo "$(GREEN)Starting frontend dev server on port $(FRONTEND_PORT)...$(NC)"
	@cd frontend && npm run dev

stop: ## Stop all background servers
	@echo "$(YELLOW)Stopping development servers...$(NC)"
	@if [ -f .backend.pid ]; then \
		kill $$(cat .backend.pid) 2>/dev/null || true; \
		rm .backend.pid; \
		echo "$(GREEN)✓ Backend stopped$(NC)"; \
	else \
		echo "$(YELLOW)No backend PID file found$(NC)"; \
	fi
	@if [ -f .frontend.pid ]; then \
		kill $$(cat .frontend.pid) 2>/dev/null || true; \
		rm .frontend.pid; \
		echo "$(GREEN)✓ Frontend stopped$(NC)"; \
	else \
		echo "$(YELLOW)No frontend PID file found$(NC)"; \
	fi

logs: ## Show development logs (tail -f both logs)
	@echo "$(BLUE)Showing development logs (Ctrl+C to stop)...$(NC)"
	@tail -f logs/backend.log logs/frontend.log

logs-backend: ## Show backend logs
	@tail -f logs/backend.log

logs-frontend: ## Show frontend logs
	@tail -f logs/frontend.log

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

status: ## Show status of background servers
	@echo "$(BLUE)Development Servers Status:$(NC)"
	@if [ -f .backend.pid ]; then \
		if ps -p $$(cat .backend.pid) > /dev/null 2>&1; then \
			echo "$(GREEN)✓ Backend running$(NC)  (PID: $$(cat .backend.pid))"; \
		else \
			echo "$(RED)✗ Backend not running$(NC) (stale PID file)"; \
			rm .backend.pid; \
		fi \
	else \
		echo "$(YELLOW)○ Backend not started$(NC)"; \
	fi
	@if [ -f .frontend.pid ]; then \
		if ps -p $$(cat .frontend.pid) > /dev/null 2>&1; then \
			echo "$(GREEN)✓ Frontend running$(NC) (PID: $$(cat .frontend.pid))"; \
		else \
			echo "$(RED)✗ Frontend not running$(NC) (stale PID file)"; \
			rm .frontend.pid; \
		fi \
	else \
		echo "$(YELLOW)○ Frontend not started$(NC)"; \
	fi

##@ Docker (Future)

docker-build: ## Build Docker image (not implemented yet)
	@echo "$(YELLOW)Docker support coming soon...$(NC)"

docker-up: ## Start services with Docker Compose (not implemented yet)
	@echo "$(YELLOW)Docker support coming soon...$(NC)"
