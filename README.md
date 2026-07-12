# fejd

App for booking haircut appointment.

## Initial environment setup

This repository uses Nix Flakes and direnv to provide a consistent development environment.

### 1. Install Nix

If you do not already have Nix installed, install it first:

```bash
curl -L https://nixos.org/nix/install | sh -s -- --no-daemon
```

Then enable flakes for your user:

```bash
mkdir -p ~/.config/nix
cat > ~/.config/nix/nix.conf <<'EOF'
experimental-features = nix-command flakes
EOF
```

### 2. Install and enable direnv

Install direnv through Nix:

```bash
nix profile install nixpkgs#direnv nixpkgs#nix-direnv
```

Load the hook in your shell:

```bash
# zsh
echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc

# bash
echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
```

Reload your shell:

```bash
source ~/.zshrc   # or ~/.bashrc
```

### 3. Enter the project environment

From the repository root:

```bash
cd /path/to/fejd
direnv allow .
```

The first time this runs, Nix will build the dev shell from the Flake. After that, entering the directory will automatically activate it.

### 4. Use it in VS Code

If you are using VS Code in WSL, make sure the workspace can see the Nix-managed binaries:

```json
{
  "direnv.path": "/home/rixon/.nix-profile/bin/direnv"
}
```

### 5. What the environment includes

The Flake provides a development shell with:

- Go
- Node.js
- latest LTS JDK
- PostgreSQL
- Nginx
- Docker / Docker Compose
- Keycloak
- Gradle / Android tools
- Corepack and pnpm available directly in the shell
- common build utilities and package managers

### 5. What the environment includes

The Flake provides a development shell with:

- Go
- Node.js
- latest LTS JDK
- PostgreSQL
- Nginx
- Docker / Docker Compose
- Keycloak
- Gradle / Android tools
- common build utilities and package managers

### 6. Useful commands

```bash
nix develop
nix flake show
nix flake update
```

Inside the shell, you can use:

```bash
corepack pnpm --version
corepack pnpm install
corepack pnpm run dev
```

### 7. Keycloak theme builder

The repository includes a project-local theme builder under [keycloak/theme-builder](keycloak/theme-builder). Developers can edit the React-style theme assets there and publish them into [keycloak/themes](keycloak/themes) with:

```bash
./keycloak-theme-converter fejd
```

The generated files are written directly into the Keycloak theme folder that is mounted into the Keycloak container.

### 8. Docker containers

The repository includes Dockerfiles for the backend, frontend, and Keycloak, plus a Compose setup that connects them automatically.

Run from the repository root:

```bash
docker compose up --build
```

This starts:
- backend on http://localhost:8080
- frontend on http://localhost:5173
- Keycloak on http://localhost:9090
- PostgreSQL on localhost:5432

If you are on WSL, Docker Desktop should be reachable through the WSL bridge. If you use Docker Engine directly in Linux, the same compose file works without changes.

If you want to inspect the current environment, run:

```bash
direnv status
```
