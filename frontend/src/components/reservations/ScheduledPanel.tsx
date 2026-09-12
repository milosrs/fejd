import type { ReactNode } from "react"
import { CalendarDate, getLocalTimeZone } from "@internationalized/date"
import { format } from "date-fns"
import {
  CalendarDaysIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ClockIcon,
  VideoIcon,
} from "lucide-react"

import { cn } from "#lib/utils"
import { formatDuration } from "#lib/calendar"
import { useI18n } from "../../lib/i18n"
import { Button } from "#components/ui/button"
import { TONE_BAR, type ScheduleEvent } from "./types"

const AVATAR_BG = [
  "bg-rose-500",
  "bg-blue-500",
  "bg-emerald-500",
  "bg-amber-500",
  "bg-violet-500",
  "bg-cyan-500",
  "bg-fuchsia-500",
  "bg-lime-500",
]

function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return "?"
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

function avatarBg(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = (hash * 31 + name.charCodeAt(i)) >>> 0
  }
  return AVATAR_BG[hash % AVATAR_BG.length]
}

export interface ScheduledPanelProps {
  date: CalendarDate
  onPrevDay: () => void
  onNextDay: () => void
  onOpenCalendar?: () => void
  events: ScheduleEvent[]
  isLoading?: boolean
  renderActions?: (event: ScheduleEvent) => ReactNode
  className?: string
}

function EventCard({
  event,
  renderActions,
}: {
  event: ScheduleEvent
  renderActions?: (event: ScheduleEvent) => ReactNode
}) {
  const { t } = useI18n()
  const names = event.attendees ?? []
  const showNameLabel = names.length > 0 && names.length <= 2 && !event.extraAttendees

  return (
    <div className="relative rounded-2xl border border-border/60 bg-card p-3 pl-4">
      <span
        className={cn(
          "absolute inset-y-2 left-0 w-1 rounded-full",
          TONE_BAR[event.tone ?? "blue"],
        )}
      />

      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate font-medium text-foreground">{event.title}</p>
          {event.subtitle && (
            <p className="mt-0.5 truncate text-xs text-muted-foreground">
              {event.subtitle}
            </p>
          )}
        </div>
        <span className="shrink-0 text-xs text-muted-foreground">
          {formatDuration(event.start, event.end)}
        </span>
      </div>

      <div className="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
        <ClockIcon className="size-3.5" />
        <span className="tabular-nums">
          {format(event.start, "HH:mm")} – {format(event.end, "HH:mm")}
        </span>
      </div>

      {(names.length > 0 || !!event.extraAttendees) && (
        <div className="mt-3 flex items-center gap-2">
          {names.length > 0 && (
            <div className="flex -space-x-2">
              {names.map((name) => (
                <span
                  key={name}
                  title={name}
                  className={cn(
                    "flex size-6 items-center justify-center rounded-full border-2 border-card text-[9px] font-medium text-white",
                    avatarBg(name),
                  )}
                >
                  {initials(name)}
                </span>
              ))}
            </div>
          )}
          {showNameLabel ? (
            <span className="truncate text-xs text-muted-foreground">
              {names.join(" and ")}
            </span>
          ) : (
            !!event.extraAttendees && (
              <span className="text-xs text-muted-foreground">
                {t("common.more", { count: event.extraAttendees })}
              </span>
            )
          )}
        </div>
      )}

      {event.meetLink && (
        <a
          href={event.meetLink}
          className="mt-3 inline-flex items-center gap-1.5 rounded-full bg-emerald-500/15 px-3 py-1 text-xs font-medium text-emerald-600 transition-colors hover:bg-emerald-500/25 dark:text-emerald-400"
        >
          <VideoIcon className="size-3.5" />
          {t("scheduled.meetLink")}
        </a>
      )}

      {renderActions && (
        <div className="mt-3 flex flex-wrap items-center gap-2 border-t border-border/60 pt-3">
          {renderActions(event)}
        </div>
      )}
    </div>
  )
}

export function ScheduledPanel({
  date,
  onPrevDay,
  onNextDay,
  onOpenCalendar,
  events,
  isLoading,
  renderActions,
  className,
}: ScheduledPanelProps) {
  const sorted = [...events].sort(
    (a, b) => a.start.getTime() - b.start.getTime(),
  )
  const { t } = useI18n()

  const groups = new Map<number, ScheduleEvent[]>()
  for (const event of sorted) {
    const hour = event.start.getHours()
    const list = groups.get(hour) ?? []
    list.push(event)
    groups.set(hour, list)
  }
  const hours = [...groups.keys()].sort((a, b) => a - b)

  return (
    <div className={cn("flex flex-col gap-4", className)}>
      <header className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-lg font-semibold text-foreground">{t("scheduled.title")}</h2>
          <p className="text-sm text-muted-foreground">
            {format(date.toDate(getLocalTimeZone()), "d MMMM, yyyy")}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          {onOpenCalendar && (
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={t("scheduled.today")}
              onPress={onOpenCalendar}
            >
              <CalendarDaysIcon className="size-4" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("scheduled.prevDay")}
            onPress={onPrevDay}
          >
            <ChevronLeftIcon className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("scheduled.nextDay")}
            onPress={onNextDay}
          >
            <ChevronRightIcon className="size-4" />
          </Button>
        </div>
      </header>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">{t("common.loading")}</p>
      ) : hours.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("scheduled.noEvents")}</p>
      ) : (
        <div className="space-y-5">
          {hours.map((hour) => (
            <div key={hour} className="flex gap-3">
              <span className="w-11 shrink-0 pt-1 text-right text-xs font-medium tabular-nums text-muted-foreground">
                {String(hour).padStart(2, "0")}:00
              </span>
              <div className="flex min-w-0 flex-1 flex-col gap-3">
                {groups.get(hour)!.map((event) => (
                  <EventCard
                    key={event.id}
                    event={event}
                    renderActions={renderActions}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
