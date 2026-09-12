import { useEffect } from "react"
import { useParams } from "react-router-dom"
import { useAuthStore } from "../../stores/authStore"
import { useInvitation } from "../../hooks/useInvitations"
import { useI18n } from "../../lib/i18n"
import { Button } from "../../components/ui/button"

export function InviteLandingPage() {
  const { token = "" } = useParams<{ token: string }>()
  const setToken = useAuthStore((s) => s.setPendingInviteToken)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const { data: invitation, isLoading, isError } = useInvitation(token || null)
  const { t } = useI18n()

  useEffect(() => {
    if (token) setToken(token)
  }, [token, setToken])

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
      ) : authenticated ? (
        <>
          <h1 className="text-xl font-semibold text-foreground text-center">
            {t("invite.landing.joining", { salon: invitation?.salon_name ?? "" })}
          </h1>
          <p className="text-muted-foreground">{t("invite.landing.redirect")}</p>
        </>
      ) : (
        <>
          <h1 className="text-xl font-semibold text-foreground text-center">
            {t("invite.landing.invitedTo", { salon: invitation?.salon_name ?? "" })}
          </h1>
          <p className="text-muted-foreground text-center max-w-sm">
            {t("invite.landing.joinPrompt")}
          </p>
          <div className="flex gap-3">
            <Button onClick={register}>{t("common.register")}</Button>
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
