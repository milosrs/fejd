import { useNavigate } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"
import { useMyAppointments } from "../hooks/useApi"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { format } from "date-fns"

export function MyAppointmentsPage() {
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { data: appointments, isLoading } = useMyAppointments()

  if (!authenticated) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to view your appointments.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <h1 className="text-xl font-semibold text-foreground">My Appointments</h1>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        {isLoading ? (
          <p className="text-muted-foreground">Loading...</p>
        ) : (appointments && (Array.isArray(appointments) ? appointments : []).length === 0) ? (
          <p className="text-muted-foreground">No appointments yet.</p>
        ) : (
          <div className="space-y-4">
            {(Array.isArray(appointments) ? appointments : []).map((apt: any) => (
              <Card key={apt.id}>
                <CardHeader>
                  <CardTitle className="text-base">
                    {format(new Date(apt.start_time), "EEEE, MMMM d, yyyy 'at' h:mm a")}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">
                    Duration: {Math.round((new Date(apt.end_time).getTime() - new Date(apt.start_time).getTime()) / 60000)} min
                    {" · "}Status: <span className={`font-medium ${apt.status === "confirmed" ? "text-green-600" : apt.status === "cancelled" ? "text-red-600" : "text-foreground"}`}>{apt.status}</span>
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
