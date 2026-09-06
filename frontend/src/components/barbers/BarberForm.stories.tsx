import type { Meta, StoryObj } from "@storybook/react"
import { BarberForm } from "./BarberForm"
import { SalonFrame } from "../../stories/salon"
import type { Employee, Service } from "../../hooks/useApi"

const meta: Meta<typeof BarberForm> = {
  title: "Salon/BarberForm",
  component: BarberForm,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof BarberForm>

const services: Service[] = [
  {
    id: "22222222-2222-4222-8222-222222222222",
    business_id: "11111111-1111-4111-8111-111111111111",
    name: "Haircut",
    duration_minutes: 30,
    price: 25,
    active: true,
    created_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "33333333-3333-4333-8333-333333333333",
    business_id: "11111111-1111-4111-8111-111111111111",
    name: "Beard Trim",
    duration_minutes: 20,
    price: 15,
    active: true,
    created_at: "2024-01-01T00:00:00Z",
  },
]

const existing: Employee = {
  id: "55555555-5555-4555-8555-555555555555",
  business_id: "11111111-1111-4111-8111-111111111111",
  user_id: "emp-1",
  role: "employee",
  display_name: "Sam Barber",
  active: true,
}

export const Create: Story = {
  render: () => (
    <SalonFrame>
      <BarberForm services={services} onClose={() => {}} onSubmit={() => {}} />
    </SalonFrame>
  ),
}

export const Edit: Story = {
  render: () => (
    <SalonFrame>
      <BarberForm
        initial={existing}
        services={services}
        onClose={() => {}}
        onSubmit={() => {}}
        onUploadAvatar={() => {}}
      />
    </SalonFrame>
  ),
}
