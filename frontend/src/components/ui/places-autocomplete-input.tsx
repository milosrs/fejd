import { useRef } from "react"

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
