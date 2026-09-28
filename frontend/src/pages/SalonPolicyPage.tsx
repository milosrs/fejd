import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { parseDate, getLocalTimeZone, today, Time } from "@internationalized/date"
import { ArrowLeft } from "lucide-react"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useSalonPolicy, updateSalonPolicy } from "../hooks/useApi"
import { useCanWrite } from "../hooks/useCanWrite"
import { useI18n } from "../lib/i18n"
import { salonPath } from "../lib/salonDomain"
import { WEEKDAY_ORDER } from "../lib/calendar"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "../components/ui/card"
import { Calendar, RangeCalendar } from "../components/ui/calendar"
import { TimeField, TimeFieldInput, TimeFieldSegment } from "../components/ui/time-field"

interface DaySchedule {
  day_of_week: number
  start_time: string
  end_time: string
  non_working: boolean
}

type ClosureType = "single" | "range" | "weekly" | "yearly"

interface ClosureRule {
  key: string
  type: ClosureType
  start_date?: string
  end_date?: string
  day_of_week?: number
  month?: number
  day?: number
  reason?: string
}

const defaultSchedule = (): DaySchedule[] =>
  Array.from({ length: 7 }, (_, i) => ({
    day_of_week: i,
    start_time: "09:00",
    end_time: "17:00",
    non_working: false,
  }))

const dateInputClass =
  "h-8 rounded-2xl border border-border bg-input/50 px-2.5 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/30"

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

function newClosureKey(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

function formatMonthDay(month: number, day: number): string {
  const d = new Date(2000, month - 1, day)
  return new Intl.DateTimeFormat(undefined, { month: "long", day: "numeric" }).format(d)
}

function describeClosure(c: ClosureRule, days: string[], t: (key: string) => string): string {
  switch (c.type) {
    case "single":
      return c.start_date ?? ""
    case "range":
      return `${c.start_date ?? ""} – ${c.end_date ?? ""}`
    case "weekly":
      return days[c.day_of_week ?? 0]
    case "yearly":
      return `${formatMonthDay(c.month ?? 1, c.day ?? 1)} (${t("policy.nonWorkingDayEveryYear")})`
    default:
      return ""
  }
}

function isValidDate(value: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(value) && !Number.isNaN(Date.parse(value))
}

export function SalonPolicyPage() {
  const { slug, salon } = useSalonContext()
  const isOwner = useIsOwner()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const canWrite = useCanWrite()
  const { t } = useI18n()

  const businessId = salon?.business.id ?? ""
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

  const [schedule, setSchedule] = useState<DaySchedule[]>(defaultSchedule())
  const [leadHours, setLeadHours] = useState("")
  const [noShowTime, setNoShowTime] = useState<Time | null>(null)
  const [slotInterval, setSlotInterval] = useState("30")
  const [autoApprove, setAutoApprove] = useState(false)
  const [dateClosures, setDateClosures] = useState<ClosureRule[]>([])
  const [closureMode, setClosureMode] = useState<"single" | "range">("single")
  const [singleDate, setSingleDate] = useState("")
  const [repeatYearly, setRepeatYearly] = useState(false)
  const [singleReason, setSingleReason] = useState("")
  const [rangeStart, setRangeStart] = useState("")
  const [rangeEnd, setRangeEnd] = useState("")
  const [rangeReason, setRangeReason] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!policy) return
    setLeadHours(String(policy.cancellation_lead_hours))
    setNoShowTime(
      new Time(
        Math.floor(policy.no_show_after_minutes / 60),
        policy.no_show_after_minutes % 60,
      ),
    )
    setSlotInterval(String(policy.slot_interval_minutes))
    setAutoApprove(policy.auto_approve)

    const next = defaultSchedule()
    for (const wh of policy.working_hours ?? []) {
      if (wh.day_of_week < 0 || wh.day_of_week > 6) continue
      next[wh.day_of_week].start_time = utcWallClockToLocal(wh.start_time)
      next[wh.day_of_week].end_time = utcWallClockToLocal(wh.end_time)
    }
    for (const c of policy.closures ?? []) {
      if (c.type === "weekly" && c.day_of_week != null && c.day_of_week >= 0 && c.day_of_week <= 6) {
        next[c.day_of_week].non_working = true
      }
    }
    setSchedule(next)

    setDateClosures(
      (policy.closures ?? [])
        .filter((c) => c.type !== "weekly")
        .map((c) => ({
          key: c.id,
          type: c.type as ClosureType,
          start_date: c.start_date,
          end_date: c.end_date,
          day_of_week: c.day_of_week,
          month: c.month,
          day: c.day,
          reason: c.reason ?? "",
        })),
    )
  }, [policy])

  if (!salon) {
    return null
  }

  if (!isOwner) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={() => navigate(salonPath(slug))}>
            <ArrowLeft className="size-4" /> {t("common.back")}
          </Button>
        </div>
        <p className="text-muted-foreground">{t("policy.notAuthorized")}</p>
      </div>
    )
  }

  const updateSchedule = (index: number, patch: Partial<DaySchedule>) => {
    setSchedule((prev) => {
      const next = [...prev]
      next[index] = { ...next[index], ...patch }
      return next
    })
  }

  const addSingleClosure = () => {
    const date = singleDate.trim()
    if (!isValidDate(date)) {
      setError(t("policy.nonWorkingDayInvalid"))
      return
    }
    setError("")
    if (repeatYearly) {
      const [year, month, day] = date.split("-").map(Number)
      setDateClosures((prev) => [
        ...prev,
        { key: newClosureKey(), type: "yearly", month, day, reason: singleReason.trim() || undefined },
      ])
    } else {
      setDateClosures((prev) => [
        ...prev,
        { key: newClosureKey(), type: "single", start_date: date, reason: singleReason.trim() || undefined },
      ])
    }
    setSingleDate("")
    setSingleReason("")
  }

  const addRangeClosure = () => {
    const start = rangeStart.trim()
    const end = rangeEnd.trim()
    if (!isValidDate(start) || !isValidDate(end)) {
      setError(t("policy.nonWorkingDayInvalid"))
      return
    }
    if (end < start) {
      setError(t("policy.nonWorkingDayInvalidRange"))
      return
    }
    setError("")
    setDateClosures((prev) => [
      ...prev,
      { key: newClosureKey(), type: "range", start_date: start, end_date: end, reason: rangeReason.trim() || undefined },
    ])
    setRangeStart("")
    setRangeEnd("")
    setRangeReason("")
  }

  const removeClosure = (key: string) => {
    setDateClosures((prev) => prev.filter((c) => c.key !== key))
  }

  const handleSave = async () => {
    const lead = parseInt(leadHours, 10)
    const interval = parseInt(slotInterval, 10)
    if (Number.isNaN(lead) || lead < 0) {
      setError(t("policy.cancellationInvalid"))
      return
    }
    if (Number.isNaN(interval) || interval <= 0) {
      setError(t("policy.intervalInvalid"))
      return
    }
    const noShow = noShowTime ? noShowTime.hour * 60 + noShowTime.minute : 0
    setSaving(true)
    setError("")
    try {
      await updateSalonPolicy(businessId, {
        cancellation_lead_hours: lead,
        no_show_after_minutes: noShow,
        slot_interval_minutes: interval,
        auto_approve: autoApprove,
        timezone: getLocalTimeZone(),
        working_hours: schedule
          .filter((d) => !d.non_working)
          .map((d) => ({
            day_of_week: d.day_of_week,
            start_time: d.start_time,
            end_time: d.end_time,
          })),
        closures: [
          ...schedule
            .filter((d) => d.non_working)
            .map((d) => ({ type: "weekly" as const, day_of_week: d.day_of_week })),
          ...dateClosures.map(({ key: _key, ...c }) => ({
            type: c.type,
            start_date: c.start_date,
            end_date: c.end_date,
            day_of_week: c.day_of_week,
            month: c.month,
            day: c.day,
            reason: c.reason || undefined,
          })),
        ],
      })
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      await queryClient.invalidateQueries({ queryKey: ["salon-policy", businessId] })
      await queryClient.invalidateQueries({ queryKey: ["slots"] })
      navigate(salonPath(slug))
    } catch {
      setError(t("policy.failed"))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-2">
        <Button variant="ghost" size="sm" onClick={() => navigate(salonPath(slug))}>
          <ArrowLeft className="size-4" /> {t("common.back")}
        </Button>
        <h1 className="text-lg font-semibold text-foreground">{t("policy.title")}</h1>
        <Button size="sm" onClick={handleSave} isDisabled={saving || !canWrite}>
          {saving ? t("common.saving") : t("common.save")}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("policy.workingHours")}</CardTitle>
          <CardDescription>{t("policy.workingHoursHelp")}</CardDescription>
        </CardHeader>
        <CardContent className="divide-y divide-border">
          {WEEKDAY_ORDER.map((dow) => (
            <div key={dow} className="space-y-2 py-3 first:pt-0 last:pb-0">
              <div className="text-sm font-medium text-foreground">{days[dow]}</div>
              <div className="flex items-center gap-2">
                <Input
                  type="time"
                  className="w-28"
                  value={schedule[dow]?.start_time || ""}
                  disabled={schedule[dow]?.non_working}
                  onChange={(e) => updateSchedule(dow, { start_time: e.target.value })}
                />
                <span className="text-sm text-muted-foreground">{t("admin.schedule.to")}</span>
              </div>
              <div className="flex items-center gap-2">
                <Input
                  type="time"
                  className="w-28"
                  value={schedule[dow]?.end_time || ""}
                  disabled={schedule[dow]?.non_working}
                  onChange={(e) => updateSchedule(dow, { end_time: e.target.value })}
                />
                <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                  <input
                    type="checkbox"
                    checked={schedule[dow]?.non_working || false}
                    onChange={(e) => updateSchedule(dow, { non_working: e.target.checked })}
                  />
                  {t("policy.workingHoursNonWorking")}
                </label>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("policy.appointments")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
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
          <TimeField
            value={noShowTime}
            onChange={setNoShowTime}
            granularity="minute"
            hourCycle={24}
            shouldForceLeadingZeros
            className="space-y-1"
          >
            <Label>{t("policy.noShowGrace")}</Label>
            <TimeFieldInput>
              {(segment) => <TimeFieldSegment segment={segment} />}
            </TimeFieldInput>
          </TimeField>
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
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("policy.nonWorkingDays")}</CardTitle>
          <CardDescription>{t("policy.nonWorkingDaysPolicyHelp")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {dateClosures.length === 0 && (
            <p className="text-sm text-muted-foreground">{t("policy.nonWorkingDaysEmpty")}</p>
          )}
          {dateClosures.map((c) => (
            <div key={c.key} className="flex items-center justify-between gap-2 rounded-xl border border-border p-2">
              <div className="min-w-0">
                <span className="font-medium">{describeClosure(c, days, t)}</span>
                {c.reason && <span className="text-muted-foreground ml-2">({c.reason})</span>}
              </div>
              <Button variant="outline" size="sm" onClick={() => removeClosure(c.key)} isDisabled={!canWrite}>
                {t("common.remove")}
              </Button>
            </div>
          ))}

          <div className="flex gap-1 rounded-xl bg-muted p-1">
            <button
              type="button"
              onClick={() => setClosureMode("single")}
              className={`flex-1 rounded-lg px-2 py-1 text-sm font-medium transition-colors ${
                closureMode === "single" ? "bg-background text-foreground shadow-sm" : "text-muted-foreground"
              }`}
            >
              {t("policy.nonWorkingDaysSingle")}
            </button>
            <button
              type="button"
              onClick={() => setClosureMode("range")}
              className={`flex-1 rounded-lg px-2 py-1 text-sm font-medium transition-colors ${
                closureMode === "range" ? "bg-background text-foreground shadow-sm" : "text-muted-foreground"
              }`}
            >
              {t("policy.nonWorkingDaysMultiple")}
            </button>
          </div>

          {closureMode === "single" ? (
            <div className="space-y-2">
              <Calendar
                value={singleDate ? parseDate(singleDate) : undefined}
                onChange={(date) => setSingleDate(date ? date.toString() : "")}
                minValue={today(getLocalTimeZone())}
                className="mx-auto"
              />
              <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <input
                  type="checkbox"
                  checked={repeatYearly}
                  onChange={(e) => setRepeatYearly(e.target.checked)}
                />
                {t("policy.nonWorkingDayRepeatYearly")}
              </label>
              <div className="flex items-center gap-2">
                <Input
                  className="flex-1"
                  placeholder={t("policy.nonWorkingDayReason")}
                  value={singleReason}
                  onChange={(e) => setSingleReason(e.target.value)}
                />
                <Button variant="outline" size="sm" onClick={addSingleClosure} isDisabled={!canWrite}>
                  {t("policy.addNonWorkingDay")}
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-2">
              <RangeCalendar
                value={
                  rangeStart && rangeEnd
                    ? { start: parseDate(rangeStart), end: parseDate(rangeEnd) }
                    : undefined
                }
                onChange={(range) => {
                  setRangeStart(range?.start?.toString() ?? "")
                  setRangeEnd(range?.end?.toString() ?? "")
                }}
                minValue={today(getLocalTimeZone())}
                className="mx-auto"
              />
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm text-muted-foreground">{t("policy.nonWorkingDayRangeStart")}</span>
                <input
                  type="date"
                  value={rangeStart}
                  onChange={(e) => setRangeStart(e.target.value)}
                  className={`${dateInputClass} w-36`}
                />
                <span className="text-sm text-muted-foreground">{t("policy.nonWorkingDayRangeEnd")}</span>
                <input
                  type="date"
                  value={rangeEnd}
                  onChange={(e) => setRangeEnd(e.target.value)}
                  className={`${dateInputClass} w-36`}
                />
              </div>
              <div className="flex items-center gap-2">
                <Input
                  className="flex-1"
                  placeholder={t("policy.nonWorkingDayReason")}
                  value={rangeReason}
                  onChange={(e) => setRangeReason(e.target.value)}
                />
                <Button variant="outline" size="sm" onClick={addRangeClosure} isDisabled={!canWrite}>
                  {t("policy.addNonWorkingRange")}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {error && <p className="text-sm text-red-500">{error}</p>}

      <div className="flex justify-end">
        <Button onClick={handleSave} isDisabled={saving || !canWrite}>
          {saving ? t("common.saving") : t("common.save")}
        </Button>
      </div>
    </div>
  )
}
