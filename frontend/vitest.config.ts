import path from "node:path"
import { defineConfig } from "vitest/config"
import { storybookTest } from "@storybook/addon-vitest/vitest-plugin"
import { playwright } from "@vitest/browser-playwright"

// Vitest runs the Storybook browser tests. This mirrors the Storybook addon's
// Vitest 4 template: the storybookTest plugin pulls the React/tailwind config
// from .storybook, so no react() plugin is added here.
export default defineConfig({
  test: {
    projects: [
      {
        extends: true,
        plugins: [
          // Runs the tests for the stories defined in the Storybook config.
          // See: https://storybook.js.org/docs/next/writing-tests/integrations/vitest-addon
          storybookTest({ configDir: path.join(import.meta.dirname, ".storybook") }),
        ],
        test: {
          name: "storybook",
          browser: {
            enabled: true,
            headless: true,
            provider: playwright({}),
            instances: [{ browser: "chromium" }],
          },
        },
      },
    ],
  },
})
