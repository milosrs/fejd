import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { keycloakify } from 'keycloakify/vite-plugin';

// The theme is built from React (this project) and keycloakify emits it as a
// provider JAR (into dist_keycloak/). keycloak-theme-converter then copies the
// JAR into keycloak/providers/, which docker compose mounts at
// /opt/keycloak/providers. We ship JAR-only, so no loose theme directory.
export default defineConfig({
  plugins: [
    react(),
    keycloakify({
      accountThemeImplementation: 'none',
      themeName: 'fejd',
      // Absolute URL of the Fejd app; the nav bar links back to it. Resolved at
      // runtime from the FEJD_APP_URL environment variable (see docker-compose).
      extraThemeProperties: ['fejdAppUrl=${env.FEJD_APP_URL:https://fejd.fyi}'],
      // keycloakify requires at least one JAR target. The repo mounts the JAR
      // directly, so keep a single target (the "26.x" range).
      keycloakVersionTargets: {
        '22-to-25': false,
        'all-other-versions': true,
      },
    }),
  ],
});
