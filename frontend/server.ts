import { createServer } from "node:http"
import { createRequestListener } from "@react-router/node"
import * as build from "./build/server/index.js"
import { createLoadContext } from "./app/lib/context.ts"

const requestListener = createRequestListener({
  build,
  getLoadContext: () => createLoadContext(),
})

const port = Number(process.env.PORT || 3000)
createServer(requestListener).listen(port, () => {
  console.log(`SSR server listening on :${port}`)
})
