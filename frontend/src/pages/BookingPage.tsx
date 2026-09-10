import { useEffect, useState } from "react"
import { Navigate, useNavigate, useSearchParams } from "react-router-dom"
import { format } from "date-fns"
import { parseDate, getLocalTimeZone, today } from "@internationalized/date"
import { useQueryClient } from "@tanstack/react-query"
import { useSalonContext } from "../context/SalonContext"
import { useServices, useServiceEmployees, createAppointment } from "../hooks/useApi"
import { useBookingStore } from "../stores/bookingStore"
import { useTimeSlotStream } from "../hooks/useTimeSlotStream"
import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"
import { salonPath } from "../lib/salonDomain"
import { resolveImageUrl } from "../lib/images"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "../components/ui/card"
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
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { slug, salon } = useSalonContext()
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

  useTimeSlotStream(slug)

  const urlService = searchParams.get("service")
  const effectiveServiceId = selectedServiceId ?? urlService

  useEffect(() => {
    if (urlService && urlService !== selectedServiceId) {
      setService(urlService)
    }
  }, [urlService, selectedServiceId, setService])

  const { data: services, isLoading: servicesLoading } = useServices(slug)
  const { data: barbers, isLoading: barbersLoading } = useServiceEmployees(
    slug,
    effectiveServiceId ?? "",
  )

  const [booking, setBooking] = useState(false)
  const [error, setError] = useState("")
  const [bookedTime, setBookedTime] = useState<string | null>(null)

  const service = (services ?? []).find((s) => s.id === effectiveServiceId)
  const barber = (barbers ?? []).find((b) => b.id === selectedEmployeeId)

  const handleBook = async () => {
    if (!authenticated) {
      login()
      return
    }
    if (!selectedSlot || !selectedEmployeeId || !effectiveServiceId) return

    setBooking(true)
    setError("")
    try {
      await createAppointment({
        business_id: salon!.business.id,
        service_id: effectiveServiceId,
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
              <Button onClick={() => navigate(salonPath(slug))} className="mt-2">
                Done
              </Button>
            </CardContent>
          </Card>
        </main>
      </div>
    )
  }

  // Booking requires a service: arriving without one redirects to the services
  // page to pick a service first.
  if (!effectiveServiceId) {
    return <Navigate to={salonPath(slug, "/services")} replace />
  }

  if (servicesLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <p className="text-muted-foreground">Loading…</p>
      </div>
    )
  }

  if (!service) {
    return <Navigate to={salonPath(slug, "/services")} replace />
  }

  const serviceImage = service.picture_id
    ? resolveImageUrl(`/api/images/${service.picture_id}`)
    : undefined

  return (
    <div className="min-h-screen bg-background pb-[env(safe-area-inset-bottom)]">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <Button variant="ghost" onClick={() => navigate(salonPath(slug, "/services"))}>
            <ArrowLeft className="size-4" /> Back
          </Button>
          <h1 className="text-lg font-semibold text-foreground">Booking</h1>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
          <aside className="space-y-4">
            <Card>
              {serviceImage && (
                <img
                  src={serviceImage}
                  alt={service.name}
                  className="h-48 w-full rounded-t-2xl object-cover"
                />
              )}
              <CardHeader>
                <CardTitle>{service.name}</CardTitle>
                {service.description && (
                  <CardDescription>{service.description}</CardDescription>
                )}
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div className="flex justify-between gap-4">
                  <span className="text-muted-foreground">
                    {t("booking.confirm.service")}
                  </span>
                  <span className="font-medium text-foreground">{service.name}</span>
                </div>
                <div className="flex justify-between gap-4">
                  <span className="text-muted-foreground">Duration</span>
                  <span className="font-medium text-foreground">
                    {service.duration_minutes} min
                  </span>
                </div>
                {service.price != null && service.price > 0 && (
                  <div className="flex justify-between gap-4">
                    <span className="text-muted-foreground">Price</span>
                    <span className="font-medium text-foreground">
                      ${service.price.toFixed(2)}
                    </span>
                  </div>
                )}
              </CardContent>
            </Card>
          </aside>

          <div className="space-y-4">
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
                        slug={slug}
                        serviceId={effectiveServiceId}
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
          </div>
        </div>
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
