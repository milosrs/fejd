import { useEffect, useState } from "react"

interface MapEmbedProps {
  address: string
  className?: string
}

const defaultClass = "h-64 w-full rounded-xl border border-border"

// MapEmbed renders a keyless map centered on the given address using
// OpenStreetMap. It geocodes the address via Nominatim (free, no API key) and
// embeds an OSM map with a marker at the result.
export function MapEmbed({ address, className }: MapEmbedProps) {
  const [coords, setCoords] = useState<{ lat: number; lon: number } | null>(null)

  useEffect(() => {
    let disposed = false
    setCoords(null)

    const timer = setTimeout(() => {
      fetch(
        `https://nominatim.openstreetmap.org/search?format=json&limit=1&q=${encodeURIComponent(address)}`,
      )
        .then((res) => res.json())
        .then((results) => {
          if (disposed || !Array.isArray(results) || results.length === 0) return
          setCoords({
            lat: parseFloat(results[0].lat),
            lon: parseFloat(results[0].lon),
          })
        })
        .catch(() => {})
    }, 400)

    return () => {
      disposed = true
      clearTimeout(timer)
    }
  }, [address])

  const cls = className ?? defaultClass

  if (!coords) {
    return <div className={cls} />
  }

  const { lat, lon } = coords
  const d = 0.006
  const src =
    `https://www.openstreetmap.org/export/embed.html` +
    `?bbox=${lon - d}%2C${lat - d}%2C${lon + d}%2C${lat + d}` +
    `&layer=mapnik&marker=${lat}%2C${lon}`

  return <iframe title="Salon location" src={src} className={cls} loading="lazy" />
}
