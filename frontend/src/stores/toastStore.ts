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

// Anti-spam limits: at most MAX_VISIBLE toasts on screen at once, and the same
// message (per variant) is not shown again within SAME_MESSAGE_DELAY_MS.
const MAX_VISIBLE = 3
const SAME_MESSAGE_DELAY_MS = 2500
const AUTO_DISMISS_MS = 6000

let counter = 0
let queue: Toast[] = []
const lastShownAt = new Map<string, number>()

function scheduleDismiss(id: string) {
  setTimeout(() => {
    useToastStore.getState().dismiss(id)
  }, AUTO_DISMISS_MS)
}

function promoteNext() {
  while (useToastStore.getState().toasts.length < MAX_VISIBLE && queue.length > 0) {
    const next = queue.shift()!
    useToastStore.setState((s) => ({ toasts: [...s.toasts, next] }))
    scheduleDismiss(next.id)
  }
}

export const useToastStore = create<ToastState>((set, get) => ({
  toasts: [],
  toast: (message, variant = "info") => {
    const key = `${variant}:${message}`
    const now = Date.now()
    const last = lastShownAt.get(key) ?? 0
    if (now - last < SAME_MESSAGE_DELAY_MS) return
    lastShownAt.set(key, now)

    const toast: Toast = { id: String(++counter), message, variant }

    if (get().toasts.length >= MAX_VISIBLE) {
      queue.push(toast)
      return
    }

    set((s) => ({ toasts: [...s.toasts, toast] }))
    scheduleDismiss(toast.id)
  },
  error: (message) => get().toast(message, "error"),
  success: (message) => get().toast(message, "success"),
  dismiss: (id) => {
    set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
    promoteNext()
  },
}))

// getErrorMessage extracts the human-readable message from a thrown API error
// ({ status, body: { error } }) or a generic Error.
export function getErrorMessage(err: unknown): string {
  const e = err as { body?: { error?: string }; message?: string }
  return e?.body?.error ?? e?.message ?? "Something went wrong."
}
