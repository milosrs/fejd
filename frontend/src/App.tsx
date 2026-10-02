import { BrowserRouter, Routes, Route, Link, useLocation, useNavigate } from "react-router"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { Home as HomeIcon, Menu as MenuIcon, Scissors as ScissorsIcon } from "lucide-react"
import { useAuthStore } from "./stores/authStore"
import { useMe } from "./hooks/useMe"
import { useHeaderHeightMeasure } from "./hooks/useHeaderHeightMeasure"
import { HomePage } from "./pages/HomePage"
import { LandingPage } from "./pages/LandingPage"
import { ServicesPage } from "./pages/ServicesPage"
import { ServiceDetailPage } from "./pages/ServiceDetailPage"
import { BarbersPage } from "./pages/BarbersPage"
import { BookingPage } from "./pages/BookingPage"
import { MyAppointmentsPage } from "./pages/MyAppointmentsPage"
import { AdminSchedulePage } from "./pages/AdminSchedulePage"
import { AdminServicesPage } from "./pages/AdminServicesPage"
import { MySchedulePage } from "./pages/MySchedulePage"
import { MyReservationsPage } from "./pages/MyReservationsPage"
import { SalonPolicyPage } from "./pages/SalonPolicyPage"
import { SalonSettingsPage, SalonGeneralSettings, SalonSettingsIndex } from "./pages/SalonSettingsPage"
import { InvitedCustomersPage } from "./pages/InvitedCustomersPage"
import { NotFoundPage } from "./pages/NotFoundPage"
import { SalonLayout } from "./components/SalonLayout"
import { Toaster } from "./components/ui/toaster"
import { Button } from "./components/ui/button"
import { I18nProvider, useI18n } from "./lib/i18n"
import { ThemeProvider } from "#components/theme-provider"
import { OnboardingGate } from "#components/OnboardingGate"
import { Loader } from "#components/Loader"
import { UserMenu } from "./components/UserMenu"
import { LanguageToggle } from "./components/LanguageToggle"
import { EmailVerificationBanner } from "./components/EmailVerificationBanner"
import { SideDrawer } from "./components/ui/drawer"
import { InviteLandingPage } from "./components/invite/InviteLandingPage"
import { InviteAcceptHandler } from "./components/invite/InviteAcceptHandler"
import { subdomainSlug, openAppHome, openSalon } from "./lib/salonDomain"
import { consumeReturnTo, auth } from "./lib/auth"
import { claimRegistrationRole } from "./lib/api"
import { registerPush, setNotificationTapHandler } from "./lib/push"

export function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: 1, staleTime: 30000 },
    },
  })
}

const queryClient = makeQueryClient()

export function AppInit({ children }: { children: React.ReactNode }) {
  const init = useAuthStore((s) => s.init)
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const isRealmAdmin = useAuthStore((s) => s.isRealmAdmin)
  const emailVerified = useAuthStore((s) => s.emailVerified)
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

  // Register this device for push notifications once authenticated. Best-effort:
  // permission denial or a missing web Firebase config are logged and ignored.
  useEffect(() => {
    if (!initialized || !authenticated) return
    registerPush().catch(() => {})
  }, [initialized, authenticated])

  return (
    <div className="animate-fade-in">
      <InviteAcceptHandler />
      <header
        ref={topHeaderRef}
        className="sticky top-0 z-40 border-b border-border bg-background"
      >
        <div className="flex items-center justify-between gap-3 px-4 pt-[calc(0.75rem+env(safe-area-inset-top))] pb-3 sm:px-6">
          <div className="flex min-w-0 items-center gap-2">
            {initialized ? (
              authenticated && (
                <button
                  type="button"
                  onClick={() => setNavOpen(true)}
                  aria-label={t("app.menu")}
                  className="shrink-0 rounded-md p-1.5 text-foreground hover:bg-muted md:hidden"
                >
                  <MenuIcon className="size-5" />
                </button>
              )
            ) : (
              <span className="shrink-0 rounded-md p-1.5 md:hidden">
                <div className="size-5 animate-pulse rounded bg-muted" />
              </span>
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
            {initialized ? (
              authenticated && (
                <div className="hidden min-w-0 items-center gap-2 md:flex">
                  <button
                    type="button"
                    onClick={() => openAppHome(navigate)}
                    className="flex shrink-0 items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted"
                  >
                    <HomeIcon className="size-4" />
                    {t("nav.home")}
                  </button>
                  {hasSalon && primaryBusiness && (
                    <button
                      type="button"
                      onClick={() => openSalon(navigate, primaryBusiness.slug)}
                      className="flex shrink-0 items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted"
                    >
                      <ScissorsIcon className="size-4" />
                      {t("nav.mySalon")}
                    </button>
                  )}
                  <nav className="flex items-center gap-1">
                    <Link
                      to="/my/appointments"
                      className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                    >
                      {t("nav.myAppointments")}
                    </Link>
                    {isRealmAdmin && (
                      <Link
                        to="/admin/invited-customers"
                        className="rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                      >
                        {t("nav.invitedCustomers")}
                      </Link>
                    )}
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
              )
            ) : (
              <div className="hidden min-w-0 items-center gap-2 md:flex">
                <div className="h-5 w-20 animate-pulse rounded bg-muted" />
                <div className="h-5 w-24 animate-pulse rounded bg-muted" />
              </div>
            )}
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <LanguageToggle />
            {initialized ? (
              authenticated ? (
                <UserMenu />
              ) : (
                <>
                  <Button variant="outline" onClick={login}>
                    {t("common.logIn")}
                  </Button>
                  <Button onClick={() => register()}>{t("common.register")}</Button>
                </>
              )
            ) : (
              <>
                <div className="h-8 w-16 animate-pulse rounded-2xl bg-muted" />
                <div className="h-8 w-20 animate-pulse rounded-2xl bg-muted" />
              </>
            )}
          </div>
        </div>
      </header>

      {authenticated && !emailVerified && <EmailVerificationBanner />}

      {authenticated && (
        <SideDrawer
          open={navOpen}
          onClose={() => setNavOpen(false)}
          title={t("app.navigation")}
        >
          <nav className="flex flex-col gap-1 p-2">
            <button
              type="button"
              onClick={() => {
                setNavOpen(false)
                openAppHome(navigate)
              }}
              className="flex items-center gap-1.5 rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted"
            >
              <HomeIcon className="size-4" />
              {t("nav.home")}
            </button>
            {hasSalon && primaryBusiness && (
              <button
                type="button"
                onClick={() => {
                  setNavOpen(false)
                  openSalon(navigate, primaryBusiness.slug)
                }}
                className="flex items-center gap-1.5 rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted"
              >
                <ScissorsIcon className="size-4" />
                {t("nav.mySalon")}
              </button>
            )}
            <Link
              to="/my/appointments"
              onClick={() => setNavOpen(false)}
              className="rounded-lg px-3 py-2.5 text-sm font-medium text-foreground hover:bg-muted"
            >
              {t("nav.myAppointments")}
            </Link>
            {isRealmAdmin && (
              <Link
                to="/admin/invited-customers"
                onClick={() => setNavOpen(false)}
                className="rounded-lg px-3 py-2.5 text-sm font-medium text-foreground hover:bg-muted"
              >
                {t("nav.invitedCustomers")}
              </Link>
            )}
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

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
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

// PushNavigationHandler routes native notification taps to the salon's
// reservation list. Web taps are handled by the firebase-messaging service
// worker (which opens the link directly).
export function PushNavigationHandler() {
  const navigate = useNavigate()

  useEffect(() => {
    setNotificationTapHandler((data) => {
      const businessId = data.business_id
      if (!businessId) return
      navigate(`/admin/business/${businessId}/my-reservations`)
    })
    return () => setNotificationTapHandler(null)
  }, [navigate])

  return null
}

function App() {  // On a salon subdomain the salon site is served from the root; on the app
  // host, native, and local dev it is served from /:slug.
  const hostSlug = subdomainSlug()

  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <Toaster />
      <QueryClientProvider client={queryClient}>
        <I18nProvider>
          <BrowserRouter>
            <PushNavigationHandler />
            <AppInit>
              <Routes>
                {hostSlug ? (
                  <Route path="/" element={<SalonLayout slug={hostSlug} />}>
                    <Route index element={<LandingPage />} />
                    <Route path="services" element={<ServicesPage />} />
                    <Route path="services/:serviceSlug" element={<ServiceDetailPage />} />
                    <Route path="barbers" element={<BarbersPage />} />
                    <Route path="book" element={<BookingPage />} />
                    <Route path="book/:serviceSlug" element={<BookingPage />} />
                    <Route path="policy" element={<SalonPolicyPage />} />
                    <Route path="settings" element={<SalonSettingsPage />}>
                      <Route index element={<SalonSettingsIndex />} />
                      <Route path="general" element={<SalonGeneralSettings />} />
                      <Route path="policy" element={<SalonPolicyPage />} />
                    </Route>
                  </Route>
                ) : (
                  <>
                    <Route path="/" element={<HomePage />} />
                    <Route path="/:slug" element={<SalonLayout />}>
                      <Route index element={<LandingPage />} />
                      <Route path="services" element={<ServicesPage />} />
                      <Route path="services/:serviceSlug" element={<ServiceDetailPage />} />
                      <Route path="barbers" element={<BarbersPage />} />
                      <Route path="book" element={<BookingPage />} />
                      <Route path="book/:serviceSlug" element={<BookingPage />} />
                      <Route path="policy" element={<SalonPolicyPage />} />
                      <Route path="settings" element={<SalonSettingsPage />}>
                        <Route index element={<SalonSettingsIndex />} />
                        <Route path="general" element={<SalonGeneralSettings />} />
                        <Route path="policy" element={<SalonPolicyPage />} />
                      </Route>
                    </Route>
                  </>
                )}

                <Route path="/invite/:token" element={<InviteLandingPage />} />

                <Route path="/my/appointments" element={<ProtectedRoute><MyAppointmentsPage /></ProtectedRoute>} />
                <Route path="/admin/invited-customers" element={<ProtectedRoute><InvitedCustomersPage /></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/schedule" element={<ProtectedRoute><OnboardingGate><AdminSchedulePage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/services" element={<ProtectedRoute><OnboardingGate><AdminServicesPage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/my-schedule" element={<ProtectedRoute><OnboardingGate><MySchedulePage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/my-reservations" element={<ProtectedRoute><OnboardingGate><MyReservationsPage /></OnboardingGate></ProtectedRoute>} />
                <Route path="*" element={<NotFoundPage />} />
              </Routes>
            </AppInit>
          </BrowserRouter>
        </I18nProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

export default App
