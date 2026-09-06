import type { Meta, StoryObj } from "@storybook/react-vite"
import { ServicesPage } from "./ServicesPage"
import { SalonFrame, mockOwnerMe } from "../stories/salon"

const meta: Meta<typeof ServicesPage> = {
  title: "Salon/Services",
  component: ServicesPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof ServicesPage>

export const Default: Story = {
  render: () => (
    <SalonFrame>
      <ServicesPage />
    </SalonFrame>
  ),
}

export const Empty: Story = {
  render: () => (
    <SalonFrame services={[]}>
      <ServicesPage />
    </SalonFrame>
  ),
}

export const Editing: Story = {
  render: () => (
    <SalonFrame authenticated me={mockOwnerMe} initialEditing>
      <ServicesPage />
    </SalonFrame>
  ),
}
