import { Outlet, useLoaderData } from "react-router"
import { SalonLayout } from "../../src/components/SalonLayout"
import { APP_HOST, BASE_DOMAIN, resolveSubdomainSlug } from "../../src/lib/salonDomain"
import { fetchPublic, prefetchPublic } from "../lib/serverData"
import { resolveQueryClient } from "../lib/context"
import { readLocaleCookie } from "../lib/locale"
import {
  hairSalonJsonLd,
  salonDescription,
  salonImageUrl,
  salonMeta,
  salonPageUrl,
  salonTitle,
  type SalonMetaData,
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
  const locale = readLocaleCookie(request)

  let salon: SalonMetaData | null = null
  let sections: unknown | null = null

  await prefetchPublic(qc, ["i18n", locale], `/api/i18n/${locale}`)
  if (slug) {
    ;[salon, sections] = await Promise.all([
      fetchPublic<SalonMetaData>(`/api/business/${slug}`).catch(() => null),
      fetchPublic(`/api/business/${slug}/sections`).catch(() => null),
    ])
    if (salon) qc.setQueryData(["salon", slug], salon)
    if (sections) qc.setQueryData(["sections", slug], sections)
  }
  return { slug, salon, sections }
}

export function meta({
  loaderData,
  location,
}: {
  loaderData?: { slug?: string; salon?: SalonMetaData | null }
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

  const path = location?.pathname ?? "/"
  const image = salonImageUrl(salon, slug)

  return salonMeta({
    title: salonTitle(salon),
    description: salonDescription(salon),
    url: salonPageUrl(slug, path),
    image,
    jsonLd: [hairSalonJsonLd(salon, slug, image)],
  })
}

export default function SalonRoute() {
  const { slug } = useLoaderData<typeof loader>()
  if (!slug) {
    return <Outlet />
  }
  return <SalonLayout slug={slug} />
}
