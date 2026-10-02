import { NotFoundPage } from "../../src/pages/NotFoundPage"

// Error pages must not be indexed.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default NotFoundPage
