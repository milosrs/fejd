import type { StorybookConfig } from "@storybook/react-vite"

const config: StorybookConfig = {
  stories: ["../src/**/*.stories.@(js|jsx|mjs|ts|tsx)"],
  addons: [
    "@storybook/addon-docs",
    "@storybook/addon-a11y",
    "@storybook/addon-vitest"
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {
      builder: {
        // Storybook must not load vite.config.ts (React Router SSR) nor
        // vite.config.spa.ts (VitePWA); it uses a minimal client-only config.
        viteConfigPath: "vite.config.storybook.ts",
      },
    },
  },
}

export default config
