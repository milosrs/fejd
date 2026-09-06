import { createContext, useContext } from "react"
import { useParams } from "react-router-dom"
import { useSalon, type Salon } from "../hooks/useSalon"
import { useMe } from "../hooks/useMe"
import { useAuthStore } from "../stores/authStore"
import { isOwnerOfBusiness } from "../lib/ownership"

interface SalonContextValue {
  slug: string
  salon: Salon | undefined
  isLoading: boolean
  error: unknown
}

const SalonContext = createContext<SalonContextValue | null>(null)

export function SalonProvider({ children }: { children: React.ReactNode }) {
  const { slug = "" } = useParams<{ slug: string }>()
  const { data, isLoading, error } = useSalon(slug)

  return (
    <SalonContext.Provider value={{ slug, salon: data, isLoading, error }}>
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

export function useIsOwner() {
  const { slug } = useSalonContext()
  const authenticated = useAuthStore((s) => s.authenticated)
  const { data: me } = useMe()

  if (!authenticated) return false
  return isOwnerOfBusiness(me, slug)
}
