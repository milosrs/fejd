import type { Meta, StoryObj } from "@storybook/react-vite"
import { useEffect, useLayoutEffect, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { UserMenu } from "./UserMenu"
import { useAuthStore } from "../stores/authStore"
import type { Me } from "../hooks/useMe"

const me: Me = {
  approval_status: "approved",
  has_salon: true,
  avatar: "https://picsum.photos/seed/fejd-avatar/200",
  businesses: [],
}

function stubMobileMatchMedia(mobile: boolean) {
  const original = window.matchMedia
  const listeners = new Set<(event: MediaQueryListEvent) => void>()
  const mql = {
    get matches() {
      return mobile
    },
    media: "(max-width: 767px)",
    onchange: null,
    addEventListener: (_type: string, cb: (event: MediaQueryListEvent) => void) => {
      listeners.add(cb)
    },
    removeEventListener: (_type: string, cb: (event: MediaQueryListEvent) => void) => {
      listeners.delete(cb)
    },
    addListener: (cb: (event: MediaQueryListEvent) => void) => {
      listeners.add(cb)
    },
    removeListener: (cb: (event: MediaQueryListEvent) => void) => {
      listeners.delete(cb)
    },
    dispatchEvent: () => false,
  } as unknown as MediaQueryList
  window.matchMedia = () => mql
  return () => {
    window.matchMedia = original
  }
}

function Frame({ mobile = false }: { mobile?: boolean }) {
  const queryClient = useQueryClient()
  const [ready, setReady] = useState(!mobile)

  useLayoutEffect(() => {
    useAuthStore.setState({
      initialized: true,
      authenticated: true,
      userInfo: { sub: "user-1", email: "owner@example.com", name: "Owner" },
      roles: ["Owner"],
    })
    queryClient.setQueryData(["me"], me)
    return () => {
      useAuthStore.setState({
        initialized: true,
        authenticated: false,
        userInfo: null,
        roles: [],
      })
      queryClient.removeQueries({ queryKey: ["me"] })
    }
  }, [queryClient])

  useEffect(() => {
    if (!mobile) return
    const restore = stubMobileMatchMedia(true)
    setReady(true)
    return restore
  }, [mobile])

  if (!ready) return null

  return (
    <div className="flex justify-end p-12">
      <UserMenu />
    </div>
  )
}

const meta: Meta<typeof UserMenu> = {
  title: "UI/UserMenu",
  component: UserMenu,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof UserMenu>

export const Desktop: Story = {
  render: () => <Frame />,
}

export const Mobile: Story = {
  render: () => <Frame mobile />,
}
