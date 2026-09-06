import { useQuery } from "@tanstack/react-query"
import { GET } from "../lib/api"
import { useAuthStore } from "../stores/authStore"
import type { components } from "../lib/api-types"

export type Me = components["schemas"]["dto.Me"]

export function useMe() {
  const authenticated = useAuthStore((s) => s.authenticated)
  return useQuery({
    queryKey: ["me"],
    queryFn: async () => {
      const { data } = await GET("/api/me")
      return data
    },
    enabled: authenticated,
  })
}
