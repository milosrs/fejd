import type { Meta, StoryObj } from "@storybook/react"
import { Select } from "./select"
import { Label } from "./label"

const meta: Meta<typeof Select> = {
  title: "UI/Select",
  component: Select,
  args: {
    disabled: false,
  },
}

export default meta
type Story = StoryObj<typeof Select>

export const Default: Story = {
  render: (args) => (
    <Select {...args}>
      <option value="">-- Select an option --</option>
      <option value="1">Option 1</option>
      <option value="2">Option 2</option>
      <option value="3">Option 3</option>
    </Select>
  ),
}

export const WithLabel: Story = {
  render: (args) => (
    <div className="grid gap-1.5 max-w-sm">
      <Label htmlFor="story-select">Employee</Label>
      <Select id="story-select" {...args}>
        <option value="">-- Select --</option>
        <option value="alice">Alice (Hairstylist)</option>
        <option value="bob">Bob (Barber)</option>
        <option value="carol">Carol (Colorist)</option>
      </Select>
    </div>
  ),
}

export const Disabled: Story = {
  render: (args) => (
    <Select {...args} disabled>
      <option>Disabled select</option>
    </Select>
  ),
}
