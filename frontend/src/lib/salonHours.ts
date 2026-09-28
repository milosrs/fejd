// Salon opening-hours helpers for the "hours" landing-page section.
//
// Working hours are stored server-side as plain wall-clock "HH:MM" strings
// (anchored to UTC by the backend), so the browser converts them to its local
// timezone for display and for the live open/closed check — the same convention
// used by the salon policy page.

export interface BusinessHours {
  day_of_week: number
  start_time: string
  end_time: string
}

export interface BusinessClosure {
  type: string
  day_of_week?: number
  month?: number
  day?: number
  start_date?: string
  end_date?: string
}

// utcWallClockToLocal converts a UTC "HH:MM" wall-clock string (as stored by the
// server) into the browser's local "HH:MM".
function utcWallClockToLocal(value: string): string {
  const match = /^(\d{1,2}):(\d{2})/.exec(value)
  if (!match) return value
  const now = new Date()
  const d = new Date(
    Date.UTC(now.getFullYear(), now.getMonth(), now.getDate(), Number(match[1]), Number(match[2])),
  )
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`
}

// hhmmToMinutes parses an HH:MM time value into total minutes.
function hhmmToMinutes(value: string): number {
  const match = /^(\d{1,2}):(\d{2})$/.exec(value.trim())
  if (!match) return Number.NaN
  const minutes = Number(match[2])
  if (minutes > 59) return Number.NaN
  return Number(match[1]) * 60 + minutes
}

// toISODate renders a Date as a local "YYYY-MM-DD" string.
function toISODate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  return `${y}-${m}-${day}`
}

// localHoursForDay returns the local "HH:MM" open/close times for a weekday, or
// null when the salon is closed that day.
export function localHoursForDay(
  hours: BusinessHours[],
  dayOfWeek: number,
): { start: string; end: string } | null {
  const wh = hours.find((h) => h.day_of_week === dayOfWeek)
  if (!wh) return null
  const start = utcWallClockToLocal(wh.start_time)
  const end = utcWallClockToLocal(wh.end_time)
  if (Number.isNaN(hhmmToMinutes(start)) || Number.isNaN(hhmmToMinutes(end))) return null
  return { start, end }
}

// isSalonOpenNow reports whether the salon is open at the given instant, taking
// working hours and non-working-day closures into account. day_of_week uses
// JavaScript's convention (0 = Sunday .. 6 = Saturday).
export function isSalonOpenNow(
  hours: BusinessHours[],
  closures: BusinessClosure[],
  now: Date = new Date(),
): boolean {
  const dow = now.getDay()
  const nowMinutes = now.getHours() * 60 + now.getMinutes()
  const todayStr = toISODate(now)

  for (const c of closures) {
    switch (c.type) {
      case "weekly":
        if (c.day_of_week === dow) return false
        break
      case "single":
        if (c.start_date === todayStr) return false
        break
      case "range":
        if (c.start_date && c.end_date && todayStr >= c.start_date && todayStr <= c.end_date) {
          return false
        }
        break
      case "yearly":
        if (c.month === now.getMonth() + 1 && c.day === now.getDate()) return false
        break
    }
  }

  const local = localHoursForDay(hours, dow)
  if (!local) return false
  const start = hhmmToMinutes(local.start)
  const end = hhmmToMinutes(local.end)
  if (Number.isNaN(start) || Number.isNaN(end)) return false
  return start <= nowMinutes && nowMinutes < end
}
