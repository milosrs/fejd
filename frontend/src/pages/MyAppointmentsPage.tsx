import { useState } from "react"
import { useNavigate } from "react-router"
import { CalendarDate, getLocalTimeZone, today } from "@internationalized/date"
import { useAuthStore } from "../stores/authStore"
import { useMyAppointments } from "../hooks/useApi"
import { MyAppointmentsList } from "../components/MyAppointmentsList"
import { CalendarGrid } from "../components/reservations/CalendarGrid"
import { toneFor, type ReservationTag } from "../components/reservations/types"
import { Button } from "../components/ui/button"
import { useI18n } from "../lib/i18n"

export function MyAppointmentsPage() {
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { data: appointments, isLoading } = useMyAppointments()
  const { t } = useI18n()

  const todayDate = today(getLocalTimeZone())
  const [month, setMonth] = useState(() => new CalendarDate(todayDate.year, todayDate.month, 1))
  const [selected, setSelected] = useState<CalendarDate>(() => todayDate)

  const eventsByDay: Record<string, ReservationTag[]> = {}
  for (const a of appointments ?? []) {
    const d = new Date(a.start_time)
    const key = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, "0")}-${String(d.getUTCDate()).padStart(2, "0")}`
    ;(eventsByDay[key] ??= []).push({
      id: a.id,
      label: a.business_name || t("appointments.title"),
      tone: toneFor(a.business_name || a.id),
      muted: a.status === "cancelled",
    })
  }

  if (!authenticated) {
    return (
      <div className="min-h-app bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">{t("appointments.loginPrompt")}</p>
        <Button onClick={login}>{t("common.logIn")}</Button>
      </div>
    )
  }

  return (
    <div className="min-h-app bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <h1 className="text-xl font-semibold text-foreground">{t("appointments.title")}</h1>
          <Button variant="outline" size="sm" onClick={() => navigate("/")}>
            {t("common.back")}
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-8">
        <CalendarGrid
          month={month}
          onMonthChange={setMonth}
          selected={selected}
          onSelect={setSelected}
          today={todayDate}
          eventsByDay={eventsByDay}
        />

        <section className="space-y-4">
          <h2 className="text-lg font-semibold text-foreground">{t("appointments.history")}</h2>
          <MyAppointmentsList appointments={appointments} isLoading={isLoading} />
        </section>
      </main>
    </div>
  )
}
