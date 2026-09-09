import type { Meta, StoryObj } from "@storybook/react-vite"
import { AddAppointmentDialog } from "./AddAppointmentDialog"
import { StaffProviders, staffBusinessId } from "../stories/staff"

const meta: Meta<typeof AddAppointmentDialog> = {
  title: "Staff/Add Appointment",
  component: AddAppointmentDialog,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof AddAppointmentDialog>

export const Default: Story = {
  render: () => (
    <StaffProviders>
      <AddAppointmentDialog businessId={staffBusinessId} onClose={() => {}} />
    </StaffProviders>
  ),
}

export const NoServices: Story = {
  render: () => (
    <StaffProviders services={[]} customers={[]}>
      <AddAppointmentDialog businessId={staffBusinessId} onClose={() => {}} />
    </StaffProviders>
  ),
}
