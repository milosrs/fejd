import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { keycloakify } from 'keycloakify/vite-plugin';
import path from 'node:path';
import fs from 'node:fs';

// The theme is built from React (this project) and keycloakify emits the FTL
// theme directly into the repo's mounted theme directory: keycloak/themes/fejd.
export default defineConfig({
  plugins: [
    react(),
    keycloakify({
      accountThemeImplementation: 'none',
      themeName: 'fejd',
      // Absolute URL of the Fejd app; the nav bar links back to it. Resolved at
      // runtime from the FEJD_APP_URL environment variable (see docker-compose).
      extraThemeProperties: ['fejdAppUrl=${env.FEJD_APP_URL:https://fejd.fyi}'],
      keycloakifyBuildDirPath: path.resolve(process.cwd(), '../themes'),
      // keycloakify requires at least one JAR target; keep a single one (the
      // repo mounts the theme directory directly, so the JAR is only a build
      // by-product and is gitignored).
      keycloakVersionTargets: {
        '22-to-25': false,
        'all-other-versions': true,
      },
      // keycloakify stages the generated theme under
      // <keycloakifyBuildDirPath>/resources/theme/<themeName> and deletes it
      // after packaging the JARs. Copy it out to the mounted theme directory
      // before that cleanup runs.
      postBuild: async (buildContext) => {
        const themeName = buildContext.themeNames[0];
        const srcDir = path.join(process.cwd(), 'theme', themeName);
        const destDir = path.join(buildContext.keycloakifyBuildDirPath, themeName);
        fs.rmSync(destDir, { recursive: true, force: true });
        fs.cpSync(srcDir, destDir, { recursive: true });
      },
    }),
  ],
});
