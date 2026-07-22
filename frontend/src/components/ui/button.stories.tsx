import type { Meta, StoryObj } from "@storybook/react"
import { Button } from "./button"

const meta: Meta<typeof Button> = {
  title: "UI/Button",
  component: Button,
  args: {
    children: "Button",
    disabled: false,
  },
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "outline", "ghost", "destructive"],
    },
    size: {
      control: "select",
      options: ["default", "sm", "lg", "icon"],
    },
  },
}

export default meta
type Story = StoryObj<typeof Button>

export const Default: Story = {
  args: { variant: "default", size: "default" },
}

export const Outline: Story = {
  args: { variant: "outline", size: "default" },
}

export const Ghost: Story = {
  args: { variant: "ghost", size: "default" },
}

export const Destructive: Story = {
  args: { variant: "destructive", size: "default" },
}

export const Small: Story = {
  args: { variant: "default", size: "sm" },
}

export const Large: Story = {
  args: { variant: "default", size: "lg" },
}

export const Icon: Story = {
  args: { variant: "default", size: "icon", children: "✓" },
}

export const Disabled: Story = {
  args: { variant: "default", disabled: true },
}

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button variant="default">Default</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="destructive">Destructive</Button>
      <Button size="sm">Small</Button>
      <Button size="lg">Large</Button>
      <Button size="icon">✓</Button>
      <Button disabled>Disabled</Button>
    </div>
  ),
}
