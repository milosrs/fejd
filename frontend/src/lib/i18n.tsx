import { createContext, useContext, useMemo, useState, type ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { GET } from "./api"

export const DEFAULT_LOCALE = "en"

interface I18nContextValue {
  locale: string
  setLocale: (locale: string) => void
  ready: boolean
  t: (key: string, params?: Record<string, string | number>) => string
  pickLocalized: <T>(localized: Record<string, T> | undefined) => T | undefined
}

const I18nContext = createContext<I18nContextValue | null>(null)

function interpolate(
  value: string,
  params?: Record<string, string | number>,
): string {
  if (!params) return value
  let out = value
  for (const [k, v] of Object.entries(params)) {
    out = out.split(`{${k}}`).join(String(v))
  }
  return out
}

const fallbackI18n: I18nContextValue = {
  locale: DEFAULT_LOCALE,
  setLocale: () => {},
  ready: true,
  t: (key, params) => interpolate(key, params),
  pickLocalized: (localized) => {
    if (!localized) return undefined
    if (DEFAULT_LOCALE in localized) return localized[DEFAULT_LOCALE]
    const first = Object.keys(localized)[0]
    return first ? localized[first] : undefined
  },
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState(DEFAULT_LOCALE)

  const { data: translations } = useQuery({
    queryKey: ["i18n", locale],
    queryFn: async () => {
      const { data } = await GET("/api/i18n/{locale}", {
        params: { path: { locale } },
      })
      return data ?? {}
    },
    staleTime: Infinity,
  })

  const value = useMemo<I18nContextValue>(
    () => ({
      locale,
      setLocale,
      ready: translations != null,
      t: (key, params) => interpolate(translations?.[key] ?? key, params),
      pickLocalized: (localized) => {
        if (!localized) return undefined
        if (locale in localized) return localized[locale]
        if (DEFAULT_LOCALE in localized) return localized[DEFAULT_LOCALE]
        const first = Object.keys(localized)[0]
        return first ? localized[first] : undefined
      },
    }),
    [locale, translations],
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const ctx = useContext(I18nContext)
  return ctx ?? fallbackI18n
}
