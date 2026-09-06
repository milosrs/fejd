import { useEffect } from "react"
import { useAuthStore } from "../stores/authStore"
import { useOnboardingStore } from "../stores/onboardingStore"
import { resolveOnboardingState, type OnboardingGateState } from "../lib/onboarding"

export function useOnboarding() {
  const authenticated = useAuthStore((s) => s.authenticated)
  const approvalStatus = useOnboardingStore((s) => s.approvalStatus)
  const hasSalon = useOnboardingStore((s) => s.hasSalon)
  const status = useOnboardingStore((s) => s.status)
  const fetchMe = useOnboardingStore((s) => s.fetchMe)
  const createBusiness = useOnboardingStore((s) => s.createBusiness)
  const reset = useOnboardingStore((s) => s.reset)

  useEffect(() => {
    if (authenticated && status === "idle") {
      fetchMe()
    }
  }, [authenticated, status, fetchMe])

  useEffect(() => {
    if (!authenticated) {
      reset()
    }
  }, [authenticated, reset])

  return {
    approvalStatus,
    hasSalon,
    gate: resolveOnboardingState(approvalStatus, hasSalon) as OnboardingGateState,
    ready: status === "ready",
    createBusiness,
  }
}
