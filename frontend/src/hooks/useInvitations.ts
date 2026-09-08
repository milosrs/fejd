import { useMutation } from "@tanstack/react-query"
import { createInvitation, type Invitation } from "./useApi"

export function useInvitations(businessId: string) {
  const invite = useMutation({
    mutationFn: (body?: { expires_in_hours?: number }) => createInvitation(businessId, body),
  })

  return { invite }
}

export type { Invitation }
