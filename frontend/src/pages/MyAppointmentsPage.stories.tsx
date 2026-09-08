import type { Meta, StoryObj } from "@storybook/react-vite"
import { MyAppointmentsPage } from "./MyAppointmentsPage"
import { CustomerAppointmentsFrame, mockCustomerAppointments } from "../stories/staff"

const meta: Meta<typeof MyAppointmentsPage> = {
  title: "Account/My Appointments",
  component: MyAppointmentsPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof MyAppointmentsPage>

export const WithAppointments: Story = {
  render: () => (
    <CustomerAppointmentsFrame appointments={mockCustomerAppointments}>
      <MyAppointmentsPage />
    </CustomerAppointmentsFrame>
  ),
}

export const Empty: Story = {
  render: () => (
    <CustomerAppointmentsFrame appointments={[]}>
      <MyAppointmentsPage />
    </CustomerAppointmentsFrame>
  ),
}
