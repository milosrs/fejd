import { Mail, MapPin, Phone } from "lucide-react"
import type { ContactContent } from "../../lib/sections"

export function ContactSection({ content }: { content: ContactContent }) {
  const hasContent = Boolean(
    content.heading || content.phone || content.email || content.address,
  )
  if (!hasContent) return null

  return (
    <section className="space-y-4">
      {content.heading && (
        <h3 className="text-xl font-semibold text-foreground">{content.heading}</h3>
      )}
      <ul className="space-y-2 text-muted-foreground">
        {content.phone && (
          <li className="flex items-center gap-2">
            <Phone className="size-4" />
            <a href={`tel:${content.phone}`}>{content.phone}</a>
          </li>
        )}
        {content.email && (
          <li className="flex items-center gap-2">
            <Mail className="size-4" />
            <a href={`mailto:${content.email}`}>{content.email}</a>
          </li>
        )}
        {content.address && (
          <li className="flex items-center gap-2">
            <MapPin className="size-4" />
            <span>{content.address}</span>
          </li>
        )}
      </ul>
    </section>
  )
}
