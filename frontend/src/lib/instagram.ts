// Instagram helpers: turn a stored Instagram link (or bare handle) into a
// canonical profile URL and a display handle like "@fejd".

export function instagramUrl(value: string): string {
  const trimmed = value.trim()
  if (!trimmed) return ""
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  if (/^(www\.)?instagram\.com\//i.test(trimmed)) return `https://${trimmed}`
  return `https://www.instagram.com/${trimmed.replace(/^@/, "").replace(/^\/+|\/+$/g, "")}`
}

export function instagramHandle(value: string): string {
  const url = instagramUrl(value)
  if (!url) return ""
  try {
    const handle = new URL(url).pathname.split("/").filter(Boolean)[0]
    return handle ? `@${handle}` : ""
  } catch {
    return ""
  }
}
