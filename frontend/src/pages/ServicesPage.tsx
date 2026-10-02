import { useState } from "react"
import { useNavigate } from "react-router"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useServices, useAdminEmployees, useServiceEmployeesAdmin, useServiceCombinationsAdmin, type Service } from "../hooks/useApi"
import { useServiceMutations } from "../hooks/useServiceMutations"
import { useBookingStore } from "../stores/bookingStore"
import { useCanWrite } from "../hooks/useCanWrite"
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
  const canWrite = useCanWrite()
  const { t } = useI18n()

  const businessId = salon?.business.id ?? ""
  const { data, isLoading, isError } = useServices(slug)
  const { create, update, remove, uploadImage, setEmployees, setCombinations } = useServiceMutations(businessId, slug)

  const [form, setForm] = useState<FormState>(null)
  const [deleting, setDeleting] = useState<Service | null>(null)
  const [deleteError, setDeleteError] = useState("")
  const [formError, setFormError] = useState("")

  const editingOn = editing && isOwner === true

  const editingServiceId = form?.mode === "edit" ? form.service.id : ""
  const { data: staffData } = useAdminEmployees(editingOn ? businessId : "")
  const { data: assignedEmployees } = useServiceEmployeesAdmin(businessId, editingServiceId)
  const { data: assignedCombinations } = useServiceCombinationsAdmin(businessId, editingServiceId)

  const staff = (staffData ?? []).filter((e) => e.active)
  const assignedEmployeeIds = (assignedEmployees ?? []).map((e) => e.id)
  const assignedCombinableIds = (assignedCombinations ?? []).map((s) => s.id)
  const combinableOptions = (data ?? []).filter((s) => s.id !== editingServiceId)

  const services = editingOn ? (data ?? []) : (data ?? []).filter((s) => s.active)

  const handleBook = (serviceId: string) => {
    reset()
    setService(serviceId)
    const svc = (data ?? []).find((s) => s.id === serviceId)
    const target = svc?.slug ?? serviceId
    navigate(`${salonPath(slug, "/book")}/${target}`)
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
            picture_id: form.service.picture_id,
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

      await setCombinations.mutateAsync({
        serviceId,
        serviceIds: values.combinable_service_ids,
      })

      setForm(null)
    } catch (err) {
      const e = err as unknown as { body?: { error?: string } }
      setFormError(e?.body?.error ?? t("services.saveError"))
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
        setDeleteError(e?.body?.error ?? t("services.deleteError"))
      },
    })
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">{t("services.title")}</h2>
        {editingOn && (
          <Button size="sm" onClick={() => openForm({ mode: "create" })} isDisabled={!canWrite}>
            <Plus className="size-3" /> {t("services.add")}
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
          {t("services.loadError")}
        </p>
      ) : services.length === 0 ? (
        <p className="py-12 text-center text-muted-foreground">
          {t("services.empty")}
        </p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {services.map((service) => (
            <ServiceCard
              key={service.id}
              service={service}
              onBook={handleBook}
              disabled={!canWrite}
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
          services={combinableOptions}
          assignedCombinableIds={assignedCombinableIds}
          onClose={() => setForm(null)}
          onSubmit={handleSubmit}
          saving={
            create.isPending ||
            update.isPending ||
            uploadImage.isPending ||
            setEmployees.isPending ||
            setCombinations.isPending
          }
          error={formError}
          disabled={!canWrite}
        />
      )}

      <ConfirmDialog
        open={deleting != null}
        title={t("services.deleteTitle")}
        description={
          deleteError ||
          (deleting ? t("common.deleteConfirm", { name: deleting.name }) : undefined)
        }
        confirmLabel={t("common.yes")}
        cancelLabel={t("common.no")}
        confirmDisabled={!canWrite}
        onConfirm={handleDelete}
        onCancel={() => {
          setDeleting(null)
          setDeleteError("")
        }}
      />
    </section>
  )
}
