import { X } from "lucide-react"
import { useToastStore, type ToastVariant } from "../../stores/toastStore"

const variantClass: Record<ToastVariant, string> = {
  error: "bg-destructive text-destructive-foreground border-destructive",
  success: "bg-green-600 text-white border-green-600",
  info: "bg-background text-foreground border-border",
}

export function Toaster() {
  const toasts = useToastStore((s) => s.toasts)
  const dismiss = useToastStore((s) => s.dismiss)

  if (toasts.length === 0) return null

  return (
    <div className="fixed left-1/2 top-[calc(1rem+env(safe-area-inset-top))] z-[100] flex w-[calc(100%-2rem)] max-w-sm -translate-x-1/2 flex-col gap-2 pointer-events-none">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`pointer-events-auto flex items-center gap-3 rounded-xl border px-4 py-3 text-sm font-medium shadow-lg ${variantClass[t.variant]}`}
        >
          <span className="flex-1">{t.message}</span>
          <button
            type="button"
            aria-label="Dismiss"
            onClick={() => dismiss(t.id)}
            className="shrink-0 opacity-70 hover:opacity-100"
          >
            <X className="size-4" />
          </button>
        </div>
      ))}
    </div>
  )
}
