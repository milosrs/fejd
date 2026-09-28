import { useRef, useState, type ReactNode } from "react"
import { createPortal } from "react-dom"
import { cn } from "#lib/utils"

// DisabledTooltip wraps a disabled control in a focusable, event-capturing
// span and shows a reason tooltip above it. A disabled <button> swallows
// pointer events (pointer-events: none), so the wrapper span — which is not
// disabled — captures hover (desktop) and touch press (mobile) instead and
// renders the reason in a portal positioned above the control.
export function DisabledTooltip({
  reason,
  className,
  children,
}: {
  reason?: string
  className?: string
  children: ReactNode
}) {
  const ref = useRef<HTMLSpanElement | null>(null)
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null)
  const [visible, setVisible] = useState(false)

  if (!reason) return <>{children}</>

  const show = () => {
    const el = ref.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    setPos({ top: rect.top, left: rect.left + rect.width / 2 })
    setVisible(true)
  }
  const hide = () => setVisible(false)

  return (
    <span
      ref={ref}
      className={cn("relative inline-flex", className)}
      onPointerEnter={show}
      onPointerLeave={hide}
      onPointerCancel={hide}
    >
      {children}
      {visible &&
        pos &&
        createPortal(
          <div
            role="tooltip"
            className="pointer-events-none fixed z-[80] max-w-xs -translate-x-1/2 -translate-y-full rounded-md bg-foreground px-2.5 py-1.5 text-center text-xs font-medium text-background shadow-lg"
            style={{ top: pos.top - 6, left: pos.left }}
          >
            {reason}
          </div>,
          document.body,
        )}
    </span>
  )
}
