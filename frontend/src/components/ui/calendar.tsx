import { DayPicker } from "react-day-picker"
import { cn } from "../../lib/utils"

function Calendar({
  className,
  showOutsideDays = true,
  classNames,
  ...props
}: React.ComponentProps<typeof DayPicker>) {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn("p-3", className)}
      classNames={{
        root: "w-full",
        months: "flex flex-col sm:flex-row gap-2",
        month: "flex flex-col gap-4",
        month_caption: "flex justify-center pt-1 relative items-center",
        caption_label: "text-sm font-medium",
        nav: "flex items-center gap-1",
        button_previous: "absolute left-1 size-7 bg-transparent p-0 opacity-50 hover:opacity-100 border border-border rounded-md",
        button_next: "absolute right-1 size-7 bg-transparent p-0 opacity-50 hover:opacity-100 border border-border rounded-md",
        chevron: "fill-foreground size-4",
        month_grid: "w-full border-collapse space-y-1",
        weekdays: "flex",
        weekday: "text-muted-foreground rounded-md w-8 font-normal text-[0.8rem]",
        week: "flex w-full mt-2",
        day: cn(
          "relative p-0 text-center text-sm focus-within:relative focus-within:z-20 [&:has([aria-selected])]:bg-primary/20 rounded-md",
        ),
        day_button: cn(
          "size-8 p-0 font-normal aria-selected:opacity-100 rounded-md hover:bg-muted",
        ),
        selected: "bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground focus:bg-primary focus:text-primary-foreground",
        today: "bg-muted text-foreground",
        outside: "text-muted-foreground opacity-50",
        disabled: "text-muted-foreground opacity-50",
        range_start: "day-range-start",
        range_end: "day-range-end",
        range_middle: "aria-selected:bg-primary/20 aria-selected:text-foreground",
        hidden: "invisible",
        ...classNames,
      }}
      {...props}
    />
  )
}

export { Calendar }
