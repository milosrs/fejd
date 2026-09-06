import { useEffect, useState } from "react"
import { useParams, useNavigate, useSearchParams } from "react-router-dom"
import { format } from "date-fns"
import { parseDate, getLocalTimeZone, today } from "@internationalized/date"
import { useQueryClient } from "@tanstack/react-query"
import { useSalonContext } from "../context/SalonContext"
import { useServices, useServiceEmployees, createAppointment } from "../hooks/useApi"
import { useBookingStore } from "../stores/bookingStore"
import { useTimeSlotStream } from "../hooks/useTimeSlotStream"
import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { Calendar } from "../components/ui/calendar"
import { BarberAvailabilityCard } from "../components/booking/BarberAvailabilityCard"
import { ArrowLeft, CheckCircle2 } from "lucide-react"

function mapBookingError(msg: string, t: (key: string) => string): string {
  switch (msg) {
    case "time slot is no longer available":
      return t("booking.error.taken")
    case "already booked today":
      return t("booking.error.alreadyBooked")
    case "employee does not offer this service":
      return t("booking.error.noService")
    default:
      return t("booking.error.generic")
  }
}

export function BookingPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { salon } = useSalonContext()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { t, ready } = useI18n()
  const queryClient = useQueryClient()

  const {
    selectedServiceId,
    selectedEmployeeId,
    selectedDate,
    selectedSlot,
    setService,
    setDate,
    selectSlot,
    clearSlot,
  } = useBookingStore()

  useTimeSlotStream(slug!)

  const urlService = searchParams.get("service")
  useEffect(() => {
    if (urlService && urlService !== selectedServiceId) {
      setService(urlService)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlService])

  const { data: services } = useServices(slug!)
  const { data: barbers, isLoading: barbersLoading } = useServiceEmployees(
    slug!,
    selectedServiceId ?? "",
  )

  const [booking, setBooking] = useState(false)
  const [error, setError] = useState("")
  const [bookedTime, setBookedTime] = useState<string | null>(null)

  const service = (services ?? []).find((s) => s.id === selectedServiceId)
  const barber = (barbers ?? []).find((b) => b.id === selectedEmployeeId)

  const handleBook = async () => {
    if (!authenticated) {
      login()
      return
    }
    if (!selectedSlot || !selectedEmployeeId || !selectedServiceId) return

    setBooking(true)
    setError("")
    try {
      await createAppointment({
        business_id: salon!.business.id,
        service_id: selectedServiceId,
        business_user_id: selectedEmployeeId,
        start_time: selectedSlot.start_time,
      })
      setBookedTime(selectedSlot.start_time)
    } catch (err) {
      const e = err as { status?: number; body?: { error?: string } }
      const msg = e?.body?.error ?? ""
      setError(mapBookingError(msg, t))
      if (msg === "time slot is no longer available") {
        clearSlot()
        queryClient.invalidateQueries({ queryKey: ["slots", slug] })
      }
    } finally {
      setBooking(false)
    }
  }

  if (bookedTime) {
    return (
    <div className="min-h-screen bg-background pb-[env(safe-area-inset-bottom)]">
        <main className="max-w-4xl mx-auto px-4 py-8">
          <Card>
            <CardContent className="flex flex-col items-center gap-3 py-10 text-center">
              <CheckCircle2 className="size-10 text-green-600" />
              <h2 className="text-xl font-semibold text-foreground">
                {t("booking.success.title")}
              </h2>
              <p className="text-muted-foreground">
                {format(new Date(bookedTime), "EEEE, MMMM d, yyyy 'at' h:mm a")}
              </p>
              <p className="text-sm text-muted-foreground">{t("booking.success.body")}</p>
              <Button onClick={() => navigate(`/${slug}`)} className="mt-2">
                Done
              </Button>
            </CardContent>
          </Card>
        </main>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background pb-[env(safe-area-inset-bottom)]">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <Button variant="ghost" onClick={() => navigate(`/${slug}`)}>
            <ArrowLeft className="size-4" /> Back
          </Button>
          <h1 className="text-lg font-semibold text-foreground">Booking</h1>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-6">
        {!service ? (
          <Card>
            <CardHeader>
              <CardTitle>{t("booking.step.service")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {(services ?? []).filter((s) => s.active).map((s) => (
                <button
                  key={s.id}
                  onClick={() => setService(s.id)}
                  className="flex w-full items-center justify-between rounded-xl border border-border bg-muted/40 px-4 py-3 text-left transition-colors hover:bg-muted"
                >
                  <div>
                    <span className="font-medium text-foreground">{s.name}</span>
                    <span className="ml-3 text-sm text-muted-foreground">
                      {s.duration_minutes} min
                    </span>
                  </div>
                  {s.price != null && s.price > 0 && (
                    <span className="font-medium">${s.price.toFixed(2)}</span>
                  )}
                </button>
              ))}
            </CardContent>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle>{t("booking.step.date")}</CardTitle>
              </CardHeader>
              <CardContent>
                <Calendar
                  value={selectedDate ? parseDate(selectedDate) : undefined}
                  onChange={(date) => date && setDate(date.toString())}
                  minValue={today(getLocalTimeZone())}
                  className="mx-auto"
                />
              </CardContent>
            </Card>

            {selectedDate && (
              <Card>
                <CardHeader>
                  <CardTitle>{t("booking.step.barber")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {barbersLoading ? (
                    <p className="text-muted-foreground">Loading…</p>
                  ) : (barbers ?? []).length === 0 ? (
                    <p className="text-muted-foreground">{t("booking.empty.barbers")}</p>
                  ) : (
                    (barbers ?? []).map((b) => (
                      <BarberAvailabilityCard
                        key={b.id}
                        barber={b}
                        slug={slug!}
                        serviceId={selectedServiceId!}
                        date={selectedDate}
                        selectedStartTime={selectedSlot?.start_time}
                        onSelectSlot={(employeeId, slot) => selectSlot(employeeId, slot)}
                      />
                    ))
                  )}
                </CardContent>
              </Card>
            )}

            {selectedSlot && service && barber && (
              <Card>
                <CardHeader>
                  <CardTitle>{t("booking.confirm.title")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <dl className="space-y-2 text-sm">
                    <Row label={t("booking.confirm.service")} value={service.name} />
                    <Row label={t("booking.confirm.barber")} value={barber.display_name || barber.user_id} />
                    <Row label={t("booking.confirm.date")} value={format(new Date(selectedSlot.start_time), "EEEE, MMMM d, yyyy")} />
                    <Row label={t("booking.confirm.time")} value={format(new Date(selectedSlot.start_time), "h:mm a")} />
                    {service.price != null && service.price > 0 && (
                      <Row label={t("booking.confirm.price")} value={`$${service.price.toFixed(2)}`} />
                    )}
                  </dl>
                  {error && <p className="text-sm text-destructive">{error}</p>}
                  <Button onClick={handleBook} isDisabled={booking || !ready} className="w-full">
                    {!authenticated
                      ? "Login to book"
                      : booking
                        ? "Booking…"
                        : t("booking.confirm.button")}
                  </Button>
                </CardContent>
              </Card>
            )}
          </>
        )}
      </main>
    </div>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-right font-medium text-foreground">{value}</dd>
    </div>
  )
}
