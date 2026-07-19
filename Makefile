APP ?= caramel
PKG ?= ./cmd/release
BIN_DIR ?= bin
INSTALL_DIR ?= $(HOME)/.local/bin
LDFLAGS ?=
RELEASE_BUILD := $(findstring r,$(firstword $(MAKEFLAGS)))
MACOS_PLATFORMS ?= darwin/amd64 darwin/arm64
LINUX_PLATFORMS ?= linux/386 linux/amd64 linux/arm linux/arm64 linux/riscv64
WINDOWS_PLATFORMS ?= windows/386 windows/amd64 windows/arm windows/arm64
PLATFORMS ?= $(MACOS_PLATFORMS) $(LINUX_PLATFORMS) $(WINDOWS_PLATFORMS)

.PHONY: all build build-all install clean list-platforms test

all: build-all

build:
	$(if $(RELEASE_BUILD),@go run ./cmd/version "$(CURDIR)/pkg/branding/version.gen.go")
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

install:
	@set -e; \
	goos=$$(go env GOOS); \
	goarch=$$(go env GOARCH); \
	ext=""; \
	if [ "$$goos" = "windows" ]; then ext=".exe"; fi; \
	out_dir="$(BIN_DIR)/$$goos/$$goarch"; \
	out="$$out_dir/$(APP)$$ext"; \
	install_path="$(INSTALL_DIR)/$(APP)$$ext"; \
	mkdir -p "$$out_dir" "$(INSTALL_DIR)"; \
	echo "building $$out"; \
	GOOS=$$goos GOARCH=$$goarch go build -trimpath -ldflags "$(LDFLAGS)" -o "$$out" $(PKG); \
	cp "$$out" "$$install_path"; \
	chmod 0755 "$$install_path"; \
	echo "installed $$install_path"; \
	case ":$$PATH:" in \
		*:"$(INSTALL_DIR)":*) ;; \
		*) echo "warning: $(INSTALL_DIR) is not in PATH";; \
	esac

list-platforms:
	@printf '%s\n' $(PLATFORMS)

test:
	@go test ./...

clean:
	@rm -rf "$(BIN_DIR)"
