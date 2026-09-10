import type { AboutContent } from "../../lib/sections"

export function AboutSection({ content }: { content: AboutContent }) {
  if (!content.heading && !content.body && !content.address) return null

  return (
    <section className="space-y-3">
      {content.heading && (
        <h3 className="text-xl font-semibold text-foreground">{content.heading}</h3>
      )}
      {content.body && (
        <p className="whitespace-pre-line text-muted-foreground">{content.body}</p>
      )}
      {content.address && (
        <iframe
          title="Salon location"
          src={`https://www.google.com/maps?q=${encodeURIComponent(content.address)}&output=embed`}
          className="h-64 w-full rounded-xl border border-border"
          loading="lazy"
          referrerPolicy="no-referrer-when-downgrade"
        />
      )}
    </section>
  )
}
