import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useSections } from "../hooks/useSections"
import { useI18n } from "../lib/i18n"
import { isSectionType, KNOWN_SECTION_TYPES, type Section } from "../lib/sections"
import { SectionRenderer } from "../components/sections/SectionRenderer"
import { SectionEditor } from "../components/sections/editor/SectionEditor"
import { useSectionMutations } from "../hooks/useSectionMutations"
import { uploadBusinessImage } from "../hooks/useApi"
import { dataUrlToFile, fileToDataUrl } from "../lib/images"
import { useSalonDraftStore, type DraftImagePurpose } from "../stores/salonDraftStore"
import { Button } from "../components/ui/button"
import { ChevronDown, ChevronUp, Pencil, Plus, Trash2 } from "lucide-react"

function AddSectionPicker({ onPick }: { onPick: (type: string) => void }) {
  return (
    <div className="flex flex-wrap justify-center gap-2 rounded-xl border border-dashed border-border p-4">
      {KNOWN_SECTION_TYPES.map((type) => (
        <Button key={type} variant="outline" size="sm" onClick={() => onPick(type)}>
          <Plus className="size-3" /> {type}
        </Button>
      ))}
    </div>
  )
}

// commitHeroImages uploads any staged hero logo/background and clears the draft
// only after both uploads succeed.
async function commitHeroImages(slug: string, businessId: string) {
  const draft = useSalonDraftStore.getState().drafts[slug]
  if (!draft || !businessId) return
  if (draft.logo) {
    await uploadBusinessImage(businessId, dataUrlToFile(draft.logo, "logo.png"), "logo")
  }
  if (draft.background) {
    await uploadBusinessImage(
      businessId,
      dataUrlToFile(draft.background, "background.png"),
      "background",
    )
  }
  useSalonDraftStore.getState().clearDraft(slug)
}

// resolveGalleryImages uploads any data-URL gallery images embedded in the
// section content and replaces them with their server URLs before saving.
async function resolveGalleryImages(
  content: Record<string, unknown>,
  businessId: string,
): Promise<Record<string, unknown>> {
  const next: Record<string, unknown> = { ...content }
  for (const [locale, value] of Object.entries(content)) {
    if (typeof value !== "object" || value === null) continue
    const fields = value as Record<string, unknown>
    const urls = fields.image_urls
    if (!Array.isArray(urls)) continue
    const resolved: unknown[] = []
    for (const url of urls) {
      if (typeof url === "string" && url.startsWith("data:")) {
        const img = await uploadBusinessImage(businessId, dataUrlToFile(url, "gallery.png"), "gallery")
        if (img) resolved.push(img.url)
      } else {
        resolved.push(url)
      }
    }
    next[locale] = { ...fields, image_urls: resolved }
  }
  return next
}

export function LandingPage() {
  const { slug, salon, editing } = useSalonContext()
  const isOwner = useIsOwner()
  const { t, ready } = useI18n()
  const { data: sections, isLoading } = useSections(slug)

  const businessId = salon?.business.id ?? ""
  const { create, update, remove, reorder } = useSectionMutations(businessId, slug)
  const queryClient = useQueryClient()

  const [editingSection, setEditingSection] = useState<Section | null>(null)
  const [showAddPicker, setShowAddPicker] = useState(false)
  const [saving, setSaving] = useState(false)

  if (!salon) return null

  const list = sections ?? []
  const hasRenderable = list.some((s) => isSectionType(s.type))

  const move = (index: number, direction: -1 | 1) => {
    const target = index + direction
    if (target < 0 || target >= list.length) return
    const next = [...list]
    ;[next[index], next[target]] = [next[target], next[index]]
    reorder.mutate(next.map((s) => s.id))
  }

  const addSection = (type: string) => {
    setShowAddPicker(false)
    setEditingSection({
      id: "",
      page_id: "",
      type,
      content: {},
      position: list.length,
    })
  }

  const isNew = editingSection?.id === ""

  // Images are staged locally (as data URLs) and only uploaded on Save. Hero
  // logo/background go to the draft store; gallery images are returned as data
  // URLs and stored in the section content until Save.
  const handleUploadImage = async (file: File, purpose: string): Promise<string | undefined> => {
    const dataUrl = await fileToDataUrl(file)
    if (purpose === "logo" || purpose === "background") {
      useSalonDraftStore.getState().setDraftImage(slug, purpose as DraftImagePurpose, dataUrl)
      return undefined
    }
    return dataUrl
  }

  const handleSave = async (content: Record<string, unknown>) => {
    if (!editingSection) return
    setSaving(true)
    try {
      const resolvedContent = await resolveGalleryImages(content, businessId)
      await commitHeroImages(slug, businessId)
      if (isNew) {
        await create.mutateAsync({ type: editingSection.type, content: resolvedContent })
      } else {
        await update.mutateAsync({ sectionId: editingSection.id, content: resolvedContent })
      }
      queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      setEditingSection(null)
    } catch (err) {
      console.error("[sections] failed to save:", err)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = () => {
    if (!editingSection || isNew) return
    remove.mutate(editingSection.id)
    setEditingSection(null)
  }

  const closeEditor = () => {
    useSalonDraftStore.getState().clearDraft(slug)
    setEditingSection(null)
  }

  if (isLoading || !ready) return null

  const editingOn = editing && isOwner

  return (
    <section className="space-y-8">
      {list.length === 0 || !hasRenderable ? (
        <div className="flex flex-col items-center gap-3 py-16 text-center">
          <h2 className="text-lg font-semibold text-foreground">
            {t("landing.empty.title")}
          </h2>
          <p className="max-w-sm text-muted-foreground">{t("landing.empty.body")}</p>
          {editingOn && (
            <Button
              variant="outline"
              size="sm"
              className="mt-2"
              onClick={() => setShowAddPicker((v) => !v)}
            >
              <Plus className="size-3" /> Add section
            </Button>
          )}
          {editingOn && showAddPicker && <AddSectionPicker onPick={addSection} />}
        </div>
      ) : (
        <>
          {list.map((section, index) => (
            <div key={section.id} className="relative">
              <SectionRenderer section={section} />
              {editingOn && (
                <div className="absolute right-0 top-0 flex gap-1 rounded-lg border border-border bg-background p-1 shadow-sm">
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => move(index, -1)}
                    isDisabled={index === 0}
                  >
                    <ChevronUp />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => move(index, 1)}
                    isDisabled={index === list.length - 1}
                  >
                    <ChevronDown />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => setEditingSection(section)}
                  >
                    <Pencil />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    className="text-destructive"
                    onClick={() => remove.mutate(section.id)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              )}
            </div>
          ))}

          {editingOn && (
            <div className="flex flex-col items-center gap-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowAddPicker((v) => !v)}
              >
                <Plus className="size-3" /> Add section
              </Button>
              {showAddPicker && <AddSectionPicker onPick={addSection} />}
            </div>
          )}
        </>
      )}

      {editingSection && (
        <SectionEditor
          section={editingSection}
          onClose={closeEditor}
          onSave={handleSave}
          onDelete={isNew ? undefined : handleDelete}
          saving={saving}
          onUploadImage={handleUploadImage}
        />
      )}
    </section>
  )
}
