# Simple Makefile for a Go project

ATLAS_VERSION ?= v0.35.0
ATLAS ?= atlas
ATLAS_ENV ?= local

# Build the application
all: build

build:
	@echo "Building..."
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

app:
	@cp .env-dist .env
	@ISLOCAL=1 GO111MODULE=auto realize start -name=medilink

# Create DB container
docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/cosmtrek/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

# --- Atlas migrations (versioned; never use `atlas schema apply` on prod) ---

# Install pinned Atlas CLI into ~/.local/bin
install-atlas:
	@mkdir -p "$$HOME/.local/bin"
	@ARCH=$$(uname -m); \
	case "$$ARCH" in x86_64) A=amd64;; arm64|aarch64) A=arm64;; *) echo "unsupported arch $$ARCH"; exit 1;; esac; \
	OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	URL="https://release.ariga.io/atlas/atlas-$${OS}-$${A}-$(ATLAS_VERSION)"; \
	echo "Downloading $$URL"; \
	curl -fsSL "$$URL" -o "$$HOME/.local/bin/atlas"; \
	chmod +x "$$HOME/.local/bin/atlas"; \
	"$$HOME/.local/bin/atlas" version

# Generate a migration from schema/schema.sql → schema/migrations/
# Usage: make migrate-diff NAME=add_foo
migrate-diff:
	@test -n "$(NAME)" || (echo "Usage: make migrate-diff NAME=add_foo"; exit 1)
	@$(ATLAS) migrate diff $(NAME) --env $(ATLAS_ENV)

# Apply pending migrations (ATLAS_ENV=local|prod). Optional: BASELINE=20260901120000
migrate-apply:
	@if [ -n "$(BASELINE)" ]; then \
		$(ATLAS) migrate apply --env $(ATLAS_ENV) --baseline "$(BASELINE)"; \
	else \
		$(ATLAS) migrate apply --env $(ATLAS_ENV); \
	fi

migrate-status:
	@$(ATLAS) migrate status --env $(ATLAS_ENV)

migrate-lint:
	@$(ATLAS) migrate lint --env $(ATLAS_ENV) --latest 1

# Record version without executing SQL (local DBs already ahead of baseline).
# Usage: make migrate-set VERSION=20260909120000
migrate-set:
	@test -n "$(VERSION)" || (echo "Usage: make migrate-set VERSION=20260909120000"; exit 1)
	@$(ATLAS) migrate set $(VERSION) --env $(ATLAS_ENV)

migrate-hash:
	@$(ATLAS) migrate hash --env $(ATLAS_ENV)

.PHONY: all build run test clean install-atlas migrate-diff migrate-apply migrate-status migrate-lint migrate-set migrate-hash
