import { defineConfig } from "vite"
import { reactRouter } from "@react-router/dev/vite"
import tailwindcss from "@tailwindcss/vite"

// Framework-mode (SSR) build for the web. Native/Capacitor uses
// vite.config.spa.ts (client-only, react() + VitePWA); Storybook uses
// vite.config.storybook.ts (client-only, react() + Tailwind).
export default defineConfig({
  plugins: [reactRouter(), tailwindcss()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: {
      "/api": {
        target: process.env.VITE_API_PROXY_TARGET || "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
})
