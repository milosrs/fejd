// Firebase Cloud Messaging service worker (web). Handles notifications while
// the tab is closed/inactive and opens the target URL when one is tapped.
//
// The Firebase web config here is public client config (it is embedded in the
// JS bundle anyway); the VAPID key and the server-side service account are the
// only credentials that stay out of the client.

importScripts("https://www.gstatic.com/firebasejs/10.14.1/firebase-app-compat.js")
importScripts("https://www.gstatic.com/firebasejs/10.14.1/firebase-messaging-compat.js")

firebase.initializeApp({
  apiKey: "AIzaSyAsnuUEUfWLn-m_1wga6utKRym39-a8iLQ",
  authDomain: "fejd-c8cab.firebaseapp.com",
  projectId: "fejd-c8cab",
  storageBucket: "fejd-c8cab.firebasestorage.app",
  messagingSenderId: "719999385218",
  appId: "1:719999385218:web:a899b004c507bcd4f03e9b",
})

const messaging = firebase.messaging()

messaging.onBackgroundMessage((payload) => {
  const notification = payload.notification || {}
  const link = payload.fcmOptions?.link || payload.data?.link || "/"

  self.registration.showNotification(notification.title || "", {
    body: notification.body || "",
    icon: notification.icon || undefined,
    image: notification.image || undefined,
    badge: notification.icon || undefined,
    data: { link },
  })
})

self.addEventListener("notificationclick", (event) => {
  event.notification.close()
  const link = event.notification.data?.link || "/"

  event.waitUntil(
    clients.matchAll({ type: "window", includeUncontrolled: true }).then((windowClients) => {
      for (const client of windowClients) {
        if ("navigate" in client && "focus" in client) {
          client.navigate(link)
          return client.focus()
        }
      }
      return clients.openWindow(link)
    }),
  )
})
