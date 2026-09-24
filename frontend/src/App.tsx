import { BrowserRouter, Routes, Route, Link, useLocation, useNavigate } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { Menu as MenuIcon } from "lucide-react"
import { useAuthStore } from "./stores/authStore"
import { useMe } from "./hooks/useMe"
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
import { Button } from "./components/ui/button"
import { I18nProvider, useI18n } from "./lib/i18n"
import { ThemeProvider } from "#components/theme-provider"
import { OnboardingGate } from "#components/OnboardingGate"
import { Loader } from "#components/Loader"
import { UserMenu } from "./components/UserMenu"
import { LanguageToggle } from "./components/LanguageToggle"
import { SideDrawer } from "./components/ui/drawer"
import { InviteLandingPage } from "./components/invite/InviteLandingPage"
import { InviteAcceptHandler } from "./components/invite/InviteAcceptHandler"
import { subdomainSlug, openAppHome } from "./lib/salonDomain"
import { consumeReturnTo, auth } from "./lib/auth"
import { claimRegistrationRole } from "./lib/api"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

function AppInit({ children }: { children: React.ReactNode }) {
  const init = useAuthStore((s) => s.init)
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const navigate = useNavigate()
  const location = useLocation()
  const { data: me } = useMe()
  const topHeaderRef = useHeaderHeightMeasure("--top-header-height")
  const { t } = useI18n()

  const businesses = me?.businesses ?? []
  const primaryBusiness =
    businesses.find((b) => b.role === "admin") ?? businesses[0]
  const hasSalon = businesses.length > 0

  const [navOpen, setNavOpen] = useState(false)

  useEffect(() => {
    init()
  }, [init])

  const [appReady, setAppReady] = useState(false)

  useEffect(() => {
    if (!initialized) return
    const t = setTimeout(() => setAppReady(true), 350)
    return () => clearTimeout(t)
  }, [initialized])

  // After authentication completes, send the user back to the route they were
  // on before logging in (relevant for the native app, which can be cold-started
  // by the redirect and lose its in-memory route).
  useEffect(() => {
    if (!initialized || !authenticated) return
    const returnTo = consumeReturnTo()
    if (!returnTo || returnTo === "/") return
    const current = location.pathname + location.search
    if (returnTo !== current) navigate(returnTo)
  }, [initialized, authenticated, navigate, location.pathname, location.search])

  // Finalize a self-registration: grant the realm role chosen on the register
  // page (carried in the token's registration_role claim), then refresh the
  // token so realm_access.roles reflects it.
  useEffect(() => {
    if (!initialized || !authenticated) return
    if (!auth.getRegistrationRole()) return
    let cancelled = false
    claimRegistrationRole()
      .then(async () => {
        if (!cancelled) await auth.refresh()
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [initialized, authenticated])

  if (!appReady) {
    return <Loader className={initialized ? "animate-fade-out" : undefined} />
  }

  const isSalonView = (() => {
    if (subdomainSlug()) return true
    const first = location.pathname.split("/").filter(Boolean)[0]
    return !!first && !["invite", "my", "admin"].includes(first)
  })()

  return (
    <div className="animate-fade-in">
      <InviteAcceptHandler />
      <header
        ref={topHeaderRef}
        className="sticky top-0 z-40 border-b border-border bg-background"
      >
        <div className="flex items-center justify-between gap-3 px-4 pt-[calc(0.75rem+env(safe-area-inset-top))] pb-3 sm:px-6">
          <div className="flex min-w-0 items-center gap-2">
            {authenticated && (
              <button
                type="button"
                onClick={() => setNavOpen(true)}
                aria-label={t("app.menu")}
                className="shrink-0 rounded-md p-1.5 text-foreground hover:bg-muted md:hidden"
              >
                <MenuIcon className="size-5" />
              </button>
            )}
            <button
              type="button"
              onClick={() => openAppHome(navigate)}
              aria-label="Go to home page"
              className="shrink-0 cursor-pointer rounded hover:opacity-80 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            >
              <img src="/logo-white.jpg" alt="fejd" className="h-7 w-auto dark:hidden" />
              <img src="/logo_dark.jpg" alt="fejd" className="hidden h-7 w-auto dark:block" />
            </button>
            {authenticated && (
              <div className="hidden min-w-0 items-center gap-2 md:flex">
                {isSalonView && (
                  <button
                    type="button"
                    onClick={() => openAppHome(navigate)}
                    className="shrink-0 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted"
                  >
                    {t("app.allSalons")}
                  </button>
                )}
                <nav className="flex items-center gap-1">
                  <Link
                    to="/my/appointments"
                    className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                  >
                    {t("nav.myAppointments")}
                  </Link>
                  {hasSalon && primaryBusiness && (
                    <>
                      <Link
                        to={`/admin/business/${primaryBusiness.id}/my-reservations`}
                        className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                      >
                        {t("nav.myReservations")}
                      </Link>
                      <Link
                        to={`/admin/business/${primaryBusiness.id}/my-schedule`}
                        className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                      >
                        {t("nav.reserveMyTime")}
                      </Link>
                    </>
                  )}
                </nav>
              </div>
            )}
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <LanguageToggle />
            {authenticated ? (
              <UserMenu />
            ) : (
              <>
                <Button variant="outline" onClick={login}>
                  {t("common.logIn")}
                </Button>
                <Button onClick={() => register()}>{t("common.register")}</Button>
              </>
            )}
          </div>
        </div>
      </header>

      {authenticated && (
        <SideDrawer
          open={navOpen}
          onClose={() => setNavOpen(false)}
          title={t("app.navigation")}
        >
          <nav className="flex flex-col gap-1 p-2">
            {isSalonView && (
              <button
                type="button"
                onClick={() => {
                  setNavOpen(false)
                  openAppHome(navigate)
                }}
                className="rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted"
              >
                {t("app.allSalons")}
              </button>
            )}
            <Link
              to="/my/appointments"
              onClick={() => setNavOpen(false)}
              className="rounded-lg px-3 py-2.5 text-sm font-medium text-foreground hover:bg-muted"
            >
              {t("nav.myAppointments")}
            </Link>
            {hasSalon && primaryBusiness && (
              <>
                <Link
                  to={`/admin/business/${primaryBusiness.id}/my-reservations`}
                  onClick={() => setNavOpen(false)}
                  className="rounded-lg px-3 py-2.5 text-sm font-medium text-foreground hover:bg-muted"
                >
                  {t("nav.myReservations")}
                </Link>
                <Link
                  to={`/admin/business/${primaryBusiness.id}/my-schedule`}
                  onClick={() => setNavOpen(false)}
                  className="rounded-lg px-3 py-2.5 text-sm font-medium text-foreground hover:bg-muted"
                >
                  {t("nav.reserveMyTime")}
                </Link>
              </>
            )}
          </nav>
        </SideDrawer>
      )}

      {children}
    </div>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { t } = useI18n()

  useEffect(() => {
    if (initialized && !authenticated) {
      login()
    }
  }, [initialized, authenticated, login])

  if (!initialized) {
    return <Loader label={t("common.loading")} />
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
