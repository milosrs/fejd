import type { ReactNode } from "react"
import { Links, Meta, Outlet, Scripts, ScrollRestoration, useLoaderData } from "react-router"
import { dehydrate, useQueryClient } from "@tanstack/react-query"
import { ThemeProvider } from "#components/theme-provider"
import { I18nProvider } from "../src/lib/i18n"
import { Toaster } from "../src/components/ui/toaster"
import { AppInit } from "../src/App"
import { readLocaleCookie } from "./lib/locale"
import "../src/styles/globals.css"

export function loader({ request }: { request: Request }) {
  return { hostname: new URL(request.url).hostname, locale: readLocaleCookie(request) }
}

export function meta() {
  return [
    { title: "fejd - Book haircut appointments" },
    { name: "description", content: "Book haircut appointments at salons near you." },
  ]
}

export function Layout({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const dehydratedState = typeof window === "undefined" ? dehydrate(queryClient) : null

  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
        <meta name="theme-color" content="#000000" />
        <script
          dangerouslySetInnerHTML={{
            __html: `;(function () {
  var theme = "system"
  try {
    var stored = localStorage.getItem("vite-ui-theme")
    if (stored === "light" || stored === "dark" || stored === "system") {
      theme = stored
    }
  } catch (e) {}
  if (theme === "system") {
    theme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
  }
  document.documentElement.classList.add(theme)
})()`,
          }}
        />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        {dehydratedState && (
          <script
            dangerouslySetInnerHTML={{
              __html: `window.__REACT_QUERY_STATE__ = ${JSON.stringify(dehydratedState)}`,
            }}
          />
        )}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  )
}

export default function App() {
  const { locale } = useLoaderData<typeof loader>()
  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <Toaster />
      <I18nProvider initialLocale={locale}>
        <AppInit>
          <Outlet />
        </AppInit>
      </I18nProvider>
    </ThemeProvider>
  )
}
