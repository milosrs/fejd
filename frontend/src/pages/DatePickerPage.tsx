import { useState } from "react"
import { Calendar } from "../components/ui/calendar"
import { CalendarDate, getLocalTimeZone, today } from "@internationalized/date"
import { format } from "date-fns"

export function DatePickerPage() {
  const [selectedDate, setSelectedDate] = useState<CalendarDate>()
  const todayDate = today(getLocalTimeZone())

  return (
    <div className="min-h-screen flex flex-col items-center bg-background p-8">
      <h1 className="text-2xl font-bold mb-8">Select a Date</h1>
      <Calendar
        value={selectedDate}
        onChange={setSelectedDate}
        minValue={todayDate}
        className="mx-auto"
      />
      {selectedDate && (
        <p className="mt-6 text-muted-foreground">
          Selected:{" "}
          {format(
            new Date(selectedDate.year, selectedDate.month - 1, selectedDate.day),
            "PPP"
          )}
        </p>
      )}
    </div>
  )
}
