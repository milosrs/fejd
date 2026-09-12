import { Menu, MenuItem, MenuTrigger, Popover } from "react-aria-components"

import { Button } from "#components/ui/button"
import { useI18n } from "../lib/i18n"

const menuItemClassName =
  "flex w-full cursor-default items-center gap-2 rounded-lg px-2 py-1.5 text-sm outline-none select-none data-focused:bg-accent data-focused:text-accent-foreground"

function Flag({ children }: { children: React.ReactNode }) {
  return (
    <span className="flex size-4 shrink-0 items-center justify-center overflow-hidden rounded-[3px] ring-1 ring-foreground/15">
      {children}
    </span>
  )
}

function EnglandFlag() {
  return (
    <Flag>
      <svg viewBox="0 0 16 16" className="size-4" aria-hidden="true">
        <rect width="16" height="16" fill="#ffffff" />
        <rect x="0" y="5" width="16" height="6" fill="#C8102E" />
        <rect x="5" y="0" width="6" height="16" fill="#C8102E" />
      </svg>
    </Flag>
  )
}

function SerbiaFlag() {
  return (
    <Flag>
      <svg viewBox="0 0 16 16" className="size-4" aria-hidden="true">
        <rect width="16" height="5.33" fill="#C6363C" />
        <rect y="5.33" width="16" height="5.33" fill="#0C4076" />
        <rect y="10.67" width="16" height="5.33" fill="#ffffff" />
      </svg>
    </Flag>
  )
}

export function LanguageToggle() {
  const { locale, setLocale, t } = useI18n()
  const CurrentFlag = locale === "rs" ? SerbiaFlag : EnglandFlag

  return (
    <MenuTrigger>
      <Button variant="outline" size="icon" aria-label={t("language.toggle")}>
        <CurrentFlag />
      </Button>
      <Popover className="min-w-40 rounded-xl bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/5 dark:ring-foreground/10">
        <Menu className="outline-none" onAction={(key) => setLocale(String(key))}>
          <MenuItem id="en" className={menuItemClassName}>
            <EnglandFlag />
            {t("language.english")}
          </MenuItem>
          <MenuItem id="rs" className={menuItemClassName}>
            <SerbiaFlag />
            {t("language.serbian")}
          </MenuItem>
        </Menu>
      </Popover>
    </MenuTrigger>
  )
}
