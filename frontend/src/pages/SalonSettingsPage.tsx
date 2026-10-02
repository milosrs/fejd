import { useState } from "react"
import { NavLink, Outlet, Navigate } from "react-router"
import { Pencil, Trash2, Settings, Globe } from "lucide-react"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useI18n } from "../lib/i18n"
import { salonPath } from "../lib/salonDomain"
import { useCanWrite } from "../hooks/useCanWrite"
import { publishSalon } from "../lib/api"
import { SalonLocationForm } from "../components/SalonLocationForm"
import { RenameSalonDialog } from "../components/RenameSalonDialog"
import { DeleteSalonDialog } from "../components/DeleteSalonDialog"
import { Button } from "../components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"

const publishMessageStyles: Record<"success" | "warning" | "error", string> = {
  success:
    "rounded-xl border border-green-500/40 bg-green-500/10 p-3 text-sm text-green-600 dark:text-green-400",
  warning:
    "rounded-xl border border-orange-500/40 bg-orange-500/15 p-3 text-sm text-orange-600 dark:text-orange-400",
  error: "rounded-xl border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive",
}

export function SalonSettingsPage() {
  const { slug, salon } = useSalonContext()
  const isOwner = useIsOwner()
  const { t } = useI18n()

  if (isOwner === undefined || !salon) {
    return (
      <div className="space-y-4">
        <div className="h-6 w-32 animate-pulse rounded bg-muted" />
        <div className="h-9 w-64 animate-pulse rounded-2xl bg-muted" />
        <div className="h-40 w-full animate-pulse rounded-2xl bg-muted" />
      </div>
    )
  }

  if (!isOwner) {
    return <p className="text-muted-foreground">{t("policy.notAuthorized")}</p>
  }

  const menu = [
    { to: salonPath(slug, "/settings/general"), label: t("settings.general") },
    { to: salonPath(slug, "/settings/policy"), label: t("salon.policy") },
  ]

  return (
    <div className="grid gap-6 md:grid-cols-[200px_minmax(0,1fr)]">
      <aside>
        <div className="mb-2 flex items-center gap-2 px-2 text-sm font-semibold text-foreground">
          <Settings className="size-4" />
          {t("salon.settings")}
        </div>
        <nav className="flex flex-col gap-1">
          {menu.map((m) => (
            <NavLink
              key={m.to}
              to={m.to}
              className={({ isActive }) =>
                `rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                }`
              }
            >
              {m.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="min-w-0">
        <Outlet />
      </div>
    </div>
  )
}

export function SalonGeneralSettings() {
  const { slug, salon } = useSalonContext()
  const { t } = useI18n()
  const canWrite = useCanWrite()
  const [renameOpen, setRenameOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [publishing, setPublishing] = useState(false)
  const [publishMessage, setPublishMessage] = useState<{
    kind: "success" | "warning" | "error"
    text: string
  } | null>(null)

  if (!salon) return null

  const handlePublish = async () => {
    setPublishing(true)
    setPublishMessage(null)
    try {
      const res = await publishSalon(salon.business.id)
      setPublishMessage(
        res.status === "no_changes"
          ? { kind: "warning", text: t("publish.noChanges") }
          : { kind: "success", text: t("publish.success") },
      )
    } catch {
      setPublishMessage({ kind: "error", text: t("publish.error") })
    } finally {
      setPublishing(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold text-foreground">{t("settings.general")}</h1>
        <p className="text-sm text-muted-foreground">{t("settings.generalHelp")}</p>
      </div>

      <SalonLocationForm
        businessId={salon.business.id}
        slug={slug}
        initial={{
          address_line: salon.business.address_line,
          city: salon.business.city,
          postal_code: salon.business.postal_code,
          country: salon.business.country,
          phone: salon.business.phone,
        }}
      />

      <Card>
        <CardHeader>
          <CardTitle>{t("settings.global")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" onClick={() => setRenameOpen(true)} isDisabled={!canWrite}>
              <Pencil /> {t("salon.rename")}
            </Button>
            <Button variant="destructive" onClick={() => setDeleteOpen(true)} isDisabled={!canWrite}>
              <Trash2 /> {t("salon.delete")}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("publish.title")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">{t("publish.help")}</p>
          <Button onClick={handlePublish} isDisabled={publishing || !canWrite}>
            <Globe /> {publishing ? t("common.saving") : t("publish.button")}
          </Button>
          {publishMessage && (
            <p className={publishMessageStyles[publishMessage.kind]}>
              {publishMessage.text}
            </p>
          )}
        </CardContent>
      </Card>

      {renameOpen && (
        <RenameSalonDialog
          businessId={salon.business.id}
          slug={slug}
          currentName={salon.business.name}
          onClose={() => setRenameOpen(false)}
        />
      )}

      {deleteOpen && (
        <DeleteSalonDialog
          businessId={salon.business.id}
          slug={slug}
          salonName={salon.business.name}
          onClose={() => setDeleteOpen(false)}
        />
      )}
    </div>
  )
}

export function SalonSettingsIndex() {
  return <Navigate to="general" replace />
}
