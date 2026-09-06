import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useServices, type Service } from "../hooks/useApi"
import { useServiceMutations } from "../hooks/useServiceMutations"
import { useBookingStore } from "../stores/bookingStore"
import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"
import { ServiceCard, ServiceCardSkeleton } from "../components/services/ServiceCard"
import { ServiceForm, type ServiceFormValues } from "../components/services/ServiceForm"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { Button } from "../components/ui/button"
import { Plus } from "lucide-react"

type FormState = { mode: "create" } | { mode: "edit"; service: Service } | null

export function ServicesPage() {
  const { slug, salon, editing } = useSalonContext()
  const isOwner = useIsOwner()
  const navigate = useNavigate()
  const setService = useBookingStore((s) => s.setService)
  const reset = useBookingStore((s) => s.reset)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { t, ready } = useI18n()

  const businessId = salon?.business.id ?? ""
  const { data, isLoading, isError } = useServices(slug)
  const { create, update, remove, uploadImage } = useServiceMutations(businessId, slug)

  const [form, setForm] = useState<FormState>(null)
  const [deleting, setDeleting] = useState<Service | null>(null)
  const [deleteError, setDeleteError] = useState("")

  const editingOn = editing && isOwner
  const services = editingOn ? (data ?? []) : (data ?? []).filter((s) => s.active)
  const registerNote = !authenticated && ready ? t("services.book.requiresAuth") : undefined

  const handleBook = (serviceId: string) => {
    if (!authenticated) {
      login()
      return
    }
    reset()
    setService(serviceId)
    navigate(`/${slug}/book?service=${serviceId}`)
  }

  const handleSubmit = (values: ServiceFormValues) => {
    if (!form) return
    if (form.mode === "create") {
      create.mutate(
        {
          name: values.name,
          description: values.description,
          duration_minutes: values.duration_minutes,
          price: values.price,
          active: true,
        },
        {
          onSuccess: (service) => {
            if (service) setForm({ mode: "edit", service })
          },
        },
      )
    } else {
      update.mutate(
        {
          serviceId: form.service.id,
          input: {
            name: values.name,
            description: values.description,
            duration_minutes: values.duration_minutes,
            price: values.price,
            active: form.service.active,
          },
        },
        { onSuccess: () => setForm(null) },
      )
    }
  }

  const handleUploadImage = (file: File) => {
    if (form?.mode === "edit") {
      uploadImage.mutate({ serviceId: form.service.id, file })
    }
  }

  const handleDelete = () => {
    if (!deleting) return
    remove.mutate(deleting.id, {
      onSuccess: () => {
        setDeleting(null)
        setDeleteError("")
      },
      onError: (err) => {
        const e = err as unknown as { body?: { error?: string } }
        setDeleteError(e?.body?.error ?? "Failed to delete service.")
      },
    })
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">Services</h2>
        {editingOn && (
          <Button size="sm" onClick={() => setForm({ mode: "create" })}>
            <Plus className="size-3" /> Add service
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <ServiceCardSkeleton key={i} />
          ))}
        </div>
      ) : isError ? (
        <p className="py-12 text-center text-muted-foreground">
          Couldn't load services. Please try again.
        </p>
      ) : services.length === 0 ? (
        <p className="py-12 text-center text-muted-foreground">
          No services available at this time.
        </p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {services.map((service) => (
            <ServiceCard
              key={service.id}
              service={service}
              onBook={handleBook}
              note={registerNote}
              onEdit={editingOn ? () => setForm({ mode: "edit", service }) : undefined}
              onDelete={editingOn ? () => setDeleting(service) : undefined}
            />
          ))}
        </div>
      )}

      {form && (
        <ServiceForm
          key={form.mode === "edit" ? form.service.id : "create"}
          initial={form.mode === "edit" ? form.service : undefined}
          onClose={() => setForm(null)}
          onSubmit={handleSubmit}
          onUploadImage={form.mode === "edit" ? handleUploadImage : undefined}
          uploading={uploadImage.isPending}
          saving={form.mode === "create" ? create.isPending : update.isPending}
        />
      )}

      <ConfirmDialog
        open={deleting != null}
        title="Delete service"
        description={
          deleteError ||
          (deleting ? `Delete "${deleting.name}"? This can't be undone.` : undefined)
        }
        onConfirm={handleDelete}
        onCancel={() => {
          setDeleting(null)
          setDeleteError("")
        }}
      />
    </section>
  )
}
