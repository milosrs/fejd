import { useState } from "react"
import { format } from "date-fns"
import { useQueryClient } from "@tanstack/react-query"
import { useMyServices, useCustomers, bookOwnAppointment } from "../hooks/useApi"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

export function AddAppointmentDialog({
  businessId,
  onClose,
}: {
  businessId: string
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const { data: services } = useMyServices(businessId)
  const { data: customers } = useCustomers(businessId)

  const [date, setDate] = useState(() => format(new Date(), "yyyy-MM-dd"))
  const [time, setTime] = useState("09:00")
  const [serviceId, setServiceId] = useState("")
  const [customerId, setCustomerId] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  const handleSave = async () => {
    if (!serviceId || !date || !time) {
      setError("Choose a service and a time.")
      return
    }
    setSaving(true)
    setError("")
    try {
      await bookOwnAppointment(businessId, {
        service_id: serviceId,
        start_time: new Date(`${date}T${time}:00`).toISOString(),
        customer_user_id: customerId || undefined,
      })
      await queryClient.invalidateQueries({ queryKey: ["my-reservations", businessId] })
      onClose()
    } catch {
      setError("Failed to add appointment.")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-semibold text-foreground">Add appointment</h3>
        <div className="mt-4 space-y-3">
          <div className="space-y-1">
            <Label htmlFor="appt-service">Service</Label>
            <select
              id="appt-service"
              className="w-full h-8 rounded-2xl border border-border bg-input/50 px-2.5 text-sm"
              value={serviceId}
              onChange={(e) => setServiceId(e.target.value)}
            >
              <option value="">-- Select --</option>
              {(services ?? []).map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label htmlFor="appt-date">Date</Label>
              <Input id="appt-date" type="date" value={date} onChange={(e) => setDate(e.target.value)} />
            </div>
            <div className="space-y-1">
              <Label htmlFor="appt-time">Time</Label>
              <Input id="appt-time" type="time" value={time} onChange={(e) => setTime(e.target.value)} />
            </div>
          </div>
          <div className="space-y-1">
            <Label htmlFor="appt-customer">Customer (optional)</Label>
            <select
              id="appt-customer"
              className="w-full h-8 rounded-2xl border border-border bg-input/50 px-2.5 text-sm"
              value={customerId}
              onChange={(e) => setCustomerId(e.target.value)}
            >
              <option value="">No customer</option>
              {(customers ?? []).map((c) => (
                <option key={c.user_id} value={c.user_id}>
                  {c.display_name || c.user_id}
                </option>
              ))}
            </select>
          </div>
        </div>
        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave} isDisabled={saving}>
            {saving ? "Adding…" : "Add appointment"}
          </Button>
        </div>
      </div>
    </div>
  )
}
