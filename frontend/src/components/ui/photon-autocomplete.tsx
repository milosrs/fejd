import { useEffect, useRef, useState } from "react"
import { Input } from "./input"

// PhotonPlace is the structured result of selecting an autocomplete suggestion.
export interface PhotonPlace {
  label: string
  addressLine: string
  city: string
  postalCode: string
  country: string
  latitude?: number
  longitude?: number
}

interface PhotonFeature {
  geometry?: { coordinates?: [number, number] }
  properties?: {
    name?: string
    housenumber?: string
    street?: string
    city?: string
    town?: string
    village?: string
    postcode?: string
    country?: string
    countrycode?: string
  }
}

function toPlace(f: PhotonFeature): PhotonPlace {
  const p = f.properties ?? {}
  const [lon, lat] = f.geometry?.coordinates ?? []
  const street = [p.street, p.housenumber].filter(Boolean).join(" ") || p.name || ""
  const city = p.city || p.town || p.village || ""
  const country = (p.countrycode || p.country || "").toUpperCase()
  const label = [street, city, country].filter(Boolean).join(", ") || p.name || ""

  return {
    label,
    addressLine: street,
    city,
    postalCode: p.postcode || "",
    country,
    latitude: typeof lat === "number" ? lat : undefined,
    longitude: typeof lon === "number" ? lon : undefined,
  }
}

// PhotonAutocomplete is a keyless address/city autocomplete backed by Photon
// (Komoot's OpenStreetMap search). It does not require Google Maps or a credit
// card. On selection it reports a structured PhotonPlace with coordinates.
export function PhotonAutocomplete({
  value,
  onChange,
  onSelect,
  placeholder,
}: {
  value: string
  onChange: (value: string) => void
  onSelect: (place: PhotonPlace) => void
  placeholder?: string
}) {
  const [results, setResults] = useState<PhotonPlace[]>([])
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const query = value.trim()
    if (!query) {
      setResults([])
      return
    }
    let disposed = false
    const timer = setTimeout(async () => {
      try {
        const res = await fetch(
          `https://photon.komoot.io/api/?q=${encodeURIComponent(query)}&limit=6&lang=en`,
        )
        if (!res.ok) throw new Error("photon request failed")
        const data = await res.json()
        if (disposed) return
        setResults((data.features ?? []).map(toPlace))
        setOpen(true)
      } catch {
        if (!disposed) setResults([])
      }
    }, 300)

    return () => {
      disposed = true
      clearTimeout(timer)
    }
  }, [value])

  useEffect(() => {
    function onDocClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", onDocClick)
    return () => document.removeEventListener("mousedown", onDocClick)
  }, [])

  return (
    <div ref={containerRef} className="relative">
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onFocus={() => setOpen(true)}
        placeholder={placeholder}
        className="w-full"
      />
      {open && results.length > 0 && (
        <ul className="absolute z-20 mt-1 max-h-60 w-full overflow-auto rounded-xl border border-border bg-popover p-1 shadow-md">
          {results.map((place, i) => (
            <li key={i}>
              <button
                type="button"
                className="w-full rounded-lg px-2.5 py-1.5 text-left text-sm text-foreground hover:bg-muted"
                onClick={() => {
                  onSelect(place)
                  setOpen(false)
                }}
              >
                {place.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
