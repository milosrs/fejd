import { useState } from "react"
import { useAuthStore } from "../stores/authStore"
import { Calendar } from "../components/ui/calendar"
import { format } from "date-fns"

export function DatePickerPage() {
  const [selectedDate, setSelectedDate] = useState<Date>()

  return (
    <div className="min-h-screen flex flex-col items-center bg-background p-8">
      <h1 className="text-2xl font-bold mb-8">Select a Date</h1>
      <Calendar
        mode="single"
        selected={selectedDate}
        onSelect={setSelectedDate}
        disabled={(date) => {
          const today = new Date()
          today.setHours(0, 0, 0, 0)
          return date < today
        }}
        className="mx-auto"
      />
      {selectedDate && (
        <p className="mt-6 text-muted-foreground">
          Selected: {format(selectedDate, "PPP")}
        </p>
      )}
    </div>
  )
}
