import type { GalleryContent } from "../../lib/sections"
import { resolveImageUrl } from "../../lib/images"

export function GallerySection({ content }: { content: GalleryContent }) {
  const images = (content.image_urls ?? [])
    .map(resolveImageUrl)
    .filter((url): url is string => Boolean(url))

  if (!content.heading && images.length === 0) return null

  return (
    <section className="space-y-4">
      {content.heading && (
        <h3 className="text-xl font-semibold text-foreground">{content.heading}</h3>
      )}
      {images.length > 0 && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          {images.map((url) => (
            <img
              key={url}
              src={url}
              alt=""
              className="aspect-square w-full rounded-lg object-cover border border-border"
            />
          ))}
        </div>
      )}
    </section>
  )
}
