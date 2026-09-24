import { useEffect, useRef, useCallback } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { API_BASE_URL } from "../lib/api"

// useReservationsStream keeps the staff "my reservations" calendar in sync with
// bookings made elsewhere (e.g. a customer booking in another window) by
// subscribing to the salon's slot/booking SSE stream and invalidating the
// queries that feed the calendar and pending-approvals card.
export function useReservationsStream(businessId: string, slug?: string) {
  const queryClient = useQueryClient()
  const eventSourceRef = useRef<EventSource | null>(null)

  const connect = useCallback(() => {
    if (!businessId || !slug) return

    const url = `${API_BASE_URL}/api/sse/business/${slug}/slots`
    const es = new EventSource(url)

    es.addEventListener("slots_updated", () => {
      queryClient.invalidateQueries({ queryKey: ["my-reservations", businessId] })
      queryClient.invalidateQueries({ queryKey: ["business-appointments", businessId] })
      queryClient.invalidateQueries({ queryKey: ["business-unavailability", businessId] })
      queryClient.invalidateQueries({ queryKey: ["customers", businessId] })
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
  }, [businessId, slug, queryClient])

  useEffect(() => {
    connect()
    return () => {
      eventSourceRef.current?.close()
      eventSourceRef.current = null
    }
  }, [connect])
}
