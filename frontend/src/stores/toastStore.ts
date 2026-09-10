import { create } from "zustand"

export type ToastVariant = "error" | "success" | "info"

export type Toast = {
  id: string
  message: string
  variant: ToastVariant
}

type ToastState = {
  toasts: Toast[]
  toast: (message: string, variant?: ToastVariant) => void
  error: (message: string) => void
  success: (message: string) => void
  dismiss: (id: string) => void
}

let counter = 0

export const useToastStore = create<ToastState>((set, get) => ({
  toasts: [],
  toast: (message, variant = "info") => {
    const id = String(++counter)
    set((s) => ({ toasts: [...s.toasts, { id, message, variant }] }))
    setTimeout(() => {
      set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
    }, 6000)
  },
  error: (message) => get().toast(message, "error"),
  success: (message) => get().toast(message, "success"),
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}))

// getErrorMessage extracts the human-readable message from a thrown API error
// ({ status, body: { error } }) or a generic Error.
export function getErrorMessage(err: unknown): string {
  const e = err as { body?: { error?: string }; message?: string }
  return e?.body?.error ?? e?.message ?? "Something went wrong."
}
