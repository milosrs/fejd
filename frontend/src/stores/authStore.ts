import { create } from "zustand"
import { auth } from "../lib/auth"
import type { AuthUserInfo } from "../lib/auth/types"

interface AuthState {
  initialized: boolean
  authenticated: boolean
  userInfo: AuthUserInfo | null
  roles: string[]
  pendingInviteToken: string | null
  init: () => Promise<void>
  login: () => Promise<void>
  register: () => Promise<void>
  logout: () => Promise<void>
  setPendingInviteToken: (token: string | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  initialized: false,
  authenticated: false,
  userInfo: null,
  roles: [],
  pendingInviteToken: null,

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

    // Native deep links: an invite link opened while the app runs lands here.
    auth.onInviteLink((token) => {
      set({ pendingInviteToken: token })
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

  register: async () => {
    await auth.register()
  },

  logout: async () => {
    await auth.logout()
    set({ authenticated: false, userInfo: null, roles: [], pendingInviteToken: null })
  },

  setPendingInviteToken: (token) => {
    set({ pendingInviteToken: token })
  },
}))
