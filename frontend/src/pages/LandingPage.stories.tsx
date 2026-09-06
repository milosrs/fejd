import type { Meta, StoryObj } from "@storybook/react"
import { LandingPage } from "./LandingPage"
import { SalonFrame, mockSections } from "../stories/salon"

const meta: Meta<typeof LandingPage> = {
  title: "Salon/Landing",
  component: LandingPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof LandingPage>

export const WithSections: Story = {
  render: () => (
    <SalonFrame sections={mockSections}>
      <LandingPage />
    </SalonFrame>
  ),
}

export const Empty: Story = {
  render: () => (
    <SalonFrame sections={[]}>
      <LandingPage />
    </SalonFrame>
  ),
}
