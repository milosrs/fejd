import { BASE_DOMAIN } from "../../src/lib/salonDomain"

export type SalonMetaBusiness = {
  name?: string
  city?: string
  address_line?: string
  postal_code?: string
  country?: string
  phone?: string
  latitude?: number | null
  longitude?: number | null
}

export type SalonMetaService = {
  name?: string
  slug?: string
  description?: string
  price?: number
  duration_minutes?: number
  active?: boolean
}

export type SalonMetaData = {
  business?: SalonMetaBusiness
  services?: SalonMetaService[]
  images?: { hero?: string; logo?: string; background?: string }
}

type SalonMetaTag = { [key: string]: unknown }

function activeServices(salon: SalonMetaData | null): SalonMetaService[] {
  return (salon?.services ?? []).filter((s) => s.active !== false)
}

export function salonName(salon: SalonMetaData | null): string {
  return salon?.business?.name ?? ""
}

export function salonCity(salon: SalonMetaData | null): string {
  return salon?.business?.city ?? ""
}

export function salonTitle(salon: SalonMetaData | null): string {
  const name = salonName(salon)
  const city = salonCity(salon)
  const services = activeServices(salon)
    .map((s) => s.name)
    .filter(Boolean)
    .join(", ")
  const parts = [name]
  if (services) parts.push(services)
  if (city) parts.push(`u ${city}`)
  return parts.join(" — ").trim() || "fejd"
}

export function salonDescription(salon: SalonMetaData | null): string {
  const name = salonName(salon)
  const city = salonCity(salon)
  const services = activeServices(salon)
    .map((s) => s.name)
    .filter(Boolean)
    .join(", ")
  return `${name}${city ? ` in ${city}` : ""}. ${services}.`.trim()
}

export function salonImageUrl(salon: SalonMetaData | null, slug: string): string | undefined {
  const path = salon?.images?.hero ?? salon?.images?.logo
  if (!path) return undefined
  if (/^https?:\/\//.test(path)) return path
  if (!BASE_DOMAIN || !slug) return undefined
  return `https://${slug}.${BASE_DOMAIN}${path.startsWith("/") ? path : `/${path}`}`
}

export function salonRootUrl(slug: string): string | undefined {
  if (!BASE_DOMAIN || !slug) return undefined
  return `https://${slug}.${BASE_DOMAIN}`
}

export function salonPageUrl(slug: string, path: string): string | undefined {
  const root = salonRootUrl(slug)
  if (!root) return undefined
  return `${root}${path.startsWith("/") ? path : `/${path}`}`
}

export function hairSalonJsonLd(
  salon: SalonMetaData | null,
  slug: string,
  image?: string,
): unknown {
  const rootUrl = salonRootUrl(slug)
  const hasGeo = salon?.business?.latitude != null && salon?.business?.longitude != null
  return {
    "@context": "https://schema.org",
    "@type": "HairSalon",
    name: salon?.business?.name,
    ...(rootUrl ? { url: rootUrl } : {}),
    ...(image ? { image } : {}),
    ...(hasGeo
      ? {
          geo: {
            "@type": "GeoCoordinates",
            latitude: salon?.business?.latitude,
            longitude: salon?.business?.longitude,
          },
        }
      : {}),
    address: {
      "@type": "PostalAddress",
      streetAddress: salon?.business?.address_line,
      addressLocality: salon?.business?.city,
      postalCode: salon?.business?.postal_code,
      addressCountry: salon?.business?.country,
    },
    telephone: salon?.business?.phone,
    makesOffer: activeServices(salon).map((s) => ({
      "@type": "Offer",
      name: s.name,
      description: s.description,
      price: s.price,
      priceCurrency: "RSD",
    })),
  }
}

export function serviceJsonLd(
  salon: SalonMetaData | null,
  service: SalonMetaService | undefined,
  slug: string,
): unknown {
  const rootUrl = salonRootUrl(slug)
  return {
    "@context": "https://schema.org",
    "@type": "Service",
    name: service?.name,
    description: service?.description,
    provider: {
      "@type": "HairSalon",
      name: salon?.business?.name,
      ...(rootUrl ? { url: rootUrl } : {}),
    },
    ...(salon?.business?.city
      ? { areaServed: { "@type": "City", name: salon.business.city } }
      : {}),
    ...(service?.price != null
      ? { offers: { "@type": "Offer", price: service.price, priceCurrency: "RSD" } }
      : {}),
  }
}

export function breadcrumbJsonLd(items: Array<{ name: string; url?: string }>): unknown {
  return {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      ...(item.url ? { item: item.url } : {}),
    })),
  }
}

export function salonMeta(args: {
  title: string
  description: string
  url?: string
  image?: string
  jsonLd?: unknown[]
}): SalonMetaTag[] {
  const { title, description, url, image, jsonLd = [] } = args
  const tags: SalonMetaTag[] = [
    { title },
    { name: "description", content: description },
    { property: "og:title", content: title },
    { property: "og:description", content: description },
    { property: "og:type", content: "website" },
    { name: "twitter:card", content: "summary_large_image" },
    { name: "twitter:title", content: title },
    { name: "twitter:description", content: description },
  ]
  if (url) {
    tags.push({ property: "og:url", content: url }, { tagName: "link", rel: "canonical", href: url })
  }
  if (image) {
    tags.push({ property: "og:image", content: image }, { name: "twitter:image", content: image })
  }
  for (const node of jsonLd) tags.push({ "script:ld+json": node })
  return tags
}

export function salonFromMatches(matches: unknown): SalonMetaData | null {
  if (!Array.isArray(matches)) return null
  for (const match of matches) {
    if (!match || typeof match !== "object") continue
    const data = (match as { loaderData?: unknown }).loaderData
    if (!data || typeof data !== "object") continue
    const salon = (data as { salon?: SalonMetaData | null }).salon
    if (salon) return salon
  }
  return null
}
