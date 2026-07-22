import type { Meta, StoryObj } from "@storybook/react"
import { Input } from "./input"
import { Label } from "./label"

const meta: Meta<typeof Input> = {
  title: "UI/Input",
  component: Input,
  args: {
    placeholder: "Enter value...",
    disabled: false,
  },
  argTypes: {
    type: {
      control: "select",
      options: ["text", "email", "password", "number", "date", "time"],
    },
  },
}

export default meta
type Story = StoryObj<typeof Input>

export const Text: Story = {
  args: { type: "text", placeholder: "Enter your name" },
}

export const Email: Story = {
  args: { type: "email", placeholder: "email@example.com" },
}

export const Password: Story = {
  args: { type: "password", placeholder: "Enter password" },
}

export const Number: Story = {
  args: { type: "number", placeholder: "0" },
}

export const Time: Story = {
  args: { type: "time" },
}

export const Disabled: Story = {
  args: { disabled: true, value: "Disabled input" },
}

export const WithLabel: Story = {
  render: (args) => (
    <div className="grid gap-1.5">
      <Label htmlFor="story-input">Label</Label>
      <Input id="story-input" {...args} />
    </div>
  ),
  args: { placeholder: "Input with label" },
}
