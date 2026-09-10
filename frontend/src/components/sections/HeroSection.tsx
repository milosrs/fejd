import { useNavigate } from "react-router-dom"
import type { HeroContent } from "../../lib/sections"
import { resolveImageUrl } from "../../lib/images"
import { salonPath } from "../../lib/salonDomain"
import { Button } from "../ui/button"

interface HeroSectionProps {
  content: HeroContent
  backgroundUrl?: string
  logoUrl?: string
  slug?: string
  contained?: boolean
}

export function HeroSection({
  content,
  backgroundUrl,
  logoUrl,
  slug,
  contained,
}: HeroSectionProps) {
  const navigate = useNavigate()
  const background = resolveImageUrl(backgroundUrl)
  const logo = resolveImageUrl(logoUrl)

  return (
    <section
      className={`relative overflow-hidden bg-background ${
        contained
          ? "rounded-xl border border-border"
          : "left-1/2 -mt-8 w-[100dvw] -translate-x-1/2 min-h-[calc(100dvh_-_var(--top-header-height,0px)_-_var(--salon-header-height,0px))]"
      }`}
    >
      {background && (
        <img
          src={background}
          alt=""
          className="absolute inset-0 h-full w-full object-cover"
        />
      )}
      <div
        className={`relative flex flex-col items-center justify-center gap-4 px-6 text-center ${
          contained
            ? "py-12"
            : "min-h-[calc(100dvh_-_var(--top-header-height,0px)_-_var(--salon-header-height,0px))] py-16"
        }`}
      >
        {logo && (
          <img
            src={logo}
            alt=""
            className="h-54 w-54 rounded-full border-2 border-border bg-background object-cover"
          />
        )}
        {(content.headline || content.subheadline) && (
          <div className="flex flex-col items-center gap-2 rounded-xl border border-border bg-background/70 px-5 py-4 backdrop-blur-sm">
            {content.headline && (
              <h2 className="text-3xl font-bold text-foreground">{content.headline}</h2>
            )}
            {content.subheadline && (
              <p className="max-w-xl text-muted-foreground">{content.subheadline}</p>
            )}
          </div>
        )}
        {content.cta_text && slug && (
          <Button onClick={() => navigate(salonPath(slug, "/book"))}>{content.cta_text}</Button>
        )}
      </div>
    </section>
  )
}
