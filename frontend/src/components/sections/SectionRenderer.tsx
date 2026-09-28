import { useI18n } from "../../lib/i18n"
import type {
  AboutContent,
  ContactContent,
  GalleryContent,
  HeroContent,
  HoursContent,
  Section,
  SectionContent,
} from "../../lib/sections"
import { useSalonContext } from "../../context/SalonContext"
import { useSalonDraftStore } from "../../stores/salonDraftStore"
import { HeroSection } from "./HeroSection"
import { AboutSection } from "./AboutSection"
import { GallerySection } from "./GallerySection"
import { ContactSection } from "./ContactSection"
import { HoursSection } from "./HoursSection"

export function SectionRenderer({
  section,
  contained,
}: {
  section: Section
  contained?: boolean
}) {
  const { pickLocalized } = useI18n()
  const { slug, salon } = useSalonContext()
  const draft = useSalonDraftStore((s) => s.drafts[slug])

  const content = pickLocalized<SectionContent>(
    section.content as Record<string, SectionContent>,
  )

  switch (section.type) {
    case "hero":
      return (
        <HeroSection
          content={(content ?? {}) as HeroContent}
          backgroundUrl={contained ? (draft?.background ?? salon?.images.background) : salon?.images.background}
          logoUrl={contained ? (draft?.logo ?? salon?.images.logo) : salon?.images.logo}
          slug={slug}
          contained={contained}
        />
      )
    case "about":
      return <AboutSection content={(content ?? {}) as AboutContent} />
    case "gallery":
      return <GallerySection content={(content ?? {}) as GalleryContent} contained={contained} />
    case "contact":
      return <ContactSection content={(content ?? {}) as ContactContent} />
    case "hours":
      return <HoursSection content={(content ?? {}) as HoursContent} />
    default:
      return null
  }
}
