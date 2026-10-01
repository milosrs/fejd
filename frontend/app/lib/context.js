import { createContext, RouterContextProvider } from "react-router"
import { QueryClient } from "@tanstack/react-query"

// The query-client context is shared via globalThis so the custom Node server
// (server.js) and the bundled app (loaders + entry.server) use the same
// RouterContext key. React Router v8's getLoadContext must return a
// RouterContextProvider, and loaders read values back with context.get().
const g = globalThis

export const queryClientContext =
  g.__fejdQueryClientContext ??
  (g.__fejdQueryClientContext = createContext(null))

function makeQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: 1, staleTime: 30000 } } })
}

export function createLoadContext() {
  const provider = new RouterContextProvider()
  provider.set(queryClientContext, makeQueryClient())
  return provider
}

// Fallback used when no RouterContextProvider is available (e.g. react-router
// dev, which does not run a custom getLoadContext). Dev is single-user, so a
// module singleton is safe there.
let fallbackQueryClient = null

export function getFallbackQueryClient() {
  return (fallbackQueryClient ??= makeQueryClient())
}

// resolveQueryClient returns the per-request query client (production) or the
// dev fallback. Accepts the loader/handler `context` which may be undefined in
// dev.
export function resolveQueryClient(context) {
  const qc = context?.get?.(queryClientContext)
  return qc ?? getFallbackQueryClient()
}
