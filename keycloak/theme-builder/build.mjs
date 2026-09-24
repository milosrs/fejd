import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

// Builds the React theme (keycloakify) and publishes the generated FTL into
// keycloak/themes/fejd via the vite.config.ts postBuild hook. Requires Maven on
// PATH (see flake.nix).
const dir = dirname(fileURLToPath(import.meta.url));

execSync("npm run build-keycloak-theme", { stdio: "inherit", cwd: dir });
