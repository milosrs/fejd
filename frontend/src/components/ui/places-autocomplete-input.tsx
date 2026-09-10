import { useEffect, useRef } from "react"
import { loadGoogleMaps } from "../../lib/googleMaps"

const inputClass =
  "h-8 w-full min-w-0 rounded-2xl border border-transparent bg-input/50 px-2.5 py-1 text-base outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/30 md:text-sm"

interface PlacesAutocompleteInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}

// PlacesAutocompleteInput is a controlled input with Google Places
// autocomplete. When a suggestion is selected, the place's formatted address is
// written back through onChange. It degrades to a plain input when the Maps API
// isn't configured or fails to load.
export function PlacesAutocompleteInput({
  value,
  onChange,
  placeholder,
}: PlacesAutocompleteInputProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  useEffect(() => {
    let autocomplete: any
    let disposed = false

    loadGoogleMaps()
      .then(() => {
        const google = (window as any).google
        if (disposed || !inputRef.current || !google?.maps?.places) return

        autocomplete = new google.maps.places.Autocomplete(inputRef.current, {
          types: ["geocode"],
        })
        autocomplete.addListener("place_changed", () => {
          const place = autocomplete.getPlace()
          if (place?.formatted_address) {
            onChangeRef.current(place.formatted_address)
          }
        })
      })
      .catch(() => {
        // Autocomplete is unavailable; the plain input still works.
      })

    return () => {
      disposed = true
      const google = (window as any).google
      if (autocomplete && google?.maps?.event) {
        google.maps.event.clearInstanceListeners(autocomplete)
      }
    }
  }, [])

  return (
    <input
      ref={inputRef}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className={inputClass}
    />
  )
}
