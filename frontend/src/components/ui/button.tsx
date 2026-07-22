import * as React from "react"
import { cn } from "../../lib/utils"

function Button({ className, variant = "default", size = "default", ...props }: React.ComponentProps<"button"> & {
  variant?: "default" | "outline" | "ghost" | "destructive"
  size?: "default" | "sm" | "lg" | "icon"
}) {
  const variants: Record<string, string> = {
    default: "bg-primary text-primary-foreground hover:bg-primary/90",
    outline: "border border-border bg-background hover:bg-muted",
    ghost: "hover:bg-muted",
    destructive: "bg-red-600 text-white hover:bg-red-700",
  }
  const sizes: Record<string, string> = {
    default: "h-10 px-4 py-2",
    sm: "h-8 px-3 text-sm",
    lg: "h-12 px-6 text-lg",
    icon: "h-10 w-10",
  }
  return (
    <button
      data-slot="button"
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-md font-medium cursor-pointer transition-colors disabled:opacity-50 disabled:cursor-not-allowed",
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    />
  )
}

export { Button }
