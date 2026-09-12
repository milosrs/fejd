import { useNavigate } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"
import { useMyAppointments } from "../hooks/useApi"
import { MyAppointmentsList } from "../components/MyAppointmentsList"
import { Button } from "../components/ui/button"

export function MyAppointmentsPage() {
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { data: appointments, isLoading } = useMyAppointments()

  if (!authenticated) {
    return (
      <div className="min-h-app bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to view your appointments.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  return (
    <div className="min-h-app bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <h1 className="text-xl font-semibold text-foreground">My Appointments</h1>
          <Button variant="outline" size="sm" onClick={() => navigate("/")}>
            Back
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <MyAppointmentsList appointments={appointments} isLoading={isLoading} />
      </main>
    </div>
  )
}
