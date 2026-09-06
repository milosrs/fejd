import "./styles/globals.css"
import App from "./App"
import { createRoot } from "react-dom/client"
import { I18nProvider } from "./lib/i18n"

createRoot(document.getElementById("root")!).render(
  <I18nProvider>
    <App />
  </I18nProvider>,
)

if ("serviceWorker" in navigator) {
  import("workbox-window").then(({ Workbox }) => {
    const wb = new Workbox("/sw.js")
    wb.register()
  })
}
