import type { Meta, StoryObj } from "@storybook/react"
import { ServicesPage } from "./ServicesPage"
import { SalonFrame } from "../stories/salon"

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
