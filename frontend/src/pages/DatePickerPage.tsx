import { useState } from "react"
import { Calendar } from "../components/ui/calendar"
import { CalendarDate, getLocalTimeZone, today } from "@internationalized/date"
import { format } from "date-fns"
import { useI18n } from "../lib/i18n"

export function DatePickerPage() {
  const [selectedDate, setSelectedDate] = useState<CalendarDate>()
  const todayDate = today(getLocalTimeZone())
  const { t } = useI18n()

  return (
    <div className="min-h-app flex flex-col items-center bg-background p-8">
      <h1 className="text-2xl font-bold mb-8">{t("datepicker.title")}</h1>
      <Calendar
        value={selectedDate}
        onChange={setSelectedDate}
        minValue={todayDate}
        className="mx-auto"
      />
      {selectedDate && (
        <p className="mt-6 text-muted-foreground">
          {t("datepicker.selected", {
            date: format(
              new Date(selectedDate.year, selectedDate.month - 1, selectedDate.day),
              "PPP",
            ),
          })}
        </p>
      )}
    </div>
  )
}
