import { create } from "zustand"
import keycloak from "../lib/keycloak"

interface AuthState {
  initialized: boolean
  authenticated: boolean
  token: string | null
  userInfo: { sub: string; email: string; name: string } | null
  roles: string[]
  init: () => Promise<void>
  login: () => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  initialized: false,
  authenticated: false,
  token: null,
  userInfo: null,
  roles: [],
  init: async () => {
    try {
      console.log("[auth] initializing keycloak (login-required)")
      const authenticated = await keycloak.init({
        onLoad: "login-required",
        pkceMethod: "S256",
      })
      console.log("[auth] keycloak.init result:", authenticated)
      if (authenticated) {
        const tokenParsed = keycloak.tokenParsed as Record<string, any>
        console.log(
          "[auth] authenticated as",
          tokenParsed?.preferred_username ?? keycloak.subject,
          "roles:",
          tokenParsed?.realm_access?.roles,
        )
        set({
          initialized: true,
          authenticated: true,
          token: keycloak.token ?? null,
          userInfo: {
            sub: keycloak.subject ?? "",
            email: tokenParsed?.email ?? "",
            name: tokenParsed?.name ?? tokenParsed?.preferred_username ?? "",
          },
          roles: tokenParsed?.realm_access?.roles ?? [],
        })

        keycloak.onTokenExpired = () => {
          console.log("[auth] token expired, attempting refresh")
          keycloak.updateToken(30).catch(() => {
            console.error("[auth] token refresh failed, redirecting to login")
            set({ authenticated: false, token: null })
            keycloak.login()
          })
        }
      } else {
        console.log("[auth] no valid token, redirecting to keycloak login")
        set({ initialized: true, authenticated: false })
        keycloak.login()
      }
    } catch (err) {
      console.error("[auth] keycloak init failed:", err)
      set({ initialized: true, authenticated: false })
      keycloak.login().catch(() => {})
    }
  },
  login: () => keycloak.login(),
  logout: () => keycloak.logout(),
}))
