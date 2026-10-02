import { SalonPolicyPage } from "../../src/pages/SalonPolicyPage"

// Policy is an owner-only settings page, not public content for search.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default SalonPolicyPage
