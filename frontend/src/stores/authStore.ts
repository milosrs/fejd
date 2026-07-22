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
      const authenticated = await keycloak.init({
        onLoad: "check-sso",
        silentCheckSsoRedirectUri: window.location.origin + "/silent-check-sso.html",
        pkceMethod: "S256",
      })
      if (authenticated) {
        const tokenParsed = keycloak.tokenParsed as Record<string, any>
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
      } else {
        set({ initialized: true, authenticated: false })
      }

      keycloak.onTokenExpired = () => {
        keycloak.updateToken(30).catch(() => {
          set({ authenticated: false, token: null })
        })
      }
    } catch {
      set({ initialized: true, authenticated: false })
    }
  },
  login: () => keycloak.login(),
  logout: () => keycloak.logout(),
}))
