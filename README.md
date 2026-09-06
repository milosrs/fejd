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

### 9. Native app (Capacitor)

The frontend is a Capacitor app: the same React code runs on web (keycloak-js)
and on native Android/iOS (Capacitor Browser + PKCE). The native projects are
generated and gitignored — regenerate them with:

```bash
cd frontend
pnpm install
npx cap add android   # or: npx cap add ios
```

After changing web code or Capacitor config, sync the web build into the native
projects:

```bash
cd frontend
pnpm build
npx cap sync
```

Run on an emulator/simulator:

```bash
cd frontend
npx cap run android   # requires Android Studio + an AVD
npx cap run ios       # requires macOS + Xcode + a simulator
```

`npx cap open android` / `npx cap open ios` opens the project in the native IDE
instead.

The native auth path needs the `fejd://` URL scheme (already registered in the
scaffolded `AndroidManifest.xml` intent-filter and the iOS `Info.plist`
`CFBundleURLTypes`) and reads Keycloak connection settings from
`VITE_KEYCLOAK_URL`, `VITE_KEYCLOAK_REALM`, `VITE_KEYCLOAK_NATIVE_CLIENT_ID`
(default `salon-mobile`), and `VITE_KEYCLOAK_NATIVE_REDIRECT_URI` (default
`fejd://callback`).

To test the full invite → password-set → app-open handoff on a device/simulator:

1. Start the backend stack: `just up-backend` (Keycloak, DB, backend, SMTP).
2. Invite a user (from `backend/`):

   ```bash
   KEYCLOAK_URL=http://localhost:9090 KEYCLOAK_REALM=fejd \
   KEYCLOAK_ADMIN_CLIENT_ID=fejd-admin \
   KEYCLOAK_ADMIN_CLIENT_SECRET=H6vKp9sQ2wXrT4yL8mN1cB3dV5fG7jZ0 \
   INVITE_REDIRECT_URI=fejd://callback \
   go run ./cmd/fejd-admin invite --email you@example.com --name "You"
   ```

3. Open the invite email (caught by the local SMTP at `docker exec fejd-smtp
   cat /var/mail/catchall.eml`) and open its link on the device — the app opens
   via the `fejd://` scheme, and the `UPDATE_PASSWORD` action sets the password.

Universal/App Link handoff for invites additionally needs the hosted
`apple-app-site-association` (iOS) and `assetlinks.json` (Android) under
`.well-known/` on the app domain — see `frontend/public/.well-known/`.

