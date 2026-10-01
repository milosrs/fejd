import { useEffect, useLayoutEffect, useState } from "react"

// useLayoutEffect on the client, useEffect on the server, so SSR does not log
// "useLayoutEffect does nothing on the server" warnings.
const useIsomorphicLayoutEffect =
  typeof window !== "undefined" ? useLayoutEffect : useEffect

// useHeaderHeightMeasure publishes the rendered height of the element it is
// attached to as a CSS variable on <html>, so descendants can size themselves
// against the remaining viewport (e.g. a full-bleed hero below the app + salon
// headers). The returned value is a stable ref callback.
export function useHeaderHeightMeasure(variable: `--${string}`) {
  const [el, setEl] = useState<HTMLElement | null>(null)

  useIsomorphicLayoutEffect(() => {
    if (!el) return

    const update = () => {
      document.documentElement.style.setProperty(variable, `${el.offsetHeight}px`)
    }
    update()

    const observer = new ResizeObserver(update)
    observer.observe(el)

    return () => {
      observer.disconnect()
      document.documentElement.style.removeProperty(variable)
    }
  }, [el, variable])

  return setEl
}
