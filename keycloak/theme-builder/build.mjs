import { execSync } from "node:child_process";
import { cpSync, existsSync, rmSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

// Builds the React theme (keycloakify) and publishes the generated FTL into
// keycloak/themes/fejd via the vite.config.ts postBuild hook. Requires Maven on
// PATH (see flake.nix).
const dir = dirname(fileURLToPath(import.meta.url));

execSync("npm run build-keycloak-theme", { stdio: "inherit", cwd: dir });

// The email theme is hand-written FTL (keycloakify only generates the login
// theme) and lives in ./email. Copy it into the built theme directory, which
// is mounted into the Keycloak container. Rebuild after editing ./email.
const emailSrc = join(dir, "email");
const emailDest = join(dir, "..", "themes", "fejd", "email");
if (existsSync(emailSrc)) {
  rmSync(emailDest, { recursive: true, force: true });
  cpSync(emailSrc, emailDest, { recursive: true });
}
