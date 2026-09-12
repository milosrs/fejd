import { useEffect } from "react"
import { useParams } from "react-router-dom"
import { useAuthStore } from "../../stores/authStore"
import { useInvitation } from "../../hooks/useInvitations"
import { Button } from "../../components/ui/button"

export function InviteLandingPage() {
  const { token = "" } = useParams<{ token: string }>()
  const setToken = useAuthStore((s) => s.setPendingInviteToken)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const { data: invitation, isLoading, isError } = useInvitation(token || null)

  useEffect(() => {
    if (token) setToken(token)
  }, [token, setToken])

  const playStore = import.meta.env.VITE_PLAY_STORE_URL
  const appStore = import.meta.env.VITE_APP_STORE_URL

  return (
    <div className="min-h-app flex flex-col items-center justify-center gap-4 bg-background p-8">
      {isLoading ? (
        <p className="text-muted-foreground">Loading invitation…</p>
      ) : isError ? (
        <>
          <h1 className="text-xl font-semibold text-foreground">Invite unavailable</h1>
          <p className="text-muted-foreground text-center max-w-sm">
            This invite link is invalid, has already been used, or has expired.
          </p>
        </>
      ) : authenticated ? (
        <>
          <h1 className="text-xl font-semibold text-foreground text-center">
            Joining {invitation?.salon_name}…
          </h1>
          <p className="text-muted-foreground">You'll be redirected shortly.</p>
        </>
      ) : (
        <>
          <h1 className="text-xl font-semibold text-foreground text-center">
            You've been invited to {invitation?.salon_name}
          </h1>
          <p className="text-muted-foreground text-center max-w-sm">
            Register or log in to join this salon.
          </p>
          <div className="flex gap-3">
            <Button onClick={register}>Register</Button>
            <Button variant="outline" onClick={login}>
              Log in
            </Button>
          </div>

          {(playStore || appStore) && (
            <div className="mt-4 flex flex-col items-center gap-2">
              <p className="text-xs text-muted-foreground">Don't have the app?</p>
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
