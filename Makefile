OS := $(shell uname)

# Path to Go compiler (can be overridden)
GO ?= go

# Version info injection
VERSION := $(shell cat VERSION 2>/dev/null || echo dev)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -s -w -X 'github.com/Cyber-Def/postq-tunnel/internal/version.Version=$(VERSION)' -X 'github.com/Cyber-Def/postq-tunnel/internal/version.BuildTime=$(BUILD_TIME)'

BINARY_CLIENT := release/qtun
BINARY_SERVER := release/qtunnel

.PHONY: all release clean rebuild check-go

all: check-go release

# Main target – builds both binaries
release: $(BINARY_CLIENT) $(BINARY_SERVER)

# Build client binary
$(BINARY_CLIENT): $(shell find ./cmd/qtun -type f)
	@mkdir -p release
	GOOS=$$( $(GO) env GOOS ) GOARCH=$$( $(GO) env GOARCH ) \
	    $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_CLIENT) ./cmd/qtun/main.go
ifeq ($(OS),Darwin)
	codesign -s - -f $(BINARY_CLIENT)
endif

# Build server binary
$(BINARY_SERVER): $(shell find ./cmd/server -type f)
	@mkdir -p release
	GOOS=$$( $(GO) env GOOS ) GOARCH=$$( $(GO) env GOARCH ) \
	    $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_SERVER) ./cmd/server/main.go
ifeq ($(OS),Darwin)
	codesign -s - -f $(BINARY_SERVER)
endif

# Clean generated files
clean:
	@rm -rf release

rebuild: clean release

check-go:
	@command -v $(GO) >/dev/null 2>&1 || { echo "Error: Go compiler not found. Install with 'brew install go' or set GO variable."; exit 1; }
