import { ProtectedRoute } from "../../src/App"
import { InvitedCustomersPage } from "../../src/pages/InvitedCustomersPage"

// Authenticated pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default function AdminInvitedCustomers() {
  return (
    <ProtectedRoute>
      <InvitedCustomersPage />
    </ProtectedRoute>
  )
}
