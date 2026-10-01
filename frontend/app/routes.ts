import { type RouteConfig, index, layout, route } from "@react-router/dev/routes"

// Salon site routes. The salon layout resolves the slug from the request host
// (subdomain mode) or the :slug param (path mode / local dev).
const salonRoutes = [
  index("routes/salon-index.tsx"),
  route("services", "routes/salon-services.tsx"),
  route("services/:serviceSlug", "routes/salon-service-detail.tsx"),
  route("barbers", "routes/salon-barbers.tsx"),
  route("book", "routes/salon-booking.tsx"),
  route("book/:serviceSlug", "routes/salon-booking.tsx", { id: "routes/salon-booking-service" }),
  route("policy", "routes/salon-policy.tsx"),
]

// Path-mode copy of the salon routes (/{slug}/*) used on local dev / hosts
// without wildcard subdomains. Distinct ids avoid clashing with subdomain mode.
const salonPathRoutes = [
  index("routes/salon-index.tsx", { id: "routes/salon-path-index" }),
  route("services", "routes/salon-services.tsx", { id: "routes/salon-path-services" }),
  route("services/:serviceSlug", "routes/salon-service-detail.tsx", { id: "routes/salon-path-service-detail" }),
  route("barbers", "routes/salon-barbers.tsx", { id: "routes/salon-path-barbers" }),
  route("book", "routes/salon-booking.tsx", { id: "routes/salon-path-booking" }),
  route("book/:serviceSlug", "routes/salon-booking.tsx", { id: "routes/salon-path-booking-service" }),
  route("policy", "routes/salon-policy.tsx", { id: "routes/salon-path-policy" }),
]

export default [
  // Salon site — subdomain mode (root-level on *.fejd.fyi).
  layout("routes/salon.tsx", salonRoutes),

  // Salon site — path mode (/{slug} on local dev).
  route(":slug", "routes/salon.tsx", { id: "routes/salon-path" }, salonPathRoutes),

  route("invite/:token", "routes/invite.tsx"),
  route("my/appointments", "routes/my-appointments.tsx"),
  route("admin/invited-customers", "routes/admin-invited-customers.tsx"),
  route("admin/business/:businessId/schedule", "routes/admin-schedule.tsx"),
  route("admin/business/:businessId/services", "routes/admin-services.tsx"),
  route("admin/business/:businessId/my-schedule", "routes/admin-my-schedule.tsx"),
  route("admin/business/:businessId/my-reservations", "routes/admin-my-reservations.tsx"),
] satisfies RouteConfig
