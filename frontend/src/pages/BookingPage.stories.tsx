import type { Meta, StoryObj } from "@storybook/react-vite"
import { BookingPage } from "./BookingPage"
import { SalonFrame } from "../stories/salon"

const meta: Meta<typeof BookingPage> = {
  title: "Salon/Booking",
  component: BookingPage,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof BookingPage>

export const ServiceSelection: Story = {
  render: () => (
    <SalonFrame>
      <BookingPage />
    </SalonFrame>
  ),
}
