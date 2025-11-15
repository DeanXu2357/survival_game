# Makefile Quick Reference Guide

This guide provides quick examples for using the project Makefile.

## Quick Start

### First Time Setup

```bash
# 1. Check if all required tools are installed
make check

# 2. Install all dependencies
make install

# 3. Start development servers
make dev
```

## Common Commands

### Development

```bash
# Start both backend and frontend (recommended)
make dev
# Opens:
#   - Backend:  ws://localhost:3033
#   - Frontend: http://localhost:5173

# Start only backend
make backend

# Start only frontend
make frontend
```

### Building

```bash
# Build everything for production
make build

# Build backend binary only
make backend-build
# Output: bin/survival

# Build frontend only
make frontend-build
# Output: frontend/dist/
```

### Testing

```bash
# Run all tests
make test

# Run backend tests with verbose output
make test-backend
```

### Cleaning

```bash
# Clean all build artifacts
make clean

# Clean only backend artifacts
make clean-backend

# Clean only frontend artifacts
make clean-frontend
```

## Detailed Commands

### make help
Displays all available commands with descriptions.

```bash
$ make help
Survival Game - Development Commands

Usage:
  make <target>

General
  help              Display this help message

Development
  dev               Start both backend and frontend in parallel (recommended)
  start             Alias for 'make dev'
  backend           Start backend Go WebSocket server (port 3033)
  frontend          Start frontend Vite dev server (port 5173)
  ...
```

### make check
Verifies that all required tools are installed.

```bash
$ make check
Checking required tools...
✓ Go:    go version go1.24.7 linux/amd64
✓ Node:  v22.21.1
✓ npm:   10.9.4
All required tools are installed
```

### make install
Installs all project dependencies.

```bash
$ make install
Installing backend dependencies...
go: downloading github.com/gorilla/websocket v1.5.3
...
Backend dependencies installed
Installing frontend dependencies...
added 123 packages, and audited 124 packages in 5s
...
Frontend dependencies installed
```

### make dev
Starts both backend and frontend development servers in parallel.

```bash
$ make dev
Starting development servers...
Backend:  http://localhost:3033
Frontend: http://localhost:5173
Starting backend server on port 3033...
Starting frontend dev server on port 5173...

# Backend logs:
2025/11/15 12:00:00 Hub is running...
2025/11/15 12:00:00 WebSocket server starting on port 3033

# Frontend logs:
VITE v7.0.4  ready in 523 ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

### make build
Builds both backend and frontend for production.

```bash
$ make build
Building backend...
Backend built: bin/survival
Building frontend...
vite v7.0.4 building for production...
✓ 234 modules transformed.
Frontend built: frontend/dist/
```

## Workflow Examples

### Typical Development Session

```bash
# Morning: Start fresh
git pull
make install          # Update dependencies
make clean           # Clean old builds
make dev             # Start development servers

# Make code changes...

# Afternoon: Test changes
make test            # Run tests
make build           # Test production build

# Evening: Clean up
make clean           # Remove build artifacts
```

### Quick Bug Fix

```bash
# Start servers
make dev

# Fix the bug in your editor...

# Test the fix
# Backend changes auto-reload with `go run`
# Frontend changes auto-reload with Vite HMR

# Ctrl+C to stop servers
```

### Production Build

```bash
# Clean previous builds
make clean

# Build everything
make build

# Run production server
./bin/survival &

# Serve frontend
cd frontend && npm run preview
```

### Update Dependencies

```bash
# Update all dependencies
make deps-update

# Verify everything still works
make test
make build
```

## Parallel vs Sequential Execution

### Parallel (Default for `make dev`)
```bash
make dev
# Runs backend and frontend simultaneously using `make -j2`
```

### Sequential (Run commands one after another)
```bash
make backend &       # Run backend in background
sleep 2              # Wait for backend to start
make frontend        # Run frontend in foreground
```

## Stopping Services

Since `make dev` runs in the foreground, simply press:
```
Ctrl+C
```

To stop background processes:
```bash
# Find process
ps aux | grep "go run\|vite"

# Kill by PID
kill <PID>

# Or kill all related processes
pkill -f "go run main.go"
pkill -f "vite"
```

## Environment Variables

You can override default ports:

```bash
# Override backend port
BACKEND_PORT=8080 make backend

# Override frontend port
FRONTEND_PORT=3000 make frontend
```

## Troubleshooting

### "command not found: make"

**macOS/Linux**: Install build tools
```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt-get install build-essential

# Fedora/RHEL
sudo yum install make
```

**Windows**: Use WSL or install via Chocolatey
```powershell
choco install make
```

### Port Already in Use

```bash
# Find what's using the port
lsof -i :3033   # Backend
lsof -i :5173   # Frontend

# Kill the process
kill -9 <PID>
```

### Go Dependencies Not Found

```bash
# Clean Go module cache
go clean -modcache

# Reinstall
make install-backend
```

### Frontend Dependencies Issues

```bash
# Remove node_modules and reinstall
rm -rf frontend/node_modules
make install-frontend
```

## Tips & Tricks

### Run in Background

```bash
# Run backend in background
nohup make backend > logs/backend.log 2>&1 &

# Run frontend in background
nohup make frontend > logs/frontend.log 2>&1 &

# View logs
tail -f logs/backend.log
tail -f logs/frontend.log
```

### Watch Tests

```bash
# Use nodemon or similar tool
# Install: npm install -g nodemon

nodemon --exec "make test" --watch internal/
```

### Quick Rebuild

```bash
# Clean and build in one command
make clean && make build
```

## Advanced Usage

### Custom Targets

You can add custom targets to the Makefile for project-specific needs:

```makefile
# Example: Run database migrations
migrate:
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go

# Example: Generate code
generate:
	@echo "Generating code..."
	@go generate ./...
```

### Makefile Variables

The Makefile uses these configurable variables:

```makefile
BACKEND_PORT := 3033
FRONTEND_PORT := 5173
```

Modify these in the Makefile to change default ports.

## See Also

- [README.md](../README.md) - Project overview
- [CLAUDE.md](../CLAUDE.md) - Development guidelines
- [PROTOCOL_SPECIFICATION.md](PROTOCOL_SPECIFICATION.md) - Protocol documentation
