import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { MySchedulePage } from "../../src/pages/MySchedulePage"

export default function AdminMySchedule() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <MySchedulePage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
