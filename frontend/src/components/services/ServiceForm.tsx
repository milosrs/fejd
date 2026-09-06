import { useRef, useState } from "react"
import { Input } from "../ui/input"
import { Textarea } from "../ui/textarea"
import { Label } from "../ui/label"
import { Button } from "../ui/button"
import type { Service } from "../../hooks/useApi"

export interface ServiceFormValues {
  name: string
  description: string
  duration_minutes: number
  price: number
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
    </div>
  )
}

export function ServiceForm({
  initial,
  onClose,
  onSubmit,
  onUploadImage,
  uploading,
  saving,
}: {
  initial?: Service
  onClose: () => void
  onSubmit: (values: ServiceFormValues) => void
  onUploadImage?: (file: File) => void
  uploading?: boolean
  saving?: boolean
}) {
  const [name, setName] = useState(initial?.name ?? "")
  const [description, setDescription] = useState(initial?.description ?? "")
  const [duration, setDuration] = useState(initial ? String(initial.duration_minutes) : "30")
  const [price, setPrice] = useState(
    initial && initial.price != null ? String(initial.price) : "",
  )
  const fileRef = useRef<HTMLInputElement>(null)

  const handleSubmit = () => {
    if (!name.trim()) return
    onSubmit({
      name: name.trim(),
      description: description.trim(),
      duration_minutes: parseInt(duration) || 0,
      price: price ? parseFloat(price) : 0,
    })
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md max-h-[90vh] overflow-y-auto rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">
            {initial ? "Edit service" : "Add service"}
          </h3>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>

        <div className="space-y-4">
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Description">
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </Field>
          <Field label="Duration (minutes)">
            <Input
              type="number"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
            />
          </Field>
          <Field label="Price ($)">
            <Input
              type="number"
              step="0.01"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
          </Field>

          {initial && onUploadImage && (
            <Field label="Picture">
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  isDisabled={uploading}
                  onClick={() => fileRef.current?.click()}
                >
                  {uploading ? "Uploading…" : "Upload image"}
                </Button>
                <input
                  ref={fileRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (file) onUploadImage(file)
                    e.target.value = ""
                  }}
                />
              </div>
            </Field>
          )}
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button isDisabled={saving || !name.trim()} onClick={handleSubmit}>
            {saving ? "Saving…" : "Save"}
          </Button>
        </div>
      </div>
    </div>
  )
}
