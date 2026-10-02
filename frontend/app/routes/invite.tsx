import { InviteLandingPage } from "../../src/components/invite/InviteLandingPage"

// Invite links are personalized and must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default InviteLandingPage
