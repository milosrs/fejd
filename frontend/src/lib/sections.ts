export type SectionType = "hero" | "about" | "gallery" | "contact"

export interface Section {
  id: string
  page_id: string
  type: string
  content: Record<string, unknown>
  position: number
}

export interface HeroContent {
  headline?: string
  subheadline?: string
  cta_text?: string
}

export interface AboutContent {
  heading?: string
  body?: string
}

export interface GalleryContent {
  heading?: string
  image_urls?: string[]
}

export interface ContactContent {
  heading?: string
  phone?: string
  email?: string
  address?: string
}

export type SectionContent =
  | HeroContent
  | AboutContent
  | GalleryContent
  | ContactContent

export const KNOWN_SECTION_TYPES: SectionType[] = [
  "hero",
  "about",
  "gallery",
  "contact",
]

export function isSectionType(value: string): value is SectionType {
  return (KNOWN_SECTION_TYPES as string[]).includes(value)
}

export function getLocalizedContent<T extends object>(
  content: Record<string, unknown>,
  locale: string,
): T {
  const c = content[locale]
  return (c && typeof c === "object" ? c : {}) as T
}

export function setLocalizedContent(
  content: Record<string, unknown>,
  locale: string,
  value: unknown,
): Record<string, unknown> {
  return { ...content, [locale]: value }
}
