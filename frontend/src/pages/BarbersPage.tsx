import { useState } from "react"
import { useSalonContext, useIsOwner, useIsMember } from "../context/SalonContext"
import { useEmployees, useServices, type Employee } from "../hooks/useApi"
import { useBarberMutations, type EmployeeRemoval } from "../hooks/useBarberMutations"
import { useInvitations } from "../hooks/useInvitations"
import { BarberCard, BarberCardSkeleton } from "../components/barbers/BarberCard"
import { BarberForm, type BarberFormValues } from "../components/barbers/BarberForm"
import { InviteDialog } from "../components/barbers/InviteDialog"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { Button } from "../components/ui/button"
import { Plus, QrCode } from "lucide-react"
import { useI18n } from "../lib/i18n"

type FormState = { mode: "create" } | { mode: "edit"; employee: Employee } | null

function removalSummary(
  result: EmployeeRemoval,
  t: (key: string, params?: Record<string, string | number>) => string,
): string {
  const parts: string[] = []
  if (result.reassigned > 0) {
    parts.push(
      t("barbers.reassigned", {
        count: result.reassigned,
        plural: result.reassigned === 1 ? "" : "s",
      }),
    )
  }
  if (result.cancelled > 0) {
    parts.push(t("barbers.cancelled", { count: result.cancelled }))
  }
  return parts.length > 0 ? `${parts.join(", ")}.` : t("barbers.removed")
}

export function BarbersPage() {
  const { slug, salon, editing } = useSalonContext()
  const isOwner = useIsOwner()
  const isMember = useIsMember()
  const { data, isLoading, isError } = useEmployees(slug)
  const { data: servicesData } = useServices(slug)
  const { t } = useI18n()

  const businessId = salon?.business.id ?? ""
  const { invite, remove, uploadAvatar } = useBarberMutations(businessId, slug)
  const { invite: inviteLink } = useInvitations(businessId)

  const [form, setForm] = useState<FormState>(null)
  const [pendingRemoval, setPendingRemoval] = useState<Employee | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [notice, setNotice] = useState("")

  const editingOn = editing && isOwner
  const employees = data ?? []
  const services = (servicesData ?? []).filter((s) => s.active)

  const handleSubmit = (values: BarberFormValues) => {
    if (form?.mode !== "create") return
    invite.mutate(values, {
      onSuccess: (employee) => {
        if (employee) setForm({ mode: "edit", employee })
      },
    })
  }

  const handleInviteClick = () => {
    setInviteOpen(true)
    inviteLink.mutate(undefined, {
      onError: () => {
        setInviteOpen(false)
        setNotice(t("barbers.inviteError"))
      },
    })
  }

  const handleUploadAvatar = (file: File) => {
    if (form?.mode === "edit") {
      uploadAvatar.mutate({ userId: form.employee.user_id, file })
    }
  }

  const handleRemove = () => {
    if (!pendingRemoval) return
    remove.mutate(pendingRemoval.user_id, {
      onSuccess: (result) => {
        setNotice(result ? removalSummary(result, t) : t("barbers.removed"))
        setPendingRemoval(null)
      },
      onError: () => {
        setPendingRemoval(null)
      },
    })
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">{t("barbers.title")}</h2>
        <div className="flex items-center gap-2">
          {isMember && (
            <Button size="sm" variant="outline" onClick={handleInviteClick}>
              <QrCode className="size-3" /> {t("barbers.invite")}
            </Button>
          )}
          {editingOn && (
            <Button size="sm" onClick={() => setForm({ mode: "create" })}>
              <Plus className="size-3" /> {t("barbers.add")}
            </Button>
          )}
        </div>
      </div>

      {notice && (
        <p className="rounded-xl border border-border bg-muted p-3 text-sm text-foreground">
          {notice}
        </p>
      )}

      {isLoading ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <BarberCardSkeleton key={i} />
          ))}
        </div>
      ) : isError ? (
        <p className="py-12 text-center text-muted-foreground">
          {t("barbers.loadError")}
        </p>
      ) : employees.length === 0 ? (
        <p className="py-12 text-center text-muted-foreground">
          {t("barbers.empty")}
        </p>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
          {employees.map((employee) => (
            <BarberCard
              key={employee.id}
              employee={employee}
              onRemove={
                editingOn && employee.role !== "admin"
                  ? () => setPendingRemoval(employee)
                  : undefined
              }
            />
          ))}
        </div>
      )}

      {form && (
        <BarberForm
          key={form.mode === "edit" ? form.employee.id : "create"}
          initial={form.mode === "edit" ? form.employee : undefined}
          services={services}
          onClose={() => setForm(null)}
          onSubmit={handleSubmit}
          onUploadAvatar={form.mode === "edit" ? handleUploadAvatar : undefined}
          uploading={uploadAvatar.isPending}
          saving={invite.isPending}
        />
      )}

      <ConfirmDialog
        open={pendingRemoval != null}
        title={t("barbers.removeTitle")}
        description={
          pendingRemoval
            ? t("barbers.removeConfirm", {
                name: pendingRemoval.display_name || pendingRemoval.user_id,
              })
            : undefined
        }
        confirmLabel={t("common.remove")}
        cancelLabel={t("common.cancel")}
        onConfirm={handleRemove}
        onCancel={() => setPendingRemoval(null)}
      />

      <InviteDialog
        open={inviteOpen}
        onClose={() => setInviteOpen(false)}
        invitation={inviteLink.data ?? null}
        loading={inviteLink.isPending}
        error={inviteLink.error ? t("barbers.inviteError") : null}
        filename={`${slug}-invite`}
      />
    </section>
  )
}
