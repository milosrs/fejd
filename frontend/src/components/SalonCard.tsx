import { useNavigate } from "react-router-dom"
import { useSalon } from "../hooks/useSalon"
import { resolveImageUrl } from "../lib/images"
import { openSalon } from "../lib/salonDomain"

interface SalonCardProps {
  business: { id: string; name: string; slug: string }
  isOwner: boolean
}

export function SalonCard({ business, isOwner }: SalonCardProps) {
  const navigate = useNavigate()
  const { data: salon } = useSalon(business.slug)

  const logoUrl = resolveImageUrl(salon?.images.logo)
  const backgroundUrl = resolveImageUrl(salon?.images.background)

  return (
    <button
      type="button"
      onClick={() => openSalon(navigate, business.slug)}
      className="group relative flex aspect-[4/3] w-full cursor-pointer flex-col items-center justify-center gap-3 overflow-hidden rounded-2xl border border-border bg-muted transition-all duration-200 hover:scale-[1.03] hover:opacity-90 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      {backgroundUrl && (
        <img
          src={backgroundUrl}
          alt=""
          className="absolute inset-0 h-full w-full object-cover"
        />
      )}
      <div className="absolute inset-0 bg-black/30" />

      {logoUrl ? (
        <img
          src={logoUrl}
          alt=""
          className="relative z-10 h-20 w-20 rounded-full border-2 border-border bg-background object-cover"
        />
      ) : (
        <span className="relative z-10 flex h-20 w-20 items-center justify-center rounded-full border-2 border-border bg-background text-2xl font-semibold text-foreground">
          {business.name.charAt(0).toUpperCase()}
        </span>
      )}

      <span className="relative z-10 text-base font-medium text-white drop-shadow-sm">
        {business.name}
      </span>

      {isOwner && (
        <span className="absolute right-3 top-3 z-10 rounded-full bg-green-600 px-2.5 py-1 text-xs font-medium text-white">
          owner
        </span>
      )}
    </button>
  )
}
