import type { Meta, StoryObj } from "@storybook/react"
import { LandingPage } from "./LandingPage"
import { SalonFrame, mockSalon } from "../stories/salon"

const meta: Meta<typeof LandingPage> = {
  title: "Salon/Landing",
  component: LandingPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof LandingPage>

export const Default: Story = {
  render: () => (
    <SalonFrame>
      <LandingPage />
    </SalonFrame>
  ),
}
