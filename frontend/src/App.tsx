import { BrowserRouter, Routes, Route } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useEffect } from "react"
import { useAuthStore } from "./stores/authStore"
import { BusinessPage } from "./pages/BusinessPage"
import { BookingPage } from "./pages/BookingPage"
import { MyAppointmentsPage } from "./pages/MyAppointmentsPage"
import { AdminSchedulePage } from "./pages/AdminSchedulePage"
import { AdminServicesPage } from "./pages/AdminServicesPage"
import { DatePickerPage } from "./pages/DatePickerPage"
import { ThemeProvider } from "#components/theme-provider"
import { ModeToggle } from "#components/mode-toggle"

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
      {authenticated && (
        <header className="flex items-center justify-end gap-4 border-b px-6 py-3">
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
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AppInit>
            <ProtectedRoute>
              <Routes>
                <Route path="/" element={<DatePickerPage />} />
                <Route path="/business/:slug" element={<BusinessPage />} />
                <Route path="/business/:slug/book" element={<BookingPage />} />
                <Route path="/my/appointments" element={<MyAppointmentsPage />} />
                <Route path="/admin/business/:businessId/schedule" element={<AdminSchedulePage />} />
                <Route path="/admin/business/:businessId/services" element={<AdminServicesPage />} />
              </Routes>
            </ProtectedRoute>
          </AppInit>
        </BrowserRouter>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

export default App
