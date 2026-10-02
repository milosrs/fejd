import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { GET } from "./api"

export const DEFAULT_LOCALE = "en"
export const SUPPORTED_LOCALES = ["en", "rs"]
const LOCALE_STORAGE_KEY = "fejd-locale"
const LOCALE_COOKIE = "fejd-locale"

function readLocaleCookie(): string | undefined {
  if (typeof document === "undefined") return undefined
  const match = document.cookie.match(new RegExp(`(?:^|;\\s*)${LOCALE_COOKIE}=([^;]+)`))
  return match?.[1]
}

function writeLocaleCookie(locale: string) {
  if (typeof document === "undefined") return
  document.cookie = `${LOCALE_COOKIE}=${locale}; path=/; max-age=31536000; SameSite=Lax`
}

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

export function I18nProvider({
  children,
  initialLocale,
}: {
  children: ReactNode
  initialLocale?: string
}) {
  // Start from the server-resolved locale (read from the cookie by the loader)
  // so SSR and the first client render match. Persisted-but-uncached locales
  // are applied in a client-only effect.
  const [locale, setLocaleState] = useState(
    initialLocale && SUPPORTED_LOCALES.includes(initialLocale) ? initialLocale : DEFAULT_LOCALE,
  )

  useEffect(() => {
    try {
      const stored = readLocaleCookie() ?? localStorage.getItem(LOCALE_STORAGE_KEY)
      if (stored && SUPPORTED_LOCALES.includes(stored) && stored !== locale) {
        setLocaleState(stored)
      }
    } catch {
      // localStorage unavailable (private/strict mode) — keep default.
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const setLocale = useCallback((next: string) => {
    writeLocaleCookie(next)
    try {
      localStorage.setItem(LOCALE_STORAGE_KEY, next)
    } catch {
      // ignore: locale still applies for this session
    }
    setLocaleState(next)
  }, [])

  const { data: translations } = useQuery({
    queryKey: ["i18n", locale],
    queryFn: async () => {
      const { data } = await GET("/api/i18n/{locale}", {
        params: { path: { locale } },
      })
      return data ?? {}
    },
    staleTime: Infinity,
    // Keep the previous locale's strings on screen while the new locale loads,
    // so switching languages never flashes raw i18n keys.
    placeholderData: (previousData) => previousData,
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
    [locale, setLocale, translations],
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const ctx = useContext(I18nContext)
  return ctx ?? fallbackI18n
}
