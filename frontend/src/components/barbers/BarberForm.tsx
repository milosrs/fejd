import { useRef, useState } from "react"
import { Input } from "../ui/input"
import { Label } from "../ui/label"
import { Button } from "../ui/button"
import type { Employee, Service } from "../../hooks/useApi"

export interface BarberFormValues {
  name: string
  email: string
  service_ids: string[]
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
    </div>
  )
}

export function BarberForm({
  initial,
  services,
  onClose,
  onSubmit,
  onUploadAvatar,
  uploading,
  saving,
}: {
  initial?: Employee
  services: Service[]
  onClose: () => void
  onSubmit: (values: BarberFormValues) => void
  onUploadAvatar?: (file: File) => void
  uploading?: boolean
  saving?: boolean
}) {
  const [name, setName] = useState(initial?.display_name ?? "")
  const [email, setEmail] = useState("")
  const [serviceIds, setServiceIds] = useState<string[]>([])
  const fileRef = useRef<HTMLInputElement>(null)

  const toggleService = (id: string) => {
    setServiceIds((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id],
    )
  }

  const handleSubmit = () => {
    if (name.trim() && email.trim()) {
      onSubmit({ name: name.trim(), email: email.trim(), service_ids: serviceIds })
    }
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
            {initial ? "Barber avatar" : "Add barber"}
          </h3>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>

        {initial ? (
          <div className="space-y-4">
            <Field label="Name">
              <Input value={initial.display_name || initial.user_id} disabled />
            </Field>
            {onUploadAvatar && (
              <Field label="Avatar">
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    isDisabled={uploading}
                    onClick={() => fileRef.current?.click()}
                  >
                    {uploading ? "Uploading…" : "Upload avatar"}
                  </Button>
                  <input
                    ref={fileRef}
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={(e) => {
                      const file = e.target.files?.[0]
                      if (file) onUploadAvatar(file)
                      e.target.value = ""
                    }}
                  />
                </div>
              </Field>
            )}
            <div className="flex justify-end">
              <Button onClick={onClose}>Done</Button>
            </div>
          </div>
        ) : (
          <>
            <div className="space-y-4">
              <Field label="Name">
                <Input value={name} onChange={(e) => setName(e.target.value)} />
              </Field>
              <Field label="Email">
                <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
              </Field>
              <Field label="Services">
                {services.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No services yet.</p>
                ) : (
                  <div className="max-h-40 space-y-2 overflow-y-auto rounded-xl border border-border p-3">
                    {services.map((s) => (
                      <label key={s.id} className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={serviceIds.includes(s.id)}
                          onChange={() => toggleService(s.id)}
                        />
                        <span className="text-foreground">{s.name}</span>
                      </label>
                    ))}
                  </div>
                )}
              </Field>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <Button variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button
                isDisabled={saving || !name.trim() || !email.trim()}
                onClick={handleSubmit}
              >
                {saving ? "Inviting…" : "Invite"}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
