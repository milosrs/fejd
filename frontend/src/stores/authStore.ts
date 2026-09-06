import { create } from "zustand"
import { auth } from "../lib/auth"
import type { AuthUserInfo } from "../lib/auth/types"

interface AuthState {
  initialized: boolean
  authenticated: boolean
  userInfo: AuthUserInfo | null
  roles: string[]
  init: () => Promise<void>
  login: () => Promise<void>
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  initialized: false,
  authenticated: false,
  userInfo: null,
  roles: [],

  init: async () => {
    // Native: the appUrlOpen redirect resolves auth asynchronously; sync the
    // store whenever the adapter signals a change.
    auth.onAuthChange(() => {
      set({
        authenticated: auth.isAuthenticated(),
        userInfo: auth.getUserInfo(),
        roles: auth.getRoles(),
      })
    })

    try {
      const authenticated = await auth.init()
      set({
        initialized: true,
        authenticated,
        userInfo: authenticated ? auth.getUserInfo() : null,
        roles: authenticated ? auth.getRoles() : [],
      })
    } catch (err) {
      console.error("[auth] init failed:", err)
      set({ initialized: true, authenticated: false })
    }
  },

  login: async () => {
    await auth.login()
  },

  logout: async () => {
    await auth.logout()
    set({ authenticated: false, userInfo: null, roles: [] })
  },
}))
