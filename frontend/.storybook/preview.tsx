import type { Preview } from "@storybook/react-vite"
import "../src/styles/globals.css"
import { ThemeProvider } from "../src/components/theme-provider"

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    a11y: {
      // 'todo' - show a11y violations in the test UI only
      // 'error' - fail CI on a11y violations
      // 'off' - skip a11y checks entirely
      test: "todo",
    },
    backgrounds: {
      options: {
        light: { name: "light", value: "#ffffff" },
        dark: { name: "dark", value: "#1c1c1c" }
      }
    },
  },

  decorators: [
    (Story) => (
      <ThemeProvider defaultTheme="dark" storageKey="storybook-theme">
        <Story />
      </ThemeProvider>
    ),
  ],

  initialGlobals: {
    backgrounds: {
      value: "light"
    }
  }
}

export default preview
