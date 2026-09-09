import { Capacitor } from "@capacitor/core"

// BASE_DOMAIN is the shared parent domain, e.g. "fejd.com". Salon sites are
// served at "<slug>.fejd.com"; leave it unset in local dev to keep path-based
// routing (/<slug>). APP_HOST is the app shell host (e.g. "www.fejd.com") that
// must never be interpreted as a salon subdomain.
export const BASE_DOMAIN = (import.meta.env.VITE_APP_DOMAIN || "").trim().toLowerCase()
export const APP_HOST = (import.meta.env.VITE_APP_HOST || "").trim().toLowerCase()

function currentHostname(): string {
  if (typeof window === "undefined") return ""
  return window.location.hostname.toLowerCase()
}

// isSalonSubdomain reports whether the current host is a salon subdomain of
// BASE_DOMAIN (as opposed to the app shell, native, or local dev).
export function isSalonSubdomain(hostname: string = currentHostname()): boolean {
  return subdomainSlug(hostname) != null
}

// subdomainSlug returns the salon slug derived from the hostname, or null when
// the host is the app shell, the apex, native, local dev, or a host that is not
// a single-label subdomain of BASE_DOMAIN.
export function subdomainSlug(hostname: string = currentHostname()): string | null {
  if (Capacitor.isNativePlatform()) return null
  return resolveSubdomainSlug(hostname, BASE_DOMAIN, APP_HOST)
}

// resolveSubdomainSlug is the pure hostname-parsing core (no env/platform
// dependency) so it can be unit-tested directly.
export function resolveSubdomainSlug(hostname: string, baseDomain: string, appHost = ""): string | null {
  const host = hostname.trim().toLowerCase()
  const base = baseDomain.trim().toLowerCase()
  if (!base) return null
  if (host === base) return null
  if (appHost && host === appHost.trim().toLowerCase()) return null
  if (!host.endsWith("." + base)) return null

  const label = host.slice(0, host.length - base.length - 1)
  if (label.length === 0 || label.includes(".")) return null
  return label
}

// salonUrl returns the absolute URL of a salon's site. On native or in local
// dev it falls back to the path-based route.
export function salonUrl(slug: string): string {
  if (Capacitor.isNativePlatform() || !BASE_DOMAIN) return `/${slug}`
  return `https://${slug}.${BASE_DOMAIN}`
}

// salonPath returns the internal route path for a salon page, correct for both
// subdomain mode (path is returned as-is, rooted at "/") and path mode (path is
// prefixed with "/<slug>"). path must start with "/" ("" means the salon root).
export function salonPath(slug: string, path = ""): string {
  if (isSalonSubdomain()) return path === "" ? "/" : path
  return path === "" ? `/${slug}` : `/${slug}${path}`
}

// openSalon navigates to a salon: a full-page navigation to the subdomain on
// web, and an internal router navigation on native/local dev.
export function openSalon(navigate: (to: string) => void, slug: string): void {
  if (!Capacitor.isNativePlatform() && BASE_DOMAIN) {
    window.location.assign(salonUrl(slug))
    return
  }
  navigate(`/${slug}`)
}
