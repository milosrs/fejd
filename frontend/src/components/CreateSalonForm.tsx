import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useOnboardingStore } from "../stores/onboardingStore"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

export function CreateSalonForm() {
  const navigate = useNavigate()
  const createBusiness = useOnboardingStore((s) => s.createBusiness)
  const [name, setName] = useState("")
  const [creating, setCreating] = useState(false)

  const handleCreate = async () => {
    if (!name.trim()) return
    setCreating(true)
    try {
      const slug = await createBusiness(name.trim())
      if (slug) navigate(`/${slug}`)
    } catch {
      // keep the form open; the backend surfaces a 4xx
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="w-full max-w-xs space-y-3">
      <Label>Salon name</Label>
      <Input
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="My Salon"
      />
      <Button
        className="w-full"
        isDisabled={creating || !name.trim()}
        onClick={handleCreate}
      >
        {creating ? "Creating…" : "Create your first salon"}
      </Button>
    </div>
  )
}
