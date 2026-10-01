import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { MyReservationsPage } from "../../src/pages/MyReservationsPage"

export default function AdminMyReservations() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <MyReservationsPage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
