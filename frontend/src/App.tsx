import { BrowserRouter, Routes, Route } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useEffect } from "react"
import { useAuthStore } from "./stores/authStore"
import { HomePage } from "./pages/HomePage"
import { LandingPage } from "./pages/LandingPage"
import { ServicesPage } from "./pages/ServicesPage"
import { BarbersPage } from "./pages/BarbersPage"
import { BookingPage } from "./pages/BookingPage"
import { MyAppointmentsPage } from "./pages/MyAppointmentsPage"
import { AdminSchedulePage } from "./pages/AdminSchedulePage"
import { AdminServicesPage } from "./pages/AdminServicesPage"
import { SalonLayout } from "./components/SalonLayout"
import { I18nProvider } from "./lib/i18n"
import { ThemeProvider } from "#components/theme-provider"
import { ModeToggle } from "#components/mode-toggle"
import { OnboardingGate } from "#components/OnboardingGate"
import { InviteLandingPage } from "./components/invite/InviteLandingPage"
import { InviteAcceptHandler } from "./components/invite/InviteAcceptHandler"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

function AppInit({ children }: { children: React.ReactNode }) {
  const init = useAuthStore((s) => s.init)
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const userInfo = useAuthStore((s) => s.userInfo)
  const logout = useAuthStore((s) => s.logout)

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

  return (
    <>
      <InviteAcceptHandler />
      {authenticated && (
        <header className="flex items-center justify-end gap-4 border-b px-6 pt-[calc(0.75rem+env(safe-area-inset-top))] pb-3">
          <span className="text-sm text-muted-foreground">
            Hello {userInfo?.name}
          </span>
          <button
            onClick={logout}
            className="text-sm text-muted-foreground underline hover:text-foreground"
          >
            Logout
          </button>
          <ModeToggle />
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
  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <QueryClientProvider client={queryClient}>
        <I18nProvider>
          <BrowserRouter>
            <AppInit>
              <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/invite/:token" element={<InviteLandingPage />} />

                <Route path="/my/appointments" element={<ProtectedRoute><OnboardingGate><MyAppointmentsPage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/schedule" element={<ProtectedRoute><OnboardingGate><AdminSchedulePage /></OnboardingGate></ProtectedRoute>} />
                <Route path="/admin/business/:businessId/services" element={<ProtectedRoute><OnboardingGate><AdminServicesPage /></OnboardingGate></ProtectedRoute>} />

                <Route path="/:slug" element={<SalonLayout />}>
                  <Route index element={<LandingPage />} />
                  <Route path="services" element={<ServicesPage />} />
                  <Route path="barbers" element={<BarbersPage />} />
                  <Route path="book" element={<BookingPage />} />
                </Route>
              </Routes>
            </AppInit>
          </BrowserRouter>
        </I18nProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

export default App
