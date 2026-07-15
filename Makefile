APP ?= candy
PKG ?= ./cmd/release
BIN_DIR ?= bin
LDFLAGS ?=
MACOS_PLATFORMS ?= darwin/amd64 darwin/arm64
LINUX_PLATFORMS ?= linux/386 linux/amd64 linux/arm linux/arm64 linux/riscv64
WINDOWS_PLATFORMS ?= windows/386 windows/amd64 windows/arm windows/arm64
PLATFORMS ?= $(MACOS_PLATFORMS) $(LINUX_PLATFORMS) $(WINDOWS_PLATFORMS)

.PHONY: all build build-all clean list-platforms test

all: build-all

build:
	@mkdir -p $(BIN_DIR)
	@echo "building $(BIN_DIR)/$(APP)"
	@go build -trimpath -ldflags "$(LDFLAGS)" -o "$(BIN_DIR)/$(APP)" $(PKG)

build-all:
	@mkdir -p $(BIN_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		goos=$${platform%/*}; \
		goarch=$${platform#*/}; \
		ext=""; \
		if [ "$$goos" = "windows" ]; then ext=".exe"; fi; \
		out_dir="$(BIN_DIR)/$$goos/$$goarch"; \
		out="$$out_dir/$(APP)$$ext"; \
		mkdir -p "$$out_dir"; \
		echo "building $$out"; \
		GOOS=$$goos GOARCH=$$goarch CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o "$$out" $(PKG); \
	done

list-platforms:
	@printf '%s\n' $(PLATFORMS)

test:
	@go test ./...

clean:
	@rm -rf "$(BIN_DIR)"
