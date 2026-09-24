import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import { deleteBusiness } from "../hooks/useApi"
import { openAppHome } from "../lib/salonDomain"
import { useI18n } from "../lib/i18n"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { Label } from "./ui/label"

export function DeleteSalonDialog({
  businessId,
  slug,
  salonName,
  onClose,
}: {
  businessId: string
  slug: string
  salonName: string
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { t } = useI18n()
  const [name, setName] = useState("")
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState("")

  const confirmed = name.trim() === salonName

  const handleDelete = async () => {
    if (!confirmed) return
    setDeleting(true)
    setError("")
    try {
      await deleteBusiness(businessId, name.trim())
      await queryClient.invalidateQueries({ queryKey: ["me"] })
      await queryClient.invalidateQueries({ queryKey: ["businesses"] })
      await queryClient.invalidateQueries({ queryKey: ["salon", slug] })
      onClose()
      openAppHome(navigate)
    } catch {
      setError(t("deleteSalon.failed"))
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-2xl border border-border bg-background p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-semibold text-foreground">{t("deleteSalon.title")}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{t("deleteSalon.confirm")}</p>
        <p className="mt-3 text-sm text-muted-foreground">{t("deleteSalon.help")}</p>

        <div className="mt-4 space-y-1">
          <Label htmlFor="delete-salon-name">{t("deleteSalon.nameLabel")}</Label>
          <div className="flex gap-2">
            <Input
              id="delete-salon-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t("deleteSalon.namePlaceholder")}
              className="flex-1"
            />
            <Button
              variant="outline"
              size="sm"
              onClick={() => setName(salonName)}
            >
              {t("deleteSalon.autofill")}
            </Button>
          </div>
        </div>

        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            isDisabled={deleting || !confirmed}
          >
            {deleting ? t("deleteSalon.deleting") : t("salon.delete")}
          </Button>
        </div>
      </div>
    </div>
  )
}
