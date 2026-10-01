import type { Meta, StoryObj } from "@storybook/react-vite"
import { SalonPolicyPage } from "./SalonPolicyPage"
import { SalonFrame, mockOwnerMe } from "../stories/salon"

const mockPolicy = {
  cancellation_lead_hours: 2,
  no_show_after_minutes: 120,
  slot_interval_minutes: 30,
  auto_approve: false,
  appointment_reminder_enabled: false,
  appointment_reminder_lead_minutes: 0,
  staff_notifications_enabled: false,
  working_hours: [
    { day_of_week: 0, start_time: "09:00", end_time: "17:00" },
    { day_of_week: 1, start_time: "09:00", end_time: "17:00" },
    { day_of_week: 2, start_time: "09:00", end_time: "17:00" },
    { day_of_week: 3, start_time: "09:00", end_time: "17:00" },
    { day_of_week: 4, start_time: "09:00", end_time: "17:00" },
    { day_of_week: 5, start_time: "09:00", end_time: "17:00" },
    { day_of_week: 6, start_time: "09:00", end_time: "17:00" },
  ],
  closures: [
    { id: "aaaaaaaa-0000-4000-8000-000000000001", business_id: "11111111-1111-4111-8111-111111111111", type: "weekly", day_of_week: 6 },
    { id: "aaaaaaaa-0000-4000-8000-000000000002", business_id: "11111111-1111-4111-8111-111111111111", type: "single", start_date: "2026-12-25", reason: "Christmas" },
    { id: "aaaaaaaa-0000-4000-8000-000000000003", business_id: "11111111-1111-4111-8111-111111111111", type: "yearly", month: 1, day: 1, reason: "New Year" },
  ],
}

const meta: Meta<typeof SalonPolicyPage> = {
  title: "Salon/Salon Policy",
  component: SalonPolicyPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof SalonPolicyPage>

export const Default: Story = {
  render: () => (
    <SalonFrame authenticated me={mockOwnerMe} policy={mockPolicy}>
      <SalonPolicyPage />
    </SalonFrame>
  ),
}

export const Empty: Story = {
  render: () => (
    <SalonFrame
      authenticated
      me={mockOwnerMe}
      policy={{ ...mockPolicy, closures: [] }}
    >
      <SalonPolicyPage />
    </SalonFrame>
  ),
}
