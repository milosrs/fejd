import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { updateBusinessLocation } from "../hooks/useApi"
import { useCanWrite } from "../hooks/useCanWrite"
import { useI18n } from "../lib/i18n"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card"
import { PhotonAutocomplete, type PhotonPlace } from "./ui/photon-autocomplete"

export interface SalonLocation {
  address_line?: string
  city?: string
  postal_code?: string
  country?: string
  phone?: string
}

// SalonLocationForm is the inline (non-modal) structured location editor for the
// salon settings page. The address and city fields autocomplete via Photon
// (OpenStreetMap, keyless) and fill the coordinates so no backend geocoding is
// needed.
export function SalonLocationForm({
  businessId,
  slug,
  initial,
  onSaved,
}: {
  businessId: string
  slug: string
  initial: SalonLocation
  onSaved?: () => void
}) {
  const queryClient = useQueryClient()
  const canWrite = useCanWrite()
  const { t } = useI18n()

  const [city, setCity] = useState(initial.city ?? "")
  const [addressLine, setAddressLine] = useState(initial.address_line ?? "")
  const [postalCode, setPostalCode] = useState(initial.postal_code ?? "")
  const [country, setCountry] = useState(initial.country ?? "RS")
  const [phone, setPhone] = useState(initial.phone ?? "")
  const [latitude, setLatitude] = useState<number | undefined>(undefined)
  const [longitude, setLongitude] = useState<number | undefined>(undefined)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  const applyPlace = (place: PhotonPlace, fillAddress: boolean) => {
    if (fillAddress && place.addressLine) {
      setAddressLine(place.addressLine)
    }
    if (place.city) {
      setCity(place.city)
    }
    if (!postalCode && place.postalCode) {
      setPostalCode(place.postalCode)
    }
    if (place.country) {
      setCountry(place.country)
    }
    if (place.latitude != null && place.longitude != null) {
      setLatitude(place.latitude)
      setLongitude(place.longitude)
    }
  }

  const handleSave = async () => {
    const trimmedCity = city.trim()
    if (!trimmedCity) {
      setError(t("location.cityRequired"))
      return
    }
    setSaving(true)
    setError("")
    try {
      await updateBusinessLocation(businessId, {
        city: trimmedCity,
        address_line: addressLine.trim(),
        postal_code: postalCode.trim(),
        country: country.trim(),
        phone: phone.trim(),
        latitude,
        longitude,
      })
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      await queryClient.invalidateQueries({ queryKey: ["businesses"] })
      onSaved?.()
    } catch {
      setError(t("location.failed"))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("location.title")}</CardTitle>
        <CardDescription>{t("location.help")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="space-y-1">
          <Label htmlFor="location-address">{t("location.addressLine")}</Label>
          <PhotonAutocomplete
            value={addressLine}
            onChange={setAddressLine}
            onSelect={(place) => applyPlace(place, true)}
          />
        </div>

        <div className="space-y-1">
          <Label htmlFor="location-city">{t("location.city")}</Label>
          <PhotonAutocomplete
            value={city}
            onChange={setCity}
            onSelect={(place) => {
              const chosenCity = place.city || place.addressLine || place.label
              if (chosenCity) setCity(chosenCity)
              applyPlace(place, false)
            }}
            placeholder={t("location.cityPlaceholder")}
          />
        </div>

        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <Label htmlFor="location-postal">{t("location.postalCode")}</Label>
            <Input
              id="location-postal"
              value={postalCode}
              onChange={(e) => setPostalCode(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="location-country">{t("location.country")}</Label>
            <Input
              id="location-country"
              value={country}
              onChange={(e) => setCountry(e.target.value)}
            />
          </div>
        </div>

        <div className="space-y-1">
          <Label htmlFor="location-phone">{t("location.phone")}</Label>
          <Input
            id="location-phone"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
          />
        </div>

        {error && <p className="text-sm text-red-500">{error}</p>}

        <div className="flex justify-end">
          <Button
            onClick={handleSave}
            isDisabled={saving || !city.trim() || !canWrite}
          >
            {saving ? t("common.saving") : t("location.save")}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
