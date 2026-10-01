import { BookingPage } from "../../src/pages/BookingPage"

// Booking is a transactional flow, not a landing page for search.
export function headers() {
  return { "X-Robots-Tag": "noindex" }
}

export default BookingPage
