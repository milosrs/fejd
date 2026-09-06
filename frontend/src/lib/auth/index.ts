import { Capacitor } from "@capacitor/core"
import type { AuthAdapter } from "./types"
import { webAdapter } from "./web"
import { nativeAdapter } from "./native"

// Platform-agnostic auth: keycloak-js on web, Capacitor Browser + PKCE on
// native (Android/iOS).
export const auth: AuthAdapter = Capacitor.isNativePlatform() ? nativeAdapter : webAdapter
