import { Link, NavLink, Outlet, useNavigate } from "react-router"
import { useState } from "react"
import { Menu as MenuIcon } from "lucide-react"
import { SalonProvider, useSalonContext, useIsOwner } from "../context/SalonContext"
import { Button } from "./ui/button"
import { salonPath } from "../lib/salonDomain"
import { useHeaderHeightMeasure } from "../hooks/useHeaderHeightMeasure"
import { useSections } from "../hooks/useSections"
import { useI18n } from "../lib/i18n"
import type { HeroContent } from "../lib/sections"
import { Loader } from "./Loader"
import { SideDrawer } from "./ui/drawer"
import { NotFoundPage } from "../pages/NotFoundPage"

const ownerActionClassName =
  "w-full rounded-lg px-3 py-2.5 text-left text-sm font-medium text-foreground hover:bg-muted"

function SalonShell() {
  const { slug, salon, isLoading, editing, setEditing } = useSalonContext()
  const isOwner = useIsOwner()
  const navigate = useNavigate()
  const { data: sections } = useSections(slug)
  const { pickLocalized, t } = useI18n()
  const [menuOpen, setMenuOpen] = useState(false)
  const salonHeaderRef = useHeaderHeightMeasure("--salon-header-height")

  if (isLoading) {
    return <Loader />
  }

  if (!salon) {
    return <NotFoundPage />
  }

  const nav = [
    { to: salonPath(slug), label: t("nav.home"), end: true },
    { to: salonPath(slug, "/services"), label: t("nav.services"), end: false },
    { to: salonPath(slug, "/barbers"), label: t("nav.barbers"), end: false },
  ]

  const heroContent = pickLocalized<HeroContent>(
    sections?.find((s) => s.type === "hero")?.content as
      | Record<string, HeroContent>
      | undefined,
  )
  const title = heroContent?.headline || salon.business.name

  return (
    <div className="min-h-app bg-background pb-[env(safe-area-inset-bottom)]">
      <header
        ref={salonHeaderRef}
        className="sticky top-[var(--top-header-height,0px)] z-30 border-b border-border bg-background"
      >
        <div className="max-w-4xl mx-auto px-4 py-3 flex items-center justify-between gap-2">
          <div className="flex min-w-0 items-center gap-2">
            <button
              type="button"
              onClick={() => setMenuOpen(true)}
              aria-label={t("app.menu")}
              className="shrink-0 rounded-md p-1.5 text-foreground hover:bg-muted md:hidden"
            >
              <MenuIcon className="size-5" />
            </button>
            <Link
              to={salonPath(slug)}
              className="truncate text-lg font-semibold text-foreground"
            >
              {title}
            </Link>
          </div>
          <nav className="hidden md:flex items-center gap-1">
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
          {isOwner === undefined ? (
            <div className="hidden md:flex items-center gap-2">
              <div className="h-8 w-20 animate-pulse rounded-2xl bg-muted" />
              <div className="h-8 w-20 animate-pulse rounded-2xl bg-muted" />
              <div className="h-8 w-20 animate-pulse rounded-2xl bg-muted" />
            </div>
          ) : isOwner ? (
            <div className="hidden md:flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate(salonPath(slug, "/settings"))}
              >
                {t("salon.settings")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setEditing(!editing)}
              >
                {editing ? t("common.done") : t("common.edit")}
              </Button>
            </div>
          ) : null}
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <Outlet />
      </main>

      <SideDrawer open={menuOpen} onClose={() => setMenuOpen(false)} title={title}>
        <nav className="flex flex-col gap-1 p-2">
          {nav.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              end={n.end}
              onClick={() => setMenuOpen(false)}
              className={({ isActive }) =>
                `rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                }`
              }
            >
              {n.label}
            </NavLink>
          ))}
        </nav>
        {isOwner === undefined ? (
          <div className="mt-2 flex flex-col gap-1 border-t border-border p-2">
            <div className="h-9 w-full animate-pulse rounded-lg bg-muted" />
            <div className="h-9 w-full animate-pulse rounded-lg bg-muted" />
            <div className="h-9 w-full animate-pulse rounded-lg bg-muted" />
            <div className="h-9 w-full animate-pulse rounded-lg bg-muted" />
          </div>
        ) : isOwner ? (
          <div className="mt-2 flex flex-col gap-1 border-t border-border p-2">
            <button
              type="button"
              className={ownerActionClassName}
              onClick={() => {
                setMenuOpen(false)
                navigate(salonPath(slug, "/settings"))
              }}
            >
              {t("salon.settings")}
            </button>
            <button
              type="button"
              className={ownerActionClassName}
              onClick={() => {
                setMenuOpen(false)
                setEditing(!editing)
              }}
            >
              {editing ? t("common.done") : t("common.edit")}
            </button>
          </div>
        ) : null}
      </SideDrawer>
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
