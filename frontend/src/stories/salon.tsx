import { useLayoutEffect, useMemo } from "react"
import { MemoryRouter, Route, Routes } from "react-router"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ThemeProvider } from "../components/theme-provider"
import { SalonProvider } from "../context/SalonContext"
import { I18nProvider } from "../lib/i18n"
import { useAuthStore } from "../stores/authStore"
import { useBookingStore } from "../stores/bookingStore"
import { mockI18nEn } from "./mockI18n"
import type { Salon } from "../hooks/useSalon"
import type { Me } from "../hooks/useMe"
import type { Service, Employee } from "../hooks/useApi"
import type { Section } from "../lib/sections"
import type { BusinessHours, BusinessClosure } from "../lib/salonHours"

export const mockSalon: Salon = {
  business: {
    id: "11111111-1111-4111-8111-111111111111",
    name: "Fejd Barbershop",
    slug: "fejd",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    cancellation_lead_hours: 2,
    no_show_after_minutes: 120,
    slot_interval_minutes: 30,
  },
  services: [
    {
      id: "22222222-2222-4222-8222-222222222222",
      business_id: "11111111-1111-4111-8111-111111111111",
      name: "Haircut",
      slug: "haircut",
      duration_minutes: 30,
      price: 25,
      active: true,
      created_at: "2024-01-01T00:00:00Z",
    },
    {
      id: "33333333-3333-4333-8333-333333333333",
      business_id: "11111111-1111-4111-8111-111111111111",
      name: "Beard Trim",
      slug: "beard-trim",
      duration_minutes: 20,
      price: 15,
      active: true,
      created_at: "2024-01-01T00:00:00Z",
    },
    {
      id: "44444444-4444-4444-8444-444444444444",
      business_id: "11111111-1111-4111-8111-111111111111",
      name: "Coloring",
      slug: "coloring",
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
  images: {
    hero: "",
    logo: "",
    background: "",
  },
}

export const mockSections: Section[] = [
  {
    id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    page_id: "99999999-9999-4999-8999-999999999999",
    type: "hero",
    content: {
      en: {
        headline: "Sharp cuts, done right",
        subheadline: "Book your next appointment in seconds.",
        cta_text: "Book now",
      },
      rs: {
        headline: "Oštre frizure, kako treba",
        subheadline: "Zakažite sledeći termin za nekoliko sekundi.",
        cta_text: "Zakaži sada",
      },
    },
    position: 0,
  },
  {
    id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
    page_id: "99999999-9999-4999-8999-999999999999",
    type: "about",
    content: {
      en: {
        heading: "About us",
        body: "A neighbourhood barbershop run by people who care about the craft.",
      },
      rs: {
        heading: "O nama",
        body: "Kvartovska berbernica koju vode ljudi kojima je stalo do zanata.",
      },
    },
    position: 1,
  },
  {
    id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
    page_id: "99999999-9999-4999-8999-999999999999",
    type: "gallery",
    content: {
      en: {
        heading: "Our work",
        image_urls: [
          "https://picsum.photos/seed/salon-1/400",
          "https://picsum.photos/seed/salon-2/400",
          "https://picsum.photos/seed/salon-3/400",
        ],
      },
    },
    position: 2,
  },
  {
    id: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
    page_id: "99999999-9999-4999-8999-999999999999",
    type: "contact",
    content: {
      en: {
        heading: "Visit us",
        phone: "+46 8 123 45 67",
        email: "hello@fejd.example",
        address: "Main Street 1, Stockholm",
        instagram_url: "https://www.instagram.com/fejdbarbershop",
      },
    },
    position: 3,
  },
  {
    id: "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee",
    page_id: "99999999-9999-4999-8999-999999999999",
    type: "hours",
    content: {
      en: {
        heading: "Opening hours",
      },
      rs: {
        heading: "Radno vreme",
      },
    },
    position: 4,
  },
]

export const mockWorkingHours: BusinessHours[] = [
  { day_of_week: 1, start_time: "09:00", end_time: "17:00" },
  { day_of_week: 2, start_time: "09:00", end_time: "17:00" },
  { day_of_week: 3, start_time: "09:00", end_time: "17:00" },
  { day_of_week: 4, start_time: "09:00", end_time: "17:00" },
  { day_of_week: 5, start_time: "09:00", end_time: "17:00" },
  { day_of_week: 6, start_time: "10:00", end_time: "16:00" },
]

export const mockClosures: BusinessClosure[] = [
  { type: "weekly", day_of_week: 0 },
]

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
  roles?: string[]
  sections?: Section[]
  services?: Service[]
  employees?: Employee[]
  policy?: any
  workingHours?: BusinessHours[]
  closures?: BusinessClosure[]
  children: React.ReactNode
}

export function SalonProviders({
  slug = "fejd",
  salon = mockSalon,
  me = null,
  authenticated = false,
  roles = ["Owner"],
  sections = [],
  services = mockSalon.services,
  employees = mockSalon.employees,
  policy,
  workingHours = mockWorkingHours,
  closures = mockClosures,
  children,
}: SalonProvidersProps) {
  const queryClient = useMemo(() => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    qc.setQueryData(["salon", slug], salon)
    qc.setQueryData(["sections", slug], sections)
    qc.setQueryData(["services", slug], services)
    qc.setQueryData(["employees", slug], employees)
    qc.setQueryData(["business-working-hours", slug], workingHours)
    qc.setQueryData(["business-closures", slug], closures)
    qc.setQueryData(["i18n", "en"], mockI18nEn)
    if (policy) {
      qc.setQueryData(["salon-policy", salon.business.id], policy)
    }
    if (me) {
      qc.setQueryData(["me"], me)
    }
    return qc
  }, [slug, salon, me, sections, services, employees, policy, workingHours, closures])

  useLayoutEffect(() => {
    useAuthStore.setState({
      initialized: true,
      authenticated,
      userInfo: authenticated
        ? { sub: "user-1", email: "owner@example.com", name: "Owner" }
        : null,
      roles: authenticated ? roles : [],
    })
    useBookingStore.getState().reset()
    return () =>
      useAuthStore.setState({
        initialized: true,
        authenticated: false,
        userInfo: null,
        roles: [],
      })
  }, [authenticated, roles])

  return (
    <ThemeProvider defaultTheme="dark" storageKey="storybook-theme">
      <QueryClientProvider client={queryClient}>
        <I18nProvider>{children}</I18nProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

interface SalonFrameProps extends SalonProvidersProps {
  initialEditing?: boolean
}

export function SalonFrame({ children, initialEditing, ...props }: SalonFrameProps) {
  const { slug = "fejd" } = props
  return (
    <SalonProviders {...props}>
      <MemoryRouter initialEntries={[`/${slug}`]}>
        <Routes>
          <Route
            path="/:slug/*"
            element={<SalonProvider initialEditing={initialEditing}>{children}</SalonProvider>}
          />
        </Routes>
      </MemoryRouter>
    </SalonProviders>
  )
}
