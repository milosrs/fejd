import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { renameBusiness } from "../hooks/useApi"
import { salonPath } from "../lib/salonDomain"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

export function RenameSalonDialog({
  businessId,
  slug,
  currentName,
  onClose,
}: {
  businessId: string
  slug: string
  currentName: string
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [name, setName] = useState(currentName)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  const handleSave = async () => {
    const trimmed = name.trim()
    if (!trimmed) {
      setError("Name is required.")
      return
    }
    setSaving(true)
    setError("")
    try {
      const business = await renameBusiness(businessId, trimmed)
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      await queryClient.invalidateQueries({ queryKey: ["businesses"] })
      await queryClient.invalidateQueries({ queryKey: ["me"] })
      onClose()
      if (business?.slug && business.slug !== slug) {
        navigate(salonPath(business.slug), { replace: true })
      }
    } catch {
      setError("Failed to rename salon.")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-semibold text-foreground">Rename salon</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Update the name shown to your customers.
        </p>

        <div className="mt-4 space-y-1">
          <Label htmlFor="salon-name">Salon name</Label>
          <Input
            id="salon-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Salon name"
          />
        </div>

        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave} isDisabled={saving || !name.trim()}>
            {saving ? "Saving…" : "Save name"}
          </Button>
        </div>
      </div>
    </div>
  )
}
