import type { Meta, StoryObj } from "@storybook/react-vite"
import { BarberCard, BarberCardSkeleton } from "./BarberCard"
import type { Employee } from "../../hooks/useApi"

const base: Employee = {
  id: "55555555-5555-4555-8555-555555555555",
  business_id: "11111111-1111-4111-8111-111111111111",
  user_id: "emp-1",
  role: "employee",
  display_name: "Sam Barber",
  active: true,
}

const meta: Meta<typeof BarberCard> = {
  title: "Salon/BarberCard",
  component: BarberCard,
}

export default meta
type Story = StoryObj<typeof BarberCard>

export const Default: Story = {
  args: {
    employee: { ...base, avatar: "https://picsum.photos/seed/barber/200" },
  },
}

export const NoAvatar: Story = {
  args: {
    employee: base,
  },
}

export const Skeleton: Story = {
  render: () => <BarberCardSkeleton />,
}
