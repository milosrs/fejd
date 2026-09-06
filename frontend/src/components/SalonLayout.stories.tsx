import type { Meta, StoryObj } from "@storybook/react"
import { MemoryRouter, Route, Routes } from "react-router-dom"
import { SalonLayout } from "./SalonLayout"
import { LandingPage } from "../pages/LandingPage"
import { ServicesPage } from "../pages/ServicesPage"
import { BarbersPage } from "../pages/BarbersPage"
import { SalonProviders, mockOwnerMe, mockEmployeeMe } from "../stories/salon"

const meta: Meta<typeof SalonLayout> = {
  title: "Salon/SalonLayout",
  component: SalonLayout,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof SalonLayout>

function Shell({
  authenticated,
  me,
}: {
  authenticated: boolean
  me?: typeof mockOwnerMe | null
}) {
  return (
    <SalonProviders authenticated={authenticated} me={me ?? null}>
      <MemoryRouter initialEntries={["/fejd"]}>
        <Routes>
          <Route path="/:slug" element={<SalonLayout />}>
            <Route index element={<LandingPage />} />
            <Route path="services" element={<ServicesPage />} />
            <Route path="barbers" element={<BarbersPage />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </SalonProviders>
  )
}

export const Visitor: Story = {
  render: () => <Shell authenticated={false} />,
}

export const Owner: Story = {
  render: () => <Shell authenticated me={mockOwnerMe} />,
}

export const Employee: Story = {
  render: () => <Shell authenticated me={mockEmployeeMe} />,
}
