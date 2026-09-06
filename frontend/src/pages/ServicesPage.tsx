import { useNavigate } from "react-router-dom"
import { useSalonContext } from "../context/SalonContext"
import { useServices } from "../hooks/useApi"
import { useBookingStore } from "../stores/bookingStore"
import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"
import { ServiceCard, ServiceCardSkeleton } from "../components/services/ServiceCard"

export function ServicesPage() {
  const { slug } = useSalonContext()
  const navigate = useNavigate()
  const setService = useBookingStore((s) => s.setService)
  const reset = useBookingStore((s) => s.reset)
  const authenticated = useAuthStore((s) => s.authenticated)
  const login = useAuthStore((s) => s.login)
  const { t, ready } = useI18n()

  const { data, isLoading, isError } = useServices(slug)

  const services = (data ?? []).filter((s) => s.active)

  const registerNote = !authenticated && ready ? t("services.book.requiresAuth") : undefined

  const handleBook = (serviceId: string) => {
    if (!authenticated) {
      login()
      return
    }
    reset()
    setService(serviceId)
    navigate(`/${slug}/book?service=${serviceId}`)
  }

  return (
    <section className="space-y-6">
      <h2 className="text-xl font-semibold text-foreground">Services</h2>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <ServiceCardSkeleton key={i} />
          ))}
        </div>
      ) : isError ? (
        <p className="py-12 text-center text-muted-foreground">
          Couldn't load services. Please try again.
        </p>
      ) : services.length === 0 ? (
        <p className="py-12 text-center text-muted-foreground">
          No services available at this time.
        </p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {services.map((service) => (
            <ServiceCard
              key={service.id}
              service={service}
              onBook={handleBook}
              note={registerNote}
            />
          ))}
        </div>
      )}
    </section>
  )
}
