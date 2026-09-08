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
  register(): Promise<void>
  logout(): Promise<void>
  isAuthenticated(): boolean
  getToken(): Promise<string | undefined>
  /** Force a token refresh; resolves to whether a valid token is present. */
  refresh(): Promise<boolean>
  getUserInfo(): AuthUserInfo | null
  getRoles(): string[]
  /** Subscribe to auth-state changes; returns an unsubscribe function. */
  onAuthChange(listener: () => void): () => void
  /**
   * Subscribe to invite deep links (e.g. `https://app.example.com/invite/{token}`).
   * The listener receives the raw invite token. Web captures invites via the
   * `/invite/:token` route instead, so this is only fired on native.
   */
  onInviteLink(listener: (token: string) => void): () => void
}
