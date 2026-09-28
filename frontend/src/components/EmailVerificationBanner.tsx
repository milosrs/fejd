import { MailWarning } from "lucide-react"
import { useI18n } from "../lib/i18n"

export function EmailVerificationBanner() {
  const { t } = useI18n()

  return (
    <div
      role="alert"
      className="border-b border-yellow-500/30 bg-yellow-500/15"
    >
      <div className="mx-auto flex w-full max-w-screen-md items-center justify-center gap-3 px-4 py-3 text-center sm:px-6">
        <MailWarning className="size-5 shrink-0 text-yellow-600 dark:text-yellow-400" aria-hidden="true" />
        <div className="text-sm">
          <p className="font-semibold text-yellow-600 dark:text-yellow-400">{t("verifyEmail.title")}</p>
          <p className="text-yellow-700/80 dark:text-yellow-400/80">{t("verifyEmail.body")}</p>
        </div>
      </div>
    </div>
  )
}
