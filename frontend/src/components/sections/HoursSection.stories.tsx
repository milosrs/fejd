import type { Meta, StoryObj } from "@storybook/react-vite"
import { HoursSection } from "./HoursSection"
import { SalonFrame, mockClosures, mockWorkingHours } from "../../stories/salon"

const meta: Meta<typeof HoursSection> = {
  title: "Salon/Sections/Hours",
  component: HoursSection,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof HoursSection>

const heading = "Opening hours"

export const OpenNow: Story = {
  render: () => (
    <SalonFrame workingHours={mockWorkingHours} closures={mockClosures}>
      <HoursSection
        content={{ heading }}
        now={new Date("2026-01-05T12:00:00")}
      />
    </SalonFrame>
  ),
}

export const ClosedNow: Story = {
  render: () => (
    <SalonFrame workingHours={mockWorkingHours} closures={mockClosures}>
      <HoursSection
        content={{ heading }}
        now={new Date("2026-01-05T20:00:00")}
      />
    </SalonFrame>
  ),
}

export const ClosedDay: Story = {
  render: () => (
    <SalonFrame workingHours={mockWorkingHours} closures={mockClosures}>
      <HoursSection
        content={{ heading }}
        now={new Date("2026-01-11T12:00:00")}
      />
    </SalonFrame>
  ),
}

export const NoHeading: Story = {
  render: () => (
    <SalonFrame workingHours={mockWorkingHours} closures={mockClosures}>
      <HoursSection content={{}} now={new Date("2026-01-05T12:00:00")} />
    </SalonFrame>
  ),
}
