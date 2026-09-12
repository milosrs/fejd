import { useOnboarding } from "#hooks/useOnboarding"
import { useI18n } from "../lib/i18n"
import { CreateSalonForm } from "./CreateSalonForm"
import { Loader } from "./Loader"

export function OnboardingGate({ children }: { children: React.ReactNode }) {
  const { gate, ready } = useOnboarding()
  const { t } = useI18n()

  if (!ready) {
    return <Loader />
  }

  if (gate === "pending") {
    // TODO: replace with the real PendingApprovalPage. The copy should
    // surface token-refresh staleness ("check back shortly") rather than
    // promising an instant unlock.
    return (
      <div className="min-h-app flex items-center justify-center bg-background">
        <p className="text-muted-foreground">{t("onboarding.pending")}</p>
      </div>
    )
  }

  if (gate === "setup") {
    return (
      <div className="min-h-app flex flex-col items-center justify-center gap-4 bg-background p-8">
        <p className="text-muted-foreground">{t("onboarding.setup")}</p>
        <CreateSalonForm />
      </div>
    )
  }

  return <>{children}</>
}
