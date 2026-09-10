import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { format } from "date-fns"
import { parseDate, getLocalTimeZone, today } from "@internationalized/date"
import { useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "../stores/authStore"
import { useMyReservations, cancelReservation, markNoShow, type Appointment } from "../hooks/useApi"
import { useI18n } from "../lib/i18n"
import { formatCancellationReason } from "../lib/cancellation"
import { Button } from "../components/ui/button"
import { Label } from "../components/ui/label"
import { Textarea } from "../components/ui/textarea"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { Calendar } from "../components/ui/calendar"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { AddAppointmentDialog } from "../components/AddAppointmentDialog"

const STATUS_STYLES: Record<string, string> = {
  pending: "bg-yellow-500/15 text-yellow-600 dark:text-yellow-400",
  confirmed: "bg-green-500/15 text-green-600 dark:text-green-400",
  completed: "bg-blue-500/15 text-blue-600 dark:text-blue-400",
  cancelled: "bg-red-500/15 text-red-600 dark:text-red-400",
  no_show: "bg-muted text-muted-foreground",
}

function StatusBadge({ status }: { status: string }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
        STATUS_STYLES[status] ?? "bg-muted text-muted-foreground"
      }`}
    >
      {status.replace("_", " ")}
    </span>
  )
}

function isPast(r: Appointment): boolean {
  return new Date().getTime() >= new Date(r.start_time).getTime()
}

function canMarkNoShow(r: Appointment): boolean {
  if (r.status !== "pending" && r.status !== "confirmed") return false
  if (!isPast(r)) return false
  const graceMs = (r.no_show_after_hours ?? 0) * 3600 * 1000
  return new Date().getTime() >= new Date(r.start_time).getTime() + graceMs
}

export function MyReservationsPage() {
  const { businessId } = useParams<{ businessId: string }>()
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const queryClient = useQueryClient()

  const [date, setDate] = useState(() => today(getLocalTimeZone()).toString())
  const { data: reservations, isLoading } = useMyReservations(businessId!, date)
  const { t } = useI18n()

  const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null)
  const [reason, setReason] = useState("")
  const [cancelling, setCancelling] = useState(false)
  const [noShowTarget, setNoShowTarget] = useState<Appointment | null>(null)
  const [addOpen, setAddOpen] = useState(false)
  const [message, setMessage] = useState("")

  if (!authenticated) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to manage your reservations.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: ["my-reservations", businessId, date] })

  const handleCancel = async () => {
    if (!cancelTarget || !reason.trim()) return
    setCancelling(true)
    setMessage("")
    try {
      await cancelReservation(businessId!, cancelTarget.id, reason.trim())
      setMessage("Reservation cancelled.")
      setCancelTarget(null)
      setReason("")
      await refresh()
    } catch {
      setMessage("Failed to cancel reservation.")
    } finally {
      setCancelling(false)
    }
  }

  const handleMarkNoShow = async () => {
    if (!noShowTarget) return
    setMessage("")
    try {
      await markNoShow(businessId!, noShowTarget.id)
      setMessage("Appointment marked as no-show.")
      setNoShowTarget(null)
      await refresh()
    } catch {
      setMessage("Failed to mark no-show.")
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex gap-4 items-center">
          <h1 className="text-xl font-semibold text-foreground">My Reservations</h1>
          <Button onClick={() => setAddOpen(true)}>Add appointment</Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/admin/business/${businessId}/my-schedule`)}
          >
            Reserve time
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Pick a date</CardTitle>
          </CardHeader>
          <CardContent>
            <Calendar
              value={parseDate(date)}
              onChange={(d) => d && setDate(d.toString())}
              minValue={today(getLocalTimeZone())}
              className="mx-auto"
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Reservations for {format(new Date(`${date}T00:00:00`), "EEEE, MMMM d")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {isLoading ? (
              <p className="text-muted-foreground">Loading…</p>
            ) : (reservations ?? []).length === 0 ? (
              <p className="text-muted-foreground">No reservations on this day.</p>
            ) : (
              (reservations ?? []).map((r) => (
                <div key={r.id} className="flex items-center justify-between p-3 bg-muted rounded-md gap-3">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium">{r.service_name || "Service"}</span>
                      <StatusBadge status={r.status} />
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {format(new Date(r.start_time), "h:mm a")} – {format(new Date(r.end_time), "h:mm a")}
                    </div>
                    {r.cancellation_reason && (
                      <div className="text-xs text-muted-foreground mt-1">
                        Reason: {formatCancellationReason(r.cancellation_reason, t)}
                      </div>
                    )}
                  </div>
                  {(r.status === "pending" || r.status === "confirmed") && (
                    <div className="flex items-center gap-2 shrink-0">
                      {!isPast(r) && (
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => {
                            setCancelTarget(r)
                            setReason("")
                          }}
                        >
                          Cancel
                        </Button>
                      )}
                      {canMarkNoShow(r) && (
                        <Button variant="outline" size="sm" onClick={() => setNoShowTarget(r)}>
                          No-show
                        </Button>
                      )}
                    </div>
                  )}
                </div>
              ))
            )}
            {message && (
              <p className={`text-sm ${message.startsWith("Failed") ? "text-red-500" : "text-green-600"}`}>
                {message}
              </p>
            )}
          </CardContent>
        </Card>
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
            <h3 className="text-lg font-semibold text-foreground">Cancel reservation</h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {cancelTarget.service_name || "Service"} at {format(new Date(cancelTarget.start_time), "h:mm a")}
            </p>
            <div className="mt-4 space-y-1">
              <Label>Reason (required)</Label>
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
              <Button
                variant="destructive"
                onClick={handleCancel}
                isDisabled={cancelling || !reason.trim()}
              >
                {cancelling ? "Cancelling…" : "Cancel reservation"}
              </Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={noShowTarget != null}
        title="Mark as no-show?"
        description={
          noShowTarget
            ? `${noShowTarget.service_name || "Service"} at ${format(new Date(noShowTarget.start_time), "h:mm a")}`
            : undefined
        }
        confirmLabel="Mark no-show"
        cancelLabel="Back"
        onConfirm={handleMarkNoShow}
        onCancel={() => setNoShowTarget(null)}
      />

      {addOpen && (
        <AddAppointmentDialog businessId={businessId!} onClose={() => setAddOpen(false)} />
      )}
    </div>
  )
}
