import { useEffect, useState } from "react"
import { useParams, useNavigate, useSearchParams } from "react-router-dom"
import { format } from "date-fns"
import { parseDate, getLocalTimeZone, today } from "@internationalized/date"
import { useQueryClient } from "@tanstack/react-query"
import { useSalonContext } from "../context/SalonContext"
import {
  useServices,
  useServiceEmployees,
  useAvailableSlots,
  createAppointment,
  type TimeSlot,
} from "../hooks/useApi"
import { useBookingStore } from "../stores/bookingStore"
import { useTimeSlotStream } from "../hooks/useTimeSlotStream"
import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { Calendar } from "../components/ui/calendar"
import { ArrowLeft, CheckCircle2, UserRound } from "lucide-react"
import { resolveImageUrl } from "../lib/images"

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
    setEmployee,
    setDate,
    setSlot,
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
  const { data: slotsData } = useAvailableSlots(
    slug!,
    selectedServiceId ?? "",
    selectedEmployeeId ?? "",
    selectedDate ?? "",
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
        setSlot(null)
        queryClient.invalidateQueries({ queryKey: ["slots", slug] })
      }
    } finally {
      setBooking(false)
    }
  }

  if (bookedTime) {
    return (
      <div className="min-h-screen bg-background">
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
    <div className="min-h-screen bg-background">
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
                <CardTitle>{t("booking.step.barber")}</CardTitle>
              </CardHeader>
              <CardContent>
                {barbersLoading ? (
                  <p className="text-muted-foreground">Loading…</p>
                ) : (barbers ?? []).length === 0 ? (
                  <p className="text-muted-foreground">{t("booking.empty.barbers")}</p>
                ) : (
                  <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
                    {(barbers ?? []).map((b) => {
                      const avatar = resolveImageUrl(b.avatar)
                      const selected = b.id === selectedEmployeeId
                      return (
                        <button
                          key={b.id}
                          onClick={() => setEmployee(b.id)}
                          className={`flex flex-col items-center gap-2 rounded-xl border p-4 transition-colors ${
                            selected
                              ? "border-primary bg-primary/10"
                              : "border-border hover:bg-muted"
                          }`}
                        >
                          {avatar ? (
                            <img
                              src={avatar}
                              alt={b.display_name || b.user_id}
                              className="h-16 w-16 rounded-full object-cover"
                            />
                          ) : (
                            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-muted text-muted-foreground">
                              <UserRound className="size-7" />
                            </div>
                          )}
                          <span className="text-sm font-medium text-foreground">
                            {b.display_name || b.user_id}
                          </span>
                        </button>
                      )
                    })}
                  </div>
                )}
              </CardContent>
            </Card>

            {barber && (
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
            )}

            {selectedDate && slotsData && (
              <Card>
                <CardHeader>
                  <CardTitle>{t("booking.step.time")}</CardTitle>
                </CardHeader>
                <CardContent>
                  {slotsData.slots.length === 0 ? (
                    <p className="text-muted-foreground">{t("booking.empty.slots")}</p>
                  ) : (
                    <div className="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-6">
                      {slotsData.slots.map((slot: TimeSlot) => (
                        <button
                          key={slot.start_time}
                          onClick={() => setSlot(slot)}
                          className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                            selectedSlot?.start_time === slot.start_time
                              ? "bg-primary text-primary-foreground"
                              : "bg-muted text-foreground hover:bg-muted/70"
                          }`}
                        >
                          {format(new Date(slot.start_time), "h:mm a")}
                        </button>
                      ))}
                    </div>
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
