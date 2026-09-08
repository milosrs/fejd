import type { Meta, StoryObj } from "@storybook/react-vite"
import { MySchedulePage } from "./MySchedulePage"
import {
  StaffFrame,
  mockUnavailability,
  staffOwnerMe,
  staffEmployeeMe,
} from "../stories/staff"

const meta: Meta<typeof MySchedulePage> = {
  title: "Staff/My Schedule",
  component: MySchedulePage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof MySchedulePage>

export const OwnerWithBlockedSlots: Story = {
  render: () => (
    <StaffFrame path="my-schedule" me={staffOwnerMe} unavailability={mockUnavailability}>
      <MySchedulePage />
    </StaffFrame>
  ),
}

export const EmployeeEmpty: Story = {
  render: () => (
    <StaffFrame path="my-schedule" me={staffEmployeeMe} unavailability={[]}>
      <MySchedulePage />
    </StaffFrame>
  ),
}

export const EmployeeWithBlockedSlots: Story = {
  render: () => (
    <StaffFrame path="my-schedule" me={staffEmployeeMe} unavailability={mockUnavailability}>
      <MySchedulePage />
    </StaffFrame>
  ),
}
