import { useNavigate } from "react-router-dom"
import type { HeroContent } from "../../lib/sections"
import { resolveImageUrl } from "../../lib/images"
import { salonPath } from "../../lib/salonDomain"
import { Button } from "../ui/button"

interface HeroSectionProps {
  content: HeroContent
  heroImageUrl?: string
  logoUrl?: string
  slug?: string
}

export function HeroSection({
  content,
  heroImageUrl,
  logoUrl,
  slug,
}: HeroSectionProps) {
  const navigate = useNavigate()
  const hero = resolveImageUrl(heroImageUrl)
  const logo = resolveImageUrl(logoUrl)

  return (
    <section className="relative overflow-hidden rounded-xl border border-border">
      {hero && (
        <img
          src={hero}
          alt=""
          className="absolute inset-0 h-full w-full object-cover opacity-25"
        />
      )}
      <div className="relative flex flex-col items-center gap-4 px-6 py-16 text-center">
        {logo && (
          <img
            src={logo}
            alt=""
            className="h-16 w-16 rounded-full object-cover ring-2 ring-border"
          />
        )}
        {content.headline && (
          <h2 className="text-3xl font-bold text-foreground">{content.headline}</h2>
        )}
        {content.subheadline && (
          <p className="max-w-xl text-muted-foreground">{content.subheadline}</p>
        )}
        {content.cta_text && slug && (
          <Button onClick={() => navigate(salonPath(slug, "/book"))}>{content.cta_text}</Button>
        )}
      </div>
    </section>
  )
}
