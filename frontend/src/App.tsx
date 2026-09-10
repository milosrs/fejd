import { BrowserRouter, Routes, Route, Link, useLocation, useNavigate } from "react-router-dom"
import { QueryClient, QueryClientProvider, useMutation, useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { UserRound } from "lucide-react"
import { useAuthStore } from "./stores/authStore"
import { useMe } from "./hooks/useMe"
import { uploadAvatar } from "./hooks/useApi"
import { useHeaderHeightMeasure } from "./hooks/useHeaderHeightMeasure"
import { HomePage } from "./pages/HomePage"
import { LandingPage } from "./pages/LandingPage"
import { ServicesPage } from "./pages/ServicesPage"
import { BarbersPage } from "./pages/BarbersPage"
import { BookingPage } from "./pages/BookingPage"
import { MyAppointmentsPage } from "./pages/MyAppointmentsPage"
import { AdminSchedulePage } from "./pages/AdminSchedulePage"
import { AdminServicesPage } from "./pages/AdminServicesPage"
import { MySchedulePage } from "./pages/MySchedulePage"
import { MyReservationsPage } from "./pages/MyReservationsPage"
import { SalonLayout } from "./components/SalonLayout"
import { Toaster } from "./components/ui/toaster"
import { I18nProvider } from "./lib/i18n"
import { ThemeProvider } from "#components/theme-provider"
import { ModeToggle } from "#components/mode-toggle"
import { OnboardingGate } from "#components/OnboardingGate"
import { InviteLandingPage } from "./components/invite/InviteLandingPage"
import { InviteAcceptHandler } from "./components/invite/InviteAcceptHandler"
import { subdomainSlug, openAppHome } from "./lib/salonDomain"
import { resolveImageUrl } from "./lib/images"
import { pickImage } from "./lib/imagePicker"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

function ProfileAvatar({ src, name }: { src?: string; name?: string }) {
  const [loaded, setLoaded] = useState(false)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setLoaded(false)
    setFailed(false)
  }, [src])

  const showImage = !!src && !failed

  return (
    <span className="relative flex h-9 w-9 items-center justify-center overflow-hidden rounded-full bg-muted text-muted-foreground">
      {showImage && (
        <>
          <img
            src={src}
            alt={name}
            onLoad={() => setLoaded(true)}
            onError={() => setFailed(true)}
            className={`absolute inset-0 h-full w-full object-cover transition-opacity ${
              loaded ? "opacity-100" : "opacity-0"
            }`}
          />
          {!loaded && (
            <span className="absolute inset-0 animate-pulse rounded-full bg-muted" />
          )}
        </>
      )}
      {!showImage && <UserRound className="size-5" />}
    </span>
  )
}

function AppInit({ children }: { children: React.ReactNode }) {
  const init = useAuthStore((s) => s.init)
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const userInfo = useAuthStore((s) => s.userInfo)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const { data: me } = useMe()
  const topHeaderRef = useHeaderHeightMeasure("--top-header-height")

  const businesses = me?.businesses ?? []
  const primaryBusiness =
    businesses.find((b) => b.role === "admin") ?? businesses[0]
  const hasSalon = businesses.length > 0

  const uploadAvatarMutation = useMutation({
    mutationFn: (file: File) => uploadAvatar(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] })
    },
  })

  useEffect(() => {
    init()
  }, [init])

  if (!initialized) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  const isSalonView = (() => {
    if (subdomainSlug()) return true
    const first = location.pathname.split("/").filter(Boolean)[0]
    return !!first && !["invite", "my", "admin"].includes(first)
  })()

  const avatarUrl = resolveImageUrl(me?.avatar)
  const handlePickAvatar = async () => {
    const file = await pickImage()
    if (file) uploadAvatarMutation.mutate(file)
  }

  return (
    <>
      <InviteAcceptHandler />
      {authenticated && (
        <header ref={topHeaderRef} className="border-b">
          <div className="flex items-center justify-between gap-4 px-6 pt-[calc(0.75rem+env(safe-area-inset-top))] pb-3">
            <div className="flex min-w-0 items-center gap-3">
              <button
                type="button"
                onClick={handlePickAvatar}
                aria-label="Upload profile picture"
                className="shrink-0 overflow-hidden rounded-full ring-2 ring-border hover:opacity-80 focus:outline-none focus-visible:ring-primary"
              >
                <ProfileAvatar src={avatarUrl} name={userInfo?.name} />
              </button>
              <span className="truncate text-sm text-foreground">
                Welcome {userInfo?.name}
              </span>
              {isSalonView && (
                <button
                  type="button"
                  onClick={() => openAppHome(navigate)}
                  className="ml-2 shrink-0 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted"
                >
                  All salons
                </button>
              )}
              <nav className="flex items-center gap-1">
                <Link
                  to="/my/appointments"
                  className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                >
                  My appointments
                </Link>
                {hasSalon && primaryBusiness && (
                  <>
                    <Link
                      to={`/admin/business/${primaryBusiness.id}/my-reservations`}
                      className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                    >
                      My reservations
                    </Link>
                    <Link
                      to={`/admin/business/${primaryBusiness.id}/my-schedule`}
                      className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                    >
                      Reserve my time
                    </Link>
                  </>
                )}
              </nav>
            </div>
            <div className="flex shrink-0 items-center gap-4">
              <button
                onClick={logout}
                className="text-sm text-muted-foreground underline hover:text-foreground"
              >
                Logout
              </button>
              <ModeToggle />
            </div>
          </div>
        </header>
      )}
      {children}
    </>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)

  useEffect(() => {
    if (initialized && !authenticated) {
      login()
    }
  }, [initialized, authenticated, login])

  if (!initialized) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!authenticated) {
    return null
  }

  return <>{children}</>
}

function App() {
  // On a salon subdomain the salon site is served from the root; on the app
  // host, native, and local dev it is served from /:slug.
  const hostSlug = subdomainSlug()

  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <Toaster />
      <QueryClientProvider client={queryClient}>
        <I18nProvider>
          <BrowserRouter>
            <AppInit>
              <Routes>
                {hostSlug ? (
                  <Route path="/" element={<SalonLayout slug={hostSlug} />}>
                    <Route index element={<LandingPage />} />
                    <Route path="services" element={<ServicesPage />} />
                    <Route path="barbers" element={<BarbersPage />} />
                    <Route path="book" element={<BookingPage />} />
                  </Route>
                ) : (
                  <>
                    <Route path="/" element={<HomePage />} />
                    <Route path="/:slug" element={<SalonLayout />}>
                      <Route index element={<LandingPage />} />
                      <Route path="services" element={<ServicesPage />} />
                      <Route path="barbers" element={<BarbersPage />} />
                      <Route path="book" element={<BookingPage />} />
                    </Route>
                  </>
                )}

                <Route path="/invite/:token" element={<InviteLandingPage />} />

                <Route path="/my/appointments" element={<ProtectedRoute><OnboardingGate><MyAppointmentsPage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/schedule" element={<ProtectedRoute><OnboardingGate><AdminSchedulePage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/services" element={<ProtectedRoute><OnboardingGate><AdminServicesPage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/my-schedule" element={<ProtectedRoute><OnboardingGate><MySchedulePage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/my-reservations" element={<ProtectedRoute><OnboardingGate><MyReservationsPage /></OnboardingGate></ProtectedRoute>} />
              </Routes>
            </AppInit>
          </BrowserRouter>
        </I18nProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

export default App
