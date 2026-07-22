import type { Meta, StoryObj } from "@storybook/react"
import { Label } from "./label"
import { Input } from "./input"

const meta: Meta<typeof Label> = {
  title: "UI/Label",
  component: Label,
  args: {
    children: "Label text",
  },
}

export default meta
type Story = StoryObj<typeof Label>

export const Default: Story = {
  args: { children: "Email address" },
}

export const WithInput: Story = {
  render: () => (
    <div className="grid gap-1.5 max-w-sm">
      <Label htmlFor="email">Email address</Label>
      <Input id="email" type="email" placeholder="you@example.com" />
    </div>
  ),
}

export const WithDisabledPeer: Story = {
  render: () => (
    <div className="grid gap-1.5 max-w-sm">
      <Label htmlFor="disabled">Disabled field</Label>
      <Input id="disabled" disabled value="Cannot edit" />
    </div>
  ),
}
