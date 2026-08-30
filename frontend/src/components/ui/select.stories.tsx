import type { Meta, StoryObj } from "@storybook/react"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select"
import { Label } from "./label"

const meta: Meta<typeof Select> = {
  title: "UI/Select",
  component: Select,
  parameters: {
    layout: "centered",
  },
}

export default meta
type Story = StoryObj<typeof Select>

export const Default: Story = {
  render: () => (
    <Select className="w-56">
      <SelectTrigger>
        <SelectValue>
          {(state) => state.selectedText || "Select an option"}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectItem id="1">Option 1</SelectItem>
        <SelectItem id="2">Option 2</SelectItem>
        <SelectItem id="3">Option 3</SelectItem>
      </SelectContent>
    </Select>
  ),
}

export const WithLabel: Story = {
  render: () => (
    <div className="grid gap-1.5 max-w-sm">
      <Label>Employee</Label>
      <Select className="w-56">
        <SelectTrigger>
          <SelectValue>
            {(state) => state.selectedText || "-- Select --"}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem id="alice">Alice (Hairstylist)</SelectItem>
          <SelectItem id="bob">Bob (Barber)</SelectItem>
          <SelectItem id="carol">Carol (Colorist)</SelectItem>
        </SelectContent>
      </Select>
    </div>
  ),
}

export const Disabled: Story = {
  render: () => (
    <Select className="w-56" isDisabled>
      <SelectTrigger>
        <SelectValue>
          {(state) => state.selectedText || "Disabled select"}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectItem id="1">Disabled</SelectItem>
      </SelectContent>
    </Select>
  ),
}
