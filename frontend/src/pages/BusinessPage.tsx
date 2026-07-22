import { useParams, useNavigate } from "react-router-dom"
import { useBusiness } from "../hooks/useApi"
import { useBookingStore } from "../stores/bookingStore"
import { Button } from "../components/ui/button"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "../components/ui/card"
import { Clock, User } from "lucide-react"

export function BusinessPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { data, isLoading } = useBusiness(slug!)
  const setService = useBookingStore((s) => s.setService)
  const reset = useBookingStore((s) => s.reset)

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!data) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <p className="text-muted-foreground">Business not found</p>
      </div>
    )
  }

  const handleSelectService = (serviceId: string) => {
    reset()
    setService(serviceId)
    navigate(`/business/${slug}/book`)
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-6">
          <h1 className="text-2xl font-bold text-foreground">{data.business.name}</h1>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <h2 className="text-xl font-semibold mb-6 text-foreground">Select a service</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {data.services
            .filter((s) => s.active)
            .map((service) => (
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
        {data.services.filter((s) => s.active).length === 0 && (
          <p className="text-muted-foreground text-center py-12">No services available at this time.</p>
        )}
      </main>
    </div>
  )
}
