import { Link } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"

export function HomePage() {
  const authenticated = useAuthStore((s) => s.authenticated)

  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-background p-8">
      <h1 className="text-2xl font-bold text-foreground">fejd</h1>
      <p className="text-muted-foreground text-center max-w-sm">
        Book haircut appointments. Open a salon by its link, or manage your
        appointments below.
      </p>
      {authenticated && (
        <Link to="/my/appointments" className="text-sm underline text-foreground">
          My appointments
        </Link>
      )}
    </div>
  )
}
