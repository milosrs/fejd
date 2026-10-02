import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { AdminSchedulePage } from "../../src/pages/AdminSchedulePage"

// Authenticated pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default function AdminSchedule() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <AdminSchedulePage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
