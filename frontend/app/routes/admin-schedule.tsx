import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { AdminSchedulePage } from "../../src/pages/AdminSchedulePage"

export default function AdminSchedule() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <AdminSchedulePage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
