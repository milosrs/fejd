import { useAuthStore } from "../stores/authStore"
import { useI18n } from "../lib/i18n"

// useDisabledReason returns the reason a control is disabled when the current
// user is in read-only mode (authenticated but email not verified), or
// undefined when the user can write. Disabled controls surface this reason as a
// tooltip so the greyed-out state is self-explanatory.
export function useDisabledReason(): string | undefined {
  const authenticated = useAuthStore((s) => s.authenticated)
  const emailVerified = useAuthStore((s) => s.emailVerified)
  const { t } = useI18n()
  if (authenticated && !emailVerified) return t("verifyEmail.disabled")
  return undefined
}
