import { useMemo, useState } from "react"
import { SectionForm } from "./SectionForm"
import { SectionRenderer } from "../SectionRenderer"
import { Button } from "../../ui/button"
import { DEFAULT_LOCALE } from "../../../lib/i18n"
import {
  getLocalizedContent,
  isSectionType,
  setLocalizedContent,
  type Section,
  type SectionContent,
} from "../../../lib/sections"

export function SectionEditor({
  section,
  onClose,
  onSave,
  onDelete,
  saving,
  onUploadImage,
  uploading,
}: {
  section: Section
  onClose: () => void
  onSave: (content: Record<string, unknown>) => void
  onDelete?: () => void
  saving?: boolean
  onUploadImage?: (file: File, purpose: string) => Promise<string | undefined>
  uploading?: boolean
}) {
  const type = isSectionType(section.type) ? section.type : "hero"

  const [content, setContent] = useState<SectionContent>(() =>
    getLocalizedContent(section.content, DEFAULT_LOCALE),
  )

  const previewSection = useMemo<Section>(
    () => ({
      ...section,
      content: setLocalizedContent(section.content, DEFAULT_LOCALE, content),
    }),
    [section, content],
  )

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 pt-[calc(1rem+env(safe-area-inset-top))] pb-[calc(1rem+env(safe-area-inset-bottom))]"
      onClick={onClose}
    >
      <div
        className="w-full max-w-3xl max-h-[90vh] overflow-y-auto rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">Edit {type} section</h3>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div>
            <SectionForm
              type={type}
              value={content}
              onChange={setContent}
              onUploadImage={onUploadImage}
              uploading={uploading}
            />
          </div>
          <div className="rounded-xl border border-dashed border-border p-4">
            <p className="mb-3 text-xs uppercase tracking-wide text-muted-foreground">
              Live preview
            </p>
            <SectionRenderer section={previewSection} contained />
          </div>
        </div>

        <div className="mt-6 flex items-center justify-between">
          {onDelete ? (
            <Button variant="destructive" size="sm" onClick={onDelete}>
              Delete
            </Button>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              isDisabled={saving}
              onClick={() =>
                onSave(setLocalizedContent(section.content, DEFAULT_LOCALE, content))
              }
            >
              {saving ? "Saving…" : "Save"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
