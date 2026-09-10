import type { Meta, StoryObj } from "@storybook/react-vite"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { SalonPolicyDialog } from "./SalonPolicyDialog"

const meta: Meta<typeof SalonPolicyDialog> = {
  title: "Salon/Salon Policy",
  component: SalonPolicyDialog,
  parameters: {
    layout: "centered",
  },
}

export default meta
type Story = StoryObj<typeof SalonPolicyDialog>

function Frame() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return (
    <QueryClientProvider client={queryClient}>
      <SalonPolicyDialog
        businessId="11111111-1111-4111-8111-111111111111"
        slug="fejd"
        onClose={() => {}}
      />
    </QueryClientProvider>
  )
}

export const Default: Story = {
  render: () => <Frame />,
}

export const LongerNotice: Story = {
  render: () => <Frame />,
}
