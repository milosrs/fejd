import { useEffect, useRef, useCallback } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { API_BASE_URL } from "../lib/api"

// Events that change the salon's bookable slots (bookings, cancellations,
// reassignments and schedule changes). Each is a distinct SSE event emitted by
// the backend, but all refresh the slot grid on the booking page.
const SLOT_EVENTS = [
  "slots_updated",
  "appointment_booked",
  "appointment_cancelled",
  "appointment_reassigned",
] as const

export function useTimeSlotStream(businessSlug: string) {
  const queryClient = useQueryClient()
  const eventSourceRef = useRef<EventSource | null>(null)

  const connect = useCallback(() => {
    if (!businessSlug) return

    const url = `${API_BASE_URL}/api/sse/business/${businessSlug}/slots`
    const es = new EventSource(url)

    const invalidateSlots = (event: Event) => {
      queryClient.invalidateQueries({ queryKey: ["slots", businessSlug] })
      try {
        const data = JSON.parse((event as MessageEvent).data)
        window.dispatchEvent(
          new CustomEvent("slot-update", {
            detail: data,
          }),
        )
      } catch {
        // ignore parse errors
      }
    }

    for (const name of SLOT_EVENTS) {
      es.addEventListener(name, invalidateSlots)
    }

    es.addEventListener("closures_updated", () => {
      // A newly closed day has no slots, and the calendar must disable it
      // immediately, so refresh both the closure list and the slot grid.
      queryClient.invalidateQueries({ queryKey: ["business-closures", businessSlug] })
      queryClient.invalidateQueries({ queryKey: ["slots", businessSlug] })
    })

    es.addEventListener("connected", () => {
      // connected
    })

    es.onerror = () => {
      es.close()
      setTimeout(() => {
        if (eventSourceRef.current === es) {
          connect()
        }
      }, 3000)
    }

    eventSourceRef.current = es
  }, [businessSlug, queryClient])

  useEffect(() => {
    connect()
    return () => {
      eventSourceRef.current?.close()
      eventSourceRef.current = null
    }
  }, [connect])
}
