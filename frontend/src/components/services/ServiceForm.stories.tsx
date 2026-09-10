import type { Meta, StoryObj } from "@storybook/react-vite"
import { ServiceForm } from "./ServiceForm"
import type { Service } from "../../hooks/useApi"

const meta: Meta<typeof ServiceForm> = {
  title: "Salon/ServiceForm",
  component: ServiceForm,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof ServiceForm>

const existing: Service = {
  id: "22222222-2222-4222-8222-222222222222",
  business_id: "11111111-1111-4111-8111-111111111111",
  name: "Haircut",
  duration_minutes: 30,
  price: 25,
  active: true,
  description: "A classic cut, tailored to you.",
  created_at: "2024-01-01T00:00:00Z",
}

export const Create: Story = {
  args: {
    onClose: () => {},
    onSubmit: () => {},
  },
}

export const Edit: Story = {
  args: {
    initial: existing,
    onClose: () => {},
    onSubmit: () => {},
  },
}
