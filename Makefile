.PHONY: build run clean test

build:
	@echo "Building cy_crawler..."
	go build -o bin/cy_crawler ./cmd/cy_crawler

run: build
	@echo "Running cy_crawler..."
	./bin/cy_crawler

clean:
	@echo "Cleaning..."
	rm -rf bin/ logs/

test:
	@echo "Running tests..."
	go test ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

install-python-deps:
	@echo "Installing Python dependencies..."
	pip install -r scripts/requirements.txt

setup: deps install-python-deps
	@echo "Setup completed"

.DEFAULT_GOAL := build