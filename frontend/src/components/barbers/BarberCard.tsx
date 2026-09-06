import { Trash2, UserRound } from "lucide-react"
import { Card, CardContent } from "../ui/card"
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
    <Card className="relative">
      {onRemove && (
        <Button
          variant="ghost"
          size="icon-xs"
          className="absolute right-2 top-2 z-10 text-destructive"
          onClick={onRemove}
        >
          <Trash2 />
        </Button>
      )}
      <CardContent className="flex flex-col items-center gap-3 pt-(--card-spacing)">
        {avatar ? (
          <img
            src={avatar}
            alt={name}
            className="h-20 w-20 rounded-full object-cover ring-2 ring-border"
          />
        ) : (
          <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted text-muted-foreground ring-2 ring-border">
            <UserRound className="size-8" />
          </div>
        )}
        <span className="text-center text-sm font-medium text-foreground">{name}</span>
      </CardContent>
    </Card>
  )
}

export function BarberCardSkeleton() {
  return (
    <Card>
      <CardContent className="flex flex-col items-center gap-3 pt-(--card-spacing)">
        <div className="h-20 w-20 animate-pulse rounded-full bg-muted" />
        <div className="h-4 w-24 animate-pulse rounded bg-muted" />
      </CardContent>
    </Card>
  )
}
