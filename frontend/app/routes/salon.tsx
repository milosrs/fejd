import { Outlet, useLoaderData } from "react-router"
import { SalonLayout } from "../../src/components/SalonLayout"
import { APP_HOST, BASE_DOMAIN, resolveSubdomainSlug } from "../../src/lib/salonDomain"
import { fetchPublic, prefetchPublic } from "../lib/serverData"
import { resolveQueryClient } from "../lib/context"
import { readLocaleCookie } from "../lib/locale"

type SalonData = {
  business?: {
    name?: string
    city?: string
    address_line?: string
    country?: string
    phone?: string
  }
  services?: Array<{ name?: string; slug?: string; description?: string; price?: number }>
  images?: {
    hero?: string
    logo?: string
    background?: string
  }
}

export async function loader({
  request,
  params,
  context,
}: {
  request: Request
  params: { slug?: string }
  context: unknown
}) {
  const hostname = new URL(request.url).hostname
  const slug = resolveSubdomainSlug(hostname, BASE_DOMAIN, APP_HOST) ?? params.slug ?? ""
  const qc = resolveQueryClient(context)
  const locale = readLocaleCookie(request)

  let salon: SalonData | null = null
  let sections: unknown | null = null

  await prefetchPublic(qc, ["i18n", locale], `/api/i18n/${locale}`)
  if (slug) {
    ;[salon, sections] = await Promise.all([
      fetchPublic<SalonData>(`/api/business/${slug}`).catch(() => null),
      fetchPublic(`/api/business/${slug}/sections`).catch(() => null),
    ])
    if (salon) qc.setQueryData(["salon", slug], salon)
    if (sections) qc.setQueryData(["sections", slug], sections)
  }
  return { slug, salon, sections }
}

function salonTitle(salon: SalonData | null): string {
  const name = salon?.business?.name ?? ""
  const city = salon?.business?.city ?? ""
  const services = (salon?.services ?? []).map((s) => s.name).filter(Boolean).join(", ")
  const parts = [name]
  if (services) parts.push(services)
  if (city) parts.push(`u ${city}`)
  return parts.join(" — ").trim() || "fejd"
}

function salonDescription(salon: SalonData | null): string {
  const name = salon?.business?.name ?? ""
  const city = salon?.business?.city ?? ""
  const services = (salon?.services ?? []).map((s) => s.name).filter(Boolean).join(", ")
  return `${name}${city ? ` in ${city}` : ""}. ${services}.`.trim()
}

function absoluteImageUrl(path: string | undefined, slug: string): string | undefined {
  if (!path) return undefined
  if (/^https?:\/\//.test(path)) return path
  if (!BASE_DOMAIN) return undefined
  return `https://${slug}.${BASE_DOMAIN}${path.startsWith("/") ? path : `/${path}`}`
}

export function meta({
  loaderData,
  location,
}: {
  loaderData?: { slug?: string; salon?: SalonData | null }
  location?: { pathname?: string }
}) {
  const slug = loaderData?.slug ?? ""
  const salon = loaderData?.salon ?? null

  if (!slug || !salon) {
    const homeUrl = `https://${BASE_DOMAIN || "fejd.fyi"}`
    return [
      { title: "fejd - Book haircut appointments" },
      { name: "description", content: "Book haircut appointments at salons near you." },
      {
        "script:ld+json": {
          "@context": "https://schema.org",
          "@type": "WebSite",
          name: "fejd",
          url: homeUrl,
          publisher: { "@type": "Organization", name: "fejd", url: homeUrl },
        },
      },
    ]
  }

  const title = salonTitle(salon)
  const description = salonDescription(salon)
  const path = location?.pathname ?? "/"
  const url = BASE_DOMAIN ? `https://${slug}.${BASE_DOMAIN}${path}` : null
  const rootUrl = BASE_DOMAIN ? `https://${slug}.${BASE_DOMAIN}` : null
  const image = absoluteImageUrl(salon.images?.hero ?? salon.images?.logo, slug)

  return [
    { title },
    { name: "description", content: description },
    { property: "og:title", content: title },
    { property: "og:description", content: description },
    { property: "og:type", content: "website" },
    ...(url ? [{ property: "og:url", content: url }] : []),
    ...(image ? [{ property: "og:image", content: image }] : []),
    { name: "twitter:card", content: "summary_large_image" },
    { name: "twitter:title", content: title },
    { name: "twitter:description", content: description },
    ...(image ? [{ name: "twitter:image", content: image }] : []),
    ...(url ? [{ tagName: "link", rel: "canonical", href: url }] : []),
    {
      "script:ld+json": {
        "@context": "https://schema.org",
        "@type": "HairSalon",
        name: salon.business?.name,
        ...(rootUrl ? { url: rootUrl } : {}),
        ...(image ? { image } : {}),
        address: {
          "@type": "PostalAddress",
          streetAddress: salon.business?.address_line,
          addressLocality: salon.business?.city,
          addressCountry: salon.business?.country,
        },
        telephone: salon.business?.phone,
        makesOffer: (salon.services ?? []).map((s) => ({
          "@type": "Offer",
          name: s.name,
          description: s.description,
          price: s.price,
          priceCurrency: "RSD",
        })),
      },
    },
  ]
}

export default function SalonRoute() {
  const { slug } = useLoaderData<typeof loader>()
  if (!slug) {
    return <Outlet />
  }
  return <SalonLayout slug={slug} />
}
