import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { acceptInvitation, createInvitation, getInvitation, type Invitation } from "./useApi"

export function useInvitations(businessId: string) {
  const invite = useMutation({
    mutationFn: (body?: { expires_in_hours?: number }) => createInvitation(businessId, body),
  })

  return { invite }
}

export function useInvitation(token: string | null) {
  return useQuery({
    queryKey: ["invitation", token],
    queryFn: async () => getInvitation(token!),
    enabled: !!token,
  })
}

export function useAcceptInvitation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (token: string) => acceptInvitation(token),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] })
    },
  })
}

export type { Invitation }
