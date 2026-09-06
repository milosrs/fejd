import { useSalonContext } from "../context/SalonContext"

export function LandingPage() {
  const { salon } = useSalonContext()

  if (!salon) return null

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">{salon.business.name}</h1>
        <p className="text-muted-foreground mt-2">
          Landing page content lands here (Chunk 2 — public render).
        </p>
      </div>
    </section>
  )
}
