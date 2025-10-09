.PHONY: all build install clean test fmt vet

# Binary output directory
BIN_DIR := bin

# Command directories
COMMANDS := git-commitAll git-evars git-exec git-replicate git-treeconfig git-update

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod
GOINSTALL := $(GOCMD) install

# Build flags
LDFLAGS := -ldflags="-s -w"

all: fmt vet build

build: $(BIN_DIR)
	@echo "Building all commands..."
	@for cmd in $(COMMANDS); do \
		echo "  Building $$cmd..."; \
		$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$$cmd ./cmd/$$cmd || exit 1; \
	done
	@echo "Build complete!"

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

install:
	@echo "Installing all commands to GOPATH/bin..."
	@for cmd in $(COMMANDS); do \
		echo "  Installing $$cmd..."; \
		$(GOINSTALL) $(LDFLAGS) ./cmd/$$cmd || exit 1; \
	done
	@echo "Install complete!"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "Clean complete!"

test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

vet:
	@echo "Checking Go version..."
	@go version | grep -qE 'go1\.(2[4-9]|[3-9][0-9])' || (echo "Error: Go 1.24 or later is required" && exit 1)
	@echo "Vetting code..."
	@$(GOVET) ./...

tidy:
	@echo "Tidying go.mod..."
	@$(GOMOD) tidy

# Individual command builds
git-commitAll: $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/git-commitAll ./cmd/git-commitAll

git-evars: $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/git-evars ./cmd/git-evars

git-exec: $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/git-exec ./cmd/git-exec

git-replicate: $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/git-replicate ./cmd/git-replicate

git-treeconfig: $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/git-treeconfig ./cmd/git-treeconfig

git-update: $(BIN_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/git-update ./cmd/git-update

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Format, vet, and build all commands (default)"
	@echo "  build         - Build all commands to bin/ directory"
	@echo "  clean         - Remove build artifacts"
	@echo "  fmt           - Format all Go code"
	@echo "  help          - Show this help message"
	@echo "  install       - Install all commands to GOPATH/bin"
	@echo "  test          - Run all tests"
	@echo "  tidy          - Tidy go.mod and go.sum"
	@echo "  vet           - Run go vet on all code"
	@echo ""
	@echo "Individual command targets:"
	@for cmd in $(COMMANDS); do \
		echo "  $$cmd"; \
	done
