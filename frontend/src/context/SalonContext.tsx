import { createContext, useContext, useState } from "react"
import { useParams } from "react-router"
import { useSalon, type Salon } from "../hooks/useSalon"
import { useMe } from "../hooks/useMe"
import { useAuthStore } from "../stores/authStore"
import { isOwnerOfBusiness, hasRole } from "../lib/ownership"
import { subdomainSlug } from "../lib/salonDomain"

interface SalonContextValue {
  slug: string
  salon: Salon | undefined
  isLoading: boolean
  error: unknown
  editing: boolean
  setEditing: (editing: boolean) => void
}

const SalonContext = createContext<SalonContextValue | null>(null)

export function SalonProvider({
  children,
  initialEditing = false,
  slug: hostSlug,
}: {
  children: React.ReactNode
  initialEditing?: boolean
  slug?: string
}) {
  const { slug: paramSlug } = useParams<{ slug: string }>()
  const slug = hostSlug ?? paramSlug ?? subdomainSlug() ?? ""
  const { data, isLoading, error } = useSalon(slug)
  const [editing, setEditing] = useState(initialEditing)

  return (
    <SalonContext.Provider
      value={{ slug, salon: data, isLoading, error, editing, setEditing }}
    >
      {children}
    </SalonContext.Provider>
  )
}

export function useSalonContext() {
  const ctx = useContext(SalonContext)
  if (!ctx) {
    throw new Error("useSalonContext must be used within a SalonProvider")
  }
  return ctx
}

// useIsOwner returns undefined while the auth state (and the current user's
// businesses) are still resolving, so callers can render a skeleton instead of
// flashing the non-owner variant before the real value is known.
export function useIsOwner(): boolean | undefined {
  const { slug } = useSalonContext()
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const roles = useAuthStore((s) => s.roles)
  const { data: me, isLoading } = useMe()

  if (!initialized) return undefined
  if (!authenticated || !hasRole(roles, "Owner")) return false
  if (isLoading || !me) return undefined
  return isOwnerOfBusiness(me, slug)
}

// useIsMember is true for any active member of the current salon (owner or
// employee). Member-level actions like generating invite links rely on it.
export function useIsMember(): boolean | undefined {
  const { slug } = useSalonContext()
  const initialized = useAuthStore((s) => s.initialized)
  const authenticated = useAuthStore((s) => s.authenticated)
  const { data: me, isLoading } = useMe()

  if (!initialized) return undefined
  if (!authenticated) return false
  if (isLoading || !me) return undefined
  return me.businesses.some((b) => b.slug === slug)
}
