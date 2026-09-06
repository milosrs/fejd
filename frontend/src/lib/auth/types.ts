export interface AuthUserInfo {
  sub: string
  email: string
  name: string
}

/**
 * Platform-agnostic auth surface. The web adapter wraps keycloak-js; the
 * native adapter implements the OIDC authorization-code + PKCE flow via
 * Capacitor Browser and stores tokens in secure storage (Keychain/Keystore).
 */
export interface AuthAdapter {
  init(): Promise<boolean>
  login(): Promise<void>
  logout(): Promise<void>
  isAuthenticated(): boolean
  getToken(): Promise<string | undefined>
  getUserInfo(): AuthUserInfo | null
  getRoles(): string[]
  /** Subscribe to auth-state changes; returns an unsubscribe function. */
  onAuthChange(listener: () => void): () => void
}
