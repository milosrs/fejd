export type OnboardingGateState = "pending" | "setup" | "ready"

/**
 * Decides which onboarding view the current user should see, based on the
 * /api/me response ({ approval_status, has_salon }).
 *
 * - "pending" -> PendingApprovalPage
 * - "setup"   -> SalonSetupPage (approved but no salon yet)
 * - "ready"   -> normal routes
 */
export function resolveOnboardingState(
  approvalStatus: string | null | undefined,
  hasSalon: boolean,
): OnboardingGateState {
  if (approvalStatus === "pending") {
    return "pending"
  }
  if (approvalStatus === "approved" && !hasSalon) {
    return "setup"
  }
  return "ready"
}
