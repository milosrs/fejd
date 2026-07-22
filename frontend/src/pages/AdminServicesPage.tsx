import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"
import { useServices, createService, updateService, deleteService } from "../hooks/useApi"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"

export function AdminServicesPage() {
  const { businessId } = useParams<{ businessId: string }>()
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { data: services, refetch } = useServices(businessId!)

  const [name, setName] = useState("")
  const [duration, setDuration] = useState("30")
  const [price, setPrice] = useState("")
  const [adding, setAdding] = useState(false)
  const [message, setMessage] = useState("")
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState("")
  const [editDuration, setEditDuration] = useState("")
  const [editPrice, setEditPrice] = useState("")

  if (!authenticated) {
    return (
      <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Please log in to access admin.</p>
        <Button onClick={login}>Login</Button>
      </div>
    )
  }

  const handleAdd = async () => {
    if (!name || !duration) return
    setAdding(true)
    setMessage("")
    try {
      await createService(businessId!, {
        name,
        duration_minutes: parseInt(duration),
        price: price ? parseFloat(price) : 0,
        active: true,
      })
      setName("")
      setDuration("30")
      setPrice("")
      setMessage("Service added.")
      refetch()
    } catch {
      setMessage("Failed to add service.")
    } finally {
      setAdding(false)
    }
  }

  const handleUpdate = async (serviceId: string) => {
    try {
      await updateService(businessId!, serviceId, {
        name: editName,
        duration_minutes: parseInt(editDuration),
        price: editPrice ? parseFloat(editPrice) : 0,
      })
      setEditingId(null)
      setMessage("Service updated.")
      refetch()
    } catch {
      setMessage("Failed to update service.")
    }
  }

  const handleDelete = async (serviceId: string) => {
    if (!confirm("Delete this service?")) return
    try {
      await deleteService(businessId!, serviceId)
      refetch()
    } catch {
      setMessage("Failed to delete service.")
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex gap-4 items-center">
          <h1 className="text-xl font-semibold text-foreground">Service Management</h1>
          <Button variant="outline" size="sm" onClick={() => navigate(`/admin/business/${businessId}/schedule`)}>
            Schedule
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Add Service</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <Label>Name</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Haircut" />
            </div>
            <div>
              <Label>Duration (minutes)</Label>
              <Input type="number" value={duration} onChange={(e) => setDuration(e.target.value)} />
            </div>
            <div>
              <Label>Price ($)</Label>
              <Input type="number" step="0.01" value={price} onChange={(e) => setPrice(e.target.value)} placeholder="25.00" />
            </div>
            {message && <p className={`text-sm ${message.includes("Failed") ? "text-red-500" : "text-green-600"}`}>{message}</p>}
            <Button onClick={handleAdd} disabled={adding} className="w-full">
              {adding ? "Adding..." : "Add Service"}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Existing Services</CardTitle>
          </CardHeader>
          <CardContent>
            {(services || []).length === 0 ? (
              <p className="text-muted-foreground">No services yet.</p>
            ) : (
              <div className="space-y-3">
                {(services || []).map((svc: any) => (
                  <div key={svc.id} className="p-3 bg-muted rounded-md">
                    {editingId === svc.id ? (
                      <div className="space-y-2">
                        <Input value={editName} onChange={(e) => setEditName(e.target.value)} placeholder="Name" />
                        <Input type="number" value={editDuration} onChange={(e) => setEditDuration(e.target.value)} placeholder="Duration" />
                        <Input type="number" step="0.01" value={editPrice} onChange={(e) => setEditPrice(e.target.value)} placeholder="Price" />
                        <div className="flex gap-2">
                          <Button size="sm" onClick={() => handleUpdate(svc.id)}>Save</Button>
                          <Button size="sm" variant="ghost" onClick={() => setEditingId(null)}>Cancel</Button>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-center justify-between">
                        <div>
                          <span className="font-medium">{svc.name}</span>
                          <span className="text-muted-foreground ml-3">{svc.duration_minutes} min</span>
                          {svc.price > 0 && <span className="text-muted-foreground ml-3">${svc.price.toFixed(2)}</span>}
                          {!svc.active && <span className="text-red-500 ml-2">(Inactive)</span>}
                        </div>
                        <div className="flex gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setEditingId(svc.id)
                              setEditName(svc.name)
                              setEditDuration(svc.duration_minutes.toString())
                              setEditPrice(svc.price?.toString() || "")
                            }}
                          >
                            Edit
                          </Button>
                          <Button size="sm" variant="destructive" onClick={() => handleDelete(svc.id)}>
                            Delete
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  )
}
