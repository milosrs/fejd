import type { Meta, StoryObj } from "@storybook/react-vite"
import { SideDrawer } from "./drawer"

const meta: Meta<typeof SideDrawer> = {
  title: "UI/SideDrawer",
  component: SideDrawer,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof SideDrawer>

export const Open: Story = {
  render: () => (
    <SideDrawer open onClose={() => {}} title="Navigation">
      <nav className="flex flex-col gap-1 p-2">
        <button className="rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted">
          Home
        </button>
        <button className="rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted">
          Services
        </button>
        <button className="rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted">
          Barbers
        </button>
      </nav>
    </SideDrawer>
  ),
}
