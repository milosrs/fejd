import { Clock, Pencil, Trash2 } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "../ui/card"
import { Button } from "../ui/button"
import { resolveImageUrl } from "../../lib/images"
import type { Service } from "../../hooks/useApi"

export function ServiceCard({
  service,
  imageOverride,
  onBook,
  note,
  onEdit,
  onDelete,
}: {
  service: Service
  imageOverride?: string
  onBook: (serviceId: string) => void
  note?: string
  onEdit?: () => void
  onDelete?: () => void
}) {
  const imageUrl =
    imageOverride ??
    (service.picture_id
      ? resolveImageUrl(`/api/images/${service.picture_id}`)
      : undefined)

  const editing = Boolean(onEdit && onDelete)

  return (
    <Card>
      {imageUrl && (
        <img src={imageUrl} alt={service.name} className="h-40 w-full object-cover" />
      )}
      <CardHeader>
        <CardTitle>{service.name}</CardTitle>
        {service.description && (
          <CardDescription className="line-clamp-2">{service.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent className="flex items-center justify-between">
        <span className="flex items-center gap-1 text-muted-foreground">
          <Clock className="size-3" />
          {service.duration_minutes} min
        </span>
        {service.price != null && service.price > 0 && (
          <span className="font-medium">${service.price.toFixed(2)}</span>
        )}
      </CardContent>
      <CardFooter className="flex-col items-stretch gap-2">
        {editing ? (
          <div className="flex gap-2">
            <Button variant="outline" className="flex-1" onClick={onEdit}>
              <Pencil className="size-3" /> Edit
            </Button>
            <Button variant="destructive" className="flex-1" onClick={onDelete}>
              <Trash2 className="size-3" /> Delete
            </Button>
          </div>
        ) : (
          <Button className="w-full" onClick={() => onBook(service.id)}>
            Book now
          </Button>
        )}
        {note && (
          <p className="text-center text-xs text-muted-foreground">{note}</p>
        )}
      </CardFooter>
    </Card>
  )
}

export function ServiceCardSkeleton() {
  return (
    <Card>
      <div className="h-40 w-full animate-pulse bg-muted" />
      <CardHeader>
        <div className="h-4 w-2/3 animate-pulse rounded bg-muted" />
        <div className="h-3 w-full animate-pulse rounded bg-muted" />
      </CardHeader>
      <CardContent>
        <div className="h-3 w-1/3 animate-pulse rounded bg-muted" />
      </CardContent>
      <CardFooter>
        <div className="h-8 w-full animate-pulse rounded-2xl bg-muted" />
      </CardFooter>
    </Card>
  )
}
