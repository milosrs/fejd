import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { MySchedulePage } from "../../src/pages/MySchedulePage"

// Authenticated pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default function AdminMySchedule() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <MySchedulePage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
