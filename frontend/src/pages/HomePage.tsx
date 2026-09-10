import { useNavigate } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"
import { useMe } from "../hooks/useMe"
import { hasRole } from "../lib/ownership"
import { openSalon, salonUrl } from "../lib/salonDomain"
import { Button } from "../components/ui/button"
import { CreateSalonForm } from "../components/CreateSalonForm"

export function HomePage() {
  const authenticated = useAuthStore((s) => s.authenticated)
  const roles = useAuthStore((s) => s.roles)
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const navigate = useNavigate()
  const { data: me, isLoading: meLoading } = useMe()

  const isOwner = hasRole(roles, "Owner")
  const hasSalon = me?.has_salon ?? false
  const businesses = me?.businesses ?? []
  const primaryBusiness =
    businesses.find((b) => b.role === "admin") ?? businesses[0]

  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-background p-8">
      <h1 className="text-2xl font-bold text-foreground">fejd</h1>
      <p className="text-muted-foreground text-center max-w-sm">
        Book haircut appointments. Open a salon by its link, or manage your
        appointments below.
      </p>

      {!authenticated ? (
        <div className="flex gap-3">
          <Button onClick={register}>Register</Button>
          <Button variant="outline" onClick={login}>
            Log in
          </Button>
        </div>
      ) : meLoading ? null : isOwner && !hasSalon ? (
        <CreateSalonForm />
      ) : businesses.length > 0 ? (
        <div className="flex w-full max-w-sm flex-col items-center gap-4">
          <a
            href={salonUrl(primaryBusiness.slug)}
            onClick={(e) => {
              e.preventDefault()
              openSalon(navigate, primaryBusiness.slug)
            }}
            className="inline-flex items-center justify-center rounded-2xl bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/80"
          >
            Open my salon
          </a>

          <div className="flex w-full flex-col items-center gap-3">
            <h2 className="text-sm font-medium text-muted-foreground">Your salons</h2>
            {businesses.map((b) => (
              <a
                key={b.id}
                href={salonUrl(b.slug)}
                onClick={(e) => {
                  e.preventDefault()
                  openSalon(navigate, b.slug)
                }}
                className="flex w-full items-center justify-between rounded-xl border border-border bg-card px-4 py-3 text-sm text-foreground hover:bg-muted"
              >
                <span>{b.name}</span>
                <span className="text-xs text-muted-foreground">
                  {b.role === "admin" ? "Owner" : "Employee"}
                </span>
              </a>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  )
}
