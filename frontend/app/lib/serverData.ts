import type { QueryClient, QueryKey } from "@tanstack/react-query"

// API origin for server-side prefetch. At runtime (production) the SSR server
// reaches the backend over the internal network; in local dev it falls back to
// localhost.
const API_ORIGIN =
  process.env.API_ORIGIN || import.meta.env.VITE_API_URL || "http://localhost:8080"

export async function fetchPublic<T = unknown>(path: string): Promise<T> {
  const res = await fetch(`${API_ORIGIN}${path}`)
  if (!res.ok) {
    throw new Error(`fetch ${path} failed: ${res.status}`)
  }
  return (await res.json()) as T
}

// prefetchPublic fetches public data and populates the query cache so the
// client hydrates with it instead of refetching. Best-effort: on failure the
// query is left empty and the client fetches it normally.
export async function prefetchPublic(
  queryClient: QueryClient,
  queryKey: QueryKey,
  path: string,
): Promise<void> {
  try {
    const data = await fetchPublic(path)
    queryClient.setQueryData(queryKey, data)
  } catch {
    // best-effort: leave the query empty
  }
}
