import { Trash2, UserRound } from "lucide-react"
import { Card, CardHeader, CardTitle, CardContent } from "../ui/card"
import { Button } from "../ui/button"
import { resolveImageUrl } from "../../lib/images"
import type { Employee } from "../../hooks/useApi"

export function BarberCard({
  employee,
  onRemove,
}: {
  employee: Employee
  onRemove?: () => void
}) {
  const avatar = resolveImageUrl(employee.avatar)
  const name = employee.display_name || employee.user_id

  return (
    <Card className="relative items-center text-center">
      {onRemove && (
        <Button
          variant="ghost"
          size="icon-xs"
          className="absolute right-2 top-2 text-destructive"
          onClick={onRemove}
        >
          <Trash2 />
        </Button>
      )}
      <CardContent className="pt-(--card-spacing)">
        {avatar ? (
          <img
            src={avatar}
            alt={name}
            className="mx-auto h-20 w-20 rounded-full object-cover ring-2 ring-border"
          />
        ) : (
          <div className="mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-muted text-muted-foreground ring-2 ring-border">
            <UserRound className="size-8" />
          </div>
        )}
      </CardContent>
      <CardHeader className="items-center">
        <CardTitle>{name}</CardTitle>
      </CardHeader>
    </Card>
  )
}

export function BarberCardSkeleton() {
  return (
    <Card className="items-center text-center">
      <CardContent className="pt-(--card-spacing)">
        <div className="mx-auto h-20 w-20 animate-pulse rounded-full bg-muted" />
      </CardContent>
      <CardHeader className="items-center">
        <div className="h-4 w-24 animate-pulse rounded bg-muted" />
      </CardHeader>
    </Card>
  )
}
