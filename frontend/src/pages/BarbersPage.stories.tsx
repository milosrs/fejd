import type { Meta, StoryObj } from "@storybook/react"
import { BarbersPage } from "./BarbersPage"
import { SalonFrame, mockSalon } from "../stories/salon"
import type { Salon } from "../hooks/useSalon"

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
  render: () => {
    const empty: Salon = { ...mockSalon, employees: [] }
    return (
      <SalonFrame salon={empty}>
        <BarbersPage />
      </SalonFrame>
    )
  },
}
