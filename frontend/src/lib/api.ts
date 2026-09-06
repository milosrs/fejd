import createClient from "openapi-fetch"
import type { Middleware } from "openapi-fetch"
import type { paths } from "./api-types"
import { auth } from "./auth"
import { useAuthStore } from "../stores/authStore"

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
      console.warn("[api] 401 from", response.url, "- re-authenticating")
      useAuthStore.setState({ authenticated: false, userInfo: null, roles: [] })
      auth.login().catch(() => {})
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
