import { CalendarDate, getDayOfWeek } from "@internationalized/date"

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

// buildMonthGrid returns a Sunday-first month grid padded with overflow days
// from the neighbouring months so the grid always covers whole weeks.
export function buildMonthGrid(year: number, month: number): MonthCell[] {
  const first = new CalendarDate(year, month, 1)
  const count = daysInMonth(year, month)
  const cells: MonthCell[] = []

  const firstWeekday = getDayOfWeek(first, "en-US")
  for (let i = firstWeekday; i > 0; i--) {
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
