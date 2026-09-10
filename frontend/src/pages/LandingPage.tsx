import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useSalonContext, useIsOwner } from "../context/SalonContext"
import { useSections } from "../hooks/useSections"
import { useI18n } from "../lib/i18n"
import { isSectionType, KNOWN_SECTION_TYPES, type Section } from "../lib/sections"
import { SectionRenderer } from "../components/sections/SectionRenderer"
import { SectionEditor } from "../components/sections/editor/SectionEditor"
import { useSectionMutations } from "../hooks/useSectionMutations"
import { uploadBusinessImage, type BusinessImagePurpose } from "../hooks/useApi"
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

export function LandingPage() {
  const { slug, salon, editing } = useSalonContext()
  const isOwner = useIsOwner()
  const { t, ready } = useI18n()
  const { data: sections, isLoading } = useSections(slug)

  const businessId = salon?.business.id ?? ""
  const { create, update, remove, reorder } = useSectionMutations(businessId, slug)

  const queryClient = useQueryClient()
  const uploadImage = useMutation({
    mutationFn: (vars: { file: File; purpose: BusinessImagePurpose }) =>
      uploadBusinessImage(businessId, vars.file, vars.purpose),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["salon", slug] })
    },
  })

  const handleUploadImage = (file: File, purpose: string) => {
    uploadImage.mutate({ file, purpose: purpose as BusinessImagePurpose })
  }

  const [editingSection, setEditingSection] = useState<Section | null>(null)
  const [showAddPicker, setShowAddPicker] = useState(false)

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
    create.mutate(
      { type, content: {} },
      {
        onSuccess: (res) => {
          setEditingSection(res.data as Section)
        },
      },
    )
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
          onClose={() => setEditingSection(null)}
          onSave={(content) => {
            update.mutate(
              { sectionId: editingSection.id, content },
              { onSuccess: () => setEditingSection(null) },
            )
          }}
          onDelete={() => {
            remove.mutate(editingSection.id)
            setEditingSection(null)
          }}
          saving={update.isPending}
          onUploadImage={handleUploadImage}
          uploading={uploadImage.isPending}
        />
      )}
    </section>
  )
}
