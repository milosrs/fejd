import { useEffect, useRef } from "react"
import "leaflet/dist/leaflet.css"

interface MapEmbedProps {
  address: string
  className?: string
}

const defaultClass = "h-64 w-full overflow-hidden rounded-xl border border-border"

// CARTO Voyager is a free, keyless, stylized basemap built on OpenStreetMap
// data. It looks much cleaner than the stock OSM "mapnik" layer.
const cartoTiles =
  "https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}.png"

const attribution =
  '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'

// MapEmbed renders a keyless, interactive map centered on the given address.
// Leaflet is imported lazily inside the effect so it never loads server-side
// (Leaflet reads `window` at module load and would crash SSR).
export function MapEmbed({ address, className }: MapEmbedProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<unknown>(null)
  const markerRef = useRef<unknown>(null)

  useEffect(() => {
    let disposed = false

    const timer = setTimeout(async () => {
      const L = (await import("leaflet")).default as typeof import("leaflet")
      if (disposed) return

      try {
        const res = await fetch(
          `https://nominatim.openstreetmap.org/search?format=json&limit=1&q=${encodeURIComponent(address)}`,
        )
        const results = (await res.json()) as Array<{ lat: string; lon: string }>
        if (disposed || !Array.isArray(results) || results.length === 0 || !containerRef.current) {
          return
        }
        const lat = parseFloat(results[0].lat)
        const lon = parseFloat(results[0].lon)
        if (Number.isNaN(lat) || Number.isNaN(lon)) return

        const map = mapRef.current as import("leaflet").Map | null
        if (!map) {
          mapRef.current = L.map(containerRef.current, {
            center: [lat, lon],
            zoom: 15,
            scrollWheelZoom: false,
          })
          L.tileLayer(cartoTiles, {
            attribution,
            subdomains: "abcd",
            maxZoom: 19,
          }).addTo(mapRef.current as import("leaflet").Map)
        } else {
          map.setView([lat, lon], 15)
        }

        const m = markerRef.current as import("leaflet").CircleMarker | null
        m?.remove()
        markerRef.current = L.circleMarker([lat, lon], {
          radius: 9,
          color: "#ffffff",
          weight: 3,
          fillColor: "#e11d48",
          fillOpacity: 1,
        }).addTo(mapRef.current as import("leaflet").Map)
        ;(markerRef.current as import("leaflet").CircleMarker).bindTooltip(address)
      } catch {
        // Best-effort: leave the placeholder if geocoding or map init fails.
      }
    }, 300)

    return () => {
      disposed = true
      clearTimeout(timer)
    }
  }, [address])

  useEffect(() => {
    return () => {
      markerRef.current = null
      ;(mapRef.current as import("leaflet").Map | null)?.remove()
      mapRef.current = null
    }
  }, [])

  return <div ref={containerRef} className={className ?? defaultClass} />
}
