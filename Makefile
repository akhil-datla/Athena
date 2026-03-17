BINARY    := athena
MODULE    := github.com/akhil-datla/athena
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
BUILD_DIR := bin

.PHONY: all build test lint clean install run

all: test build

build:
	@echo "Building $(BINARY) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/athena

test:
	go test -race -cover ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/athena

run: build
	$(BUILD_DIR)/$(BINARY) serve -f examples/faq.csv
