import { useEffect, useRef } from "react"
import L from "leaflet"
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

// MapEmbed renders a keyless, interactive map centered on the given address. It
// geocodes the address via Nominatim (free, no API key) and shows a marker at
// the result using Leaflet + CARTO Voyager tiles.
export function MapEmbed({ address, className }: MapEmbedProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<L.Map | null>(null)
  const markerRef = useRef<L.CircleMarker | null>(null)

  useEffect(() => {
    let disposed = false

    const timer = setTimeout(() => {
      fetch(
        `https://nominatim.openstreetmap.org/search?format=json&limit=1&q=${encodeURIComponent(address)}`,
      )
        .then((res) => res.json())
        .then((results) => {
          if (
            disposed ||
            !Array.isArray(results) ||
            results.length === 0 ||
            !containerRef.current
          ) {
            return
          }
          const lat = parseFloat(results[0].lat)
          const lon = parseFloat(results[0].lon)
          if (Number.isNaN(lat) || Number.isNaN(lon)) return

          if (!mapRef.current) {
            mapRef.current = L.map(containerRef.current, {
              center: [lat, lon],
              zoom: 15,
              scrollWheelZoom: false,
            })
            L.tileLayer(cartoTiles, {
              attribution,
              subdomains: "abcd",
              maxZoom: 19,
            }).addTo(mapRef.current)
          } else {
            mapRef.current.setView([lat, lon], 15)
          }

          markerRef.current?.remove()
          markerRef.current = L.circleMarker([lat, lon], {
            radius: 9,
            color: "#ffffff",
            weight: 3,
            fillColor: "#e11d48",
            fillOpacity: 1,
          }).addTo(mapRef.current)
          markerRef.current.bindTooltip(address)
        })
        .catch(() => {})
    }, 300)

    return () => {
      disposed = true
      clearTimeout(timer)
    }
  }, [address])

  useEffect(() => {
    return () => {
      markerRef.current = null
      mapRef.current?.remove()
      mapRef.current = null
    }
  }, [])

  return <div ref={containerRef} className={className ?? defaultClass} />
}
