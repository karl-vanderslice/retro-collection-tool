{
  description = "retro-collection-tool development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks.url = "github:cachix/git-hooks.nix";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, git-hooks, treefmt-nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        treefmtEval = treefmt-nix.lib.evalModule pkgs {
          projectRootFile = "flake.nix";
          programs = {
            alejandra.enable = true;
            biome = {
              enable = true;
              includes = ["*.json"];
            };
            gofmt.enable = true;
          };
        };

        preCommit = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            check-merge-conflicts.enable = true;
            end-of-file-fixer.enable = true;
            trim-trailing-whitespace.enable = true;
            golangci-lint.enable = true;
            gofmt.enable = true;
            shellcheck.enable = true;
            markdownlint-cli2 = {
              enable = true;
              name = "markdownlint-cli2";
              entry = "${pkgs.markdownlint-cli2}/bin/markdownlint-cli2";
              language = "system";
              files = "\\.md$";
            };
            yamllint = {
              enable = true;
              settings.configuration = ''
                ---
                extends: relaxed
                rules:
                  line-length: disable
              '';
            };
            flake-lock-required = {
              enable = true;
              name = "flake-lock-required";
              entry = "bash scripts/check-flake-lock-tracked.sh";
              language = "system";
              pass_filenames = false;
              always_run = true;
            };
          };
        };
      in {
        formatter = treefmtEval.config.build.wrapper;

        checks = {
          pre-commit = preCommit;
          formatting = treefmtEval.config.build.check self;
        };

        devShells.default = pkgs.mkShell {
          inherit (preCommit) shellHook;
          packages = with pkgs; [
            biome
            gnumake
            go
            gopls
            golangci-lint
            igir
            mame-tools
            markdownlint-cli2
            ripgrep
            nodejs_22
            shellcheck
            shfmt
            yamllint
            zensical
            gh
            findutils
            coreutils
            git
          ] ++ preCommit.enabledPackages;
        };
      });
}
