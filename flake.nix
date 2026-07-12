{
  description = "Development environment for fejd";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
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
            go_1_22
            nodejs_20
            jdk21
            gnumake
            git
            curl
            wget
            openssl
            cacert
            gomod2nix
            nodePackages.npm
            pkg-config
            gcc
            postgresql_16
            nginx
            docker
            docker-compose
            certbot
            keycloak
            android-tools
            gradle
            direnv
            nodePackages.vite
            nodePackages.typescript
            nodePackages.react
            nodePackages.react-dom
            nodePackages.react-router-dom
            nodePackages.zustand
            nodePackages.axios
            nodePackages.vitest
            nodePackages."@vitejs/plugin-react"
            nodePackages."vite-plugin-pwa"
            nodePackages."workbox-window"
            nodePackages."@types/react"
            nodePackages."@types/react-dom"
            nodePackages."@testing-library/react"
            nodePackages."@testing-library/user-event"
            nodePackages."@testing-library/jest-dom"
            nodePackages."keycloak-js"
            nodePackages."@tanstack/react-query"
            nodePackages."@capacitor/core"
            nodePackages."@capacitor/cli"
            nodePackages."@capacitor/android"
            nodePackages."@capacitor/ios"
          ];

          shellHook = ''
            export JAVA_HOME="${pkgs.jdk21}"
            export PATH="$PATH:${pkgs.jdk21}/bin"
            export PGDATA="$PWD/.tmp/postgres"
            export PGHOST="/tmp"
            mkdir -p "$PWD/.tmp/postgres"
          '';
        };
      }
    );
}
