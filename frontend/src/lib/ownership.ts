import type { Me } from "../hooks/useMe"

export function isOwnerOfBusiness(me: Me | undefined, slug: string): boolean {
  if (!me) return false
  return me.businesses.some((b) => b.slug === slug && b.role === "admin")
}
