import { useState } from "react"
import { Input } from "../ui/input"
import { Label } from "../ui/label"
import { Button } from "../ui/button"
import { ImageUploadButton } from "../ui/image-upload-button"
import { useI18n } from "../../lib/i18n"
import type { Employee, Service } from "../../hooks/useApi"

export interface BarberFormValues {
  name: string
  email: string
  service_ids: string[]
}

function Field({
  label,
  error,
  children,
}: {
  label: string
  error?: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

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
  const { t } = useI18n()
  const [name, setName] = useState(initial?.display_name ?? "")
  const [email, setEmail] = useState("")
  const [serviceIds, setServiceIds] = useState<string[]>([])
  const [touched, setTouched] = useState({ name: false, email: false })

  const nameValid = name.trim().length > 0
  const emailValid = EMAIL_PATTERN.test(email.trim())

  const showNameError = touched.name && !nameValid
  const showEmailError = touched.email && !emailValid
  const emailError =
    email.trim() === "" ? t("barberForm.emailRequired") : t("barberForm.emailInvalid")

  const toggleService = (id: string) => {
    setServiceIds((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id],
    )
  }

  const handleSubmit = () => {
    setTouched({ name: true, email: true })
    if (nameValid && emailValid) {
      onSubmit({ name: name.trim(), email: email.trim(), service_ids: serviceIds })
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 pt-[calc(1rem+env(safe-area-inset-top))] pb-[calc(1rem+env(safe-area-inset-bottom))]"
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
                <ImageUploadButton
                  onPicked={onUploadAvatar}
                  label="Upload avatar"
                  uploading={uploading}
                />
              </Field>
            )}
            <div className="flex justify-end">
              <Button onClick={onClose}>Done</Button>
            </div>
          </div>
        ) : (
          <>
            <div className="space-y-4">
              <Field label="Name" error={showNameError ? t("barberForm.nameRequired") : undefined}>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  onBlur={() => setTouched((prev) => ({ ...prev, name: true }))}
                />
              </Field>
              <Field label="Email" error={showEmailError ? emailError : undefined}>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  onBlur={() => setTouched((prev) => ({ ...prev, email: true }))}
                />
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
              <Button isDisabled={saving} onClick={handleSubmit}>
                {saving ? "Inviting…" : "Invite"}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
