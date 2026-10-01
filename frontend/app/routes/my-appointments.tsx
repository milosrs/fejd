import { ProtectedRoute } from "../../src/App"
import { MyAppointmentsPage } from "../../src/pages/MyAppointmentsPage"

export default function MyAppointments() {
  return (
    <ProtectedRoute>
      <MyAppointmentsPage />
    </ProtectedRoute>
  )
}
