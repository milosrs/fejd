import { Link, Navigate, useNavigate, useParams } from "react-router"
import { Clock } from "lucide-react"
import { useSalonContext } from "../context/SalonContext"
import { useServices } from "../hooks/useApi"
import { useI18n } from "../lib/i18n"
import { salonPath } from "../lib/salonDomain"
import { resolveImageUrl } from "../lib/images"
import { Button } from "../components/ui/button"
import { Card, CardContent } from "../components/ui/card"
import { Loader } from "../components/Loader"

// ServiceDetailPage renders a single service by its slug. It exists so each
// service has a stable, human-readable URL (/services/{serviceSlug}) that can be
// indexed for queries like "fade haircut <city>".
export function ServiceDetailPage() {
  const { serviceSlug } = useParams<{ serviceSlug: string }>()
  const { slug, salon, isLoading: salonLoading } = useSalonContext()
  const { data: services, isLoading: servicesLoading } = useServices(slug)
  const navigate = useNavigate()
  const { t } = useI18n()

  if (salonLoading || servicesLoading) {
    return <Loader />
  }

  const service = (services ?? []).find((s) => s.slug === serviceSlug)

  if (!salon || !service) {
    return <Navigate to={salonPath(slug, "/services")} replace />
  }

  const image = service.picture_id
    ? resolveImageUrl(`/api/images/${service.picture_id}`)
    : undefined

  return (
    <div className="space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link to={salonPath(slug, "/services")} className="hover:text-foreground">
          {t("services.title")}
        </Link>
      </nav>

      <Card>
        {image && (
          <img
            src={image}
            alt={service.name}
            className="h-56 w-full rounded-t-2xl object-cover"
          />
        )}
        <CardContent className="space-y-4 p-6">
          <div>
            <h1 className="text-2xl font-semibold text-foreground">{service.name}</h1>
            <p className="text-sm text-muted-foreground">{salon.business.name}</p>
          </div>

          {service.description && (
            <p className="text-foreground">{service.description}</p>
          )}

          <div className="flex items-center gap-4 text-sm">
            <span className="flex items-center gap-1 text-muted-foreground">
              <Clock className="size-4" />
              {service.duration_minutes} {t("booking.minutes")}
            </span>
            {service.price != null && service.price > 0 && (
              <span className="font-medium text-foreground">
                ${service.price.toFixed(2)}
              </span>
            )}
          </div>

          <Button
            className="w-full"
            onClick={() => navigate(`${salonPath(slug, "/book")}/${service.slug}`)}
          >
            {t("services.bookNow")}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
