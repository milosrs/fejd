import { useAuthStore } from "../stores/authStore"

// useCanWrite reports whether the current user is allowed to attempt write
// requests (POST/PUT/DELETE). Unauthenticated users are allowed so their action
// can trigger a login/register redirect; authenticated users must have a
// verified email, mirroring the backend's read-only gate for unverified emails.
export function useCanWrite() {
  const authenticated = useAuthStore((s) => s.authenticated)
  const emailVerified = useAuthStore((s) => s.emailVerified)
  return !authenticated || emailVerified
}
