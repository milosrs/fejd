import "./index.css"
import App from "./App"
import { createRoot } from "react-dom/client"

createRoot(document.getElementById("root")!).render(<App />)

if ("serviceWorker" in navigator) {
  import("workbox-window").then(({ Workbox }) => {
    const wb = new Workbox("/sw.js")
    wb.register()
  })
}
