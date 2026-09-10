import { Input } from "../../ui/input"
import { Textarea } from "../../ui/textarea"
import { Label } from "../../ui/label"
import { ImageUploadButton } from "../../ui/image-upload-button"
import { PlacesAutocompleteInput } from "../../ui/places-autocomplete-input"
import { resolveImageUrl } from "../../../lib/images"
import type {
  AboutContent,
  ContactContent,
  GalleryContent,
  HeroContent,
  SectionContent,
  SectionType,
} from "../../../lib/sections"

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      {children}
    </div>
  )
}

export function HeroSectionForm({
  value,
  onChange,
  onUploadImage,
  uploading,
}: {
  value: HeroContent
  onChange: (value: HeroContent) => void
  onUploadImage?: (file: File, purpose: string) => Promise<string | undefined>
  uploading?: boolean
}) {
  return (
    <div className="space-y-4">
      <Field label="Headline">
        <Input
          value={value.headline ?? ""}
          onChange={(e) => onChange({ ...value, headline: e.target.value })}
        />
      </Field>
      <Field label="Subheadline">
        <Textarea
          value={value.subheadline ?? ""}
          onChange={(e) => onChange({ ...value, subheadline: e.target.value })}
        />
      </Field>
      <Field label="CTA text">
        <Input
          value={value.cta_text ?? ""}
          onChange={(e) => onChange({ ...value, cta_text: e.target.value })}
        />
      </Field>
      {onUploadImage && (
        <>
          <Field label="Logo">
            <ImageUploadButton
              label="Upload logo"
              onPicked={(file) => onUploadImage(file, "logo")}
              uploading={uploading}
            />
          </Field>
          <Field label="Background">
            <ImageUploadButton
              label="Upload background"
              onPicked={(file) => onUploadImage(file, "background")}
              uploading={uploading}
            />
          </Field>
        </>
      )}
    </div>
  )
}

export function AboutSectionForm({
  value,
  onChange,
}: {
  value: AboutContent
  onChange: (value: AboutContent) => void
}) {
  return (
    <div className="space-y-4">
      <Field label="Heading">
        <Input
          value={value.heading ?? ""}
          onChange={(e) => onChange({ ...value, heading: e.target.value })}
        />
      </Field>
      <Field label="Body">
        <Textarea
          value={value.body ?? ""}
          onChange={(e) => onChange({ ...value, body: e.target.value })}
        />
      </Field>
    </div>
  )
}

export function GallerySectionForm({
  value,
  onChange,
  onUploadImage,
  uploading,
}: {
  value: GalleryContent
  onChange: (value: GalleryContent) => void
  onUploadImage?: (file: File, purpose: string) => Promise<string | undefined>
  uploading?: boolean
}) {
  const images = value.image_urls ?? []

  const addImage = async (file: File) => {
    if (!onUploadImage) return
    const url = await onUploadImage(file, "gallery")
    if (url) onChange({ ...value, image_urls: [...images, url] })
  }

  const removeImage = (index: number) => {
    onChange({ ...value, image_urls: images.filter((_, i) => i !== index) })
  }

  return (
    <div className="space-y-4">
      <Field label="Heading">
        <Input
          value={value.heading ?? ""}
          onChange={(e) => onChange({ ...value, heading: e.target.value })}
        />
      </Field>
      <Field label="Images">
        <div className="space-y-3">
          {images.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {images.map((url, index) => (
                <div key={`${url}-${index}`} className="relative">
                  <img
                    src={resolveImageUrl(url)}
                    alt=""
                    className="h-16 w-16 rounded-lg border border-border object-cover"
                  />
                  <button
                    type="button"
                    aria-label="Remove image"
                    onClick={() => removeImage(index)}
                    className="absolute -right-1.5 -top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-destructive text-destructive-foreground text-xs leading-none hover:bg-destructive/90"
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
          )}
          {onUploadImage && (
            <ImageUploadButton
              label="Add image"
              onPicked={addImage}
              uploading={uploading}
            />
          )}
        </div>
      </Field>
    </div>
  )
}

export function ContactSectionForm({
  value,
  onChange,
}: {
  value: ContactContent
  onChange: (value: ContactContent) => void
}) {
  return (
    <div className="space-y-4">
      <Field label="Heading">
        <Input
          value={value.heading ?? ""}
          onChange={(e) => onChange({ ...value, heading: e.target.value })}
        />
      </Field>
      <Field label="Phone">
        <Input
          value={value.phone ?? ""}
          onChange={(e) => onChange({ ...value, phone: e.target.value })}
        />
      </Field>
      <Field label="Email">
        <Input
          value={value.email ?? ""}
          onChange={(e) => onChange({ ...value, email: e.target.value })}
        />
      </Field>
      <Field label="Address">
        <PlacesAutocompleteInput
          value={value.address ?? ""}
          onChange={(address) => onChange({ ...value, address })}
          placeholder="Street, city, country"
        />
      </Field>
      <Field label="Google review URL">
        <Input
          value={value.rating_url ?? ""}
          onChange={(e) => onChange({ ...value, rating_url: e.target.value })}
          placeholder="https://..."
        />
      </Field>
    </div>
  )
}

export function SectionForm({
  type,
  value,
  onChange,
  onUploadImage,
  uploading,
}: {
  type: SectionType
  value: SectionContent
  onChange: (value: SectionContent) => void
  onUploadImage?: (file: File, purpose: string) => Promise<string | undefined>
  uploading?: boolean
}) {
  switch (type) {
    case "hero":
      return (
        <HeroSectionForm
          value={value as HeroContent}
          onChange={onChange}
          onUploadImage={onUploadImage}
          uploading={uploading}
        />
      )
    case "about":
      return <AboutSectionForm value={value as AboutContent} onChange={onChange} />
    case "gallery":
      return (
        <GallerySectionForm
          value={value as GalleryContent}
          onChange={onChange}
          onUploadImage={onUploadImage}
          uploading={uploading}
        />
      )
    case "contact":
      return <ContactSectionForm value={value as ContactContent} onChange={onChange} />
    default:
      return null
  }
}
