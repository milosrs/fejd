import { useOnboarding } from "#hooks/useOnboarding"

export function OnboardingGate({ children }: { children: React.ReactNode }) {
  const { gate, ready } = useOnboarding()

  if (!ready) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (gate === "pending") {
    // TODO: replace with the real PendingApprovalPage. The copy should
    // surface token-refresh staleness ("check back shortly") rather than
    // promising an instant unlock.
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Your account is awaiting approval. Check back shortly.</p>
      </div>
    )
  }

  if (gate === "setup") {
    // TODO: replace with the real SalonSetupPage (name form -> createBusiness).
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-muted-foreground">Set up your salon to continue.</p>
      </div>
    )
  }

  return <>{children}</>
}
