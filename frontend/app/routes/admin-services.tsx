import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { AdminServicesPage } from "../../src/pages/AdminServicesPage"

// Authenticated pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default function AdminServices() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <AdminServicesPage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
