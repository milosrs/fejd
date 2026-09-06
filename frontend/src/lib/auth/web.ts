import keycloak from "../keycloak"
import type { AuthAdapter, AuthUserInfo } from "./types"

function parsedToken(): Record<string, any> | undefined {
  return keycloak.tokenParsed as Record<string, any> | undefined
}

export const webAdapter: AuthAdapter = {
  async init() {
    return keycloak.init({ onLoad: "login-required", pkceMethod: "S256" })
  },

  async login() {
    await keycloak.login()
  },

  async logout() {
    await keycloak.logout()
  },

  isAuthenticated() {
    return keycloak.authenticated ?? false
  },

  async getToken() {
    return keycloak.token ?? undefined
  },

  getUserInfo(): AuthUserInfo | null {
    if (!keycloak.authenticated) return null
    const t = parsedToken()
    return {
      sub: keycloak.subject ?? "",
      email: t?.email ?? "",
      name: t?.name ?? t?.preferred_username ?? "",
    }
  },

  getRoles() {
    return parsedToken()?.realm_access?.roles ?? []
  },

  onAuthChange() {
    // keycloak-js drives its own redirect flow; init() resolves it.
    return () => {}
  },
}
