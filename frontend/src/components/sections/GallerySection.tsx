import { useEffect, useState } from "react"
import { ChevronLeft, ChevronRight, X } from "lucide-react"
import type { GalleryContent } from "../../lib/sections"
import { resolveImageUrl } from "../../lib/images"

export function GallerySection({
  content,
  contained,
}: {
  content: GalleryContent
  contained?: boolean
}) {
  const images = (content.image_urls ?? [])
    .map(resolveImageUrl)
    .filter((url): url is string => Boolean(url))

  const [activeIndex, setActiveIndex] = useState<number | null>(null)

  useEffect(() => {
    if (activeIndex === null) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setActiveIndex(null)
      if (e.key === "ArrowLeft") {
        setActiveIndex((i) => (i === null ? i : (i - 1 + images.length) % images.length))
      }
      if (e.key === "ArrowRight") {
        setActiveIndex((i) => (i === null ? i : (i + 1) % images.length))
      }
    }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [activeIndex, images.length])

  if (!content.heading && images.length === 0) return null

  const step = (dir: -1 | 1) => {
    setActiveIndex((i) => (i === null ? i : (i + dir + images.length) % images.length))
  }

  return (
    <section className="space-y-4">
      {content.heading && (
        <h3 className="text-xl font-semibold text-foreground">{content.heading}</h3>
      )}
      {images.length > 0 && (
        <div
          className={`flex gap-3 overflow-x-auto pb-2 snap-x [justify-content:safe_center] ${
            contained ? "" : "relative left-1/2 w-[100dvw] -translate-x-1/2 px-4"
          }`}
        >
          {images.map((url, index) => (
            <button
              key={url}
              type="button"
              onClick={() => setActiveIndex(index)}
              className="shrink-0 snap-start"
            >
              <img
                src={url}
                alt=""
                className="h-56 w-56 rounded-xl border border-border object-cover sm:h-72 sm:w-72"
              />
            </button>
          ))}
        </div>
      )}

      {activeIndex !== null && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 pt-[calc(1rem+env(safe-area-inset-top))] pb-[calc(1rem+env(safe-area-inset-bottom))]"
          onClick={() => setActiveIndex(null)}
        >
          <button
            type="button"
            aria-label="Close"
            onClick={() => setActiveIndex(null)}
            className="absolute right-4 top-[calc(1rem+env(safe-area-inset-top))] z-10 flex h-10 w-10 items-center justify-center rounded-full bg-black/60 text-white hover:bg-black/80"
          >
            <X className="size-6" />
          </button>

          {images.length > 1 && (
            <button
              type="button"
              aria-label="Previous image"
              onClick={(e) => {
                e.stopPropagation()
                step(-1)
              }}
              className="absolute left-4 z-10 flex h-10 w-10 items-center justify-center rounded-full bg-black/60 text-white hover:bg-black/80"
            >
              <ChevronLeft className="size-6" />
            </button>
          )}

          <img
            src={images[activeIndex]}
            alt=""
            onClick={(e) => e.stopPropagation()}
            className="max-h-[85vh] max-w-[92vw] rounded-lg object-contain"
          />

          {images.length > 1 && (
            <button
              type="button"
              aria-label="Next image"
              onClick={(e) => {
                e.stopPropagation()
                step(1)
              }}
              className="absolute right-4 z-10 flex h-10 w-10 items-center justify-center rounded-full bg-black/60 text-white hover:bg-black/80"
            >
              <ChevronRight className="size-6" />
            </button>
          )}
        </div>
      )}
    </section>
  )
}
