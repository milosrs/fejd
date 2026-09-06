import { useLayoutEffect, useMemo } from "react"
import { MemoryRouter, Route, Routes } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ThemeProvider } from "../components/theme-provider"
import { SalonProvider } from "../context/SalonContext"
import { useAuthStore } from "../stores/authStore"
import type { Salon } from "../hooks/useSalon"
import type { Me } from "../hooks/useMe"

export const mockSalon: Salon = {
  business: {
    id: "11111111-1111-4111-8111-111111111111",
    name: "Fejd Barbershop",
    slug: "fejd",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
  services: [
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
    {
      id: "44444444-4444-4444-8444-444444444444",
      business_id: "11111111-1111-4111-8111-111111111111",
      name: "Coloring",
      duration_minutes: 90,
      price: 80,
      active: false,
      created_at: "2024-01-01T00:00:00Z",
    },
  ],
  employees: [
    {
      id: "55555555-5555-4555-8555-555555555555",
      business_id: "11111111-1111-4111-8111-111111111111",
      user_id: "emp-1",
      role: "employee",
      display_name: "Sam Barber",
      active: true,
    },
    {
      id: "66666666-6666-4666-8666-666666666666",
      business_id: "11111111-1111-4111-8111-111111111111",
      user_id: "emp-2",
      role: "employee",
      display_name: "Alex Scissor",
      active: true,
    },
  ],
}

export const mockOwnerMe: Me = {
  approval_status: "approved",
  has_salon: true,
  businesses: [
    {
      id: "11111111-1111-4111-8111-111111111111",
      name: "Fejd Barbershop",
      slug: "fejd",
      role: "admin",
    },
  ],
}

export const mockEmployeeMe: Me = {
  approval_status: "approved",
  has_salon: false,
  businesses: [
    {
      id: "11111111-1111-4111-8111-111111111111",
      name: "Fejd Barbershop",
      slug: "fejd",
      role: "employee",
    },
  ],
}

interface SalonProvidersProps {
  slug?: string
  salon?: Salon
  me?: Me | null
  authenticated?: boolean
  children: React.ReactNode
}

export function SalonProviders({
  slug = "fejd",
  salon = mockSalon,
  me = null,
  authenticated = false,
  children,
}: SalonProvidersProps) {
  const queryClient = useMemo(() => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    qc.setQueryData(["salon", slug], salon)
    if (me) {
      qc.setQueryData(["me"], me)
    }
    return qc
  }, [slug, salon, me])

  useLayoutEffect(() => {
    useAuthStore.setState({
      initialized: true,
      authenticated,
      userInfo: authenticated
        ? { sub: "user-1", email: "owner@example.com", name: "Owner" }
        : null,
      roles: [],
    })
    return () =>
      useAuthStore.setState({
        initialized: true,
        authenticated: false,
        userInfo: null,
        roles: [],
      })
  }, [authenticated])

  return (
    <ThemeProvider defaultTheme="dark" storageKey="storybook-theme">
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </ThemeProvider>
  )
}

interface SalonFrameProps extends SalonProvidersProps {}

export function SalonFrame({ children, ...props }: SalonFrameProps) {
  const { slug = "fejd" } = props
  return (
    <SalonProviders {...props}>
      <MemoryRouter initialEntries={[`/${slug}`]}>
        <Routes>
          <Route path="/:slug" element={<SalonProvider>{children}</SalonProvider>} />
        </Routes>
      </MemoryRouter>
    </SalonProviders>
  )
}
