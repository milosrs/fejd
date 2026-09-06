import { useQuery } from "@tanstack/react-query"
import { GET } from "../lib/api"
import type { Section } from "../lib/sections"

export function useSections(slug: string) {
  return useQuery({
    queryKey: ["sections", slug],
    queryFn: async () => {
      const { data } = await GET("/api/business/{slug}/sections", {
        params: { path: { slug } },
      })
      return (data ?? []) as Section[]
    },
    enabled: !!slug,
  })
}
