import { startTransition, StrictMode } from "react"
import { hydrateRoot } from "react-dom/client"
import { HydratedRouter } from "react-router/dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

declare global {
  interface Window {
    __REACT_QUERY_STATE__?: unknown
  }
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30000 },
  },
})

if (typeof window !== "undefined" && window.__REACT_QUERY_STATE__) {
  queryClient.hydrate(window.__REACT_QUERY_STATE__)
}

startTransition(() => {
  hydrateRoot(
    document,
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <HydratedRouter />
      </QueryClientProvider>
    </StrictMode>,
  )
})

if ("serviceWorker" in navigator) {
  import("workbox-window").then(({ Workbox }) => {
    const wb = new Workbox("/sw.js")
    wb.register()
  })
}
