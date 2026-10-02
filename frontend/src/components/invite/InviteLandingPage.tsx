import { useEffect } from "react"
import { useParams } from "react-router"
import { useAuthStore } from "../../stores/authStore"
import { useInvitation } from "../../hooks/useInvitations"
import { useI18n } from "../../lib/i18n"
import { Button } from "../../components/ui/button"

export function InviteLandingPage() {
  const { token = "" } = useParams<{ token: string }>()
  const setToken = useAuthStore((s) => s.setPendingInviteToken)
  const authenticated = useAuthStore((s) => s.authenticated)
  const initialized = useAuthStore((s) => s.initialized)
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const { data: invitation, isLoading, isError } = useInvitation(token || null)
  const { t } = useI18n()

  useEffect(() => {
    if (token) setToken(token)
  }, [token, setToken])

  // Map the invitation's DB role to the Keycloak realm role used by the
  // register-page dropdown. realm-admin invites have no dropdown equivalent and
  // fall back to a normal (selectable) registration.
  const inviteRole = invitation?.role
  const role =
    inviteRole === "customer" ? "Customer"
    : inviteRole === "employee" ? "Employee"
    : inviteRole === "owner" ? "Owner"
    : undefined

  const playStore = import.meta.env.VITE_PLAY_STORE_URL
  const appStore = import.meta.env.VITE_APP_STORE_URL

  return (
    <div className="min-h-app flex flex-col items-center justify-center gap-4 bg-background p-8">
      {isLoading ? (
        <p className="text-muted-foreground">{t("invite.landing.loading")}</p>
      ) : isError ? (
        <>
          <h1 className="text-xl font-semibold text-foreground">{t("invite.landing.unavailable")}</h1>
          <p className="text-muted-foreground text-center max-w-sm">
            {t("invite.landing.invalid")}
          </p>
        </>
      ) : !initialized ? (
        <>
          <div className="h-6 w-64 animate-pulse rounded bg-muted" />
          <div className="h-4 w-80 animate-pulse rounded bg-muted" />
          <div className="flex gap-3">
            <div className="h-9 w-20 animate-pulse rounded-2xl bg-muted" />
            <div className="h-9 w-24 animate-pulse rounded-2xl bg-muted" />
          </div>
        </>
      ) : authenticated ? (
        <>
          <h1 className="text-xl font-semibold text-foreground text-center">
            {invitation?.salon_name
              ? t("invite.landing.joining", { salon: invitation.salon_name })
              : t("invite.landing.joiningPlatform")}
          </h1>
          <p className="text-muted-foreground">{t("invite.landing.redirect")}</p>
        </>
      ) : (
        <>
          <h1 className="text-xl font-semibold text-foreground text-center">
            {invitation?.salon_name
              ? t("invite.landing.invitedTo", { salon: invitation.salon_name })
              : t("invite.landing.invitedToPlatform")}
          </h1>
          <p className="text-muted-foreground text-center max-w-sm">
            {t("invite.landing.joinPrompt")}
          </p>
          <div className="flex gap-3">
            <Button onClick={() => register(role)}>{t("common.register")}</Button>
            <Button variant="outline" onClick={login}>
              {t("common.logIn")}
            </Button>
          </div>

          {(playStore || appStore) && (
            <div className="mt-4 flex flex-col items-center gap-2">
              <p className="text-xs text-muted-foreground">{t("invite.landing.noApp")}</p>
              <div className="flex gap-3">
                {playStore && (
                  <a href={playStore} className="text-sm underline text-foreground">
                    Google Play
                  </a>
                )}
                {appStore && (
                  <a href={appStore} className="text-sm underline text-foreground">
                    App Store
                  </a>
                )}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
