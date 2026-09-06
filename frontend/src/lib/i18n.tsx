import { createContext, useContext, useMemo, useState, type ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { GET } from "./api"

export const DEFAULT_LOCALE = "en"

interface I18nContextValue {
  locale: string
  setLocale: (locale: string) => void
  ready: boolean
  t: (key: string) => string
  pickLocalized: <T>(localized: Record<string, T> | undefined) => T | undefined
}

const I18nContext = createContext<I18nContextValue | null>(null)

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
      t: (key) => translations?.[key] ?? key,
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
  if (!ctx) {
    throw new Error("useI18n must be used within an I18nProvider")
  }
  return ctx
}
