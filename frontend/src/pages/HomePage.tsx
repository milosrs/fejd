import { useAuthStore } from "../stores/authStore"
import { useMe } from "../hooks/useMe"
import { useBusinesses } from "../hooks/useApi"
import { hasRole } from "../lib/ownership"
import { useI18n } from "../lib/i18n"
import { CreateSalonForm } from "../components/CreateSalonForm"
import { SalonCard } from "../components/SalonCard"

export function HomePage() {
  const authenticated = useAuthStore((s) => s.authenticated)
  const roles = useAuthStore((s) => s.roles)
  const { data: me, isLoading: meLoading } = useMe()
  const { data: directory, isLoading: directoryLoading } = useBusinesses()
  const { t } = useI18n()

  const isOwner = hasRole(roles, "Owner")
  const hasSalon = me?.has_salon ?? false
  const businesses = me?.businesses ?? []
  const visibleBusinesses = isOwner
    ? businesses.filter((b) => b.role === "admin")
    : businesses

  return (
    <div className="min-h-app flex flex-col items-center justify-center gap-4 bg-background p-8">
      <img src="/logo-white.jpg" alt="fejd" className="h-36 w-auto dark:hidden" />
      <img src="/logo_dark.jpg" alt="fejd" className="hidden h-36 w-auto dark:block" />
      <p className="text-muted-foreground text-center max-w-sm">
        {isOwner
          ? t("home.owner.label")
          : "Book haircut appointments. Open a salon by its link, or manage your appointments below."}
      </p>

      {!authenticated ? (
        directoryLoading ? null : (
          <div className="grid w-full max-w-4xl grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {(directory ?? []).map((b) => (
              <SalonCard key={b.id} business={b} isOwner={false} />
            ))}
          </div>
        )
      ) : meLoading ? null : isOwner && !hasSalon ? (
        <CreateSalonForm />
      ) : visibleBusinesses.length > 0 ? (
        <div className="grid w-full max-w-4xl grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {visibleBusinesses.map((b) => (
            <SalonCard key={b.id} business={b} isOwner={b.role === "admin"} />
          ))}
        </div>
      ) : null}
    </div>
  )
}
