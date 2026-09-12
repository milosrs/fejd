import { useState } from "react"
import { Capacitor } from "@capacitor/core"
import { Button } from "./button"
import { pickImage } from "../../lib/imagePicker"
import { useI18n } from "../../lib/i18n"

export function ImageUploadButton({
  onPicked,
  label,
  uploading,
}: {
  onPicked: (file: File) => void
  label?: string
  uploading?: boolean
}) {
  const native = Capacitor.isNativePlatform()
  const [picking, setPicking] = useState(false)
  const { t } = useI18n()
  const resolvedLabel = label ?? t("upload.uploadImage")

  const handlePick = async (source: "library" | "camera") => {
    setPicking(true)
    try {
      const file = await pickImage(source)
      if (file) onPicked(file)
    } finally {
      setPicking(false)
    }
  }

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        isDisabled={uploading || picking}
        onClick={() => handlePick("library")}
      >
        {uploading || picking ? t("upload.working") : resolvedLabel}
      </Button>
      {native && (
        <Button
          variant="outline"
          size="sm"
          isDisabled={uploading || picking}
          onClick={() => handlePick("camera")}
        >
          {t("upload.camera")}
        </Button>
      )}
    </div>
  )
}
