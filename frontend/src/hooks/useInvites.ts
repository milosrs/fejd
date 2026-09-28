import { useMutation } from "@tanstack/react-query"
import { useState } from "react"
import { useAuthStore } from "../stores/authStore"
import { useMe } from "./useMe"
import {
  createCustomerInvitation,
  createInvitation,
  createPlatformInvitation,
  type Invitation,
} from "./useApi"

export type InviteKind = "employee" | "customer" | "owner" | "realm-admin"

export interface InviteOption {
  kind: InviteKind
  labelKey: string
  filename: string
}

async function createInvite(
  kind: InviteKind,
  isRealmAdmin: boolean,
  businessId?: string,
): Promise<Invitation | undefined> {
  switch (kind) {
    case "employee":
      if (!businessId) throw new Error("no business")
      return createInvitation(businessId, { role: "employee" })
    case "customer":
      if (isRealmAdmin) return createPlatformInvitation({ role: "customer" })
      if (businessId) return createInvitation(businessId, { role: "customer" })
      return createCustomerInvitation()
    case "owner":
      return createPlatformInvitation({ role: "owner" })
    case "realm-admin":
      return createPlatformInvitation({ role: "realm-admin" })
  }
}

// useInvites computes which invite options the current user may mint (Owner →
// employee/customer; realm admin → owner/customer/realm-admin; everyone else →
// customer) and drives the shared invite dialog shown in the profile menu.
export function useInvites() {
  const isRealmAdmin = useAuthStore((s) => s.isRealmAdmin)
  const { data: me } = useMe()

  const ownerBusiness = (me?.businesses ?? []).find((b) => b.role === "admin")
  const businessId = ownerBusiness?.id

  const options: InviteOption[] = []
  if (isRealmAdmin) {
    options.push(
      { kind: "owner", labelKey: "invite.menu.owner", filename: "fejd-owner-invite" },
      { kind: "customer", labelKey: "invite.menu.customer", filename: "fejd-customer-invite" },
      { kind: "realm-admin", labelKey: "invite.menu.realmAdmin", filename: "fejd-realm-admin-invite" },
    )
  } else if (businessId) {
    options.push(
      { kind: "employee", labelKey: "invite.menu.employee", filename: "fejd-employee-invite" },
      { kind: "customer", labelKey: "invite.menu.customer", filename: "fejd-customer-invite" },
    )
  } else {
    options.push({ kind: "customer", labelKey: "invite.friend", filename: "fejd-customer-invite" })
  }

  const [open, setOpen] = useState(false)
  const [kind, setKind] = useState<InviteKind>("customer")

  const mutation = useMutation({
    mutationFn: (k: InviteKind) => createInvite(k, isRealmAdmin, businessId),
  })

  const start = (k: InviteKind) => {
    setKind(k)
    setOpen(true)
    mutation.mutate(k)
  }

  const active = options.find((o) => o.kind === kind)

  return {
    options,
    open,
    kind,
    filename: active?.filename ?? "fejd-invite",
    close: () => setOpen(false),
    start,
    invitation: mutation.data ?? null,
    loading: mutation.isPending,
    error: mutation.isError,
  }
}
