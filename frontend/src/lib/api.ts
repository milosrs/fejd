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
  async onResponse({ request, response }) {
    if (response.status === 401) {
      // A 401 usually means a stale token, not a dead session: refresh it and
      // let the query retry with the fresh token. Only fall back to a full
      // login redirect when the refresh fails, which terminates the loop.
      const refreshed = await auth.refresh().catch(() => false)
      if (!refreshed) {
        console.warn("[api] 401 from", response.url, "- session expired, re-authenticating")
        useAuthStore.setState({ authenticated: false, userInfo: null, roles: [] })
        auth.login().catch(() => {})
      }
    }
    if (!response.ok) {
      const body = await response.clone().json().catch(() => undefined)
      // Surface write failures as a toast. Reads (GET) are left to the UI's
      // loading/empty states and are retried, so toasting them would spam.
      if (response.status !== 401 && request.method !== "GET") {
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
