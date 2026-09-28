import { useEffect, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useSalonPolicy, updateSalonPolicy } from "../hooks/useApi"
import { useI18n } from "../lib/i18n"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

interface HourRow {
  day_of_week: number
  start_time: string
  end_time: string
}

interface ClosureRow {
  closure_date: string
  reason: string
}

const defaultHours = (): HourRow[] =>
  Array.from({ length: 7 }, (_, i) => ({ day_of_week: i, start_time: "09:00", end_time: "17:00" }))

// getLocalTimeZone returns the browser's IANA timezone (e.g. "Europe/Stockholm").
function getLocalTimeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC"
  } catch {
    return "UTC"
  }
}

// utcWallClockToLocal converts a UTC "HH:MM" wall-clock string (as stored by the
// server) into the browser's local "HH:MM" for display in a time input.
function utcWallClockToLocal(value: string): string {
  const match = /^(\d{1,2}):(\d{2})/.exec(value)
  if (!match) return value
  const today = new Date()
  const d = new Date(
    Date.UTC(today.getFullYear(), today.getMonth(), today.getDate(), Number(match[1]), Number(match[2])),
  )
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`
}

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
  const { t } = useI18n()
  const { data: policy } = useSalonPolicy(businessId)

  const days = [
    t("days.sunday"),
    t("days.monday"),
    t("days.tuesday"),
    t("days.wednesday"),
    t("days.thursday"),
    t("days.friday"),
    t("days.saturday"),
  ]

  const [leadHours, setLeadHours] = useState("")
  const [noShowHours, setNoShowHours] = useState("")
  const [slotInterval, setSlotInterval] = useState("30")
  const [autoApprove, setAutoApprove] = useState(false)
  const [hours, setHours] = useState<HourRow[]>(defaultHours())
  const [closures, setClosures] = useState<ClosureRow[]>([])
  const [newClosureDate, setNewClosureDate] = useState("")
  const [newClosureReason, setNewClosureReason] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!policy) return
    setLeadHours(String(policy.cancellation_lead_hours))
    setNoShowHours(String(policy.no_show_after_hours))
    setSlotInterval(String(policy.slot_interval_minutes))
    setAutoApprove(policy.auto_approve)
    if (policy.working_hours?.length) {
      setHours(
        policy.working_hours.map((wh) => ({
          day_of_week: wh.day_of_week,
          start_time: utcWallClockToLocal(wh.start_time),
          end_time: utcWallClockToLocal(wh.end_time),
        })),
      )
    }
    setClosures(
      (policy.closures ?? []).map((c) => ({
        closure_date: c.closure_date,
        reason: c.reason ?? "",
      })),
    )
  }, [policy])

  const updateRow = (index: number, field: keyof HourRow, value: string) => {
    setHours((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], [field]: value }
      return next
    })
  }

  const addClosure = () => {
    const date = newClosureDate.trim()
    if (!/^\d{4}-\d{2}-\d{2}$/.test(date) || Number.isNaN(Date.parse(date))) {
      setError(t("policy.nonWorkingDayInvalid"))
      return
    }
    setError("")
    setClosures((prev) => {
      if (prev.some((c) => c.closure_date === date)) return prev
      return [...prev, { closure_date: date, reason: newClosureReason.trim() }]
    })
    setNewClosureDate("")
    setNewClosureReason("")
  }

  const removeClosure = (date: string) => {
    setClosures((prev) => prev.filter((c) => c.closure_date !== date))
  }

  const handleSave = async () => {
    const lead = parseInt(leadHours, 10)
    const noShow = parseInt(noShowHours, 10)
    const interval = parseInt(slotInterval, 10)
    if (Number.isNaN(lead) || lead < 0) {
      setError(t("policy.cancellationInvalid"))
      return
    }
    if (Number.isNaN(noShow) || noShow < 0) {
      setError(t("policy.noShowInvalid"))
      return
    }
    if (Number.isNaN(interval) || interval <= 0) {
      setError(t("policy.intervalInvalid"))
      return
    }
    setSaving(true)
    setError("")
    try {
      await updateSalonPolicy(businessId, {
        cancellation_lead_hours: lead,
        no_show_after_hours: noShow,
        slot_interval_minutes: interval,
        auto_approve: autoApprove,
        timezone: getLocalTimeZone(),
        working_hours: hours,
        closures: closures.map((c) => ({
          closure_date: c.closure_date,
          reason: c.reason || undefined,
        })),
      })
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      await queryClient.invalidateQueries({ queryKey: ["salon-policy", businessId] })
      await queryClient.invalidateQueries({ queryKey: ["slots"] })
      onClose()
    } catch {
      setError(t("policy.failed"))
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
        <h3 className="text-lg font-semibold text-foreground">{t("policy.title")}</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("policy.help")}
        </p>

        <div className="mt-4 space-y-4">
          <div className="space-y-1">
            <Label htmlFor="cancellation-lead-hours">{t("policy.cancellationNotice")}</Label>
            <Input
              id="cancellation-lead-hours"
              type="number"
              min={0}
              value={leadHours}
              onChange={(e) => setLeadHours(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="no-show-after-hours">{t("policy.noShowGrace")}</Label>
            <Input
              id="no-show-after-hours"
              type="number"
              min={0}
              value={noShowHours}
              onChange={(e) => setNoShowHours(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="slot-interval">{t("policy.slotInterval")}</Label>
            <Input
              id="slot-interval"
              type="number"
              min={5}
              step={5}
              value={slotInterval}
              onChange={(e) => setSlotInterval(e.target.value)}
            />
          </div>

          <label className="flex items-start gap-3 rounded-xl border border-border p-3">
            <input
              type="checkbox"
              checked={autoApprove}
              onChange={(e) => setAutoApprove(e.target.checked)}
              className="mt-0.5 shrink-0"
            />
            <span className="min-w-0">
              <span className="block text-sm font-medium text-foreground">
                {t("policy.autoApprove")}
              </span>
              <span className="block text-sm text-muted-foreground">
                {t("policy.autoApproveHelp")}
              </span>
            </span>
          </label>

          <div className="space-y-2">
            <Label>{t("policy.workingHours")}</Label>
            <div className="space-y-2 rounded-xl border border-border p-3">
              {days.map((day, i) => (
                <div key={day} className="flex items-center gap-2">
                  <span className="w-24 text-sm text-muted-foreground">{day}</span>
                  <Input
                    type="time"
                    className="w-28"
                    value={hours[i]?.start_time || ""}
                    onChange={(e) => updateRow(i, "start_time", e.target.value)}
                  />
                  <span className="text-muted-foreground">{t("admin.schedule.to")}</span>
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

          <div className="space-y-2">
            <Label>{t("policy.nonWorkingDays")}</Label>
            <p className="text-sm text-muted-foreground">{t("policy.nonWorkingDaysHelp")}</p>
            <div className="space-y-2 rounded-xl border border-border p-3">
              {closures.length === 0 && (
                <p className="text-sm text-muted-foreground">{t("policy.nonWorkingDaysEmpty")}</p>
              )}
              {closures.map((c) => (
                <div key={c.closure_date} className="flex items-center justify-between gap-2">
                  <div className="min-w-0">
                    <span className="font-medium">{c.closure_date}</span>
                    {c.reason && (
                      <span className="text-muted-foreground ml-2">({c.reason})</span>
                    )}
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => removeClosure(c.closure_date)}
                  >
                    {t("common.remove")}
                  </Button>
                </div>
              ))}
              <div className="flex items-center gap-2">
                <input
                  type="date"
                  value={newClosureDate}
                  onChange={(e) => setNewClosureDate(e.target.value)}
                  className="h-8 w-36 rounded-2xl border border-border bg-input/50 px-2.5 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/30"
                />
                <Input
                  className="flex-1"
                  placeholder={t("policy.nonWorkingDayReason")}
                  value={newClosureReason}
                  onChange={(e) => setNewClosureReason(e.target.value)}
                />
                <Button variant="outline" size="sm" onClick={addClosure}>
                  {t("policy.addNonWorkingDay")}
                </Button>
              </div>
            </div>
          </div>
        </div>

        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSave} isDisabled={saving}>
            {saving ? t("common.saving") : t("policy.savePolicy")}
          </Button>
        </div>
      </div>
    </div>
  )
}
