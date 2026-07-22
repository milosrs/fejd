import { useState, useEffect } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"
import { useAdminEmployees, useAdminWorkingHours, updateWorkingHours, addOverride, deleteOverride } from "../hooks/useApi"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"
import { Select } from "../components/ui/select"

const DAYS = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]

interface HourRow {
  day_of_week: number
  start_time: string
  end_time: string
}

export function AdminSchedulePage() {
  const { businessId } = useParams<{ businessId: string }>()
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)

  const { data: employees } = useAdminEmployees(businessId!)
  const [selectedUserId, setSelectedUserId] = useState("")
  const { data: whData } = useAdminWorkingHours(businessId!, selectedUserId)

  const [hours, setHours] = useState<HourRow[]>([])
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState("")

  useEffect(() => {
    if (whData?.working_hours) {
      setHours(
        whData.working_hours.map((wh: any) => ({
          day_of_week: wh.day_of_week,
          start_time: wh.start_time,
          end_time: wh.end_time,
        })),
      )
    } else {
      setHours(
        DAYS.map((_, i) => ({
          day_of_week: i,
          start_time: "09:00",
          end_time: "17:00",
        })),
      )
    }
  }, [whData])

  if (!authenticated) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to access admin.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  const updateRow = (index: number, field: keyof HourRow, value: string) => {
    const newHours = [...hours]
    newHours[index] = { ...newHours[index], [field]: value }
    setHours(newHours)
  }

  const handleSave = async () => {
    setSaving(true)
    setMessage("")
    try {
      await updateWorkingHours(businessId!, selectedUserId, hours)
      setMessage("Working hours saved.")
    } catch {
      setMessage("Failed to save.")
    } finally {
      setSaving(false)
    }
  }

  const handleAddOverride = async () => {
    const date = prompt("Enter date (YYYY-MM-DD):")
    if (!date) return
    const isOff = confirm("Is this a full day off? Click OK for yes, Cancel for no.")
    let startTime: string | null = null
    let endTime: string | null = null
    if (!isOff) {
      startTime = prompt("Start time (HH:MM):", "09:00")
      endTime = prompt("End time (HH:MM):", "17:00")
    }
    const reason = prompt("Reason (optional):") || ""
    try {
      await addOverride(businessId!, selectedUserId, {
        override_date: date,
        start_time: startTime ?? undefined,
        end_time: endTime ?? undefined,
        is_off: isOff,
        reason,
      })
      setMessage("Override added.")
    } catch {
      setMessage("Failed to add override.")
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex gap-4 items-center">
          <h1 className="text-xl font-semibold text-foreground">Schedule Management</h1>
          <Button variant="outline" size="sm" onClick={() => navigate(`/admin/business/${businessId}/services`)}>
            Services
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Select Employee</CardTitle>
          </CardHeader>
          <CardContent>
            <select
              className="w-full h-10 rounded-md border border-border bg-background px-3 py-2 text-sm"
              value={selectedUserId}
              onChange={(e) => setSelectedUserId(e.target.value)}
            >
              <option value="">-- Select --</option>
              {(employees || []).map((emp: any) => (
                <option key={emp.user_id} value={emp.user_id}>
                  {emp.display_name || emp.user_id}
                </option>
              ))}
            </select>
          </CardContent>
        </Card>

        {selectedUserId && (
          <>
            <Card>
              <CardHeader>
                <CardTitle>Weekly Working Hours</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {DAYS.map((day, i) => (
                  <div key={day} className="flex items-center gap-3">
                    <Label className="w-24">{day}</Label>
                    <Input
                      type="time"
                      className="w-32"
                      value={hours[i]?.start_time || ""}
                      onChange={(e) => updateRow(i, "start_time", e.target.value)}
                    />
                    <span className="text-muted-foreground">to</span>
                    <Input
                      type="time"
                      className="w-32"
                      value={hours[i]?.end_time || ""}
                      onChange={(e) => updateRow(i, "end_time", e.target.value)}
                    />
                  </div>
                ))}
                {message && (
                  <p className={`text-sm ${message.includes("Failed") ? "text-red-500" : "text-green-600"}`}>
                    {message}
                  </p>
                )}
                <Button onClick={handleSave} disabled={saving} className="w-full">
                  {saving ? "Saving..." : "Save Working Hours"}
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Date Overrides</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {whData?.overrides && whData.overrides.length > 0 && (
                  <div className="space-y-2 mb-4">
                    {whData.overrides.map((o: any) => (
                      <div key={o.id} className="flex items-center justify-between p-2 bg-muted rounded-md">
                        <div>
                          <span className="font-medium">{o.override_date}</span>
                          {o.is_off ? (
                            <span className="text-red-500 ml-2">Day off</span>
                          ) : (
                            <span className="text-muted-foreground ml-2">
                              {o.start_time} - {o.end_time}
                            </span>
                          )}
                          {o.reason && <span className="text-muted-foreground ml-2">({o.reason})</span>}
                        </div>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={async () => {
                            try {
                              await deleteOverride(businessId!, selectedUserId, o.id)
                              setMessage("Override removed.")
                            } catch {
                              setMessage("Failed to remove override.")
                            }
                          }}
                        >
                          Remove
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
                <Button variant="outline" onClick={handleAddOverride} className="w-full">
                  Add Date Override
                </Button>
              </CardContent>
            </Card>
          </>
        )}
      </main>
    </div>
  )
}
