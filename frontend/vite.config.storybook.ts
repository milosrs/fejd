import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

// Storybook-only Vite config. Storybook must not load vite.config.ts (React
// Router SSR) nor vite.config.spa.ts (VitePWA), so it uses this minimal
// client config: React + Tailwind.
export default defineConfig({
  plugins: [react(), tailwindcss()],
})
