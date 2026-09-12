import type { Meta, StoryObj } from "@storybook/react-vite"
import { Loader } from "./Loader"

const meta: Meta<typeof Loader> = {
  title: "UI/Loader",
  component: Loader,
  parameters: {
    layout: "fullscreen",
  },
  argTypes: {
    speed: {
      control: { type: "number", min: 0.5, max: 8, step: 0.1 },
      description: "Breathing cycle duration in seconds",
    },
    scale: {
      control: { type: "number", min: 1, max: 1.3, step: 0.01 },
      description: "Peak breathing scale factor",
    },
    label: { control: "text" },
    fullScreen: { control: "boolean" },
  },
  args: {
    speed: 2.1,
    scale: 1.2,
    label: "Loading…",
    fullScreen: true,
  },
}

export default meta
type Story = StoryObj<typeof Loader>

export const Default: Story = {}

export const Inline: Story = {
  parameters: {
    layout: "centered",
  },
  args: {
    fullScreen: false,
  },
  render: (args) => (
    <div className="flex h-80 w-96 items-center justify-center rounded-xl border border-border">
      <Loader {...args} />
    </div>
  ),
}
