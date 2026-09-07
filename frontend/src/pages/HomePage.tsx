import { useState } from "react"
import { Link } from "react-router-dom"
import { useAuthStore } from "../stores/authStore"
import { useMe } from "../hooks/useMe"
import { hasRole } from "../lib/ownership"
import { Button } from "../components/ui/button"
import { CreateSalonForm } from "../components/CreateSalonForm"

export function HomePage() {
  const authenticated = useAuthStore((s) => s.authenticated)
  const roles = useAuthStore((s) => s.roles)
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const { data: me, isLoading: meLoading } = useMe()

  const isOwner = hasRole(roles, "Owner")
  const hasSalon = me?.has_salon ?? false
  const ownerBusiness = me?.businesses?.find((b) => b.role === "admin")

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
      ) : ownerBusiness ? (
        <div className="flex flex-col items-center gap-3">
          <Link
            to={`/${ownerBusiness.slug}`}
            className="text-sm underline text-foreground"
          >
            Open my salon
          </Link>
          <Link to="/my/appointments" className="text-sm underline text-foreground">
            My appointments
          </Link>
        </div>
      ) : (
        <Link to="/my/appointments" className="text-sm underline text-foreground">
          My appointments
        </Link>
      )}
    </div>
  )
}
