import { useMutation, useQueryClient } from "@tanstack/react-query"
import { createService, updateService, deleteService, uploadServiceImage, setServiceEmployees } from "./useApi"
import type { components } from "../lib/api-types"

type ServiceInput = components["schemas"]["handler.ServiceInput"]

export function useServiceMutations(businessId: string, slug: string) {
  const queryClient = useQueryClient()

  const create = useMutation({
    mutationFn: (input: ServiceInput) => createService(businessId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services", slug] })
    },
  })

  const update = useMutation({
    mutationFn: (vars: { serviceId: string; input: Partial<ServiceInput> }) =>
      updateService(businessId, vars.serviceId, vars.input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services", slug] })
    },
  })

  const remove = useMutation({
    mutationFn: (serviceId: string) => deleteService(businessId, serviceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services", slug] })
    },
  })

  const uploadImage = useMutation({
    mutationFn: (vars: { serviceId: string; file: File }) =>
      uploadServiceImage(businessId, vars.serviceId, vars.file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["services", slug] })
    },
  })

  const setEmployees = useMutation({
    mutationFn: (vars: { serviceId: string; businessUserIds: string[] }) =>
      setServiceEmployees(businessId, vars.serviceId, vars.businessUserIds),
    onSuccess: (_data, vars) => {
      queryClient.invalidateQueries({ queryKey: ["services", slug] })
      queryClient.invalidateQueries({
        queryKey: ["admin-service-employees", businessId, vars.serviceId],
      })
    },
  })

  return { create, update, remove, uploadImage, setEmployees }
}
