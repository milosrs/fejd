import { useSalonContext } from "../context/SalonContext"
import { useSections } from "../hooks/useSections"
import { useI18n } from "../lib/i18n"
import { isSectionType } from "../lib/sections"
import { SectionRenderer } from "../components/sections/SectionRenderer"

export function LandingPage() {
  const { slug, salon } = useSalonContext()
  const { t, ready } = useI18n()
  const { data: sections, isLoading } = useSections(slug)

  if (!salon) return null

  const list = sections ?? []
  const hasRenderable = list.some((s) => isSectionType(s.type))

  if (isLoading || !ready) return null

  return (
    <section className="space-y-8">
      {list.length === 0 || !hasRenderable ? (
        <div className="flex flex-col items-center gap-2 py-16 text-center">
          <h2 className="text-lg font-semibold text-foreground">
            {t("landing.empty.title")}
          </h2>
          <p className="max-w-sm text-muted-foreground">{t("landing.empty.body")}</p>
        </div>
      ) : (
        list.map((section) => <SectionRenderer key={section.id} section={section} />)
      )}
    </section>
  )
}
