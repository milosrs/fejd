import { Outlet, useLoaderData } from "react-router"
import { SalonLayout } from "../../src/components/SalonLayout"
import { APP_HOST, BASE_DOMAIN, resolveSubdomainSlug } from "../../src/lib/salonDomain"
import { fetchPublic, prefetchPublic } from "../lib/serverData"
import { resolveQueryClient } from "../lib/context"

type SalonData = {
  business?: {
    name?: string
    city?: string
    address_line?: string
    country?: string
    phone?: string
    image?: string
  }
  services?: Array<{ name?: string; slug?: string; description?: string; price?: number }>
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
  const slug = params.slug ?? resolveSubdomainSlug(hostname, BASE_DOMAIN, APP_HOST) ?? ""
  const qc = resolveQueryClient(context)

  let salon: SalonData | null = null
  let sections: unknown | null = null

  await prefetchPublic(qc, ["i18n", "en"], "/api/i18n/en")
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
    return [
      { title: "fejd - Book haircut appointments" },
      { name: "description", content: "Book haircut appointments at salons near you." },
    ]
  }

  const title = salonTitle(salon)
  const description = salonDescription(salon)
  const path = location?.pathname ?? "/"
  const url = `https://${slug}.${BASE_DOMAIN}${path}`

  return [
    { title },
    { name: "description", content: description },
    { property: "og:title", content: title },
    { property: "og:description", content: description },
    { property: "og:type", content: "website" },
    { property: "og:url", content: url },
    { name: "twitter:card", content: "summary_large_image" },
    { name: "twitter:title", content: title },
    { name: "twitter:description", content: description },
    { tagName: "link", rel: "canonical", href: url },
    { tagName: "link", rel: "alternate", hrefLang: "en", href: url },
    { tagName: "link", rel: "alternate", hrefLang: "sr", href: url },
    {
      "script:ld+json": {
        "@context": "https://schema.org",
        "@type": "HairSalon",
        name: salon.business?.name,
        url: `https://${slug}.${BASE_DOMAIN}`,
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
