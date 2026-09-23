import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { QrCode } from "lucide-react"
import { createPlatformInvitation } from "../../hooks/useApi"
import { InviteDialog } from "../barbers/InviteDialog"
import { Button } from "../ui/button"
import { useI18n } from "../../lib/i18n"
import type { Invitation } from "../../hooks/useInvitations"

type PlatformRole = "customer" | "owner" | "realm-admin"

export function RealmAdminInvite() {
  const { t } = useI18n()
  const [open, setOpen] = useState(false)
  const [role, setRole] = useState<PlatformRole>("owner")

  const invite = useMutation({
    mutationFn: (r: PlatformRole) => createPlatformInvitation({ role: r }),
  })

  const start = (r: PlatformRole) => {
    setRole(r)
    setOpen(true)
    invite.mutate(r)
  }

  return (
    <>
      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" onClick={() => start("owner")}>
          <QrCode className="size-3" /> {t("admin.inviteOwner")}
        </Button>
        <Button size="sm" variant="outline" onClick={() => start("customer")}>
          <QrCode className="size-3" /> {t("admin.inviteCustomer")}
        </Button>
        <Button size="sm" variant="outline" onClick={() => start("realm-admin")}>
          <QrCode className="size-3" /> {t("admin.inviteRealmAdmin")}
        </Button>
      </div>

      <InviteDialog
        open={open}
        onClose={() => setOpen(false)}
        invitation={(invite.data as Invitation | null) ?? null}
        loading={invite.isPending}
        error={invite.error ? t("admin.inviteError") : null}
        filename={`fejd-${role}-invite`}
        role={role}
      />
    </>
  )
}
