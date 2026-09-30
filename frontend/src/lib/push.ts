import { Capacitor } from "@capacitor/core"
import { FirebaseMessaging } from "@capacitor-firebase/messaging"
import { getApps, initializeApp } from "firebase/app"

import { auth } from "./auth"
import { API_BASE_URL } from "./api"

const TOKEN_STORAGE_KEY = "fejd.push_token"

let tokenListenerRegistered = false

// On web the messaging plugin delegates to the Firebase JS SDK, which expects a
// default Firebase app to exist (the plugin calls getMessaging() without an
// explicit app). On native the Firebase SDK is initialized from the
// google-services.json / GoogleService-Info.plist files instead.
function ensureWebFirebase() {
  if (Capacitor.getPlatform() !== "web") return
  if (getApps().length > 0) return

  initializeApp({
    apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
    authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN ?? "fejd-c8cab.firebaseapp.com",
    projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID ?? "fejd-c8cab",
    storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET ?? "fejd-c8cab.firebasestorage.app",
    messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID ?? "719999385218",
    appId: import.meta.env.VITE_FIREBASE_APP_ID,
  })
}

async function persistToken(token: string) {
  const accessToken = await auth.getToken()
  if (!accessToken) return

  try {
    localStorage.setItem(TOKEN_STORAGE_KEY, token)
  } catch {
    // localStorage may be unavailable; the server registration still succeeds.
  }

  await fetch(`${API_BASE_URL}/api/me/push-token`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ token, platform: Capacitor.getPlatform() }),
  })
}

async function requestTokenAndRegister() {
  const options =
    Capacitor.getPlatform() === "web"
      ? { vapidKey: import.meta.env.VITE_FIREBASE_VAPID_KEY }
      : undefined

  const { token } = await FirebaseMessaging.getToken(options)
  if (token) await persistToken(token)
}

// registerPush requests notification permission and registers this device's FCM
// token with the backend so the user can receive push notifications. Failures
// (permission denied, web not configured) are non-fatal and logged only.
export async function registerPush(): Promise<void> {
  try {
    ensureWebFirebase()

    const supported = await FirebaseMessaging.isSupported()
    if (!supported.isSupported) return

    const permission = await FirebaseMessaging.requestPermissions()
    if (permission.receive !== "granted") return

    await requestTokenAndRegister()

    // FCM refreshes the registration token occasionally on native; re-register
    // it so the backend always holds the current token.
    if (!tokenListenerRegistered) {
      tokenListenerRegistered = true
      await FirebaseMessaging.addListener("tokenReceived", (event) => {
        if (event.token) persistToken(event.token).catch(() => {})
      })
    }
  } catch (err) {
    console.warn("[push] registration failed:", err)
  }
}

// unregisterPush removes this device's token from the backend and deletes the
// FCM registration so the device stops receiving notifications (used on logout).
export async function unregisterPush(): Promise<void> {
  let stored: string | null = null
  try {
    stored = localStorage.getItem(TOKEN_STORAGE_KEY)
  } catch {
    // ignore
  }

  try {
    if (stored) {
      const accessToken = await auth.getToken()
      if (accessToken) {
        await fetch(`${API_BASE_URL}/api/me/push-token`, {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${accessToken}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ token: stored }),
        })
      }
      try {
        localStorage.removeItem(TOKEN_STORAGE_KEY)
      } catch {
        // ignore
      }
    }

    await FirebaseMessaging.deleteToken()
  } catch (err) {
    console.warn("[push] unregister failed:", err)
  }
}
