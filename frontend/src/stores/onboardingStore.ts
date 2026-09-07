import { create } from "zustand"
import { GET, POST } from "../lib/api"

export interface OnboardingState {
  approvalStatus: string | null
  hasSalon: boolean
  status: "idle" | "loading" | "ready"
  fetchMe: () => Promise<void>
  createBusiness: (name: string) => Promise<string | undefined>
  reset: () => void
}

export const useOnboardingStore = create<OnboardingState>((set) => ({
  approvalStatus: null,
  hasSalon: false,
  status: "idle",
  fetchMe: async () => {
    set({ status: "loading" })
    try {
      const { data } = await GET("/api/me")
      set({
        approvalStatus: data?.approval_status ?? null,
        hasSalon: data?.has_salon ?? false,
        status: "ready",
      })
    } catch (err) {
      console.error("[onboarding] failed to fetch /api/me:", err)
      set({ status: "ready" })
    }
  },
  createBusiness: async (name: string) => {
    const { data } = await POST("/api/me/business", { body: { name } })
    set({ hasSalon: true })
    return data?.slug
  },
  reset: () => set({ approvalStatus: null, hasSalon: false, status: "idle" }),
}))
