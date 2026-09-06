import type { Meta, StoryObj } from "@storybook/react"
import { ServiceCard, ServiceCardSkeleton } from "./ServiceCard"
import type { Service } from "../../hooks/useApi"

const base: Service = {
  id: "22222222-2222-4222-8222-222222222222",
  business_id: "11111111-1111-4111-8111-111111111111",
  name: "Haircut",
  duration_minutes: 30,
  price: 25,
  active: true,
  description: "A classic cut, tailored to you.",
  created_at: "2024-01-01T00:00:00Z",
}

const meta: Meta<typeof ServiceCard> = {
  title: "Salon/ServiceCard",
  component: ServiceCard,
}

export default meta
type Story = StoryObj<typeof ServiceCard>

export const Default: Story = {
  args: {
    service: { ...base, picture_id: "img-1" },
    onBook: () => {},
  },
}

export const NoImage: Story = {
  args: {
    service: base,
    onBook: () => {},
  },
}

export const RequiresAuth: Story = {
  args: {
    service: base,
    onBook: () => {},
    note: "To book, you have to register.",
  },
}

export const Skeleton: Story = {
  render: () => <ServiceCardSkeleton />,
}
