import { SalonSettingsPage } from "../../src/pages/SalonSettingsPage"

// Settings are owner-only and must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default SalonSettingsPage
