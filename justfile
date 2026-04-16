set shell := ["bash", "-eu", "-o", "pipefail", "-c"]
set quiet

binary := "retro-collection-tool"
pkg := "./cmd/retro-collection-tool"
bin_dir := "bin"

default:
    @just --list

build:
    nix develop -c go build -trimpath -ldflags "-s -w" -o {{bin_dir}}/{{binary}} {{pkg}}

run:
    nix develop -c go run {{pkg}} --help

test:
    nix develop -c go test ./...

pre-commit:
    nix develop -c pre-commit run --all-files

hooks-install:
    nix develop -c pre-commit install --install-hooks

fmt:
    nix fmt

lint:
    nix develop -c golangci-lint run ./...
    nix develop -c shellcheck scripts/*.sh
    nix fmt -- --fail-on-change

tidy:
    nix develop -c go mod tidy

docs-serve:
    nix develop -c zensical serve

docs-build:
    nix develop -c zensical build

clean:
    rm -rf {{bin_dir}} dist site coverage.out
