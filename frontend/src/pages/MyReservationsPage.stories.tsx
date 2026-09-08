import type { Meta, StoryObj } from "@storybook/react-vite"
import { MyReservationsPage } from "./MyReservationsPage"
import {
  StaffFrame,
  mockReservations,
  staffOwnerMe,
  staffEmployeeMe,
} from "../stories/staff"

const meta: Meta<typeof MyReservationsPage> = {
  title: "Staff/My Reservations",
  component: MyReservationsPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof MyReservationsPage>

export const OwnerWithReservations: Story = {
  render: () => (
    <StaffFrame path="my-reservations" me={staffOwnerMe} reservations={mockReservations}>
      <MyReservationsPage />
    </StaffFrame>
  ),
}

export const EmployeeWithReservations: Story = {
  render: () => (
    <StaffFrame path="my-reservations" me={staffEmployeeMe} reservations={mockReservations}>
      <MyReservationsPage />
    </StaffFrame>
  ),
}

export const Empty: Story = {
  render: () => (
    <StaffFrame path="my-reservations" me={staffEmployeeMe} reservations={[]}>
      <MyReservationsPage />
    </StaffFrame>
  ),
}
