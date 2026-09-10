import { useI18n } from "../../lib/i18n"
import type {
  AboutContent,
  ContactContent,
  GalleryContent,
  HeroContent,
  Section,
  SectionContent,
} from "../../lib/sections"
import { useSalonContext } from "../../context/SalonContext"
import { HeroSection } from "./HeroSection"
import { AboutSection } from "./AboutSection"
import { GallerySection } from "./GallerySection"
import { ContactSection } from "./ContactSection"

export function SectionRenderer({
  section,
  contained,
}: {
  section: Section
  contained?: boolean
}) {
  const { pickLocalized } = useI18n()
  const { slug, salon } = useSalonContext()

  const content = pickLocalized<SectionContent>(
    section.content as Record<string, SectionContent>,
  )

  switch (section.type) {
    case "hero":
      return (
        <HeroSection
          content={(content ?? {}) as HeroContent}
          backgroundUrl={salon?.images.background}
          logoUrl={salon?.images.logo}
          slug={slug}
          contained={contained}
        />
      )
    case "about":
      return <AboutSection content={(content ?? {}) as AboutContent} />
    case "gallery":
      return <GallerySection content={(content ?? {}) as GalleryContent} />
    case "contact":
      return <ContactSection content={(content ?? {}) as ContactContent} />
    default:
      return null
  }
}
