import { useEffect, useRef, useState } from "react"
import { UserRound } from "lucide-react"
import { Input } from "../ui/input"
import { Textarea } from "../ui/textarea"
import { Label } from "../ui/label"
import { Button } from "../ui/button"
import { ImageUploadButton } from "../ui/image-upload-button"
import { ServiceCard } from "./ServiceCard"
import { resolveImageUrl } from "../../lib/images"
import type { Employee, Service } from "../../hooks/useApi"

export interface ServiceFormValues {
  name: string
  description: string
  duration_minutes: number
  price: number
  imageFile?: File | null
  employee_ids: string[]
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
  staff,
  assignedEmployeeIds,
  onClose,
  onSubmit,
  saving,
  error,
}: {
  initial?: Service
  staff: Employee[]
  assignedEmployeeIds: string[]
  onClose: () => void
  onSubmit: (values: ServiceFormValues) => void
  saving?: boolean
  error?: string
}) {
  const [name, setName] = useState(initial?.name ?? "")
  const [description, setDescription] = useState(initial?.description ?? "")
  const [duration, setDuration] = useState(initial ? String(initial.duration_minutes) : "30")
  const [price, setPrice] = useState(
    initial && initial.price != null ? String(initial.price) : "",
  )
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string>()
  const [employeeIds, setEmployeeIds] = useState<string[]>(assignedEmployeeIds)
  const employeesTouched = useRef(false)

  useEffect(() => {
    if (!employeesTouched.current) {
      setEmployeeIds(assignedEmployeeIds)
    }
  }, [assignedEmployeeIds])

  useEffect(() => {
    if (!imageFile) {
      setPreviewUrl(undefined)
      return
    }
    const url = URL.createObjectURL(imageFile)
    setPreviewUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [imageFile])

  const toggleEmployee = (id: string) => {
    employeesTouched.current = true
    setEmployeeIds((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id],
    )
  }

  const handleSubmit = () => {
    if (!name.trim()) return
    onSubmit({
      name: name.trim(),
      description: description.trim(),
      duration_minutes: parseInt(duration) || 0,
      price: price ? parseFloat(price) : 0,
      imageFile,
      employee_ids: employeeIds,
    })
  }

  const existingImage = initial?.picture_id
    ? resolveImageUrl(`/api/images/${initial.picture_id}`)
    : undefined

  const previewService: Service = {
    id: initial?.id ?? "",
    business_id: "",
    created_at: "",
    name: name.trim() || "Service name",
    description: description.trim() || undefined,
    duration_minutes: parseInt(duration) || 0,
    price: price ? parseFloat(price) : undefined,
    active: true,
    picture_id: initial?.picture_id,
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 pt-[calc(1rem+env(safe-area-inset-top))] pb-[calc(1rem+env(safe-area-inset-bottom))]"
      onClick={onClose}
    >
      <div
        className="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-2xl border border-border bg-background p-6"
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

        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
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

            <Field label="Barbers">
              {staff.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No barbers yet. Invite barbers from the Barbers page.
                </p>
              ) : (
                <div className="max-h-40 space-y-1 overflow-y-auto rounded-xl border border-border p-2">
                  {staff.map((employee) => {
                    const avatar = resolveImageUrl(employee.avatar)
                    const name =
                      employee.display_name?.trim() ||
                      (employee.role === "admin" ? "You" : "Barber")
                    return (
                      <label
                        key={employee.id}
                        className="flex items-center gap-3 rounded-lg px-2 py-1.5 text-sm hover:bg-muted"
                      >
                        {avatar ? (
                          <img
                            src={avatar}
                            alt={name}
                            className="h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-border"
                          />
                        ) : (
                          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground ring-1 ring-border">
                            <UserRound className="size-4" />
                          </div>
                        )}
                        <span className="min-w-0 flex-1 truncate text-foreground">
                          {name}
                        </span>
                        <input
                          type="checkbox"
                          checked={employeeIds.includes(employee.id)}
                          onChange={() => toggleEmployee(employee.id)}
                          className="shrink-0"
                        />
                      </label>
                    )
                  })}
                </div>
              )}
            </Field>

            <Field label="Picture">
              <div className="space-y-2">
                {(previewUrl || existingImage) && (
                  <img
                    src={previewUrl ?? existingImage}
                    alt=""
                    className="h-24 w-full rounded-lg border border-border object-cover"
                  />
                )}
                <ImageUploadButton
                  label={previewUrl || existingImage ? "Replace image" : "Add image"}
                  onPicked={setImageFile}
                />
              </div>
            </Field>
          </div>

          <div className="rounded-xl border border-dashed border-border p-4">
            <p className="mb-3 text-xs uppercase tracking-wide text-muted-foreground">
              Live preview
            </p>
            <ServiceCard
              service={previewService}
              imageOverride={previewUrl ?? undefined}
              onBook={() => {}}
            />
          </div>
        </div>

        <div className="mt-6 flex items-center justify-between gap-2">
          {error ? (
            <p className="text-sm text-destructive">{error}</p>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button isDisabled={saving || !name.trim()} onClick={handleSubmit}>
              {saving ? "Saving…" : "Save"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
