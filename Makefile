# Monad Dashboard Makefile
.PHONY: all frontend backend clean install dev build run

# Default target
all: build

# Install dependencies
install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Installing backend dependencies..."
	cd backend && go mod download

# Development server
dev: frontend-dev

frontend-dev:
	@echo "Starting frontend development server..."
	cd frontend && npm run dev

backend-dev:
	@echo "Starting backend development server..."
	cd backend && go run .

# Build frontend
frontend:
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Frontend build completed"

# Build backend with embedded frontend
backend: frontend
	@echo "Copying frontend build to backend..."
	rm -rf backend/frontend
	mkdir -p backend/frontend/dist
	cp -r frontend/dist/* backend/frontend/dist/
	@echo "Building backend with embedded assets..."
	cd backend && go build -o ../monad-dashboard .
	@echo "Backend build completed"

# Complete build
build: backend
	@echo "✅ Monad Dashboard build completed!"
	@echo "Run './monad-dashboard' to start the server"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf frontend/dist
	rm -rf backend/frontend
	rm -f monad-dashboard
	@echo "Clean completed"

# Run the application
run: build
	@echo "Starting Monad Dashboard on http://localhost:8080"
	./monad-dashboard

# Development mode with auto-reload
watch:
	@echo "Starting development mode..."
	@echo "Frontend: http://localhost:5173"
	@echo "Backend: http://localhost:8080"
	@echo ""
	@echo "Run 'make backend-dev' in another terminal"
	make frontend-dev

# Docker build (optional)
docker-build:
	@echo "Building Docker image..."
	docker build -t monad-dashboard .

# Help
help:
	@echo "Monad Dashboard Build System"
	@echo ""
	@echo "Available commands:"
	@echo "  install      - Install all dependencies"
	@echo "  dev          - Start frontend development server"
	@echo "  backend-dev  - Start backend development server"
	@echo "  build        - Build complete application"
	@echo "  run          - Build and run application"
	@echo "  clean        - Clean build artifacts"
	@echo "  watch        - Start development mode with auto-reload"
	@echo "  help         - Show this help message"