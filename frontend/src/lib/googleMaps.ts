let loadPromise: Promise<void> | null = null

// loadGoogleMaps loads the Google Maps JavaScript API (Places library) once.
// It resolves immediately when no API key is configured or the API is already
// loaded, so callers can attach autocomplete without special-casing.
export function loadGoogleMaps(): Promise<void> {
  const key = import.meta.env.VITE_GOOGLE_MAPS_API_KEY
  if (!key) return Promise.resolve()
  if ((window as any).google?.maps?.places) return Promise.resolve()
  if (loadPromise) return loadPromise

  loadPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script")
    script.src = `https://maps.googleapis.com/maps/api/js?key=${encodeURIComponent(key)}&libraries=places`
    script.async = true
    script.defer = true
    script.onload = () => resolve()
    script.onerror = () => {
      loadPromise = null
      reject(new Error("Failed to load Google Maps"))
    }
    document.head.appendChild(script)
  })
  return loadPromise
}
