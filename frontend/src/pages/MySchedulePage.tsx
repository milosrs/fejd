import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { format } from "date-fns"
import { useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "../stores/authStore"
import { useMyUnavailability, reserveOwnSlot, deleteOwnSlot } from "../hooks/useApi"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"

function toRFC3339(date: string, time: string): string {
  return new Date(`${date}T${time}:00`).toISOString()
}

export function MySchedulePage() {
  const { businessId } = useParams<{ businessId: string }>()
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const queryClient = useQueryClient()

  const { data: blocks, isLoading } = useMyUnavailability(businessId!)

  const [date, setDate] = useState(() => format(new Date(), "yyyy-MM-dd"))
  const [startTime, setStartTime] = useState("09:00")
  const [endTime, setEndTime] = useState("10:00")
  const [reason, setReason] = useState("")
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState("")
  const [isError, setIsError] = useState(false)

  if (!authenticated) {
    return (
      <div className="min-h-app bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to manage your schedule.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: ["my-unavailability", businessId] })

  const handleReserve = async () => {
    if (!date || !startTime || !endTime) return
    if (endTime <= startTime) {
      setMessage("End time must be after start time.")
      setIsError(true)
      return
    }
    setSaving(true)
    setMessage("")
    try {
      await reserveOwnSlot(businessId!, {
        start_time: toRFC3339(date, startTime),
        end_time: toRFC3339(date, endTime),
        reason: reason || undefined,
      })
      setReason("")
      setMessage("Time slot reserved.")
      setIsError(false)
      await refresh()
    } catch (err) {
      const e = err as { body?: { error?: string } }
      setMessage(e?.body?.error ?? "Failed to reserve. The slot may already be blocked.")
      setIsError(true)
    } finally {
      setSaving(false)
    }
  }

  const handleRemove = async (id: string) => {
    try {
      await deleteOwnSlot(businessId!, id)
      setMessage("Blocked slot removed.")
      setIsError(false)
      await refresh()
    } catch {
      setMessage("Failed to remove blocked slot.")
      setIsError(true)
    }
  }

  const upcoming = (blocks ?? []).filter((b) => new Date(b.end_time) > new Date())

  return (
    <div className="min-h-app bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex gap-4 items-center">
          <h1 className="text-xl font-semibold text-foreground">My Schedule</h1>
          <Button variant="outline" size="sm" onClick={() => navigate("/my/appointments")}>
            My appointments
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Reserve a time slot</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Reserved slots are hidden from customers, so they cannot book during
              this time.
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div className="space-y-1">
                <Label>Date</Label>
                <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
              </div>
              <div className="space-y-1">
                <Label>Start time</Label>
                <Input type="time" value={startTime} onChange={(e) => setStartTime(e.target.value)} />
              </div>
              <div className="space-y-1">
                <Label>End time</Label>
                <Input type="time" value={endTime} onChange={(e) => setEndTime(e.target.value)} />
              </div>
            </div>
            <div className="space-y-1">
              <Label>Reason (optional)</Label>
              <Input
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="e.g. urgent errand"
              />
            </div>
            {message && (
              <p className={`text-sm ${isError ? "text-red-500" : "text-green-600"}`}>
                {message}
              </p>
            )}
            <Button onClick={handleReserve} isDisabled={saving} className="w-full">
              {saving ? "Reserving..." : "Reserve slot"}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Reserved time slots</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {isLoading ? (
              <p className="text-muted-foreground">Loading…</p>
            ) : upcoming.length === 0 ? (
              <p className="text-muted-foreground">No reserved time slots.</p>
            ) : (
              upcoming.map((b) => (
                <div key={b.id} className="flex items-center justify-between p-3 bg-muted rounded-md">
                  <div>
                    <span className="font-medium">
                      {format(new Date(b.start_time), "EEE, MMM d yyyy")}
                    </span>
                    <span className="text-muted-foreground ml-2">
                      {format(new Date(b.start_time), "h:mm a")} – {format(new Date(b.end_time), "h:mm a")}
                    </span>
                    {b.reason && <span className="text-muted-foreground ml-2">({b.reason})</span>}
                    {b.status === "pending" && (
                      <span className="ml-2 inline-flex items-center rounded-full bg-yellow-500/15 px-2 py-0.5 text-xs font-medium text-yellow-600 dark:text-yellow-400">
                        awaiting approval
                      </span>
                    )}
                  </div>
                  <Button variant="destructive" size="sm" onClick={() => handleRemove(b.id)}>
                    Remove
                  </Button>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  )
}
