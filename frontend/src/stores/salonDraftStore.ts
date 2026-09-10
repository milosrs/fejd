import { create } from "zustand"
import { persist } from "zustand/middleware"

export type DraftImagePurpose = "logo" | "background"

interface SalonDraft {
  logo?: string
  background?: string
}

interface SalonDraftState {
  drafts: Record<string, SalonDraft>
  setDraftImage: (slug: string, purpose: DraftImagePurpose, dataUrl: string) => void
  clearDraft: (slug: string) => void
}

// useSalonDraftStore holds unsaved salon edits (currently the hero logo and
// background images) as data URLs, persisted to localStorage so in-progress
// edits survive a reload without being committed to the server. "Done" uploads
// and clears the draft; "Cancel Edit" just clears it.
export const useSalonDraftStore = create<SalonDraftState>()(
  persist(
    (set) => ({
      drafts: {},
      setDraftImage: (slug, purpose, dataUrl) =>
        set((state) => ({
          drafts: {
            ...state.drafts,
            [slug]: {
              ...state.drafts[slug],
              [purpose]: dataUrl,
            },
          },
        })),
      clearDraft: (slug) =>
        set((state) => {
          const next = { ...state.drafts }
          delete next[slug]
          return { drafts: next }
        }),
    }),
    { name: "fejd-salon-drafts" },
  ),
)
