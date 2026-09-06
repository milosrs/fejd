import { useQuery } from "@tanstack/react-query"
import { GET } from "../lib/api"
import type { components } from "../lib/api-types"

export type Salon = components["schemas"]["handler.BusinessResponse"]

export function useSalon(slug: string) {
  return useQuery({
    queryKey: ["salon", slug],
    queryFn: async () => {
      const { data } = await GET("/api/business/{slug}", {
        params: { path: { slug } },
      })
      return data
    },
    enabled: !!slug,
  })
}
