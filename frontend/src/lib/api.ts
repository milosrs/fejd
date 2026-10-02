import createClient from "openapi-fetch"
import type { Middleware } from "openapi-fetch"
import type { paths } from "./api-types"
import { auth } from "./auth"
import { useAuthStore } from "../stores/authStore"
import { useToastStore } from "../stores/toastStore"
import { Capacitor } from "@capacitor/core"

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    const token = await auth.getToken()
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`)
    }
    return request
  },
}

const errorMiddleware: Middleware = {
  async onResponse({ response }) {
    if (response.status === 401) {
      // A 401 usually means a stale token, not a dead session: refresh it and
      // let the query retry with the fresh token. If the refresh fails, just
      // mark the user unauthenticated — the ProtectedRoute will prompt for login
      // when the user actually tries to reach a protected page. Auto-redirecting
      // here causes a full-page bounce on the initial load.
      const refreshed = await auth.refresh().catch(() => false)
      if (!refreshed) {
        console.warn("[api] 401 from", response.url, "- session expired")
        useAuthStore.setState({ authenticated: false, userInfo: null, roles: [] })
      }
    }
    if (!response.ok) {
      const body = await response.clone().json().catch(() => undefined)
      // Surface every 4xx/5xx failure as a toast. 401 is handled above via the
      // refresh/login redirect rather than a toast. The toast store dedupes
      // identical messages and caps concurrent toasts, so retried reads don't
      // spam the screen.
      if (response.status !== 401) {
        useToastStore.getState().error(body?.error ?? `Request failed (${response.status})`)
      }
      throw { status: response.status, body }
    }
    return response
  },
}

// On web the API is same-origin (served through the reverse proxy alongside the
// SPA), so the base is empty and requests resolve against the current subdomain.
// The native app has no reverse proxy and needs the absolute API URL.
const isNative = Capacitor.isNativePlatform()
const configuredBase = import.meta.env.VITE_API_URL || ""

// No default Content-Type: openapi-fetch sets application/json for JSON bodies
// automatically, and leaves it unset for FormData so the browser can add the
// multipart boundary. A fixed application/json default would break multipart
// uploads.
const apiClient = createClient<paths>({
  baseUrl: isNative ? configuredBase : "",
})

export const API_BASE_URL = isNative ? configuredBase : ""

apiClient.use(authMiddleware)
apiClient.use(errorMiddleware)

const { GET, POST, PUT, DELETE } = apiClient

export { GET, POST, PUT, DELETE, apiClient }
export type { paths }

/**
 * Finalize a self-registration: asks the backend to grant the realm role
 * carried in the token's registration_role claim (see /api/me/claim-role).
 * Uses raw fetch because the endpoint is not part of the generated
 * openapi-types client surface.
 */
export async function claimRegistrationRole(): Promise<boolean> {
  const token = await auth.getToken()
  const res = await fetch(`${API_BASE_URL}/api/me/claim-role`, {
    method: "POST",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) throw new Error(`claim-role failed: ${res.status}`)
  const body = await res.json().catch(() => ({}))
  return Boolean(body?.claimed)
}
