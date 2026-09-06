import { Input } from "../../ui/input"
import { Textarea } from "../../ui/textarea"
import { Label } from "../../ui/label"
import { ImageUploadButton } from "../../ui/image-upload-button"
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
  onUploadImage?: (file: File) => void
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
        <Field label="Hero image">
          <ImageUploadButton onPicked={onUploadImage} uploading={uploading} />
        </Field>
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
}: {
  value: GalleryContent
  onChange: (value: GalleryContent) => void
}) {
  return (
    <div className="space-y-4">
      <Field label="Heading">
        <Input
          value={value.heading ?? ""}
          onChange={(e) => onChange({ ...value, heading: e.target.value })}
        />
      </Field>
      <Field label="Image URLs (one per line)">
        <Textarea
          value={(value.image_urls ?? []).join("\n")}
          onChange={(e) =>
            onChange({
              ...value,
              image_urls: e.target.value.split("\n").filter((s) => s.trim() !== ""),
            })
          }
        />
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
        <Input
          value={value.address ?? ""}
          onChange={(e) => onChange({ ...value, address: e.target.value })}
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
  onUploadImage?: (file: File) => void
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
      return <GallerySectionForm value={value as GalleryContent} onChange={onChange} />
    case "contact":
      return <ContactSectionForm value={value as ContactContent} onChange={onChange} />
    default:
      return null
  }
}
