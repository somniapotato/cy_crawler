# 定义变量
BINARY_NAME=cy_crawler
BIN_DIR=bin
LOG_DIR=logs
GO=go

# 检测操作系统并设置相应的命令
ifeq ($(OS),Windows_NT)
    BINARY_NAME:=$(BINARY_NAME).exe
    RM=rd /s /q
    MKDIR=if not exist
else
    RM=rm -rf
    MKDIR=mkdir -p
endif

.PHONY: build run clean test deps install-python-deps setup

build:
	@echo "Building $(BINARY_NAME)..."
	@$(MKDIR) $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/cy_crawler

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BIN_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@if [ -d "$(BIN_DIR)" ]; then $(RM) $(BIN_DIR); fi
	@if [ -d "$(LOG_DIR)" ]; then $(RM) $(LOG_DIR); fi

test:
	@echo "Running tests..."
	$(GO) test ./...

deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

install-python-deps:
	@echo "Installing Python dependencies..."
	@which pip >/dev/null 2>&1 || (echo "Error: pip not found" && exit 1)
	@if [ -f "scripts/requirements.txt" ]; then \
		pip install -r scripts/requirements.txt; \
	else \
		echo "Warning: scripts/requirements.txt not found"; \
	fi

setup: deps install-python-deps
	@echo "Setup completed"

# 交叉编译目标
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN_DIR)/$(BINARY_NAME)-linux ./cmd/cy_crawler

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BIN_DIR)/$(BINARY_NAME)-windows.exe ./cmd/cy_crawler

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BIN_DIR)/$(BINARY_NAME)-darwin ./cmd/cy_crawler

# 构建所有平台
build-all: clean
	@$(MKDIR) $(BIN_DIR)
	$(MAKE) build-linux
	$(MAKE) build-windows
	$(MAKE) build-darwin

.DEFAULT_GOAL := build