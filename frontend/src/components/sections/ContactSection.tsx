import { Mail, MapPin, Phone, Star } from "lucide-react"
import type { ContactContent } from "../../lib/sections"
import { instagramHandle, instagramUrl } from "../../lib/instagram"
import { openExternalUrl } from "../../lib/externalUrl"
import { useI18n } from "../../lib/i18n"
import { useSalonContext } from "../../context/SalonContext"
import { InstagramIcon } from "../icons/InstagramIcon"
import { MapEmbed } from "./MapEmbed"

// Address shown in the Contact section. Prefers the structured salon location
// (set via the Location dialog), falling back to the legacy contact-section
// address for salons that have not been migrated yet.
function addressFor(
  content: ContactContent,
  salon: { business: { address_line?: string; city?: string; country?: string } } | undefined,
): string {
  const structured = [
    salon?.business.address_line,
    salon?.business.city,
    salon?.business.country,
  ]
    .filter(Boolean)
    .join(", ")
  return structured || content.address || ""
}

// Phone shown in the Contact section. Prefers the structured salon phone (set
// via the Location dialog), falling back to the legacy contact-section phone.
function phoneFor(
  content: ContactContent,
  salon: { business: { phone?: string } } | undefined,
): string {
  return salon?.business.phone || content.phone || ""
}

export function ContactSection({ content }: { content: ContactContent }) {
  const { t } = useI18n()
  const { salon } = useSalonContext()
  const igUrl = content.instagram_url ? instagramUrl(content.instagram_url) : ""
  const igHandle = content.instagram_url ? instagramHandle(content.instagram_url) : ""
  const address = addressFor(content, salon)
  const phone = phoneFor(content, salon)

  const hasContent = Boolean(
    content.heading ||
      phone ||
      content.email ||
      address ||
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
            {phone && (
              <li className="flex items-center gap-2">
                <Phone className="size-4" />
                <a href={`tel:${phone}`}>{phone}</a>
              </li>
            )}
            {content.email && (
              <li className="flex items-center gap-2">
                <Mail className="size-4" />
                <a href={`mailto:${content.email}`}>{content.email}</a>
              </li>
            )}
            {address && (
              <li className="flex items-center gap-2">
                <MapPin className="size-4" />
                <a
                  href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  {address}
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
              {t("sections.contact.leaveRating")}
            </a>
          )}
        </div>
        {address && (
          <div className="w-full md:min-w-[50%] md:flex-1">
            <MapEmbed address={address} />
          </div>
        )}
      </div>
    </section>
  )
}
