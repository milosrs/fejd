import { Link, NavLink, Outlet } from "react-router-dom"
import { useState } from "react"
import { SalonProvider, useSalonContext, useIsOwner } from "../context/SalonContext"
import { Button } from "./ui/button"
import { SalonPolicyDialog } from "./SalonPolicyDialog"
import { RenameSalonDialog } from "./RenameSalonDialog"
import { salonPath } from "../lib/salonDomain"
import { useHeaderHeightMeasure } from "../hooks/useHeaderHeightMeasure"
import { useSections } from "../hooks/useSections"
import { useI18n } from "../lib/i18n"
import type { HeroContent } from "../lib/sections"
import { Loader } from "./Loader"

function SalonShell() {
  const { slug, salon, isLoading, editing, setEditing } = useSalonContext()
  const isOwner = useIsOwner()
  const { data: sections } = useSections(slug)
  const { pickLocalized, t } = useI18n()
  const [policyOpen, setPolicyOpen] = useState(false)
  const [renameOpen, setRenameOpen] = useState(false)
  const salonHeaderRef = useHeaderHeightMeasure("--salon-header-height")

  if (isLoading) {
    return <Loader />
  }

  if (!salon) {
    return (
      <div className="min-h-app flex items-center justify-center bg-background">
        <p className="text-muted-foreground">{t("salon.notFound")}</p>
      </div>
    )
  }

  const nav = [
    { to: salonPath(slug), label: t("nav.home"), end: true },
    { to: salonPath(slug, "/services"), label: t("nav.services"), end: false },
    { to: salonPath(slug, "/barbers"), label: t("nav.barbers"), end: false },
    { to: salonPath(slug, "/book"), label: t("nav.book"), end: false },
  ]

  const heroContent = pickLocalized<HeroContent>(
    sections?.find((s) => s.type === "hero")?.content as
      | Record<string, HeroContent>
      | undefined,
  )
  const title = heroContent?.headline || salon.business.name

  return (
    <div className="min-h-app bg-background pb-[env(safe-area-inset-bottom)]">
      <header ref={salonHeaderRef} className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex flex-wrap items-center justify-between gap-2">
          <Link to={salonPath(slug)} className="text-lg font-semibold text-foreground">
            {title}
          </Link>
          <nav className="flex flex-wrap items-center gap-1">
            {nav.map((n) => (
              <NavLink
                key={n.to}
                to={n.to}
                end={n.end}
                className={({ isActive }) =>
                  `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                    isActive
                      ? "bg-primary text-primary-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`
                }
              >
                {n.label}
              </NavLink>
            ))}
          </nav>
          {isOwner && (
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPolicyOpen(true)}
              >
                {t("salon.policy")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setRenameOpen(true)}
              >
                {t("salon.rename")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setEditing(!editing)}
              >
                {editing ? t("common.done") : t("common.edit")}
              </Button>
            </div>
          )}
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <Outlet />
      </main>

      {policyOpen && (
        <SalonPolicyDialog
          businessId={salon.business.id}
          slug={slug}
          onClose={() => setPolicyOpen(false)}
        />
      )}

      {renameOpen && (
        <RenameSalonDialog
          businessId={salon.business.id}
          slug={slug}
          currentName={salon.business.name}
          onClose={() => setRenameOpen(false)}
        />
      )}
    </div>
  )
}

export function SalonLayout({ initialEditing, slug }: { initialEditing?: boolean; slug?: string }) {
  return (
    <SalonProvider initialEditing={initialEditing} slug={slug}>
      <SalonShell />
    </SalonProvider>
  )
}
