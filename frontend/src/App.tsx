import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useEffect } from "react"
import { useAuthStore } from "./stores/authStore"
import { BusinessPage } from "./pages/BusinessPage"
import { BookingPage } from "./pages/BookingPage"
import { MyAppointmentsPage } from "./pages/MyAppointmentsPage"
import { AdminSchedulePage } from "./pages/AdminSchedulePage"
import { AdminServicesPage } from "./pages/AdminServicesPage"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

function AppInit({ children }: { children: React.ReactNode }) {
  const init = useAuthStore((s) => s.init)
  const initialized = useAuthStore((s) => s.initialized)

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

  return <>{children}</>
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppInit>
          <Routes>
            <Route path="/" element={<Navigate to="/business/demo" replace />} />
            <Route path="/business/:slug" element={<BusinessPage />} />
            <Route path="/business/:slug/book" element={<BookingPage />} />
            <Route path="/my/appointments" element={<MyAppointmentsPage />} />
            <Route path="/admin/business/:businessId/schedule" element={<AdminSchedulePage />} />
            <Route path="/admin/business/:businessId/services" element={<AdminServicesPage />} />
          </Routes>
        </AppInit>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
