import { useNavigate } from "react-router"
import { useAuthStore } from "../stores/authStore"
import { useInvitedCustomersReport } from "../hooks/useApi"
import { resolveImageUrl } from "../lib/images"
import { useI18n } from "../lib/i18n"
import { Button } from "../components/ui/button"

function InviteeAvatar({ src, name }: { src?: string; name: string }) {
  const url = resolveImageUrl(src)
  const fallback = (name || "?").trim().charAt(0).toUpperCase()
  if (url) {
    return (
      <img
        src={url}
        alt=""
        className="size-10 shrink-0 rounded-full border border-border bg-background object-cover"
      />
    )
  }
  return (
    <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-muted text-base font-semibold text-foreground">
      {fallback}
    </span>
  )
}

function UserRow({ name, avatar }: { name?: string; avatar?: string }) {
  return (
    <li className="flex items-center gap-3 rounded-xl border border-border bg-card p-3">
      <InviteeAvatar src={avatar} name={name ?? ""} />
      <span className="min-w-0 truncate text-sm font-medium text-foreground">{name}</span>
    </li>
  )
}

export function InvitedCustomersPage() {
  const navigate = useNavigate()
  const isRealmAdmin = useAuthStore((s) => s.isRealmAdmin)
  const { data: report, isLoading } = useInvitedCustomersReport()
  const { t } = useI18n()

  if (!isRealmAdmin) {
    return (
      <div className="min-h-app flex items-center justify-center bg-background p-8">
        <p className="text-muted-foreground">{t("invitedCustomers.notAuthorized")}</p>
      </div>
    )
  }

  const salons = report?.salons ?? []
  const selfRegistered = report?.self_registered ?? []
  const isEmpty = salons.length === 0 && selfRegistered.length === 0

  return (
    <div className="min-h-app bg-background">
      <header className="border-b border-border">
        <div className="mx-auto flex max-w-4xl items-center gap-4 px-4 py-4">
          <h1 className="text-xl font-semibold text-foreground">{t("invitedCustomers.title")}</h1>
          <Button variant="outline" size="sm" onClick={() => navigate("/")}>
            {t("common.back")}
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-4xl space-y-8 px-4 py-8">
        {isLoading ? (
          <p className="text-muted-foreground">{t("common.loading")}</p>
        ) : isEmpty ? (
          <p className="text-muted-foreground">{t("invitedCustomers.empty")}</p>
        ) : (
          <>
            {salons.map((salon) => (
              <section key={salon.business_id} className="space-y-3">
                <div className="flex items-center gap-3">
                  <InviteeAvatar src={salon.logo} name={salon.name} />
                  <h2 className="text-lg font-semibold text-foreground">{salon.name}</h2>
                </div>
                <ul className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {salon.customers.map((customer) => (
                    <UserRow
                      key={customer.user_id}
                      name={customer.display_name}
                      avatar={customer.avatar}
                    />
                  ))}
                </ul>
              </section>
            ))}

            {selfRegistered.length > 0 && (
              <section className="space-y-3">
                <h2 className="text-lg font-semibold text-foreground">
                  {t("invitedCustomers.ownRegistration")}
                </h2>
                <ul className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {selfRegistered.map((user) => (
                    <UserRow
                      key={user.user_id}
                      name={user.display_name}
                      avatar={user.avatar}
                    />
                  ))}
                </ul>
              </section>
            )}
          </>
        )}
      </main>
    </div>
  )
}
