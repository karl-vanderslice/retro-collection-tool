{
  description = "retro-collection-tool development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            gnumake
            go
            gopls
            golangci-lint
            nodejs_22
            nodePackages.prettier
            shellcheck
            shfmt
            mkdocs
            findutils
            coreutils
            git
          ];
        };
      });
}
