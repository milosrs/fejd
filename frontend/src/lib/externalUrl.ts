import { Capacitor } from "@capacitor/core"
import { AppLauncher } from "@capacitor/app-launcher"

// openExternalUrl opens a URL in a new tab on web and via the system on native,
// which lets the OS route to a native app (e.g. Instagram) when it is installed.
export async function openExternalUrl(url: string): Promise<void> {
  if (Capacitor.isNativePlatform()) {
    await AppLauncher.openUrl({ url })
    return
  }
  window.open(url, "_blank", "noopener,noreferrer")
}
