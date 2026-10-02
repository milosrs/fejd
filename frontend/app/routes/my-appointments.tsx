import { ProtectedRoute } from "../../src/App"
import { MyAppointmentsPage } from "../../src/pages/MyAppointmentsPage"

// Authenticated pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default function MyAppointments() {
  return (
    <ProtectedRoute>
      <MyAppointmentsPage />
    </ProtectedRoute>
  )
}
