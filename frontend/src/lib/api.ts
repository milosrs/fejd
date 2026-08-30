import createClient from "openapi-fetch"
import type { Middleware } from "openapi-fetch"
import type { paths } from "./api-types"
import keycloak from "./keycloak"
import { useAuthStore } from "../stores/authStore"

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    if (keycloak.authenticated) {
      try {
        await keycloak.updateToken(30)
        console.log("[api] token refreshed, expires:", keycloak.tokenParsed?.exp)
        if (keycloak.token) {
          request.headers.set("Authorization", `Bearer ${keycloak.token}`)
        }
      } catch (err) {
        console.error("[api] token refresh failed, redirecting to login:", err)
        keycloak.login()
      }
    } else {
      console.log("[api] not authenticated, no Authorization header for", request.url)
    }
    return request
  },
}

const errorMiddleware: Middleware = {
  async onResponse({ response }) {
    if (response.status === 401) {
      console.warn("[api] 401 from", response.url, "- token invalid/revoked, redirecting to login")
      keycloak.clearToken()
      useAuthStore.setState({ authenticated: false, token: null })
      keycloak.login().catch(() => {})
    }
    if (!response.ok) {
      const body = await response.clone().json().catch(() => undefined)
      throw { status: response.status, body }
    }
    return response
  },
}

const apiClient = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_URL || "http://localhost:8080",
  headers: { "Content-Type": "application/json" },
})

apiClient.use(authMiddleware)
apiClient.use(errorMiddleware)

const { GET, POST, PUT, DELETE } = apiClient

export { GET, POST, PUT, DELETE, apiClient }
export type { paths }
