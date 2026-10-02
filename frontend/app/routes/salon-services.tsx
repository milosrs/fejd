import { ServicesPage } from "../../src/pages/ServicesPage"
import { APP_HOST, BASE_DOMAIN, resolveSubdomainSlug } from "../../src/lib/salonDomain"
import { fetchPublic } from "../lib/serverData"
import { resolveQueryClient } from "../lib/context"
import {
  breadcrumbJsonLd,
  hairSalonJsonLd,
  salonCity,
  salonFromMatches,
  salonImageUrl,
  salonMeta,
  salonName,
  salonPageUrl,
  salonRootUrl,
  type SalonMetaService,
} from "../lib/salonMeta"

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

  let services: SalonMetaService[] | null = null
  if (slug) {
    services = await fetchPublic<SalonMetaService[]>(`/api/business/${slug}/services`).catch(
      () => null,
    )
    if (services) qc.setQueryData(["services", slug], services)
  }
  return { slug, services }
}

export function meta({
  loaderData,
  location,
  matches,
}: {
  loaderData?: { slug?: string; services?: SalonMetaService[] | null }
  location?: { pathname?: string }
  matches?: unknown
}) {
  const slug = loaderData?.slug ?? ""
  const services = loaderData?.services ?? null
  const salon = salonFromMatches(matches)

  if (!slug || !salon) {
    return [{ title: "Services - fejd" }]
  }

  const name = salonName(salon)
  const city = salonCity(salon)
  const active = (services ?? salon?.services ?? []).filter((s) => s.active !== false)
  const serviceNames = active.map((s) => s.name).filter(Boolean).join(", ")
  const image = salonImageUrl(salon, slug)
  const description = [
    `Services at ${name}${city ? ` in ${city}` : ""}`,
    serviceNames,
    "Book online.",
  ]
    .filter(Boolean)
    .join(". ")

  return salonMeta({
    title: `${name} — Services${city ? ` — u ${city}` : ""}`.trim(),
    description,
    url: salonPageUrl(slug, location?.pathname ?? "/services"),
    image,
    jsonLd: [
      hairSalonJsonLd(salon, slug, image),
      breadcrumbJsonLd([
        { name, url: salonRootUrl(slug) },
        { name: "Services", url: salonPageUrl(slug, "/services") },
      ]),
    ],
  })
}

export { ServicesPage as default }
