import { ServiceDetailPage } from "../../src/pages/ServiceDetailPage"
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
  serviceJsonLd,
  type SalonMetaService,
} from "../lib/salonMeta"

export async function loader({
  request,
  params,
  context,
}: {
  request: Request
  params: { slug?: string; serviceSlug?: string }
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
  return { slug, serviceSlug: params.serviceSlug ?? "", services }
}

export function meta({
  loaderData,
  matches,
}: {
  loaderData?: { slug?: string; serviceSlug?: string; services?: SalonMetaService[] | null }
  matches?: unknown
}) {
  const slug = loaderData?.slug ?? ""
  const serviceSlug = loaderData?.serviceSlug ?? ""
  const services = loaderData?.services ?? null
  const salon = salonFromMatches(matches)

  const service = (services ?? salon?.services ?? []).find((s) => s.slug === serviceSlug)

  if (!slug || !salon || !service) {
    return [{ title: "Service - fejd" }]
  }

  const name = salonName(salon)
  const city = salonCity(salon)
  const image = salonImageUrl(salon, slug)
  const serviceUrl = salonPageUrl(slug, `/services/${serviceSlug}`)
  const description =
    service.description?.trim() ||
    `${service.name} at ${name}${city ? ` in ${city}` : ""}. Book online.`.trim()

  return salonMeta({
    title: `${service.name} — ${name}${city ? ` — u ${city}` : ""}`.trim(),
    description,
    url: serviceUrl,
    image,
    jsonLd: [
      hairSalonJsonLd(salon, slug, image),
      serviceJsonLd(salon, service, slug),
      breadcrumbJsonLd([
        { name, url: salonRootUrl(slug) },
        { name: "Services", url: salonPageUrl(slug, "/services") },
        { name: service.name ?? "Service", url: serviceUrl },
      ]),
    ],
  })
}

export { ServiceDetailPage as default }
