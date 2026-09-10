import { useEffect, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useSalonPolicy, updateSalonPolicy } from "../hooks/useApi"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

const DAYS = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]

interface HourRow {
  day_of_week: number
  start_time: string
  end_time: string
}

const defaultHours = (): HourRow[] =>
  DAYS.map((_, i) => ({ day_of_week: i, start_time: "09:00", end_time: "17:00" }))

export function SalonPolicyDialog({
  businessId,
  slug,
  onClose,
}: {
  businessId: string
  slug: string
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const { data: policy } = useSalonPolicy(businessId)

  const [leadHours, setLeadHours] = useState("")
  const [noShowHours, setNoShowHours] = useState("")
  const [slotInterval, setSlotInterval] = useState("30")
  const [hours, setHours] = useState<HourRow[]>(defaultHours())
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!policy) return
    setLeadHours(String(policy.cancellation_lead_hours))
    setNoShowHours(String(policy.no_show_after_hours))
    setSlotInterval(String(policy.slot_interval_minutes))
    if (policy.working_hours?.length) {
      setHours(policy.working_hours.map((wh) => ({ ...wh })))
    }
  }, [policy])

  const updateRow = (index: number, field: keyof HourRow, value: string) => {
    setHours((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], [field]: value }
      return next
    })
  }

  const handleSave = async () => {
    const lead = parseInt(leadHours, 10)
    const noShow = parseInt(noShowHours, 10)
    const interval = parseInt(slotInterval, 10)
    if (Number.isNaN(lead) || lead < 0) {
      setError("Cancellation notice must be zero or greater.")
      return
    }
    if (Number.isNaN(noShow) || noShow < 0) {
      setError("No-show grace must be zero or greater.")
      return
    }
    if (Number.isNaN(interval) || interval <= 0) {
      setError("Slot interval must be greater than zero.")
      return
    }
    setSaving(true)
    setError("")
    try {
      await updateSalonPolicy(businessId, {
        cancellation_lead_hours: lead,
        no_show_after_hours: noShow,
        slot_interval_minutes: interval,
        working_hours: hours,
      })
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      await queryClient.invalidateQueries({ queryKey: ["salon-policy", businessId] })
      onClose()
    } catch {
      setError("Failed to save policy.")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 pt-[calc(1rem+env(safe-area-inset-top))] pb-[calc(1rem+env(safe-area-inset-bottom))]"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md max-h-[90vh] overflow-y-auto rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-semibold text-foreground">Salon policy</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Booking rules, time slots and default working hours.
        </p>

        <div className="mt-4 space-y-4">
          <div className="space-y-1">
            <Label htmlFor="cancellation-lead-hours">Cancellation notice (hours before)</Label>
            <Input
              id="cancellation-lead-hours"
              type="number"
              min={0}
              value={leadHours}
              onChange={(e) => setLeadHours(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="no-show-after-hours">No-show grace (hours after)</Label>
            <Input
              id="no-show-after-hours"
              type="number"
              min={0}
              value={noShowHours}
              onChange={(e) => setNoShowHours(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="slot-interval">Time slot interval (minutes)</Label>
            <Input
              id="slot-interval"
              type="number"
              min={5}
              step={5}
              value={slotInterval}
              onChange={(e) => setSlotInterval(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label>Working hours</Label>
            <div className="space-y-2 rounded-xl border border-border p-3">
              {DAYS.map((day, i) => (
                <div key={day} className="flex items-center gap-2">
                  <span className="w-24 text-sm text-muted-foreground">{day}</span>
                  <Input
                    type="time"
                    className="w-28"
                    value={hours[i]?.start_time || ""}
                    onChange={(e) => updateRow(i, "start_time", e.target.value)}
                  />
                  <span className="text-muted-foreground">to</span>
                  <Input
                    type="time"
                    className="w-28"
                    value={hours[i]?.end_time || ""}
                    onChange={(e) => updateRow(i, "end_time", e.target.value)}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>

        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave} isDisabled={saving}>
            {saving ? "Saving…" : "Save policy"}
          </Button>
        </div>
      </div>
    </div>
  )
}
