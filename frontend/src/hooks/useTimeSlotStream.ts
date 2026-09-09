import { useEffect, useRef, useCallback } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { API_BASE_URL } from "../lib/api"

export function useTimeSlotStream(businessSlug: string) {
  const queryClient = useQueryClient()
  const eventSourceRef = useRef<EventSource | null>(null)

  const connect = useCallback(() => {
    if (!businessSlug) return

    const url = `${API_BASE_URL}/api/sse/business/${businessSlug}/slots`
    const es = new EventSource(url)

    es.addEventListener("slots_updated", (event) => {
      try {
        const data = JSON.parse(event.data)
        queryClient.invalidateQueries({ queryKey: ["slots", businessSlug] })
        window.dispatchEvent(
          new CustomEvent("slot-update", {
            detail: data,
          }),
        )
      } catch {
        // ignore parse errors
      }
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
