{
  description = "retro-collection-tool development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs = { self, nixpkgs, flake-utils, git-hooks }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        preCommit = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            check-merge-conflicts.enable = true;
            end-of-file-fixer.enable = true;
            trim-trailing-whitespace.enable = true;
            golangci-lint.enable = true;
            gofmt.enable = true;
            shellcheck.enable = true;
            prettier.enable = true;
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
        checks.pre-commit = preCommit;

        devShells.default = pkgs.mkShell {
          inherit (preCommit) shellHook;
          packages = with pkgs; [
            gnumake
            go
            gopls
            golangci-lint
            ripgrep
            nodejs_22
            nodePackages.prettier
            shellcheck
            shfmt
            mkdocs
            findutils
            coreutils
            git
          ] ++ preCommit.enabledPackages;
        };
      });
}
