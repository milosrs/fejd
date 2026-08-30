import type { Meta, StoryObj } from "@storybook/react"
import { Calendar, RangeCalendar } from "./calendar"
import { CalendarDate, getLocalTimeZone, today } from "@internationalized/date"

const meta: Meta<typeof Calendar> = {
  title: "UI/Calendar",
  component: Calendar,
  parameters: {
    layout: "centered",
  },
  argTypes: {
    captionLayout: {
      control: "select",
      options: ["label", "dropdown"],
    },
    numberOfMonths: {
      control: "number",
      min: 1,
      max: 3,
    },
  },
}

export default meta
type Story = StoryObj<typeof Calendar>

export const Default: Story = {
  args: {},
}

export const WithSelectedDate: Story = {
  args: {
    value: new CalendarDate(2026, 7, 22),
  },
}

export const WithDropdownCaption: Story = {
  args: {
    captionLayout: "dropdown",
  },
}

export const WithDisabledPast: Story = {
  args: {
    minValue: today(getLocalTimeZone()),
  },
}

export const MultipleMonths: Story = {
  args: {
    numberOfMonths: 2,
  },
}

export const BookingCalendar: Story = {
  render: () => (
    <div className="border border-border rounded-lg p-4 max-w-sm">
      <h3 className="text-sm font-medium mb-2">Select a date</h3>
      <Calendar
        minValue={today(getLocalTimeZone())}
        isDateUnavailable={(date) =>
          new Date(date.year, date.month - 1, date.day).getDay() === 0
        }
      />
    </div>
  ),
}

export const Range: Story = {
  render: () => (
    <RangeCalendar
      minValue={today(getLocalTimeZone())}
      defaultValue={{
        start: new CalendarDate(2026, 7, 20),
        end: new CalendarDate(2026, 7, 25),
      }}
    />
  ),
}
