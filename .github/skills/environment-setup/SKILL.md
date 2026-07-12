---
name: fejd-environment-setup
description: Use when working on the FEJD local development environment, Nix/direnv shell, Docker Compose stack, or Keycloak theme build workflow.
---

# FEJD environment setup

## Purpose
This repository is intended to work through a reproducible local development environment based on Nix Flakes, direnv, Docker Compose, and a local Keycloak theme build pipeline.

## Core workflow
Developers should not rely on manual Linux exports or ad-hoc package installs. The repository root contains [.envrc](../../../.envrc), which activates the Nix flake automatically when the project directory is entered.

### 1. Enter the environment
Run this once from the repository root:

```bash
direnv allow .
```

After that, entering the repository will automatically load the dev shell. If needed, the explicit fallback is:

```bash
nix develop
```

## What the flake provides
The flake in [flake.nix](../../../flake.nix) provides a development shell with:

- Go
- Node.js
- Corepack and pnpm
- Java and the JDK
- Gradle and Android tools
- PostgreSQL
- Nginx
- Docker and Docker Compose
- Keycloak
- direnv and nix-direnv
- common build utilities such as curl, wget, openssl, gcc, make, and pkg-config

The shell hook also sets:

- JAVA_HOME to the JDK provided by Nix
- PGDATA to .tmp/postgres
- PGHOST to /tmp
- a local postgres data directory under .tmp/postgres

## Repository services
The compose file in [docker-compose.yml](../../../docker-compose.yml) wires up these services:

- backend
- frontend
- PostgreSQL
- Keycloak

The stack is designed to work over Docker’s internal network, so services can reach each other by container name without extra host configuration.

### Start everything
```bash
docker compose up --build
```

### Stop everything
```bash
docker compose down
```

## Local service ports
- backend: http://localhost:8080
- frontend: http://localhost:5173
- Keycloak: http://localhost:9090
- PostgreSQL: localhost:5432

## Keycloak theme workflow
The repository includes a repo-local Keycloak theme pipeline:

- Theme source builder: [keycloak/theme-builder](../../../keycloak/theme-builder)
- Build/publish script: [keycloak-theme-converter](../../../keycloak-theme-converter)
- Mounted theme output directory: [keycloak/themes](../../../keycloak/themes)
- Mounted provider directory: [keycloak/providers](../../../keycloak/providers)

To build and publish the theme into the mounted Keycloak theme directory:

```bash
./keycloak-theme-converter fejd
```

This is the preferred mechanism for making changes visible inside the Keycloak container.

## Helpful commands
```bash
direnv status
nix flake show
nix flake update
corepack pnpm --version
corepack pnpm install
```

## VS Code notes
If you are using VS Code with direnv, make sure the shell integration and direnv binary are available in your PATH. The environment is designed to be activated automatically when the workspace is opened.
