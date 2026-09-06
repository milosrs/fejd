import { useLayoutEffect, useMemo } from "react"
import { MemoryRouter, Route, Routes } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ThemeProvider } from "../components/theme-provider"
import { SalonProvider } from "../context/SalonContext"
import { I18nProvider } from "../lib/i18n"
import { useAuthStore } from "../stores/authStore"
import type { Salon } from "../hooks/useSalon"
import type { Me } from "../hooks/useMe"
import type { Section } from "../lib/sections"

const mockI18nEn = {
  "landing.empty.title": "This page isn't set up yet",
  "landing.empty.body": "The salon hasn't published any content yet. Check back soon.",
}

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
      },
    },
    position: 3,
  },
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
  sections?: Section[]
  children: React.ReactNode
}

export function SalonProviders({
  slug = "fejd",
  salon = mockSalon,
  me = null,
  authenticated = false,
  sections = [],
  children,
}: SalonProvidersProps) {
  const queryClient = useMemo(() => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    })
    qc.setQueryData(["salon", slug], salon)
    qc.setQueryData(["sections", slug], sections)
    qc.setQueryData(["i18n", "en"], mockI18nEn)
    if (me) {
      qc.setQueryData(["me"], me)
    }
    return qc
  }, [slug, salon, me, sections])

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
      <I18nProvider>
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      </I18nProvider>
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
            path="/:slug"
            element={<SalonProvider initialEditing={initialEditing}>{children}</SalonProvider>}
          />
        </Routes>
      </MemoryRouter>
    </SalonProviders>
  )
}
