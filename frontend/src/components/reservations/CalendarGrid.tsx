import { CalendarDate, type DateValue } from "@internationalized/date"
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"

import { cn } from "#lib/utils"
import { buildMonthGrid } from "#lib/calendar"
import { useI18n } from "../../lib/i18n"
import { Button } from "#components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "#components/ui/select"
import { TONE_DOT, type ReservationTag } from "./types"

const WEEKDAY_KEYS = [
  "days.sunday",
  "days.monday",
  "days.tuesday",
  "days.wednesday",
  "days.thursday",
  "days.friday",
  "days.saturday",
]

const MONTH_KEYS = [
  "months.january",
  "months.february",
  "months.march",
  "months.april",
  "months.may",
  "months.june",
  "months.july",
  "months.august",
  "months.september",
  "months.october",
  "months.november",
  "months.december",
]

export interface CalendarGridProps {
  month: CalendarDate
  onMonthChange: (month: CalendarDate) => void
  selected: CalendarDate | null
  onSelect: (date: CalendarDate) => void
  today?: CalendarDate
  eventsByDay?: Record<string, ReservationTag[]>
  minValue?: DateValue
  maxValue?: DateValue
  className?: string
}

function yearRange(anchor: number): number[] {
  const years: number[] = []
  for (let y = anchor - 10; y <= anchor + 10; y++) {
    years.push(y)
  }
  return years
}

export function CalendarGrid({
  month,
  onMonthChange,
  selected,
  onSelect,
  today,
  eventsByDay,
  minValue,
  maxValue,
  className,
}: CalendarGridProps) {
  const { t } = useI18n()
  const weekdays = WEEKDAY_KEYS.map((k) => t(k))
  const months = MONTH_KEYS.map((k) => t(k))
  const cells = buildMonthGrid(month.year, month.month)

  return (
    <div className={cn("flex flex-col gap-3", className)}>
      <header className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1">
          <Select
            aria-label={t("calendar.month")}
            className="w-fit"
            selectedKey={String(month.month)}
            onSelectionChange={(key) =>
              onMonthChange(new CalendarDate(month.year, Number(key), 1))
            }
          >
            <SelectTrigger className="gap-1 px-2 font-medium">
              <SelectValue>
                {(state) => state.selectedText || (months[month.month - 1] ?? "")}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {months.map((name, i) => (
                <SelectItem key={i + 1} id={String(i + 1)}>
                  {name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select
            aria-label={t("calendar.year")}
            className="w-fit"
            selectedKey={String(month.year)}
            onSelectionChange={(key) =>
              onMonthChange(new CalendarDate(Number(key), month.month, 1))
            }
          >
            <SelectTrigger className="gap-1 px-2 font-medium">
              <SelectValue>
                {(state) => state.selectedText || String(month.year)}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {yearRange(month.year).map((y) => (
                <SelectItem key={y} id={String(y)}>
                  {String(y)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("calendar.prevMonth")}
            onPress={() => onMonthChange(month.subtract({ months: 1 }))}
          >
            <ChevronLeftIcon className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("calendar.nextMonth")}
            onPress={() => onMonthChange(month.add({ months: 1 }))}
          >
            <ChevronRightIcon className="size-4" />
          </Button>
        </div>
      </header>

      <div className="grid grid-cols-7 gap-1.5">
        {weekdays.map((day) => (
          <div
            key={day}
            className="px-1 text-center text-[11px] font-medium uppercase tracking-wide text-muted-foreground"
          >
            {day}
          </div>
        ))}
      </div>

      <div className="grid auto-rows-fr grid-cols-7 gap-1.5">
        {cells.map((cell) => {
          const key = cell.date.toString()
          const tags = eventsByDay?.[key] ?? []
          const isToday = !!today && cell.date.compare(today) === 0
          const isSelected = !!selected && cell.date.compare(selected) === 0
          const isDisabled =
            (!!minValue && cell.date.compare(minValue) < 0) ||
            (!!maxValue && cell.date.compare(maxValue) > 0)

          return (
            <button
              key={key}
              type="button"
              disabled={isDisabled}
              aria-label={key}
              aria-pressed={isSelected}
              onClick={() => onSelect(cell.date)}
              className={cn(
                "group relative flex min-h-24 flex-col items-start gap-1.5 rounded-xl p-2 text-left transition",
                !cell.inMonth && "opacity-40",
                isDisabled && "opacity-40",
                isToday
                  ? "bg-blue-600 text-white"
                  : "bg-muted hover:bg-accent",
                isSelected &&
                  (isToday ? "ring-2 ring-blue-400" : "ring-2 ring-blue-500/70"),
                !isDisabled && !isToday && "cursor-pointer",
              )}
            >
              <span
                className={cn(
                  "text-sm font-semibold leading-none",
                  isToday ? "text-white" : "text-foreground",
                )}
              >
                {cell.date.day}
              </span>

              <div className="flex w-full flex-col gap-0.5">
                {tags.slice(0, 2).map((tag) => (
                  <CalendarTagRow key={tag.id} tag={tag} onToday={isToday} />
                ))}
                {tags.length > 2 && (
                  <span
                    className={cn(
                      "text-[11px] leading-tight",
                      isToday ? "text-blue-100" : "text-muted-foreground",
                    )}
                  >
                    {t("common.more", { count: tags.length - 2 })}
                  </span>
                )}
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}

function CalendarTagRow({
  tag,
  onToday,
}: {
  tag: ReservationTag
  onToday: boolean
}) {
  return (
    <span className="flex min-w-0 items-center gap-1">
      <span
        className={cn(
          "size-1.5 shrink-0 rounded-full",
          TONE_DOT[tag.tone ?? "blue"],
        )}
      />
      <span
        className={cn(
          "truncate text-[11px] leading-tight",
          onToday ? "text-blue-100" : "text-muted-foreground",
        )}
      >
        {tag.label}
      </span>
    </span>
  )
}
