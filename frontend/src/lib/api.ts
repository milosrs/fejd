import createClient from "openapi-fetch"
import type { Middleware } from "openapi-fetch"
import type { paths } from "./api-types"
import keycloak from "./keycloak"

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    if (keycloak.authenticated) {
      try {
        await keycloak.updateToken(30)
        if (keycloak.token) {
          request.headers.set("Authorization", `Bearer ${keycloak.token}`)
        }
      } catch {
        keycloak.login()
      }
    }
    return request
  },
}

const errorMiddleware: Middleware = {
  async onResponse({ response }) {
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
