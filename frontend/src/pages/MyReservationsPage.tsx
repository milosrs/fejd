import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { format } from "date-fns"
import { CalendarDate, getLocalTimeZone, today } from "@internationalized/date"
import { useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "../stores/authStore"
import { useMe } from "../hooks/useMe"
import {
  useMyReservationsMonth,
  useMyAppointments,
  useBusinessAppointments,
  useBusinessUnavailability,
  useAdminEmployees,
  useCustomers,
  cancelReservation,
  markNoShow,
  acceptAppointment,
  acceptUnavailability,
  rejectAppointment,
  rejectUnavailability,
  type Appointment,
} from "../hooks/useApi"
import { useI18n } from "../lib/i18n"
import { formatCancellationReason } from "../lib/cancellation"
import { Button } from "../components/ui/button"
import { Label } from "../components/ui/label"
import { Textarea } from "../components/ui/textarea"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { AddAppointmentDialog } from "../components/AddAppointmentDialog"
import { MyAppointmentsList } from "../components/MyAppointmentsList"
import { CalendarGrid } from "../components/reservations/CalendarGrid"
import { ScheduledPanel } from "../components/reservations/ScheduledPanel"
import { toneFor, type ReservationTag, type ScheduleEvent } from "../components/reservations/types"

const STATUS_STYLES: Record<string, string> = {
  pending: "bg-yellow-500/15 text-yellow-600 dark:text-yellow-400",
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
      {t(`status.${status}`)}
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

  const todayDate = today(getLocalTimeZone())
  const [month, setMonth] = useState(() => new CalendarDate(todayDate.year, todayDate.month, 1))
  const [selected, setSelected] = useState<CalendarDate>(() => todayDate)

  const { byDay, isLoading: reservationsLoading } = useMyReservationsMonth(businessId!, month)
  const { data: ownAppointments, isLoading: ownLoading } = useMyAppointments()
  const { data: me } = useMe()
  const isOwner = me?.businesses.some((b) => b.id === businessId && b.role === "admin")
  const { data: businessAppointments } = useBusinessAppointments(isOwner ? businessId! : "")
  const { data: businessUnavailability } = useBusinessUnavailability(isOwner ? businessId! : "")
  const { data: employees } = useAdminEmployees(isOwner ? businessId! : "")
  const { data: customers } = useCustomers(businessId!)
  const { t } = useI18n()

  const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null)
  const [reason, setReason] = useState("")
  const [cancelling, setCancelling] = useState(false)
  const [noShowTarget, setNoShowTarget] = useState<Appointment | null>(null)
  const [rejectTarget, setRejectTarget] = useState<{
    kind: "appointment" | "unavailability"
    id: string
    label: string
  } | null>(null)
  const [rejectReason, setRejectReason] = useState("")
  const [rejecting, setRejecting] = useState(false)
  const [addOpen, setAddOpen] = useState(false)
  const [message, setMessage] = useState("")

  if (!authenticated) {
    return (
      <div className="min-h-app bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">{t("reservations.loginPrompt")}</p>
        <Button onClick={login}>{t("common.logIn")}</Button>
      </div>
    )
  }

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: ["my-reservations", businessId] })

  const handleCancel = async () => {
    if (!cancelTarget || !reason.trim()) return
    setCancelling(true)
    setMessage("")
    try {
      await cancelReservation(businessId!, cancelTarget.id, reason.trim())
      setMessage(t("reservations.cancelled"))
      setCancelTarget(null)
      setReason("")
      await refresh()
    } catch {
      setMessage(t("reservations.cancelFailed"))
    } finally {
      setCancelling(false)
    }
  }

  const handleMarkNoShow = async () => {
    if (!noShowTarget) return
    setMessage("")
    try {
      await markNoShow(businessId!, noShowTarget.id)
      setMessage(t("reservations.noShowMarked"))
      setNoShowTarget(null)
      await refresh()
    } catch {
      setMessage(t("reservations.noShowFailed"))
    }
  }

  const handleAccept = async (r: Appointment) => {
    setMessage("")
    try {
      await acceptAppointment(businessId!, r.id)
      setMessage(t("reservations.accepted"))
      await refresh()
      await queryClient.invalidateQueries({ queryKey: ["business-appointments", businessId] })
    } catch {
      setMessage(t("reservations.acceptFailed"))
    }
  }

  const handleAcceptUnavailability = async (id: string) => {
    setMessage("")
    try {
      await acceptUnavailability(businessId!, id)
      setMessage(t("reservations.timeAccepted"))
      await queryClient.invalidateQueries({ queryKey: ["business-unavailability", businessId] })
    } catch {
      setMessage(t("reservations.timeAcceptFailed"))
    }
  }

  const handleReject = async () => {
    if (!rejectTarget || !rejectReason.trim()) return
    setRejecting(true)
    setMessage("")
    try {
      if (rejectTarget.kind === "appointment") {
        await rejectAppointment(businessId!, rejectTarget.id, rejectReason.trim())
        await refresh()
        await queryClient.invalidateQueries({ queryKey: ["business-appointments", businessId] })
      } else {
        await rejectUnavailability(businessId!, rejectTarget.id, rejectReason.trim())
        await queryClient.invalidateQueries({ queryKey: ["business-unavailability", businessId] })
      }
      setMessage(t("reservations.rejected"))
      setRejectTarget(null)
      setRejectReason("")
    } catch {
      setMessage(t("reservations.rejectFailed"))
    } finally {
      setRejecting(false)
    }
  }

  const employeeName = (businessUserId: string) =>
    (employees ?? []).find((e) => e.id === businessUserId)?.display_name ?? t("reservations.employee")

  const customerName = (userId?: string) =>
    (customers ?? []).find((c) => c.user_id === userId)?.display_name?.trim() || t("reservations.customer")

  const pendingAppointments = (businessAppointments ?? []).filter((a) => a.status === "pending")
  const pendingUnavailability = (businessUnavailability ?? []).filter((u) => u.status === "pending")

  const goToDate = (d: CalendarDate) => {
    setSelected(d)
    if (d.year !== month.year || d.month !== month.month) {
      setMonth(new CalendarDate(d.year, d.month, 1))
    }
  }

  const selectedKey = selected.toString()
  const reservations = byDay[selectedKey] ?? []

  const eventsByDay: Record<string, ReservationTag[]> = {}
  for (const [key, list] of Object.entries(byDay)) {
    eventsByDay[key] = list
      .filter((r) => r.status !== "cancelled")
      .map((r) => ({
        id: r.id,
        label: r.service_name || t("reservations.service"),
        tone: toneFor(r.service_name || r.id),
      }))
  }

  const reservationsById = new Map(reservations.map((r) => [r.id, r]))
  const events: ScheduleEvent[] = reservations.map((r) => {
    const subtitle =
      r.status === "cancelled" && r.cancellation_reason
        ? `${customerName(r.customer_user_id)} · ${formatCancellationReason(r.cancellation_reason, t)}`
        : customerName(r.customer_user_id)
    return {
      id: r.id,
      title: r.service_name || t("reservations.service"),
      subtitle,
      start: new Date(r.start_time),
      end: new Date(r.end_time),
      tone: toneFor(r.service_name || r.id),
    }
  })

  const renderActions = (event: ScheduleEvent) => {
    const r = reservationsById.get(event.id)
    if (!r) return null
    return (
      <>
        <StatusBadge status={r.status} />
        {r.status === "pending" && (
          <>
            <Button size="sm" onClick={() => handleAccept(r)}>
              {t("reservations.accept")}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() =>
                setRejectTarget({
                  kind: "appointment",
                  id: r.id,
                  label: `${r.service_name || t("reservations.service")} at ${format(new Date(r.start_time), "h:mm a")}`,
                })
              }
            >
              {t("reservations.reject")}
            </Button>
          </>
        )}
        {!isPast(r) && (
          <Button
            variant="destructive"
            size="sm"
            onClick={() => {
              setCancelTarget(r)
              setReason("")
            }}
          >
            {t("reservations.cancel")}
          </Button>
        )}
        {canMarkNoShow(r) && (
          <Button variant="outline" size="sm" onClick={() => setNoShowTarget(r)}>
            {t("reservations.noShow")}
          </Button>
        )}
      </>
    )
  }

  return (
    <div className="min-h-app bg-background">
      <header className="border-b border-border">
        <div className="max-w-6xl mx-auto px-4 py-4 flex gap-4 items-center">
          <h1 className="text-xl font-semibold text-foreground">{t("reservations.title")}</h1>
          <Button onClick={() => setAddOpen(true)}>{t("reservations.addAppointment")}</Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/admin/business/${businessId}/my-schedule`)}
          >
            {t("reservations.reserveTime")}
          </Button>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8 space-y-6">
        {message && (
          <p className={`text-sm ${message.startsWith("Failed") ? "text-red-500" : "text-green-600"}`}>
            {message}
          </p>
        )}

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,1fr)]">
          <CalendarGrid
            month={month}
            onMonthChange={setMonth}
            selected={selected}
            onSelect={goToDate}
            today={todayDate}
            eventsByDay={eventsByDay}
          />
          <ScheduledPanel
            date={selected}
            onPrevDay={() => goToDate(selected.subtract({ days: 1 }))}
            onNextDay={() => goToDate(selected.add({ days: 1 }))}
            onOpenCalendar={() => goToDate(todayDate)}
            events={events}
            isLoading={reservationsLoading}
            renderActions={renderActions}
          />
        </div>

        <Card>
          <CardHeader>
            <CardTitle>{t("reservations.myAppointments")}</CardTitle>
          </CardHeader>
          <CardContent>
            <MyAppointmentsList appointments={ownAppointments} isLoading={ownLoading} />
          </CardContent>
        </Card>

        {isOwner && (
          <Card>
            <CardHeader>
              <CardTitle>{t("reservations.pendingApprovals")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {pendingAppointments.length === 0 && pendingUnavailability.length === 0 ? (
                <p className="text-muted-foreground">{t("reservations.nothingToApprove")}</p>
              ) : (
                <>
                  {pendingAppointments.length > 0 && (
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-muted-foreground">{t("reservations.customerAppointments")}</p>
                      {pendingAppointments.map((a) => (
                        <div key={a.id} className="flex items-center justify-between p-3 bg-muted rounded-md gap-3">
                          <div className="min-w-0">
                            <div className="font-medium">{a.service_name || t("reservations.service")}</div>
                            <div className="text-sm text-muted-foreground">
                              {employeeName(a.business_user_id)} · {format(new Date(a.start_time), "EEE, MMM d · h:mm a")}
                            </div>
                          </div>
                          <div className="flex items-center gap-2 shrink-0">
                            <Button size="sm" onClick={() => handleAccept(a)}>
                              {t("reservations.accept")}
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                setRejectTarget({
                                  kind: "appointment",
                                  id: a.id,
                                  label: `${a.service_name || t("reservations.service")} · ${employeeName(a.business_user_id)}`,
                                })
                              }
                            >
                              {t("reservations.reject")}
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                  {pendingUnavailability.length > 0 && (
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-muted-foreground">{t("reservations.reservedTime")}</p>
                      {pendingUnavailability.map((u) => (
                        <div key={u.id} className="flex items-center justify-between p-3 bg-muted rounded-md gap-3">
                          <div className="min-w-0">
                            <div className="font-medium">{employeeName(u.business_user_id)}</div>
                            <div className="text-sm text-muted-foreground">
                              {format(new Date(u.start_time), "EEE, MMM d · h:mm a")} – {format(new Date(u.end_time), "h:mm a")}
                              {u.reason ? ` (${u.reason})` : ""}
                            </div>
                          </div>
                          <div className="flex items-center gap-2 shrink-0">
                            <Button size="sm" onClick={() => handleAcceptUnavailability(u.id)}>
                              {t("reservations.accept")}
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() =>
                                setRejectTarget({
                                  kind: "unavailability",
                                  id: u.id,
                                  label: `${employeeName(u.business_user_id)} · ${format(new Date(u.start_time), "EEE, MMM d · h:mm a")}`,
                                })
                              }
                            >
                              {t("reservations.reject")}
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>
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
            <h3 className="text-lg font-semibold text-foreground">{t("reservations.cancelReservation")}</h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {cancelTarget.service_name || t("reservations.service")} at {format(new Date(cancelTarget.start_time), "h:mm a")}
            </p>
            <div className="mt-4 space-y-1">
              <Label>{t("reservations.reasonRequired")}</Label>
              <Textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder={t("reservations.cancelPlaceholder")}
              />
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCancelTarget(null)}>
                {t("common.back")}
              </Button>
              <Button
                variant="destructive"
                onClick={handleCancel}
                isDisabled={cancelling || !reason.trim()}
              >
                {cancelling ? t("reservations.cancelling") : t("reservations.cancelReservation")}
              </Button>
            </div>
          </div>
        </div>
      )}

      {rejectTarget && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
          onClick={() => setRejectTarget(null)}
        >
          <div
            className="w-full max-w-sm rounded-2xl border border-border bg-background p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-lg font-semibold text-foreground">{t("reservations.reject")}</h3>
            <p className="mt-2 text-sm text-muted-foreground">{rejectTarget.label}</p>
            <div className="mt-4 space-y-1">
              <Label>{t("reservations.reasonRequired")}</Label>
              <Textarea
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                placeholder={t("reservations.rejectPlaceholder")}
              />
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setRejectTarget(null)}>
                {t("common.back")}
              </Button>
              <Button
                variant="destructive"
                onClick={handleReject}
                isDisabled={rejecting || !rejectReason.trim()}
              >
                {rejecting ? t("reservations.rejecting") : t("reservations.reject")}
              </Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={noShowTarget != null}
        title={t("reservations.noShowTitle")}
        description={
          noShowTarget
            ? `${noShowTarget.service_name || t("reservations.service")} at ${format(new Date(noShowTarget.start_time), "h:mm a")}`
            : undefined
        }
        confirmLabel={t("reservations.markNoShow")}
        cancelLabel={t("common.back")}
        onConfirm={handleMarkNoShow}
        onCancel={() => setNoShowTarget(null)}
      />

      {addOpen && (
        <AddAppointmentDialog businessId={businessId!} onClose={() => setAddOpen(false)} />
      )}
    </div>
  )
}
