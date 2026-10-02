import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { MyReservationsPage } from "../../src/pages/MyReservationsPage"

// Authenticated pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default function AdminMyReservations() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <MyReservationsPage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
