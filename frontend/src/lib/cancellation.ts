import { format, parseISO } from "date-fns"

export const CANCELLATION_NO_SERVICE_REASON = "cancellation.reason.noService"

const FALLBACK_EN = "Does not offer this service as of {date}"

// formatCancellationReason renders a stored cancellation reason. System reasons
// are stored as "<translation-key>|<YYYY-MM-DD>" and localized here; free-text
// reasons are returned unchanged.
export function formatCancellationReason(
  reason: string,
  t: (key: string) => string,
): string {
  if (!reason.startsWith(`${CANCELLATION_NO_SERVICE_REASON}|`)) {
    return reason
  }

  const date = reason.slice(CANCELLATION_NO_SERVICE_REASON.length + 1)
  const label = t(CANCELLATION_NO_SERVICE_REASON)
  const template = label === CANCELLATION_NO_SERVICE_REASON ? FALLBACK_EN : label
  const formatted = format(parseISO(date), "MMMM d, yyyy")
  return template.replace("{date}", formatted)
}
