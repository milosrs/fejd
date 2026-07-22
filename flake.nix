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

        swag = pkgs.buildGoModule {
          pname = "swag";
          version = "1.16.6";
          src = pkgs.fetchFromGitHub {
            owner = "swaggo";
            repo = "swag";
            rev = "v1.16.6";
            hash = "sha256-ixeHj+bqskQJOCxnJaU0IG9Qoe4SQk+McNY0Sy1tUwI=";
          };
            vendorHash = "sha256-P3WH4SrGL4Ejn4U34EEJA21Fne/UlOWg8jiI94Bp7Ms=";
          subPackages = [ "cmd/swag" ];
        };

        openapi-typescript' = pkgs.writeShellApplication {
          name = "openapi-typescript";
          runtimeInputs = [ pkgs.nodejs ];
          text = ''exec npx --yes openapi-typescript@7 "$@"'';
        };

        swagger2openapi' = pkgs.writeShellApplication {
          name = "swagger2openapi";
          runtimeInputs = [ pkgs.nodejs ];
          text = ''exec npx --yes swagger2openapi@7 "$@"'';
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            android-tools
            cacert
            corepack
            curl
            direnv
            docker
            docker-compose
            gcc
            git
            gnumake
            go
            gomod2nix.packages.${system}.default
            gradle
            jdk
            just
            nginx
            nix-direnv
            nodejs
            openssl
            pkg-config
            playwright
            pnpm
            postgresql
            swag
            openapi-typescript'
            swagger2openapi'
          ];

          shellHook = ''
            unset IN_NIX_SHELL
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
