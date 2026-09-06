import { useNavigate } from "react-router-dom"
import { useSalonContext } from "../context/SalonContext"
import { useBookingStore } from "../stores/bookingStore"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "../components/ui/card"
import { Clock } from "lucide-react"

export function ServicesPage() {
  const { slug, salon } = useSalonContext()
  const navigate = useNavigate()
  const setService = useBookingStore((s) => s.setService)
  const reset = useBookingStore((s) => s.reset)

  const services = (salon?.services ?? []).filter((s) => s.active)

  const handleSelectService = (serviceId: string) => {
    reset()
    setService(serviceId)
    navigate(`/${slug}/book`)
  }

  return (
    <section className="space-y-6">
      <h2 className="text-xl font-semibold text-foreground">Services</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
        {services.map((service) => (
          <Card
            key={service.id}
            className="cursor-pointer transition-all hover:shadow-md hover:border-primary"
            onClick={() => handleSelectService(service.id)}
          >
            <CardHeader>
              <CardTitle>{service.name}</CardTitle>
              <CardDescription className="flex items-center gap-4">
                <span className="flex items-center gap-1">
                  <Clock className="size-3" />
                  {service.duration_minutes} min
                </span>
                {service.price != null && service.price > 0 && <span>${service.price.toFixed(2)}</span>}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant="outline" className="w-full">
                Book
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
      {services.length === 0 && (
        <p className="text-muted-foreground text-center py-12">No services available at this time.</p>
      )}
    </section>
  )
}
