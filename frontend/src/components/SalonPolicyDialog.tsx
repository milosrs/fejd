import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { updateSalonPolicy } from "../hooks/useApi"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

export function SalonPolicyDialog({
  businessId,
  slug,
  cancellationLeadHours,
  noShowAfterHours,
  onClose,
}: {
  businessId: string
  slug: string
  cancellationLeadHours: number
  noShowAfterHours: number
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const [leadHours, setLeadHours] = useState(String(cancellationLeadHours))
  const [noShowHours, setNoShowHours] = useState(String(noShowAfterHours))
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  const handleSave = async () => {
    const lead = parseInt(leadHours, 10)
    const noShow = parseInt(noShowHours, 10)
    if (Number.isNaN(lead) || lead < 0 || Number.isNaN(noShow) || noShow < 0) {
      setError("Enter a number of zero or greater.")
      return
    }
    setSaving(true)
    setError("")
    try {
      await updateSalonPolicy(businessId, lead, noShow)
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      onClose()
    } catch {
      setError("Failed to save policy.")
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
        <h3 className="text-lg font-semibold text-foreground">Salon policy</h3>
        <p className="mt-2 text-sm text-muted-foreground">
          Cancellation and no-show rules applied to customer bookings.
        </p>
        <div className="mt-4 space-y-3">
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
