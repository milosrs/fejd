import { useMutation, useQueryClient } from "@tanstack/react-query"
import { inviteEmployee, removeEmployee, uploadEmployeeImage } from "./useApi"
import type { components } from "../lib/api-types"

export type EmployeeRemoval = components["schemas"]["handler.RemoveEmployeeResponse"]

export function useBarberMutations(businessId: string, slug: string) {
  const queryClient = useQueryClient()

  const invite = useMutation({
    mutationFn: (vars: { name: string; email: string; service_ids: string[] }) =>
      inviteEmployee(businessId, vars),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["employees", slug] })
    },
  })

  const remove = useMutation({
    mutationFn: (userId: string) => removeEmployee(businessId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["employees", slug] })
    },
  })

  const uploadAvatar = useMutation({
    mutationFn: (vars: { userId: string; file: File }) =>
      uploadEmployeeImage(businessId, vars.userId, vars.file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["employees", slug] })
    },
  })

  return { invite, remove, uploadAvatar }
}
