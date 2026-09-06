import { format } from "date-fns"
import { UserRound } from "lucide-react"
import { useAvailableSlots, type Employee, type TimeSlot } from "../../hooks/useApi"
import { useI18n } from "../../lib/i18n"
import { resolveImageUrl } from "../../lib/images"

export function BarberAvailabilityCard({
  barber,
  slug,
  serviceId,
  date,
  selectedStartTime,
  onSelectSlot,
}: {
  barber: Employee
  slug: string
  serviceId: string
  date: string
  selectedStartTime?: string
  onSelectSlot: (employeeId: string, slot: TimeSlot) => void
}) {
  const { t } = useI18n()
  const { data } = useAvailableSlots(slug, serviceId, barber.id, date)
  const slots = data?.slots ?? []
  const avatar = resolveImageUrl(barber.avatar)
  const name = barber.display_name || barber.user_id

  return (
    <div className="rounded-xl border border-border p-4">
      <div className="mb-3 flex items-center gap-3">
        {avatar ? (
          <img src={avatar} alt={name} className="h-10 w-10 rounded-full object-cover" />
        ) : (
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted text-muted-foreground">
            <UserRound className="size-5" />
          </div>
        )}
        <span className="font-medium text-foreground">{name}</span>
      </div>
      {slots.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("booking.empty.slots")}</p>
      ) : (
        <div className="flex flex-wrap gap-2">
          {slots.map((slot) => (
            <button
              key={slot.start_time}
              onClick={() => onSelectSlot(barber.id, slot)}
              className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                selectedStartTime === slot.start_time
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-foreground hover:bg-muted/70"
              }`}
            >
              {format(new Date(slot.start_time), "h:mm a")}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
