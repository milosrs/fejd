import { ProtectedRoute } from "../../src/App"
import { InvitedCustomersPage } from "../../src/pages/InvitedCustomersPage"

export default function AdminInvitedCustomers() {
  return (
    <ProtectedRoute>
      <InvitedCustomersPage />
    </ProtectedRoute>
  )
}
