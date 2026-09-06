import { useSalonContext } from "../context/SalonContext"
import { Card, CardHeader, CardTitle, CardContent } from "../components/ui/card"

export function BarbersPage() {
  const { salon } = useSalonContext()
  const employees = salon?.employees ?? []

  return (
    <section className="space-y-6">
      <h2 className="text-xl font-semibold text-foreground">Barbers</h2>
      {employees.length === 0 ? (
        <p className="text-muted-foreground text-center py-12">No barbers listed yet.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {employees.map((emp) => (
            <Card key={emp.id}>
              <CardHeader>
                <CardTitle>{emp.display_name || "Barber"}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">Full profiles land in Chunk 6.</p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </section>
  )
}
