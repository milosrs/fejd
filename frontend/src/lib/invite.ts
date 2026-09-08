/**
 * Parses an invite deep link into its raw token.
 *
 * Accepts `https://app.example.com/invite/<token>` (and any scheme/host), and
 * returns `<token>`; returns null for non-invite URLs.
 */
export function parseInviteToken(url: string): string | null {
  try {
    const u = new URL(url)
    const parts = u.pathname.split("/").filter(Boolean)
    if (parts[0] === "invite" && parts[1]) {
      return parts[1]
    }
  } catch {
    return null
  }
  return null
}
