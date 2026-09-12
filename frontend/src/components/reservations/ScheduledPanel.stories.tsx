import type { Meta, StoryObj } from "@storybook/react-vite"
import { CalendarDate } from "@internationalized/date"
import { ScheduledPanel } from "./ScheduledPanel"
import { Button } from "#components/ui/button"
import type { ScheduleEvent } from "./types"

const meta: Meta<typeof ScheduledPanel> = {
  title: "Staff/Scheduled Panel",
  component: ScheduledPanel,
  parameters: {
    layout: "centered",
  },
}

export default meta
type Story = StoryObj<typeof ScheduledPanel>

function at(day: number, hour: number, minute = 0): Date {
  return new Date(2025, 8, day, hour, minute, 0, 0)
}

const demoDate = new CalendarDate(2025, 9, 8)

const demoEvents: ScheduleEvent[] = [
  {
    id: "1",
    title: "English Lesson",
    subtitle: "Online class with tutor.",
    start: at(8, 9, 0),
    end: at(8, 10, 15),
    tone: "orange",
    attendees: ["Alex W", "Ivan M"],
  },
  {
    id: "2",
    title: "Job Interview",
    subtitle: "Frontend Developer position.",
    start: at(8, 10, 0),
    end: at(8, 11, 0),
    tone: "lime",
    meetLink: "https://example.com/meet",
  },
  {
    id: "3",
    title: "Team Sync Call",
    subtitle: "Weekly updates.",
    start: at(8, 13, 0),
    end: at(8, 15, 0),
    tone: "blue",
    attendees: ["Julia K.", "Ivan M."],
    extraAttendees: 5,
  },
]

export const Default: Story = {
  render: () => (
    <div className="w-[420px] rounded-2xl border border-border p-4">
      <ScheduledPanel
        date={demoDate}
        onPrevDay={() => {}}
        onNextDay={() => {}}
        onOpenCalendar={() => {}}
        events={demoEvents}
      />
    </div>
  ),
}

export const Empty: Story = {
  render: () => (
    <div className="w-[420px] rounded-2xl border border-border p-4">
      <ScheduledPanel
        date={demoDate}
        onPrevDay={() => {}}
        onNextDay={() => {}}
        events={[]}
      />
    </div>
  ),
}

export const WithActions: Story = {
  render: () => (
    <div className="w-[420px] rounded-2xl border border-border p-4">
      <ScheduledPanel
        date={demoDate}
        onPrevDay={() => {}}
        onNextDay={() => {}}
        events={demoEvents}
        renderActions={(event) =>
          event.id === "2" ? (
            <Button size="sm">Accept</Button>
          ) : undefined
        }
      />
    </div>
  ),
}
