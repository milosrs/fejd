import { CalendarDate, getDayOfWeek, parseDate } from "@internationalized/date"

export interface MonthCell {
  date: CalendarDate
  inMonth: boolean
}

export function daysInMonth(year: number, month: number): number {
  const next =
    month === 12
      ? new CalendarDate(year + 1, 1, 1)
      : new CalendarDate(year, month + 1, 1)
  return next.subtract({ days: 1 }).day
}

export function getMonthDays(year: number, month: number): CalendarDate[] {
  const count = daysInMonth(year, month)
  const days: CalendarDate[] = []
  for (let day = 1; day <= count; day++) {
    days.push(new CalendarDate(year, month, day))
  }
  return days
}

// WEEKDAY_ORDER lists day_of_week values in Monday-first display order. The
// day_of_week value itself keeps the app-wide convention 0 = Sunday .. 6 =
// Saturday (matching JS Date#getDay and the DB column).
export const WEEKDAY_ORDER: number[] = [1, 2, 3, 4, 5, 6, 0]

// buildMonthGrid returns a Monday-first month grid padded with overflow days
// from the neighbouring months so the grid always covers whole weeks.
export function buildMonthGrid(year: number, month: number): MonthCell[] {
  const first = new CalendarDate(year, month, 1)
  const count = daysInMonth(year, month)
  const cells: MonthCell[] = []

  const firstWeekday = getDayOfWeek(first, "en-US")
  const mondayOffset = (firstWeekday + 6) % 7
  for (let i = mondayOffset; i > 0; i--) {
    cells.push({ date: first.subtract({ days: i }), inMonth: false })
  }

  for (let day = 1; day <= count; day++) {
    cells.push({ date: new CalendarDate(year, month, day), inMonth: true })
  }

  const last = new CalendarDate(year, month, count)
  const remaining = (7 - (cells.length % 7)) % 7
  for (let i = 1; i <= remaining; i++) {
    cells.push({ date: last.add({ days: i }), inMonth: false })
  }

  return cells
}

export function formatDuration(start: Date, end: Date): string {
  const total = Math.max(0, Math.round((end.getTime() - start.getTime()) / 60000))
  const hours = Math.floor(total / 60)
  const minutes = total % 60
  if (hours === 0) return `${minutes} min`
  if (minutes === 0) return hours === 1 ? "1 hour" : `${hours} hours`
  return `${hours}h ${minutes}m`
}

// BusinessClosure is the subset of a salon closure needed to derive the
// concrete calendar dates the salon is closed on. Single-day and range closures
// map to specific dates; weekly/yearly recurrences are not expanded here.
export interface BusinessClosure {
  type: string
  start_date?: string
  end_date?: string
}

// closureDateKeys returns the set of "YYYY-MM-DD" keys the salon is closed on,
// expanding single-day and inclusive date-range closures.
export function closureDateKeys(closures: readonly BusinessClosure[] | undefined): Set<string> {
  const keys = new Set<string>()
  for (const c of closures ?? []) {
    if (c.type === "single" && c.start_date) {
      keys.add(c.start_date)
    } else if (c.type === "range" && c.start_date && c.end_date) {
      const start = parseDate(c.start_date)
      const end = parseDate(c.end_date)
      for (let d = start; d.compare(end) <= 0; d = d.add({ days: 1 })) {
        keys.add(d.toString())
      }
    }
  }
  return keys
}
