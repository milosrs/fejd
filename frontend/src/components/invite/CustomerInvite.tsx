import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { QrCode } from "lucide-react"
import { createCustomerInvitation } from "../../hooks/useApi"
import { useCanWrite } from "../../hooks/useCanWrite"
import { InviteDialog } from "../barbers/InviteDialog"
import { Button } from "../ui/button"
import { useI18n } from "../../lib/i18n"
import type { Invitation } from "../../hooks/useInvitations"

export function CustomerInvite() {
  const { t } = useI18n()
  const canWrite = useCanWrite()
  const [open, setOpen] = useState(false)

  const invite = useMutation({
    mutationFn: () => createCustomerInvitation(),
  })

  const start = () => {
    setOpen(true)
    invite.mutate()
  }

  return (
    <>
      <Button size="sm" variant="outline" onClick={start} isDisabled={!canWrite}>
        <QrCode className="size-3" /> {t("invite.friend")}
      </Button>

      <InviteDialog
        open={open}
        onClose={() => setOpen(false)}
        invitation={(invite.data as Invitation | null) ?? null}
        loading={invite.isPending}
        error={invite.error ? t("invite.friendError") : null}
        filename="fejd-customer-invite"
        role="customer"
      />
    </>
  )
}
