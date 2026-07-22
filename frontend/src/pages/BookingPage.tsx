import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { format } from "date-fns"
import { useBusiness, useAvailableSlots, createAppointment, TimeSlot } from "../hooks/useApi"
import { useBookingStore } from "../stores/bookingStore"
import { useTimeSlotStream } from "../hooks/useTimeSlotStream"
import { useAuthStore } from "../stores/authStore"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { Calendar } from "../components/ui/calendar"
import { ArrowLeft, Clock } from "lucide-react"

export function BookingPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { data: businessData } = useBusiness(slug!)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)

  const {
    selectedServiceId,
    selectedEmployeeId,
    selectedDate,
    selectedSlot,
    setEmployee,
    setDate,
    setSlot,
  } = useBookingStore()

  useTimeSlotStream(slug!)

  const { data: slotsData } = useAvailableSlots(
    slug!,
    selectedServiceId!,
    selectedEmployeeId!,
    selectedDate!,
  )

  const [booking, setBooking] = useState(false)
  const [booked, setBooked] = useState(false)
  const [error, setError] = useState("")

  const service = businessData?.services.find((s) => s.id === selectedServiceId)
  const employee = businessData?.employees.find((e) => e.id === selectedEmployeeId)

  const handleBook = async () => {
    if (!authenticated) {
      login()
      return
    }
    if (!selectedSlot || !selectedEmployeeId || !selectedServiceId || !selectedDate) return

    setBooking(true)
    setError("")
    try {
      await createAppointment({
        business_id: businessData!.business.id,
        service_id: selectedServiceId,
        business_user_id: selectedEmployeeId,
        start_time: selectedSlot.start_time,
      })
      setBooked(true)
    } catch (err: any) {
      setError(err.response?.data?.error || "Booking failed. Slot may no longer be available.")
    } finally {
      setBooking(false)
    }
  }

  if (!service) {
    return (
      <div className="min-h-screen bg-background p-4">
        <Button variant="ghost" onClick={() => navigate(`/business/${slug}`)}>
          <ArrowLeft className="size-4" /> Back
        </Button>
        <p className="text-center text-muted-foreground mt-8">Service not found.</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-4">
          <Button variant="ghost" onClick={() => navigate(`/business/${slug}`)}>
            <ArrowLeft className="size-4" /> Back
          </Button>
          <div>
            <h1 className="text-lg font-semibold text-foreground">Book: {service.name}</h1>
            <p className="text-sm text-muted-foreground">{service.duration_minutes} min</p>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-8">
        {booked ? (
          <Card>
            <CardContent className="pt-6 text-center">
              <h2 className="text-xl font-semibold text-green-600 mb-2">Appointment booked!</h2>
              <p className="text-muted-foreground mb-4">
                {format(new Date(selectedSlot!.start_time), "EEEE, MMMM d, yyyy 'at' h:mm a")}
              </p>
              <Button onClick={() => navigate(`/business/${slug}`)}>Book another</Button>
            </CardContent>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle>1. Select employee</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {(businessData?.employees ?? []).map((emp) => (
                    <Button
                      key={emp.id}
                      variant={selectedEmployeeId === emp.id ? "default" : "outline"}
                      onClick={() => setEmployee(emp.id)}
                    >
                      {emp.display_name || emp.user_id}
                    </Button>
                  ))}
                </div>
              </CardContent>
            </Card>

            {selectedEmployeeId && (
              <Card>
                <CardHeader>
                  <CardTitle>2. Select date</CardTitle>
                </CardHeader>
                <CardContent>
                  <Calendar
                    mode="single"
                    selected={selectedDate ? new Date(selectedDate + "T00:00:00") : undefined}
                    onSelect={(date) => date && setDate(format(date, "yyyy-MM-dd"))}
                    disabled={(date) => {
                      const today = new Date()
                      today.setHours(0, 0, 0, 0)
                      return date < today
                    }}
                    className="mx-auto"
                  />
                </CardContent>
              </Card>
            )}

            {selectedDate && slotsData && (
              <Card>
                <CardHeader>
                  <CardTitle>3. Select time slot</CardTitle>
                </CardHeader>
                <CardContent>
                  {slotsData.slots.length === 0 ? (
                    <p className="text-muted-foreground">No available slots for this date.</p>
                  ) : (
                    <>
                      <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-2">
                        {slotsData.slots.map((slot: TimeSlot) => (
                          <SlotButton
                            key={slot.start_time}
                            slot={slot}
                            selected={selectedSlot?.start_time === slot.start_time}
                            onSelect={() => setSlot(slot)}
                          />
                        ))}
                      </div>

                      {selectedSlot && (
                        <div className="mt-6 border-t border-border pt-4">
                          <p className="text-sm text-muted-foreground mb-2">
                            Selected: {format(new Date(selectedSlot.start_time), "EEEE, MMMM d, yyyy 'at' h:mm a")}
                          </p>
                          {error && <p className="text-sm text-red-500 mb-2">{error}</p>}
                          <Button onClick={handleBook} disabled={booking} className="w-full">
                            {booking ? "Booking..." : authenticated ? "Confirm Booking" : "Login to Book"}
                          </Button>
                        </div>
                      )}
                    </>
                  )}
                </CardContent>
              </Card>
            )}
          </>
        )}
      </main>
    </div>
  )
}

function SlotButton({
  slot,
  selected,
  onSelect,
}: {
  slot: TimeSlot
  selected: boolean
  onSelect: () => void
}) {
  const [removing, setRemoving] = useState(false)
  const [removed, setRemoved] = useState(false)

  const startTime = new Date(slot.start_time)

  return (
    <button
      className={`px-3 py-2 rounded-md text-sm font-medium transition-all duration-300
        ${removing ? "opacity-0 scale-95 pointer-events-none" : "opacity-100"}
        ${removed ? "hidden" : ""}
        ${selected ? "bg-primary text-primary-foreground" : "bg-muted hover:bg-muted/80 text-foreground"}`}
      onClick={removed ? undefined : onSelect}
      disabled={removing || removed}
    >
      {format(startTime, "h:mm a")}
    </button>
  )
}
