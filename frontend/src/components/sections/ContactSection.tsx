import { Mail, MapPin, Phone, Star } from "lucide-react"
import type { ContactContent } from "../../lib/sections"
import { instagramHandle, instagramUrl } from "../../lib/instagram"
import { openExternalUrl } from "../../lib/externalUrl"
import { InstagramIcon } from "../icons/InstagramIcon"
import { MapEmbed } from "./MapEmbed"

export function ContactSection({ content }: { content: ContactContent }) {
  const igUrl = content.instagram_url ? instagramUrl(content.instagram_url) : ""
  const igHandle = content.instagram_url ? instagramHandle(content.instagram_url) : ""

  const hasContent = Boolean(
    content.heading ||
      content.phone ||
      content.email ||
      content.address ||
      content.rating_url ||
      igUrl,
  )
  if (!hasContent) return null

  return (
    <section className="space-y-4">
      {content.heading && (
        <h3 className="text-xl font-semibold text-foreground">{content.heading}</h3>
      )}
      <div className="flex flex-col gap-6 md:flex-row md:items-start">
        <div className="space-y-4 md:max-w-[50%]">
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
                <a
                  href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(content.address)}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  {content.address}
                </a>
              </li>
            )}
            {igUrl && (
              <li className="flex items-center gap-2">
                <InstagramIcon className="size-4" />
                <a
                  href={igUrl}
                  target="_blank"
                  rel="noreferrer"
                  onClick={(event) => {
                    event.preventDefault()
                    void openExternalUrl(igUrl)
                  }}
                >
                  {igHandle}
                </a>
              </li>
            )}
          </ul>
          {content.rating_url && (
            <a
              href={content.rating_url}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1.5 rounded-2xl border border-border bg-background px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
            >
              <Star className="size-4" />
              Leave a rating
            </a>
          )}
        </div>
        {content.address && (
          <div className="w-full md:min-w-[50%] md:flex-1">
            <MapEmbed address={content.address} />
          </div>
        )}
      </div>
    </section>
  )
}
