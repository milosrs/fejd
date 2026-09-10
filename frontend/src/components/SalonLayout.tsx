import { Link, NavLink, Outlet } from "react-router-dom"
import { useState } from "react"
import { SalonProvider, useSalonContext, useIsOwner } from "../context/SalonContext"
import { Button } from "./ui/button"
import { SalonPolicyDialog } from "./SalonPolicyDialog"
import { salonPath } from "../lib/salonDomain"
import { useHeaderHeightMeasure } from "../hooks/useHeaderHeightMeasure"
import { useSections } from "../hooks/useSections"
import { useI18n } from "../lib/i18n"
import type { HeroContent } from "../lib/sections"

function SalonShell() {
  const { slug, salon, isLoading, editing, setEditing } = useSalonContext()
  const isOwner = useIsOwner()
  const { data: sections } = useSections(slug)
  const { pickLocalized } = useI18n()
  const [policyOpen, setPolicyOpen] = useState(false)
  const salonHeaderRef = useHeaderHeightMeasure("--salon-header-height")

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!salon) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Salon not found</p>
      </div>
    )
  }

  const nav = [
    { to: salonPath(slug), label: "Home", end: true },
    { to: salonPath(slug, "/services"), label: "Services", end: false },
    { to: salonPath(slug, "/barbers"), label: "Barbers", end: false },
    { to: salonPath(slug, "/book"), label: "Book", end: false },
  ]

  const heroContent = pickLocalized<HeroContent>(
    sections?.find((s) => s.type === "hero")?.content as
      | Record<string, HeroContent>
      | undefined,
  )
  const title = heroContent?.headline || salon.business.name

  return (
    <div className="min-h-screen bg-background pb-[env(safe-area-inset-bottom)]">
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
                Salon policy
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setEditing(!editing)}
              >
                {editing ? "Done" : "Edit"}
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
