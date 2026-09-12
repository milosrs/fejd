import type { CSSProperties } from "react"
import { cn } from "#lib/utils"

interface LoaderProps {
  className?: string
  fullScreen?: boolean
  label?: string
  speed?: number
  scale?: number
}

export function Loader({
  className,
  fullScreen = true,
  label = "",
  speed = 2.1,
  scale = 1.2,
}: LoaderProps) {
  return (
    <div
      role="status"
      aria-live="polite"
      className={cn(
        "flex flex-col items-center justify-center gap-6 bg-background",
        fullScreen && "min-h-screen",
        className
      )}
      style={
        {
          "--breathing-speed": `${speed}s`,
          "--breathing-scale": scale,
        } as CSSProperties
      }
    >
      <img
        src="/logo-white.jpg"
        alt=""
        width={753}
        height={289}
        className="w-104 select-none animate-breathing motion-reduce:animate-none dark:hidden"
      />
      <img
        src="/logo_dark.jpg"
        alt=""
        width={762}
        height={301}
        className="hidden w-104 select-none animate-breathing motion-reduce:animate-none dark:block"
      />
      {label && <p className="text-sm text-muted-foreground">{label}</p>}
    </div>
  )
}
