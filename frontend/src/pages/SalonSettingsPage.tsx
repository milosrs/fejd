import { useState } from "react"
import { NavLink, Outlet, Navigate } from "react-router"
import { Pencil, Trash2, Settings } from "lucide-react"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useI18n } from "../lib/i18n"
import { salonPath } from "../lib/salonDomain"
import { useCanWrite } from "../hooks/useCanWrite"
import { SalonLocationForm } from "../components/SalonLocationForm"
import { RenameSalonDialog } from "../components/RenameSalonDialog"
import { DeleteSalonDialog } from "../components/DeleteSalonDialog"
import { Button } from "../components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"

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

  if (!salon) return null

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
