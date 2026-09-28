"use client"

import * as React from "react"
import {
  composeRenderProps,
  DateInput as DateInputPrimitive,
  DateSegment as DateSegmentPrimitive,
  TimeField as TimeFieldPrimitive,
  type DateInputProps,
  type DateSegmentProps,
  type TimeFieldProps,
  type TimeValue,
} from "react-aria-components"

import { cn } from "#lib/utils"

function TimeField<T extends TimeValue>({
  className,
  ...props
}: TimeFieldProps<T>) {
  return (
    <TimeFieldPrimitive
      data-slot="time-field"
      className={composeRenderProps(className, (className) => cn(className))}
      {...props}
    />
  )
}

function TimeFieldInput({ className, ...props }: DateInputProps) {
  return (
    <DateInputPrimitive
      data-slot="time-field-input"
      className={composeRenderProps(className, (className) =>
        cn(
          "flex h-8 w-full min-w-0 items-center gap-0.5 rounded-2xl border border-transparent bg-input/50 px-2.5 py-1 text-base tabular-nums transition-[color,box-shadow] duration-200 outline-none data-focus-within:border-ring data-focus-within:ring-3 data-focus-within:ring-ring/30 data-disabled:cursor-not-allowed data-disabled:opacity-50 data-invalid:border-destructive data-invalid:ring-3 data-invalid:ring-destructive/20 md:text-sm dark:data-invalid:border-destructive/50 dark:data-invalid:ring-destructive/40",
          className,
        ),
      )}
      {...props}
    />
  )
}

function TimeFieldSegment({ className, ...props }: DateSegmentProps) {
  return (
    <DateSegmentPrimitive
      data-slot="time-field-segment"
      className={composeRenderProps(className, (className) =>
        cn(
          "rounded-sm tabular-nums caret-transparent outline-none data-focused:bg-foreground data-focused:text-background data-placeholder:text-muted-foreground data-[type=literal]:px-0 data-[type=literal]:text-muted-foreground data-[type=literal]:[&.is-placeholder]:text-muted-foreground",
          className,
        ),
      )}
      {...props}
    />
  )
}

export { TimeField, TimeFieldInput, TimeFieldSegment }
