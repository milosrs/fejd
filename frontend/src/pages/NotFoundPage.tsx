import { useLocation, useNavigate } from "react-router"
import { useI18n } from "../lib/i18n"
import { Button } from "../components/ui/button"

// NotFoundPage renders inside the app shell (below the top navigation bar) and
// mirrors the home page's centered layout so a broken link still looks like the
// rest of the app. The back button returns to the previous in-app route when
// there is history, and falls back to the home page for direct links.
export function NotFoundPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { t } = useI18n()

  const handleBack = () => {
    if (location.key !== "default") {
      navigate(-1)
    } else {
      navigate("/")
    }
  }

  return (
    <div className="min-h-app flex flex-col items-center justify-center gap-4 bg-background p-8">
      <img src="/logo-white.jpg" alt="fejd" className="h-36 w-auto dark:hidden" />
      <img src="/logo_dark.jpg" alt="fejd" className="hidden h-36 w-auto dark:block" />
      <h1 className="text-2xl font-semibold text-foreground">{t("notFound.title")}</h1>
      <p className="max-w-sm text-center text-muted-foreground">{t("notFound.message")}</p>
      <Button onClick={handleBack}>{t("common.back")}</Button>
    </div>
  )
}
