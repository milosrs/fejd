import { PassThrough } from "node:stream"
import { renderToPipeableStream } from "react-dom/server"
import { ServerRouter } from "react-router"
import { createReadableStreamFromReadable } from "@react-router/node"
import { QueryClientProvider } from "@tanstack/react-query"
import { isbot } from "isbot"
import { resolveQueryClient } from "./lib/context"

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  routerContext: unknown,
  loadContext: unknown,
) {
  const queryClient = resolveQueryClient(loadContext)
  // Crawlers get the fully-rendered HTML (meta + JSON-LD); browsers get the
  // shell early so a slow backend can't block the initial paint.
  const readyOption: "onShellReady" | "onAllReady" = isbot(request.headers.get("user-agent"))
    ? "onAllReady"
    : "onShellReady"

  return new Promise((resolve, reject) => {
    const { pipe, abort } = renderToPipeableStream(
      <QueryClientProvider client={queryClient}>
        <ServerRouter context={routerContext} url={request.url} />
      </QueryClientProvider>,
      {
        [readyOption]() {
          const body = new PassThrough()
          responseHeaders.set("Content-Type", "text/html")
          resolve(
            new Response(createReadableStreamFromReadable(body), {
              headers: responseHeaders,
              status: responseStatusCode,
            }),
          )
          pipe(body)
        },
        onShellError(error: unknown) {
          reject(error)
        },
        onError(error: unknown) {
          responseStatusCode = 500
          console.error(error)
        },
      },
    )

    setTimeout(() => abort(), 30_000)
  })
}
