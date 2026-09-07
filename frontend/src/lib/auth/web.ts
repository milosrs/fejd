import keycloak from "../keycloak"
import type { AuthAdapter, AuthUserInfo } from "./types"

function parsedToken(): Record<string, any> | undefined {
  return keycloak.tokenParsed as Record<string, any> | undefined
}

const listeners = new Set<() => void>()

function notify() {
  for (const listener of listeners) listener()
}

export const webAdapter: AuthAdapter = {
  async init() {
    // keycloak-js drives its own auth lifecycle (silent SSO check, token
    // refresh, session expiry). Keep the store in sync by forwarding these
    // events to any registered listeners.
    keycloak.onAuthSuccess = notify
    keycloak.onAuthError = notify
    keycloak.onAuthRefreshSuccess = notify
    keycloak.onAuthLogout = notify
    keycloak.onTokenExpired = () => {
      keycloak
        .updateToken(30)
        .then(() => notify())
        .catch(() => notify())
    }

    return keycloak.init({
      onLoad: "check-sso",
      pkceMethod: "S256",
      silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
    })
  },

  async login() {
    await keycloak.login({ redirectUri: window.location.href })
  },

  async register() {
    await keycloak.register({ redirectUri: window.location.href })
  },

  async logout() {
    await keycloak.logout({ redirectUri: window.location.origin })
  },

  isAuthenticated() {
    return keycloak.authenticated ?? false
  },

  async getToken() {
    try {
      await keycloak.updateToken(30)
    } catch {
      return undefined
    }
    return keycloak.token ?? undefined
  },

  async refresh() {
    try {
      await keycloak.updateToken(-1)
      return keycloak.authenticated ?? false
    } catch {
      return false
    }
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

  onAuthChange(listener: () => void) {
    listeners.add(listener)
    return () => {
      listeners.delete(listener)
    }
  },
}
