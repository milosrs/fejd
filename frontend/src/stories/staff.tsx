import { useLayoutEffect, useMemo } from "react"
import { MemoryRouter, Route, Routes } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { today, getLocalTimeZone } from "@internationalized/date"
import { ThemeProvider } from "../components/theme-provider"
import { useAuthStore } from "../stores/authStore"
import type { Me } from "../hooks/useMe"
import type { EmployeeUnavailability, Appointment, Service, Customer } from "../hooks/useApi"

export const staffBusinessId = "11111111-1111-4111-8111-111111111111"

function futureISO(days: number, hour: number, minute = 0): string {
  const now = new Date()
  const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() + days, hour, minute, 0, 0)
  return d.toISOString()
}

function inHours(hours: number): string {
  return new Date(Date.now() + hours * 3600 * 1000).toISOString()
}

export const mockUnavailability: EmployeeUnavailability[] = [
  {
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1",
    business_user_id: "55555555-5555-4555-8555-555555555555",
    start_time: futureISO(1, 12),
    end_time: futureISO(1, 13),
    reason: "Dentist appointment",
    status: "confirmed",
  },
  {
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2",
    business_user_id: "55555555-5555-4555-8555-555555555555",
    start_time: futureISO(3, 15),
    end_time: futureISO(3, 16, 30),
    status: "pending",
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
    no_show_after_hours: 2,
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
    no_show_after_hours: 2,
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
    no_show_after_hours: 2,
  },
  {
    id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb4",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "customer-4",
    service_id: "22222222-2222-4222-8222-222222222222",
    service_name: "Haircut",
    start_time: inHours(-3),
    end_time: inHours(-2.5),
    status: "confirmed",
    created_by: "customer-4",
    created_at: inHours(-4),
    no_show_after_hours: 2,
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

export const mockCustomerAppointments: Appointment[] = [
  {
    id: "cccccccc-cccc-4ccc-8ccc-ccccccccccc1",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "user-1",
    service_id: "22222222-2222-4222-8222-222222222222",
    service_name: "Haircut",
    start_time: futureISO(5, 10),
    end_time: futureISO(5, 10, 30),
    status: "confirmed",
    created_by: "user-1",
    created_at: futureISO(0, 8),
    cancellation_lead_hours: 2,
  },
  {
    id: "cccccccc-cccc-4ccc-8ccc-ccccccccccc2",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "user-1",
    service_id: "33333333-3333-4333-8333-333333333333",
    service_name: "Beard Trim",
    start_time: inHours(1),
    end_time: inHours(1.5),
    status: "confirmed",
    created_by: "user-1",
    created_at: futureISO(0, 8),
    cancellation_lead_hours: 2,
  },
  {
    id: "cccccccc-cccc-4ccc-8ccc-ccccccccccc3",
    business_id: staffBusinessId,
    business_user_id: "55555555-5555-4555-8555-555555555555",
    customer_user_id: "user-1",
    service_id: "22222222-2222-4222-8222-222222222222",
    service_name: "Haircut",
    start_time: futureISO(3, 11),
    end_time: futureISO(3, 11, 30),
    status: "cancelled",
    created_by: "user-1",
    created_at: futureISO(0, 8),
    cancellation_reason: "Something came up",
    cancellation_lead_hours: 2,
  },
]

export const mockMyServices: Service[] = [
  {
    id: "22222222-2222-4222-8222-222222222222",
    business_id: staffBusinessId,
    name: "Haircut",
    duration_minutes: 30,
    price: 25,
    active: true,
    created_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "33333333-3333-4333-8333-333333333333",
    business_id: staffBusinessId,
    name: "Beard Trim",
    duration_minutes: 20,
    price: 15,
    active: true,
    created_at: "2024-01-01T00:00:00Z",
  },
]

export const mockCustomers: Customer[] = [
  { user_id: "customer-1", display_name: "Alice Johnson" },
  { user_id: "customer-2", display_name: "Bob Smith" },
  { user_id: "customer-3", display_name: "" },
]

interface StaffProvidersProps {
  businessId?: string
  me?: Me | null
  roles?: string[]
  unavailability?: EmployeeUnavailability[]
  reservations?: Appointment[]
  services?: Service[]
  customers?: Customer[]
  children: React.ReactNode
}

export function StaffProviders({
  businessId = staffBusinessId,
  me = null,
  roles = ["Owner"],
  unavailability = [],
  reservations = [],
  services = mockMyServices,
  customers = mockCustomers,
  children,
}: StaffProvidersProps) {
  const todayKey = today(getLocalTimeZone()).toString()

  const queryClient = useMemo(() => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    qc.setQueryData(["my-unavailability", businessId], unavailability)
    qc.setQueryData(["my-reservations", businessId, todayKey], reservations)
    qc.setQueryData(["my-services", businessId], services)
    qc.setQueryData(["customers", businessId], customers)
    qc.setQueryData(["my-appointments"], [])
    if (me) {
      qc.setQueryData(["me"], me)
    }
    return qc
  }, [businessId, me, unavailability, reservations, services, customers, todayKey])

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

export function CustomerAppointmentsFrame({
  appointments = [],
  children,
}: {
  appointments?: Appointment[]
  children: React.ReactNode
}) {
  const queryClient = useMemo(() => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    qc.setQueryData(["my-appointments"], appointments)
    return qc
  }, [appointments])

  useLayoutEffect(() => {
    useAuthStore.setState({
      initialized: true,
      authenticated: true,
      userInfo: { sub: "user-1", email: "customer@example.com", name: "Customer" },
      roles: [],
    })
    return () =>
      useAuthStore.setState({
        initialized: true,
        authenticated: false,
        userInfo: null,
        roles: [],
      })
  }, [])

  return (
    <ThemeProvider defaultTheme="dark" storageKey="storybook-theme">
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/my/appointments"]}>
          <Routes>
            <Route path="/my/appointments" element={children} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </ThemeProvider>
  )
}
