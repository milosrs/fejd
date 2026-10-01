import { useLoaderData } from "react-router"
import { HomePage } from "../../src/pages/HomePage"
import { LandingPage } from "../../src/pages/LandingPage"
import { APP_HOST, BASE_DOMAIN, resolveSubdomainSlug } from "../../src/lib/salonDomain"
import { prefetchPublic } from "../lib/serverData"
import { resolveQueryClient } from "../lib/context"

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
  if (!slug) {
    await prefetchPublic(qc, ["businesses"], "/api/businesses")
  }
  return { slug }
}

const homeJsonLd = {
  "@context": "https://schema.org",
  "@type": "WebSite",
  name: "fejd",
  url: `https://${BASE_DOMAIN || "fejd.fyi"}`,
  publisher: {
    "@type": "Organization",
    name: "fejd",
    url: `https://${BASE_DOMAIN || "fejd.fyi"}`,
  },
}

export default function SalonIndex() {
  const { slug } = useLoaderData<typeof loader>()
  if (slug) {
    return <LandingPage />
  }
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(homeJsonLd) }}
      />
      <HomePage />
    </>
  )
}
