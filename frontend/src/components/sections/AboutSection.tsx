import type { AboutContent } from "../../lib/sections"

export function AboutSection({ content }: { content: AboutContent }) {
  if (!content.heading && !content.body) return null

  return (
    <section className="space-y-3">
      {content.heading && (
        <h3 className="text-xl font-semibold text-foreground">{content.heading}</h3>
      )}
      {content.body && (
        <p className="whitespace-pre-line text-muted-foreground">{content.body}</p>
      )}
    </section>
  )
}
