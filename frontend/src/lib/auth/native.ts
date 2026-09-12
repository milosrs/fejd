import { Browser } from "@capacitor/browser"
import { App } from "@capacitor/app"
import { SecureStorage } from "@aparajita/capacitor-secure-storage"
import { parseInviteToken } from "../invite"
import type { AuthAdapter, AuthUserInfo } from "./types"

const KEYCLOAK_URL = import.meta.env.VITE_KEYCLOAK_URL || "http://localhost:9090"
const KEYCLOAK_REALM = import.meta.env.VITE_KEYCLOAK_REALM || "fejd"
const CLIENT_ID = import.meta.env.VITE_KEYCLOAK_NATIVE_CLIENT_ID || "salon-mobile"
const REDIRECT_URI = import.meta.env.VITE_KEYCLOAK_NATIVE_REDIRECT_URI || "fejd://callback"

const AUTH_URL = `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/auth`
const TOKEN_URL = `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token`

const K_REFRESH = "fejd.refresh_token"
const K_CODE_VERIFIER = "fejd.code_verifier"
const K_RETURN_TO = "fejd.return_to"

// --- PKCE / crypto helpers -------------------------------------------------

function base64UrlEncode(input: Uint8Array): string {
  let binary = ""
  for (const b of input) binary += String.fromCharCode(b)
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "")
}

function randomString(length: number): string {
  const bytes = new Uint8Array(length)
  crypto.getRandomValues(bytes)
  return base64UrlEncode(bytes)
}

async function sha256(input: string): Promise<ArrayBuffer> {
  return crypto.subtle.digest("SHA-256", new TextEncoder().encode(input))
}

async function generateChallenge(verifier: string): Promise<string> {
  return base64UrlEncode(new Uint8Array(await sha256(verifier)))
}

function decodeJwt(token: string): Record<string, any> {
  const payload = token.split(".")[1]
  const base64 = payload.replace(/-/g, "+").replace(/_/g, "/")
  const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=")
  return JSON.parse(atob(padded))
}

// rememberReturnTo records the current in-app route so the user can be sent
// back to it after authentication, even if the app was cold-started by the
// redirect (which loses the in-memory router state).
function rememberReturnTo() {
  try {
    const { pathname, search, hash } = window.location
    localStorage.setItem(K_RETURN_TO, pathname + search + hash)
  } catch {
    // localStorage may be unavailable; return-to is best-effort.
  }
}

export function consumeReturnTo(): string | undefined {
  try {
    const path = localStorage.getItem(K_RETURN_TO)
    if (path) localStorage.removeItem(K_RETURN_TO)
    return path || undefined
  } catch {
    return undefined
  }
}

// --- state -----------------------------------------------------------------

let accessToken: string | null = null
let refreshToken: string | null = null
let userInfo: AuthUserInfo | null = null
let roles: string[] = []
let codeVerifier: string | null = null
let listenerRegistered = false
const changeListeners = new Set<() => void>()
const inviteListeners = new Set<(token: string) => void>()

function notify() {
  for (const l of changeListeners) l()
}

function notifyInvite(token: string) {
  for (const l of inviteListeners) l(token)
}

function handleOpenURL(url: string) {
  const inviteToken = parseInviteToken(url)
  if (inviteToken) {
    notifyInvite(inviteToken)
    return
  }

  handleRedirect(url)
    .then(async (handled) => {
      await Browser.close()
      if (!handled) {
        // A non-auth redirect (e.g. logout) still means state changed.
        notify()
      }
    })
    .catch(async (err) => {
      console.error("[auth] native redirect failed:", err)
      await Browser.close()
    })
}

interface TokenResponse {
  access_token: string
  refresh_token?: string
  expires_in: number
}

function applyTokens(t: TokenResponse) {
  accessToken = t.access_token
  if (t.refresh_token) {
    refreshToken = t.refresh_token
    SecureStorage.setItem(K_REFRESH, t.refresh_token).catch(() => {})
  }

  const claims = decodeJwt(t.access_token)
  userInfo = {
    sub: claims.sub ?? "",
    email: claims.email ?? "",
    name: claims.name ?? claims.preferred_username ?? "",
  }
  roles = claims.realm_access?.roles ?? []
}

async function exchangeToken(body: Record<string, string>): Promise<TokenResponse> {
  const params = new URLSearchParams({ client_id: CLIENT_ID, ...body })
  const res = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: params.toString(),
  })
  if (!res.ok) {
    throw new Error(`token exchange failed: ${res.status}`)
  }
  return res.json()
}

async function refreshTokens(): Promise<void> {
  if (!refreshToken) throw new Error("no refresh token")
  const tokens = await exchangeToken({
    grant_type: "refresh_token",
    refresh_token: refreshToken,
  })
  applyTokens(tokens)
}

async function handleRedirect(url: string): Promise<boolean> {
  const u = new URL(url)
  const error = u.searchParams.get("error")
  if (error) throw new Error(`authorization failed: ${error}`)

  const code = u.searchParams.get("code")
  if (!code) return false

  // The verifier may live in secure storage if the app was cold-started while
  // the browser was open (the in-memory variable is lost on process death).
  const verifier =
    codeVerifier ?? (await SecureStorage.getItem(K_CODE_VERIFIER).catch(() => null))
  if (!verifier) return false
  codeVerifier = null
  await SecureStorage.removeItem(K_CODE_VERIFIER).catch(() => {})

  const tokens = await exchangeToken({
    grant_type: "authorization_code",
    code,
    redirect_uri: REDIRECT_URI,
    code_verifier: verifier,
  })
  applyTokens(tokens)
  notify()
  return true
}

export const nativeAdapter: AuthAdapter = {
  async init() {
    if (!listenerRegistered) {
      listenerRegistered = true
      App.addListener("appUrlOpen", (state) => {
        handleOpenURL(state.url)
      })
      // Cold start: the app may have been launched by the invite link itself.
      App.getLaunchUrl()
        .then((launch) => {
          if (launch?.url) handleOpenURL(launch.url)
        })
        .catch(() => {})
    }

    refreshToken = await SecureStorage.getItem(K_REFRESH)
    if (refreshToken) {
      try {
        await refreshTokens()
        return true
      } catch {
        refreshToken = null
        await SecureStorage.removeItem(K_REFRESH).catch(() => {})
        return false
      }
    }
    return false
  },

  async login() {
    codeVerifier = randomString(64)
    const challenge = await generateChallenge(codeVerifier)

    // Persist the verifier and remember the current route so a cold-started
    // app (killed while the system browser was open) can still complete the
    // exchange and return the user to where they were.
    await SecureStorage.setItem(K_CODE_VERIFIER, codeVerifier).catch(() => {})
    rememberReturnTo()

    const params = new URLSearchParams({
      client_id: CLIENT_ID,
      redirect_uri: REDIRECT_URI,
      response_type: "code",
      scope: "openid profile email",
      code_challenge: challenge,
      code_challenge_method: "S256",
      state: randomString(32),
    })

    await Browser.open({ url: `${AUTH_URL}?${params.toString()}`, windowName: "_self" })
  },

  async register() {
    codeVerifier = randomString(64)
    const challenge = await generateChallenge(codeVerifier)
    await SecureStorage.setItem(K_CODE_VERIFIER, codeVerifier).catch(() => {})
    rememberReturnTo()

    const registrationUrl =
      `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/registrations?` +
      new URLSearchParams({
        client_id: CLIENT_ID,
        redirect_uri: REDIRECT_URI,
        response_type: "code",
        scope: "openid profile email",
        code_challenge: challenge,
        code_challenge_method: "S256",
      }).toString()

    await Browser.open({ url: registrationUrl, windowName: "_self" })
  },

  async logout() {
    accessToken = null
    refreshToken = null
    userInfo = null
    roles = []
    codeVerifier = null
    await SecureStorage.removeItem(K_REFRESH).catch(() => {})
    notify()
  },

  isAuthenticated() {
    return accessToken != null
  },

  async getToken() {
    if (!accessToken) return undefined
    const claims = decodeJwt(accessToken)
    if (claims.exp * 1000 < Date.now() - 30_000) {
      await refreshTokens()
    }
    return accessToken ?? undefined
  },

  async refresh() {
    try {
      await refreshTokens()
      return true
    } catch {
      return false
    }
  },

  getUserInfo() {
    return userInfo
  },

  getRoles() {
    return roles
  },

  onAuthChange(listener: () => void) {
    changeListeners.add(listener)
    return () => changeListeners.delete(listener)
  },

  onInviteLink(listener: (token: string) => void) {
    inviteListeners.add(listener)
    return () => inviteListeners.delete(listener)
  },
}
