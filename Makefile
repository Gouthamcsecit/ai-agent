.PHONY: help build run up down logs clean test go-run python-run dashboard-run

help:
	@echo "AI Agent Evaluation Pipeline - Go + Python"
	@echo "============================================"
	@echo ""
	@echo "Available commands:"
	@echo "  make build          - Build all Docker images"
	@echo "  make up             - Start all services with Docker Compose"
	@echo "  make down           - Stop all services"
	@echo "  make logs           - View logs"
	@echo "  make clean          - Clean up containers and volumes"
	@echo "  make test           - Run tests"
	@echo "  make go-run         - Run Go API locally"
	@echo "  make python-run     - Run Python evaluator locally"
	@echo "  make dashboard-run  - Run Streamlit dashboard locally"
	@echo "  make sample-data    - Generate sample data"

build:
	docker-compose build

up:
	docker-compose up -d
	@echo ""
	@echo "Services started!"
	@echo "================="
	@echo "Go API:           http://localhost:8080"
	@echo "Python Evaluator: http://localhost:8081"
	@echo "Dashboard:        http://localhost:8501"
	@echo ""

down:
	docker-compose down

logs:
	docker-compose logs -f

clean:
	docker-compose down -v
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	rm -rf bin/

test:
	@echo "Running Go tests..."
	go test ./...
	@echo ""
	@echo "Running Python tests..."
	cd python_evaluator && python -m pytest

# Local development commands
go-run:
	go run ./cmd/api

python-run:
	cd python_evaluator && python -m uvicorn main:app --reload --port 8081

dashboard-run:
	streamlit run dashboard/app.py

# Go commands
go-build:
	go build -o bin/api ./cmd/api

go-tidy:
	go mod tidy

# Sample data
sample-data:
	python scripts/sample_data.py
