import { ProtectedRoute } from "../../src/App"
import { OnboardingGate } from "#components/OnboardingGate"
import { AdminServicesPage } from "../../src/pages/AdminServicesPage"

export default function AdminServices() {
  return (
    <ProtectedRoute>
      <OnboardingGate>
        <AdminServicesPage />
      </OnboardingGate>
    </ProtectedRoute>
  )
}
