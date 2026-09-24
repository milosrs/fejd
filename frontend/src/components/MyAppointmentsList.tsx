import { useState } from "react"
import { format } from "date-fns"
import { useQueryClient } from "@tanstack/react-query"
import { cancelAppointment, type Appointment } from "../hooks/useApi"
import { useI18n } from "../lib/i18n"
import { formatCancellationReason } from "../lib/cancellation"
import { resolveImageUrl } from "../lib/images"
import { Button } from "./ui/button"
import { Label } from "./ui/label"
import { Textarea } from "./ui/textarea"
import { Card, CardTitle, CardContent } from "./ui/card"

function canCancel(apt: Appointment): boolean {
  if (apt.status !== "confirmed" && apt.status !== "pending") return false
  const leadMs = (apt.cancellation_lead_hours ?? 0) * 3600 * 1000
  return new Date().getTime() <= new Date(apt.start_time).getTime() - leadMs
}

const STATUS_LABELS: Record<string, string> = {
  pending: "appointments.status.pending",
  confirmed: "appointments.status.confirmed",
  completed: "appointments.status.completed",
  cancelled: "appointments.status.cancelled",
  no_show: "appointments.status.noShow",
}

const STATUS_STYLES: Record<string, string> = {
  pending: "bg-orange-500/15 text-orange-600 dark:text-orange-400",
  confirmed: "bg-green-500/15 text-green-600 dark:text-green-400",
  completed: "bg-blue-500/15 text-blue-600 dark:text-blue-400",
  cancelled: "bg-red-500/15 text-red-600 dark:text-red-400",
  no_show: "bg-muted text-muted-foreground",
}

function StatusBadge({ status }: { status: string }) {
  const { t } = useI18n()
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
        STATUS_STYLES[status] ?? "bg-muted text-muted-foreground"
      }`}
    >
      {t(STATUS_LABELS[status] ?? `status.${status}`)}
    </span>
  )
}

// MyAppointmentsList renders the appointments the current user booked
// themselves (as a customer), shared by the account and staff views.
export function MyAppointmentsList({
  appointments,
  isLoading,
}: {
  appointments?: Appointment[]
  isLoading?: boolean
}) {
  const queryClient = useQueryClient()
  const { t } = useI18n()
  const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null)
  const [reason, setReason] = useState("")
  const [cancelling, setCancelling] = useState(false)
  const [message, setMessage] = useState("")

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["my-appointments"] })

  const handleCancel = async () => {
    if (!cancelTarget) return
    setCancelling(true)
    setMessage("")
    try {
      await cancelAppointment(cancelTarget.id, reason.trim())
      setMessage(t("appointments.cancelled"))
      setCancelTarget(null)
      setReason("")
      await refresh()
    } catch {
      setMessage(t("appointments.cancelFailed"))
    } finally {
      setCancelling(false)
    }
  }

  const list = Array.isArray(appointments) ? appointments : []

  return (
    <>
      {isLoading ? (
        <p className="text-muted-foreground">{t("common.loading")}</p>
      ) : list.length === 0 ? (
        <p className="text-muted-foreground">{t("appointments.empty")}</p>
      ) : (
        <div className="space-y-4">
          {list.map((apt: Appointment) => {
            const logoUrl = resolveImageUrl(apt.business_logo)
            const fallback = (apt.business_name ?? "?").charAt(0).toUpperCase()
            return (
              <Card key={apt.id}>
                <CardContent className="flex items-center gap-4">
                  <div className="shrink-0">
                    {logoUrl ? (
                      <img
                        src={logoUrl}
                        alt={apt.business_name ?? ""}
                        className="size-12 rounded-full border border-border bg-background object-cover"
                      />
                    ) : (
                      <span className="flex size-12 items-center justify-center rounded-full bg-muted text-lg font-semibold text-foreground">
                        {fallback}
                      </span>
                    )}
                  </div>
                  <div className="min-w-0 flex-1 space-y-2">
                    <CardTitle>
                      {format(new Date(apt.start_time), "EEEE, MMMM d, yyyy 'at' h:mm a")}
                    </CardTitle>
                    <div className="flex flex-wrap items-center gap-2">
                      <StatusBadge status={apt.status} />
                      <span className="text-sm text-muted-foreground">
                        {t("appointments.duration", {
                          duration: Math.round((new Date(apt.end_time).getTime() - new Date(apt.start_time).getTime()) / 60000),
                        })}
                      </span>
                    </div>
                    {apt.cancellation_reason && (
                      <p className="text-xs text-muted-foreground">
                        {t("appointments.reason", { reason: formatCancellationReason(apt.cancellation_reason, t) })}
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
                          {t("reservations.cancel")}
                        </Button>
                      ) : (
                        <p className="text-xs text-muted-foreground">
                          {t("appointments.deadlinePassed")}
                          {apt.cancellation_lead_hours ? ` (${t("appointments.noticeRequired", { hours: apt.cancellation_lead_hours })})` : ""}.
                        </p>
                      )
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
          {message && (
            <p className={`text-sm ${message.startsWith("Failed") ? "text-red-500" : "text-green-600"}`}>
              {message}
            </p>
          )}
        </div>
      )}

      {cancelTarget && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
          onClick={() => setCancelTarget(null)}
        >
          <div
            className="w-full max-w-sm rounded-2xl border border-border bg-background p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-lg font-semibold text-foreground">{t("appointments.cancelTitle")}</h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {format(new Date(cancelTarget.start_time), "EEEE, MMMM d 'at' h:mm a")}
            </p>
            {cancelTarget.status === "confirmed" && (
              <div className="mt-4 space-y-1">
                <Label>{t("appointments.reasonLabel")}</Label>
                <Textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder={t("appointments.reasonPlaceholder")}
                />
              </div>
            )}
            <div className="mt-6 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCancelTarget(null)}>
                {t("common.back")}
              </Button>
              <Button variant="destructive" onClick={handleCancel} isDisabled={cancelling}>
                {cancelling ? t("appointments.cancelling") : t("appointments.cancelTitle")}
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
