import type { Meta, StoryObj } from "@storybook/react"
import { BarbersPage } from "./BarbersPage"
import { SalonFrame, mockOwnerMe } from "../stories/salon"

const meta: Meta<typeof BarbersPage> = {
  title: "Salon/Barbers",
  component: BarbersPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof BarbersPage>

export const Default: Story = {
  render: () => (
    <SalonFrame>
      <BarbersPage />
    </SalonFrame>
  ),
}

export const Empty: Story = {
  render: () => (
    <SalonFrame employees={[]}>
      <BarbersPage />
    </SalonFrame>
  ),
}

export const Editing: Story = {
  render: () => (
    <SalonFrame authenticated me={mockOwnerMe} initialEditing>
      <BarbersPage />
    </SalonFrame>
  ),
}
