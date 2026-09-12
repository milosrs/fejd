import type { Meta, StoryObj } from "@storybook/react-vite"
import { useState } from "react"
import { CalendarDate, getLocalTimeZone, today } from "@internationalized/date"
import { CalendarGrid } from "./CalendarGrid"
import type { ReservationTag } from "./types"

const meta: Meta<typeof CalendarGrid> = {
  title: "Staff/Calendar Grid",
  component: CalendarGrid,
  parameters: {
    layout: "centered",
  },
}

export default meta
type Story = StoryObj<typeof CalendarGrid>

const demoMonth = new CalendarDate(2025, 9, 1)
const demoToday = new CalendarDate(2025, 9, 4)
const demoSelected = new CalendarDate(2025, 9, 8)

const demoTags: Record<string, ReservationTag[]> = {
  "2025-09-01": [
    { id: "1", label: "Finance Meeting", tone: "violet" },
    { id: "2", label: "Weekly Stand-up", tone: "blue" },
  ],
  "2025-09-02": [{ id: "3", label: "Marketing Review", tone: "lime" }],
  "2025-09-03": [{ id: "4", label: "Call with Client", tone: "sky" }],
  "2025-09-04": [{ id: "5", label: "Project Deadline", tone: "rose" }],
  "2025-09-08": [
    { id: "6", label: "Yoga Session", tone: "teal" },
    { id: "7", label: "Travel Planning", tone: "amber" },
  ],
  "2025-09-11": [{ id: "8", label: "Finance Meeting", tone: "violet" }],
  "2025-09-15": [{ id: "9", label: "Weekly Stand-up", tone: "blue" }],
  "2025-09-22": [{ id: "10", label: "Yoga Session", tone: "teal" }],
  "2025-09-26": [{ id: "11", label: "Project Deadline", tone: "rose" }],
}

export const Default: Story = {
  render: () => (
    <div className="w-[560px] rounded-2xl border border-border p-4">
      <CalendarGrid
        month={demoMonth}
        onMonthChange={() => {}}
        selected={demoSelected}
        onSelect={() => {}}
        today={demoToday}
        eventsByDay={demoTags}
      />
    </div>
  ),
}

export const EmptyMonth: Story = {
  render: () => (
    <div className="w-[560px] rounded-2xl border border-border p-4">
      <CalendarGrid
        month={demoMonth}
        onMonthChange={() => {}}
        selected={null}
        onSelect={() => {}}
        today={demoToday}
      />
    </div>
  ),
}

export const Interactive: Story = {
  render: () => {
    const todayDate = today(getLocalTimeZone())
    const [month, setMonth] = useState(
      () => new CalendarDate(todayDate.year, todayDate.month, 1),
    )
    const [selected, setSelected] = useState(todayDate)

    return (
      <div className="w-[560px] rounded-2xl border border-border p-4">
        <CalendarGrid
          month={month}
          onMonthChange={setMonth}
          selected={selected}
          onSelect={(d) => {
            setSelected(d)
            if (d.year !== month.year || d.month !== month.month) {
              setMonth(new CalendarDate(d.year, d.month, 1))
            }
          }}
          today={todayDate}
          eventsByDay={{}}
        />
      </div>
    )
  },
}
