import { useLayoutEffect, useMemo } from "react"
import { MemoryRouter, Route, Routes } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { today, getLocalTimeZone } from "@internationalized/date"
import { ThemeProvider } from "../components/theme-provider"
import { useAuthStore } from "../stores/authStore"
import type { Me } from "../hooks/useMe"
import type { EmployeeUnavailability, Appointment } from "../hooks/useApi"

export const staffBusinessId = "11111111-1111-4111-8111-111111111111"

function futureISO(days: number, hour: number, minute = 0): string {
  const now = new Date()
  const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() + days, hour, minute, 0, 0)
  return d.toISOString()
}

export const mockUnavailability: EmployeeUnavailability[] = [
  {
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1",
    business_user_id: "55555555-5555-4555-8555-555555555555",
    start_time: futureISO(1, 12),
    end_time: futureISO(1, 13),
    reason: "Dentist appointment",
  },
  {
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2",
    business_user_id: "55555555-5555-4555-8555-555555555555",
    start_time: futureISO(3, 15),
    end_time: futureISO(3, 16, 30),
  },
]

export const mockReservations: Appointment[] = [
  {
    id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "customer-1",
    service_id: "22222222-2222-4222-8222-222222222222",
    service_name: "Haircut",
    start_time: futureISO(0, 9),
    end_time: futureISO(0, 9, 30),
    status: "confirmed",
    created_by: "customer-1",
    created_at: futureISO(0, 8),
  },
  {
    id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "customer-2",
    service_id: "33333333-3333-4333-8333-333333333333",
    service_name: "Beard Trim",
    start_time: futureISO(0, 10),
    end_time: futureISO(0, 10, 20),
    status: "pending",
    created_by: "customer-2",
    created_at: futureISO(0, 8),
  },
  {
    id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "customer-3",
    service_id: "22222222-2222-4222-8222-222222222222",
    service_name: "Haircut",
    start_time: futureISO(0, 11),
    end_time: futureISO(0, 11, 30),
    status: "cancelled",
    created_by: "customer-3",
    created_at: futureISO(0, 8),
    cancellation_reason: "customer asked to reschedule",
  },
]

export const staffOwnerMe: Me = {
  approval_status: "approved",
  has_salon: true,
  businesses: [
    {
      id: staffBusinessId,
      name: "Fejd Barbershop",
      slug: "fejd",
      role: "admin",
    },
  ],
}

export const staffEmployeeMe: Me = {
  approval_status: "approved",
  has_salon: false,
  businesses: [
    {
      id: staffBusinessId,
      name: "Fejd Barbershop",
      slug: "fejd",
      role: "employee",
    },
  ],
}

interface StaffProvidersProps {
  businessId?: string
  me?: Me | null
  roles?: string[]
  unavailability?: EmployeeUnavailability[]
  reservations?: Appointment[]
  children: React.ReactNode
}

export function StaffProviders({
  businessId = staffBusinessId,
  me = null,
  roles = ["Owner"],
  unavailability = [],
  reservations = [],
  children,
}: StaffProvidersProps) {
  const todayKey = today(getLocalTimeZone()).toString()

  const queryClient = useMemo(() => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    qc.setQueryData(["my-unavailability", businessId], unavailability)
    qc.setQueryData(["my-reservations", businessId, todayKey], reservations)
    if (me) {
      qc.setQueryData(["me"], me)
    }
    return qc
  }, [businessId, me, unavailability, reservations, todayKey])

  useLayoutEffect(() => {
    useAuthStore.setState({
      initialized: true,
      authenticated: true,
      userInfo: { sub: "user-1", email: "staff@example.com", name: "Staff" },
      roles,
    })
    return () =>
      useAuthStore.setState({
        initialized: true,
        authenticated: false,
        userInfo: null,
        roles: [],
      })
  }, [roles])

  return (
    <ThemeProvider defaultTheme="dark" storageKey="storybook-theme">
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </ThemeProvider>
  )
}

interface StaffFrameProps extends StaffProvidersProps {
  path?: string
}

export function StaffFrame({ children, path = "my-schedule", ...props }: StaffFrameProps) {
  const businessId = props.businessId ?? staffBusinessId
  return (
    <StaffProviders businessId={businessId} {...props}>
      <MemoryRouter initialEntries={[`/admin/business/${businessId}/${path}`]}>
        <Routes>
          <Route path="/admin/business/:businessId/*" element={children} />
        </Routes>
      </MemoryRouter>
    </StaffProviders>
  )
}
