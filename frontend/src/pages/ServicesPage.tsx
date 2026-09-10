import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useServices, useAdminEmployees, useServiceEmployeesAdmin, type Service } from "../hooks/useApi"
import { useServiceMutations } from "../hooks/useServiceMutations"
import { useBookingStore } from "../stores/bookingStore"
import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"
import { salonPath } from "../lib/salonDomain"
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
  const { create, update, remove, uploadImage, setEmployees } = useServiceMutations(businessId, slug)
  const { data: staffData } = useAdminEmployees(businessId)

  const [form, setForm] = useState<FormState>(null)
  const [deleting, setDeleting] = useState<Service | null>(null)
  const [deleteError, setDeleteError] = useState("")
  const [formError, setFormError] = useState("")

  const editingServiceId = form?.mode === "edit" ? form.service.id : ""
  const { data: assignedEmployees } = useServiceEmployeesAdmin(businessId, editingServiceId)

  const staff = (staffData ?? []).filter((e) => e.active)
  const assignedEmployeeIds = (assignedEmployees ?? []).map((e) => e.id)

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
    navigate(`${salonPath(slug, "/book")}?service=${serviceId}`)
  }

  const handleSubmit = async (values: ServiceFormValues) => {
    if (!form) return
    const imageFile = values.imageFile ?? null
    setFormError("")

    try {
      let serviceId = form.mode === "edit" ? form.service.id : ""

      if (form.mode === "create") {
        const service = await create.mutateAsync({
          name: values.name,
          description: values.description,
          duration_minutes: values.duration_minutes,
          price: values.price,
          active: true,
        })
        if (!service) return
        serviceId = service.id
      } else {
        await update.mutateAsync({
          serviceId,
          input: {
            name: values.name,
            description: values.description,
            duration_minutes: values.duration_minutes,
            price: values.price,
            active: form.service.active,
          },
        })
      }

      if (imageFile) {
        await uploadImage.mutateAsync({ serviceId, file: imageFile })
      }

      await setEmployees.mutateAsync({
        serviceId,
        businessUserIds: values.employee_ids,
      })

      setForm(null)
    } catch (err) {
      const e = err as unknown as { body?: { error?: string } }
      setFormError(e?.body?.error ?? "Failed to save service.")
    }
  }

  const openForm = (state: FormState) => {
    setFormError("")
    setForm(state)
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
          <Button size="sm" onClick={() => openForm({ mode: "create" })}>
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
              onEdit={editingOn ? () => openForm({ mode: "edit", service }) : undefined}
              onDelete={editingOn ? () => setDeleting(service) : undefined}
            />
          ))}
        </div>
      )}

      {form && (
        <ServiceForm
          key={form.mode === "edit" ? form.service.id : "create"}
          initial={form.mode === "edit" ? form.service : undefined}
          staff={staff}
          assignedEmployeeIds={assignedEmployeeIds}
          onClose={() => setForm(null)}
          onSubmit={handleSubmit}
          saving={
            create.isPending ||
            update.isPending ||
            uploadImage.isPending ||
            setEmployees.isPending
          }
          error={formError}
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
