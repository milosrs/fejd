import type { ReactNode } from "react"
import { Links, Meta, Outlet, Scripts, ScrollRestoration } from "react-router"
import { dehydrate, useQueryClient } from "@tanstack/react-query"
import { ThemeProvider } from "#components/theme-provider"
import { I18nProvider } from "../src/lib/i18n"
import { Toaster } from "../src/components/ui/toaster"
import { AppInit } from "../src/App"
import "../src/styles/globals.css"

export function loader({ request }: { request: Request }) {
  return { hostname: new URL(request.url).hostname }
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
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
        <meta name="theme-color" content="#000000" />
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
  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <Toaster />
      <I18nProvider>
        <AppInit>
          <Outlet />
        </AppInit>
      </I18nProvider>
    </ThemeProvider>
  )
}
