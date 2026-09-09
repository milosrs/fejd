import { useEffect, useRef } from "react"
import { useNavigate } from "react-router-dom"
import { useAuthStore } from "../../stores/authStore"
import { useAcceptInvitation } from "../../hooks/useInvitations"
import { openSalon } from "../../lib/salonDomain"

/**
 * Watches the pending invite token and, once the user is authenticated, redeems
 * it and navigates to the salon. Mounted once at the app root so it works for
 * both the web `/invite/:token` route and native deep links.
 */
export function InviteAcceptHandler() {
  const navigate = useNavigate()
  const authenticated = useAuthStore((s) => s.authenticated)
  const token = useAuthStore((s) => s.pendingInviteToken)
  const setToken = useAuthStore((s) => s.setPendingInviteToken)
  const accept = useAcceptInvitation()
  const acceptedRef = useRef<string | null>(null)

  useEffect(() => {
    if (!authenticated || !token || acceptedRef.current === token) return
    acceptedRef.current = token
    accept.mutate(token, {
      onSuccess: (business) => {
        setToken(null)
        if (business) openSalon(navigate, business.slug)
      },
      onError: () => {
        setToken(null)
      },
    })
  }, [authenticated, token, accept, setToken, navigate])

  return null
}
