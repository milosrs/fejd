{
  description = "Development environment for fejd";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            nodejs
            corepack
            pnpm
            jdk
            gnumake
            git
            curl
            wget
            openssl
            cacert
            gomod2nix.packages.${system}.default
            pkg-config
            gcc
            postgresql
            nginx
            docker
            docker-compose
            certbot
            keycloak
            android-tools
            gradle
            direnv
            nix-direnv
          ];

          shellHook = ''
            export JAVA_HOME="${pkgs.jdk}"
            export PATH="$PATH:${pkgs.jdk}/bin"
            export PGDATA="$PWD/.tmp/postgres"
            export PGHOST="/tmp"
            mkdir -p "$PWD/.tmp/postgres"
          '';
        };
      }
    );
}
