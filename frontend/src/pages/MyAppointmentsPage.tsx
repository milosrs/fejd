import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { format } from "date-fns"
import { useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "../stores/authStore"
import { useMyAppointments, cancelAppointment, type Appointment } from "../hooks/useApi"
import { useI18n } from "../lib/i18n"
import { formatCancellationReason } from "../lib/cancellation"
import { Button } from "../components/ui/button"
import { Label } from "../components/ui/label"
import { Textarea } from "../components/ui/textarea"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"

function canCancel(apt: Appointment): boolean {
  if (apt.status !== "confirmed" && apt.status !== "pending") return false
  const leadMs = (apt.cancellation_lead_hours ?? 0) * 3600 * 1000
  return new Date().getTime() <= new Date(apt.start_time).getTime() - leadMs
}

export function MyAppointmentsPage() {
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const queryClient = useQueryClient()
  const { data: appointments, isLoading } = useMyAppointments()
  const { t } = useI18n()

  const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null)
  const [reason, setReason] = useState("")
  const [cancelling, setCancelling] = useState(false)
  const [message, setMessage] = useState("")

  if (!authenticated) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to view your appointments.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["my-appointments"] })

  const handleCancel = async () => {
    if (!cancelTarget) return
    setCancelling(true)
    setMessage("")
    try {
      await cancelAppointment(cancelTarget.id, reason.trim())
      setMessage("Appointment cancelled.")
      setCancelTarget(null)
      setReason("")
      await refresh()
    } catch {
      setMessage("Failed to cancel appointment.")
    } finally {
      setCancelling(false)
    }
  }

  const list = Array.isArray(appointments) ? appointments : []

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <h1 className="text-xl font-semibold text-foreground">My Appointments</h1>
          <Button variant="outline" size="sm" onClick={() => navigate("/")}>
            Back
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        {isLoading ? (
          <p className="text-muted-foreground">Loading...</p>
        ) : list.length === 0 ? (
          <p className="text-muted-foreground">No appointments yet.</p>
        ) : (
          <div className="space-y-4">
            {list.map((apt: Appointment) => (
              <Card key={apt.id}>
                <CardHeader>
                  <CardTitle className="text-base">
                    {format(new Date(apt.start_time), "EEEE, MMMM d, yyyy 'at' h:mm a")}
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <p className="text-sm text-muted-foreground">
                    Duration: {Math.round((new Date(apt.end_time).getTime() - new Date(apt.start_time).getTime()) / 60000)} min
                    {" · "}Status: <span className={`font-medium ${apt.status === "confirmed" ? "text-green-600" : apt.status === "cancelled" ? "text-red-600" : "text-foreground"}`}>{apt.status}</span>
                  </p>
                  {apt.cancellation_reason && (
                    <p className="text-xs text-muted-foreground">
                      Reason: {formatCancellationReason(apt.cancellation_reason, t)}
                    </p>
                  )}
                  {(apt.status === "confirmed" || apt.status === "pending") && (
                    canCancel(apt) ? (
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => {
                          setCancelTarget(apt)
                          setReason("")
                        }}
                      >
                        Cancel
                      </Button>
                    ) : (
                      <p className="text-xs text-muted-foreground">
                        Cancellation deadline has passed
                        {apt.cancellation_lead_hours ? ` (${apt.cancellation_lead_hours}h notice required)` : ""}.
                      </p>
                    )
                  )}
                </CardContent>
              </Card>
            ))}
            {message && (
              <p className={`text-sm ${message.startsWith("Failed") ? "text-red-500" : "text-green-600"}`}>
                {message}
              </p>
            )}
          </div>
        )}
      </main>

      {cancelTarget && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
          onClick={() => setCancelTarget(null)}
        >
          <div
            className="w-full max-w-sm rounded-2xl border border-border bg-background p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-lg font-semibold text-foreground">Cancel appointment</h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {format(new Date(cancelTarget.start_time), "EEEE, MMMM d 'at' h:mm a")}
            </p>
            <div className="mt-4 space-y-1">
              <Label>Reason (optional)</Label>
              <Textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Why are you cancelling?"
              />
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCancelTarget(null)}>
                Back
              </Button>
              <Button variant="destructive" onClick={handleCancel} isDisabled={cancelling}>
                {cancelling ? "Cancelling…" : "Cancel appointment"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
