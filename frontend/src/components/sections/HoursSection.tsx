import { useI18n } from "../../lib/i18n"
import { useSalonContext } from "../../context/SalonContext"
import { useBusinessWorkingHours, useBusinessClosures } from "../../hooks/useApi"
import { isSalonOpenNow, localHoursForDay } from "../../lib/salonHours"
import { WEEKDAY_ORDER } from "../../lib/calendar"
import type { HoursContent } from "../../lib/sections"

// Indexed by day_of_week (0 = Sunday .. 6 = Saturday).
const DAY_KEYS = [
  "days.sunday",
  "days.monday",
  "days.tuesday",
  "days.wednesday",
  "days.thursday",
  "days.friday",
  "days.saturday",
]

export function HoursSection({
  content,
  now,
}: {
  content: HoursContent
  now?: Date
}) {
  const { t } = useI18n()
  const { slug } = useSalonContext()
  const { data: hours = [], isLoading: hoursLoading } = useBusinessWorkingHours(slug)
  const { data: closures = [], isLoading: closuresLoading } = useBusinessClosures(slug)

  if (hoursLoading || closuresLoading) return null

  const current = now ?? new Date()
  const today = current.getDay()
  const openNow = isSalonOpenNow(hours, closures, current)

  return (
    <section className="space-y-4">
      <h3 className="text-xl font-semibold text-foreground">
        {content.heading || t("sections.hours")}
      </h3>
      <ul className="space-y-1.5">
        {WEEKDAY_ORDER.map((dow) => {
          const key = DAY_KEYS[dow]
          const range = localHoursForDay(hours, dow)
          const isToday = dow === today
          return (
            <li key={dow} className="flex items-center gap-1.5">
              <span
                className={`w-24 shrink-0 text-sm ${
                  isToday ? "font-semibold text-foreground" : "text-muted-foreground"
                }`}
              >
                {t(key)}
              </span>
              <span
                className={`text-sm ${
                  isToday ? "font-semibold text-foreground" : "text-muted-foreground"
                }`}
              >
                {range ? `${range.start} – ${range.end}` : t("sections.hours.closed")}
              </span>
              {isToday && (
                <span
                  className={`ml-0.5 size-2.5 shrink-0 rounded-full ${
                    openNow ? "bg-green-500" : "bg-red-500"
                  }`}
                  role="img"
                  aria-label={openNow ? t("sections.hours.openNow") : t("sections.hours.closedNow")}
                  title={openNow ? t("sections.hours.openNow") : t("sections.hours.closedNow")}
                />
              )}
            </li>
          )
        })}
      </ul>
    </section>
  )
}
