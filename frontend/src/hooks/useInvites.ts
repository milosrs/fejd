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

// InviteKind describes the concrete invite the user asked to mint. "friend" is a
// platform customer invite (no salon); "salon-customer" links the invitee to a
// specific salon as a customer.
export type InviteKind = "employee" | "salon-customer" | "friend" | "owner" | "realm-admin"

export type InviteDialogRole = "employee" | "customer" | "owner" | "realm-admin" | "friend"

export interface InviteOption {
  kind: InviteKind
  labelKey: string
  filename: string
  dialogRole: InviteDialogRole
}

async function createInvite(
  kind: InviteKind,
  businessId?: string,
): Promise<Invitation | undefined> {
  switch (kind) {
    case "employee":
      if (!businessId) throw new Error("no business")
      return createInvitation(businessId, { role: "employee" })
    case "salon-customer":
      if (!businessId) throw new Error("no business")
      return createInvitation(businessId, { role: "customer" })
    case "friend":
      return createCustomerInvitation()
    case "owner":
      return createPlatformInvitation({ role: "owner" })
    case "realm-admin":
      return createPlatformInvitation({ role: "realm-admin" })
  }
}

// useInvites computes which invite options the current user may mint:
// - realm admin → owner / customer / realm-admin (platform)
// - owner       → employee / customer (salon-scoped)
// - employee    → friend (platform) / customer (salon-scoped)
// - everyone else → friend (platform customer)
// It also drives the shared invite dialog shown in the profile menu.
export function useInvites() {
  const isRealmAdmin = useAuthStore((s) => s.isRealmAdmin)
  const { data: me } = useMe()

  const businesses = me?.businesses ?? []
  const ownerBusiness = businesses.find((b) => b.role === "admin")
  const employeeBusiness = businesses.find((b) => b.role === "employee")

  const options: InviteOption[] = []
  if (isRealmAdmin) {
    options.push(
      { kind: "owner", labelKey: "invite.menu.owner", filename: "fejd-owner-invite", dialogRole: "owner" },
      { kind: "friend", labelKey: "invite.menu.customer", filename: "fejd-customer-invite", dialogRole: "friend" },
      { kind: "realm-admin", labelKey: "invite.menu.realmAdmin", filename: "fejd-realm-admin-invite", dialogRole: "realm-admin" },
    )
  } else if (ownerBusiness) {
    options.push(
      { kind: "employee", labelKey: "invite.menu.employee", filename: "fejd-employee-invite", dialogRole: "employee" },
      { kind: "salon-customer", labelKey: "invite.menu.customer", filename: "fejd-customer-invite", dialogRole: "customer" },
    )
  } else if (employeeBusiness) {
    options.push(
      { kind: "friend", labelKey: "invite.friend", filename: "fejd-customer-invite", dialogRole: "friend" },
      { kind: "salon-customer", labelKey: "invite.menu.salonCustomer", filename: "fejd-customer-invite", dialogRole: "customer" },
    )
  } else {
    options.push({ kind: "friend", labelKey: "invite.friend", filename: "fejd-customer-invite", dialogRole: "friend" })
  }

  const businessId = (ownerBusiness ?? employeeBusiness)?.id

  const [open, setOpen] = useState(false)
  const [kind, setKind] = useState<InviteKind>("friend")

  const mutation = useMutation({
    mutationFn: (k: InviteKind) => createInvite(k, businessId),
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
    dialogRole: active?.dialogRole ?? "customer",
    filename: active?.filename ?? "fejd-invite",
    close: () => setOpen(false),
    start,
    invitation: mutation.data ?? null,
    loading: mutation.isPending,
    error: mutation.isError,
  }
}
