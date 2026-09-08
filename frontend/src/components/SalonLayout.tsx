import { Link, NavLink, Outlet } from "react-router-dom"
import { useState } from "react"
import { SalonProvider, useSalonContext, useIsOwner } from "../context/SalonContext"
import { Button } from "./ui/button"
import { SalonPolicyDialog } from "./SalonPolicyDialog"

function SalonShell() {
  const { slug, salon, isLoading, editing, setEditing } = useSalonContext()
  const isOwner = useIsOwner()
  const [policyOpen, setPolicyOpen] = useState(false)

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
    { to: `/${slug}`, label: "Home", end: true },
    { to: `/${slug}/services`, label: "Services", end: false },
    { to: `/${slug}/barbers`, label: "Barbers", end: false },
    { to: `/${slug}/book`, label: "Book", end: false },
  ]

  return (
    <div className="min-h-screen bg-background pb-[env(safe-area-inset-bottom)]">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex flex-wrap items-center justify-between gap-2">
          <Link to={`/${slug}`} className="text-lg font-semibold text-foreground">
            {salon.business.name}
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
          cancellationLeadHours={salon.business.cancellation_lead_hours}
          noShowAfterHours={salon.business.no_show_after_hours}
          onClose={() => setPolicyOpen(false)}
        />
      )}
    </div>
  )
}

export function SalonLayout({ initialEditing }: { initialEditing?: boolean }) {
  return (
    <SalonProvider initialEditing={initialEditing}>
      <SalonShell />
    </SalonProvider>
  )
}
