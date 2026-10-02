import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  Button as RACButton,
  Header,
  Menu,
  MenuItem,
  MenuSection,
  MenuTrigger,
  Popover,
  Separator,
} from "react-aria-components"
import { Check, ImagePlus, LogOut, Monitor, Moon, QrCode, Sun } from "lucide-react"

import { useAuthStore } from "../stores/authStore"
import { useMe } from "../hooks/useMe"
import { useInvites } from "../hooks/useInvites"
import { useCanWrite } from "../hooks/useCanWrite"
import { useDisabledReason } from "../hooks/useDisabledReason"
import { uploadAvatar } from "../hooks/useApi"
import { useTheme } from "./theme-provider"
import { useI18n } from "../lib/i18n"
import { resolveImageUrl } from "../lib/images"
import { pickImage } from "../lib/imagePicker"
import { ProfileAvatar } from "./ProfileAvatar"
import { SideDrawer } from "./ui/drawer"
import { DisabledTooltip } from "./ui/disabled-tooltip"
import { InviteDialog } from "./barbers/InviteDialog"
import { useIsMobile } from "../hooks/useIsMobile"

const menuItemClassName =
  "flex w-full cursor-default items-center gap-2 rounded-lg px-2 py-1.5 text-sm outline-none select-none data-focused:bg-accent data-focused:text-accent-foreground"

const sectionHeaderClassName =
  "px-2 pb-1 pt-2 text-xs font-medium text-muted-foreground"

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

function DrawerItem({
  onClick,
  children,
  destructive,
  disabled,
}: {
  onClick: () => void
  children: React.ReactNode
  destructive?: boolean
  disabled?: boolean
}) {
  const reason = useDisabledReason()
  const button = (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-sm transition-colors hover:bg-muted disabled:opacity-50 disabled:pointer-events-none ${
        destructive ? "text-destructive" : "text-foreground"
      }`}
    >
      {children}
    </button>
  )
  if (disabled && reason) return <DisabledTooltip reason={reason} className="w-full">{button}</DisabledTooltip>
  return button
}

function MenuItemContent({
  reason,
  children,
}: {
  reason?: string
  children: React.ReactNode
}) {
  if (!reason) return <>{children}</>
  return (
    <DisabledTooltip reason={reason} className="w-full">
      <span className="flex w-full items-center gap-2">{children}</span>
    </DisabledTooltip>
  )
}

export function UserMenu() {
  const isMobile = useIsMobile()
  const [drawerOpen, setDrawerOpen] = useState(false)
  const queryClient = useQueryClient()

  const userInfo = useAuthStore((s) => s.userInfo)
  const logout = useAuthStore((s) => s.logout)
  const canWrite = useCanWrite()
  const disabledReason = useDisabledReason()
  const { data: me } = useMe()
  const invites = useInvites()
  const { locale, setLocale, t } = useI18n()
  const { theme, setTheme } = useTheme()

  const uploadAvatarMutation = useMutation({
    mutationFn: (file: File) => uploadAvatar(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] })
    },
  })

  const name = userInfo?.name ?? ""
  const avatarUrl = resolveImageUrl(me?.avatar)

  const close = () => setDrawerOpen(false)

  const handleChangePicture = () => {
    close()
    pickImage().then((file) => {
      if (file) uploadAvatarMutation.mutate(file)
    })
  }

  const handleAction = (key: string) => {
    switch (key) {
      case "change-picture":
        handleChangePicture()
        break
      case "logout":
        close()
        logout()
        break
      case "theme-light":
        setTheme("light")
        close()
        break
      case "theme-dark":
        setTheme("dark")
        close()
        break
      case "theme-system":
        setTheme("system")
        close()
        break
      case "lang-en":
        setLocale("en")
        close()
        break
      case "lang-rs":
        setLocale("rs")
        close()
        break
      case "invite-employee":
        invites.start("employee")
        break
      case "invite-salon-customer":
        invites.start("salon-customer")
        break
      case "invite-friend":
        invites.start("friend")
        break
      case "invite-owner":
        invites.start("owner")
        break
      case "invite-realm-admin":
        invites.start("realm-admin")
        break
    }
  }

  const inviteDialog = (
    <InviteDialog
      open={invites.open}
      onClose={invites.close}
      invitation={invites.invitation}
      loading={invites.loading}
      error={invites.error ? t("invite.friendError") : null}
      filename={invites.filename}
      role={invites.dialogRole}
    />
  )

  const trigger = (
    <RACButton
      aria-label={t("app.profileMenu")}
      className="shrink-0 overflow-hidden rounded-full ring-2 ring-border hover:opacity-80 focus:outline-none focus-visible:ring-primary"
    >
      <ProfileAvatar src={avatarUrl} name={name} />
    </RACButton>
  )

  if (isMobile) {
    return (
      <>
        <RACButton
          onPress={() => setDrawerOpen(true)}
          aria-label={t("app.profileMenu")}
          className="shrink-0 overflow-hidden rounded-full ring-2 ring-border hover:opacity-80 focus:outline-none focus-visible:ring-primary"
        >
          <ProfileAvatar src={avatarUrl} name={name} />
        </RACButton>
        <SideDrawer open={drawerOpen} onClose={close} title={t("app.profileMenu")}>
          <div className="border-b border-border px-4 py-4">
            <div className="flex items-center gap-3">
              <ProfileAvatar src={avatarUrl} name={name} />
              <span className="truncate text-sm font-medium text-foreground">
                {t("app.welcome", { name })}
              </span>
            </div>
          </div>
          <div className="flex flex-col gap-1 p-2">
            <DrawerItem onClick={handleChangePicture} disabled={!canWrite}>
              <ImagePlus className="size-5" />
              {t("app.changePicture")}
            </DrawerItem>
            {invites.options.length > 0 && (
              <>
                <div className="mt-2 px-3 text-xs font-medium text-muted-foreground">
                  {t("invite.menu.section")}
                </div>
                {invites.options.map((o) => (
                  <DrawerItem
                    key={o.kind}
                    disabled={!canWrite}
                    onClick={() => {
                      close()
                      invites.start(o.kind)
                    }}
                  >
                    <QrCode className="size-5" />
                    {t(o.labelKey)}
                  </DrawerItem>
                ))}
              </>
            )}
            <div className="mt-2 px-3 text-xs font-medium text-muted-foreground">
              {t("app.theme")}
            </div>
            <DrawerItem onClick={() => handleAction("theme-light")}>
              <Sun className="size-5" />
              <span className="flex-1">{t("theme.light")}</span>
              {theme === "light" && <Check className="size-4" />}
            </DrawerItem>
            <DrawerItem onClick={() => handleAction("theme-dark")}>
              <Moon className="size-5" />
              <span className="flex-1">{t("theme.dark")}</span>
              {theme === "dark" && <Check className="size-4" />}
            </DrawerItem>
            <DrawerItem onClick={() => handleAction("theme-system")}>
              <Monitor className="size-5" />
              <span className="flex-1">{t("theme.system")}</span>
              {theme === "system" && <Check className="size-4" />}
            </DrawerItem>
            <div className="mt-2 px-3 text-xs font-medium text-muted-foreground">
              {t("app.language")}
            </div>
            <DrawerItem onClick={() => handleAction("lang-en")}>
              <EnglandFlag />
              <span className="flex-1">{t("language.english")}</span>
              {locale === "en" && <Check className="size-4" />}
            </DrawerItem>
            <DrawerItem onClick={() => handleAction("lang-rs")}>
              <SerbiaFlag />
              <span className="flex-1">{t("language.serbian")}</span>
              {locale === "rs" && <Check className="size-4" />}
            </DrawerItem>
            <div className="my-2 h-px bg-border" />
            <DrawerItem onClick={() => handleAction("logout")} destructive>
              <LogOut className="size-5" />
              {t("app.logout")}
            </DrawerItem>
          </div>
        </SideDrawer>
        {inviteDialog}
      </>
    )
  }

  return (
    <>
      <MenuTrigger>
        {trigger}
        <Popover
          placement="bottom end"
          className="min-w-56 rounded-xl bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/5 dark:ring-foreground/10"
        >
          <Menu className="outline-none" onAction={(key) => handleAction(String(key))}>
            <MenuSection className="px-1">
              <Header className={sectionHeaderClassName}>
                {t("app.welcome", { name })}
              </Header>
              <MenuItem id="change-picture" className={menuItemClassName} isDisabled={!canWrite}>
                <MenuItemContent reason={disabledReason}>
                  <ImagePlus className="size-4" />
                  {t("app.changePicture")}
                </MenuItemContent>
              </MenuItem>
            </MenuSection>
            <Separator className="my-1 h-px bg-border" />
            {invites.options.length > 0 && (
              <>
                <MenuSection className="px-1">
                  <Header className={sectionHeaderClassName}>{t("invite.menu.section")}</Header>
                  {invites.options.map((o) => (
                    <MenuItem
                      key={o.kind}
                      id={`invite-${o.kind}`}
                      className={menuItemClassName}
                      isDisabled={!canWrite}
                    >
                      <MenuItemContent reason={disabledReason}>
                        <QrCode className="size-4" />
                        {t(o.labelKey)}
                      </MenuItemContent>
                    </MenuItem>
                  ))}
                </MenuSection>
                <Separator className="my-1 h-px bg-border" />
              </>
            )}
          <MenuSection className="px-1">
            <Header className={sectionHeaderClassName}>{t("app.theme")}</Header>
            <MenuItem id="theme-light" className={menuItemClassName}>
              <Sun className="size-4" />
              <span className="flex-1">{t("theme.light")}</span>
              {theme === "light" && <Check className="size-4" />}
            </MenuItem>
            <MenuItem id="theme-dark" className={menuItemClassName}>
              <Moon className="size-4" />
              <span className="flex-1">{t("theme.dark")}</span>
              {theme === "dark" && <Check className="size-4" />}
            </MenuItem>
            <MenuItem id="theme-system" className={menuItemClassName}>
              <Monitor className="size-4" />
              <span className="flex-1">{t("theme.system")}</span>
              {theme === "system" && <Check className="size-4" />}
            </MenuItem>
          </MenuSection>
          <Separator className="my-1 h-px bg-border" />
          <MenuSection className="px-1">
            <Header className={sectionHeaderClassName}>{t("app.language")}</Header>
            <MenuItem id="lang-en" className={menuItemClassName}>
              <EnglandFlag />
              <span className="flex-1">{t("language.english")}</span>
              {locale === "en" && <Check className="size-4" />}
            </MenuItem>
            <MenuItem id="lang-rs" className={menuItemClassName}>
              <SerbiaFlag />
              <span className="flex-1">{t("language.serbian")}</span>
              {locale === "rs" && <Check className="size-4" />}
            </MenuItem>
          </MenuSection>
          <Separator className="my-1 h-px bg-border" />
          <MenuSection className="px-1">
            <MenuItem id="logout" className={`${menuItemClassName} text-destructive`}>
              <LogOut className="size-4" />
              {t("app.logout")}
            </MenuItem>
          </MenuSection>
        </Menu>
      </Popover>
      </MenuTrigger>
      {inviteDialog}
    </>
  )
}
