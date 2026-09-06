import { useMutation, useQueryClient } from "@tanstack/react-query"
import { POST, PUT, DELETE } from "../lib/api"
import type { Section } from "../lib/sections"

const URL_ADMIN_SECTIONS = "/api/admin/business/{businessID}/sections" as const
const URL_ADMIN_SECTIONS_ITEM =
  "/api/admin/business/{businessID}/sections/{sectionID}" as const
const URL_ADMIN_SECTIONS_REORDER =
  "/api/admin/business/{businessID}/sections/reorder" as const

type JsonObject = Record<string, unknown>

function createSectionRequest(businessId: string, type: string, content: JsonObject) {
  return POST(URL_ADMIN_SECTIONS, {
    params: { path: { businessID: businessId } },
    body: { type, content } as any,
  })
}

function updateSectionRequest(businessId: string, sectionId: string, content: JsonObject) {
  return PUT(URL_ADMIN_SECTIONS_ITEM, {
    params: { path: { businessID: businessId, sectionID: sectionId } },
    body: { content } as any,
  })
}

function deleteSectionRequest(businessId: string, sectionId: string) {
  return DELETE(URL_ADMIN_SECTIONS_ITEM, {
    params: { path: { businessID: businessId, sectionID: sectionId } },
  })
}

function reorderSectionsRequest(businessId: string, sectionIds: string[]) {
  return PUT(URL_ADMIN_SECTIONS_REORDER, {
    params: { path: { businessID: businessId } },
    body: { section_ids: sectionIds } as any,
  })
}

export function useSectionMutations(businessId: string, slug: string) {
  const queryClient = useQueryClient()

  const create = useMutation({
    mutationFn: (vars: { type: string; content: JsonObject }) =>
      createSectionRequest(businessId, vars.type, vars.content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sections", slug] })
    },
  })

  const update = useMutation({
    mutationFn: (vars: { sectionId: string; content: JsonObject }) =>
      updateSectionRequest(businessId, vars.sectionId, vars.content),
    onMutate: async (vars) => {
      await queryClient.cancelQueries({ queryKey: ["sections", slug] })
      const previous = queryClient.getQueryData<Section[]>(["sections", slug])
      queryClient.setQueryData<Section[]>(["sections", slug], (old) =>
        (old ?? []).map((s) =>
          s.id === vars.sectionId ? { ...s, content: vars.content } : s,
        ),
      )
      return { previous }
    },
    onError: (_error, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["sections", slug], context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["sections", slug] })
    },
  })

  const remove = useMutation({
    mutationFn: (sectionId: string) => deleteSectionRequest(businessId, sectionId),
    onMutate: async (sectionId) => {
      await queryClient.cancelQueries({ queryKey: ["sections", slug] })
      const previous = queryClient.getQueryData<Section[]>(["sections", slug])
      queryClient.setQueryData<Section[]>(["sections", slug], (old) =>
        (old ?? []).filter((s) => s.id !== sectionId),
      )
      return { previous }
    },
    onError: (_error, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["sections", slug], context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["sections", slug] })
    },
  })

  const reorder = useMutation({
    mutationFn: (sectionIds: string[]) => reorderSectionsRequest(businessId, sectionIds),
    onMutate: async (sectionIds) => {
      await queryClient.cancelQueries({ queryKey: ["sections", slug] })
      const previous = queryClient.getQueryData<Section[]>(["sections", slug])
      queryClient.setQueryData<Section[]>(["sections", slug], (old) => {
        if (!old) return old
        const byId = new Map(old.map((s) => [s.id, s]))
        return sectionIds
          .map((id, index) => {
            const section = byId.get(id)
            return section ? { ...section, position: index } : null
          })
          .filter((s): s is Section => s !== null)
      })
      return { previous }
    },
    onError: (_error, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["sections", slug], context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["sections", slug] })
    },
  })

  return { create, update, remove, reorder }
}
