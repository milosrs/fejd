import type { Meta, StoryObj } from "@storybook/react"
import { InviteDialog } from "./InviteDialog"
import type { Invitation } from "../../hooks/useInvitations"

const meta: Meta<typeof InviteDialog> = {
  title: "Salon/InviteDialog",
  component: InviteDialog,
}

export default meta
type Story = StoryObj<typeof InviteDialog>

const invitation: Invitation = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  url: "https://app.example.com/invite/abc123def456",
  token: "abc123def456",
  expires_at: "2024-12-31T00:00:00Z",
}

export const Ready: Story = {
  args: {
    open: true,
    onClose: () => {},
    invitation,
  },
}

export const Loading: Story = {
  args: {
    open: true,
    onClose: () => {},
    invitation: null,
    loading: true,
  },
}

export const Error: Story = {
  args: {
    open: true,
    onClose: () => {},
    invitation: null,
    error: "Couldn't generate an invite link.",
  },
}
