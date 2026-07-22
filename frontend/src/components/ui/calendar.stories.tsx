import type { Meta, StoryObj } from "@storybook/react"
import { Calendar } from "./calendar"

const meta: Meta<typeof Calendar> = {
  title: "UI/Calendar",
  component: Calendar,
  parameters: {
    layout: "centered",
  },
  argTypes: {
    mode: {
      control: "select",
      options: ["single", "multiple", "range"],
    },
    showOutsideDays: {
      control: "boolean",
    },
  },
}

export default meta
type Story = StoryObj<typeof Calendar>

export const Default: Story = {
  args: {
    mode: "single",
    showOutsideDays: true,
  },
}

export const WithSelectedDate: Story = {
  args: {
    mode: "single",
    showOutsideDays: true,
    selected: new Date(2026, 6, 22),
  },
}

export const DateRange: Story = {
  args: {
    mode: "range",
    showOutsideDays: true,
    selected: {
      from: new Date(2026, 6, 20),
      to: new Date(2026, 6, 25),
    },
  },
}

export const WithDisabledPast: Story = {
  args: {
    mode: "single",
    showOutsideDays: true,
    disabled: (date: Date) => {
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      return date < today
    },
  },
}

export const BookingCalendar: Story = {
  render: () => (
    <div className="border border-border rounded-lg p-4 max-w-sm">
      <h3 className="text-sm font-medium mb-2">Select a date</h3>
      <Calendar
        mode="single"
        showOutsideDays
        disabled={(date) => {
          const today = new Date()
          today.setHours(0, 0, 0, 0)
          return date < today || date.getDay() === 0
        }}
      />
    </div>
  ),
}
