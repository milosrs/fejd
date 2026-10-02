// Server-side locale resolution for SSR. The locale is stored in the
// `fejd-locale` cookie by the client's I18nProvider, so the server can render
// the right language without a hydration mismatch.
export function readLocaleCookie(request: Request): string {
  const cookie = request.headers.get("cookie") ?? ""
  const match = cookie.match(/(?:^|;\s*)fejd-locale=([^;]+)/)
  const locale = match?.[1]
  return locale === "en" || locale === "rs" ? locale : "en"
}
