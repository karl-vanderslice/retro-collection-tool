SHELL := bash

BINARY := retro-collection-tool
PKG := ./cmd/retro-collection-tool
BIN_DIR := bin
NIX := nix

# Re-enter the same Make target inside `nix develop` unless we're already in a nix shell.
define ensure_nix_shell
	@if [ -z "$$IN_NIX_SHELL" ]; then \
		echo "[hermetic] entering nix develop for target '$@'"; \
		exec $(NIX) develop path:. --accept-flake-config -c $(MAKE) $@ IN_NIX_SHELL=1; \
	fi
endef

.PHONY: all
all: build

.PHONY: build
build:
	$(ensure_nix_shell)
	go build -trimpath -ldflags "-s -w" -o $(BIN_DIR)/$(BINARY) $(PKG)

.PHONY: run
run:
	$(ensure_nix_shell)
	go run $(PKG) --help

.PHONY: test
test:
	$(ensure_nix_shell)
	go test ./...

.PHONY: pre-commit
pre-commit:
	$(ensure_nix_shell)
	pre-commit run --all-files

.PHONY: hooks-install
hooks-install:
	$(ensure_nix_shell)
	pre-commit install --install-hooks

.PHONY: fmt
fmt:
	$(ensure_nix_shell)
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')
	prettier --write "**/*.{md,json,yaml,yml}"

.PHONY: lint
lint:
	$(ensure_nix_shell)
	golangci-lint run ./...
	shellcheck scripts/*.sh
	prettier --check "**/*.{md,json,yaml,yml}"

.PHONY: tidy
tidy:
	$(ensure_nix_shell)
	go mod tidy

.PHONY: docs-serve
docs-serve:
	$(ensure_nix_shell)
	mkdocs serve

.PHONY: docs-build
docs-build:
	$(ensure_nix_shell)
	mkdocs build

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) dist site coverage.out
