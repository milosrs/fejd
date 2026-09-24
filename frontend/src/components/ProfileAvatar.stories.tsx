import type { Meta, StoryObj } from "@storybook/react-vite"
import { ProfileAvatar } from "./ProfileAvatar"

const meta: Meta<typeof ProfileAvatar> = {
  title: "UI/ProfileAvatar",
  component: ProfileAvatar,
  parameters: {
    layout: "centered",
  },
  args: {
    name: "Owner",
  },
}

export default meta
type Story = StoryObj<typeof ProfileAvatar>

export const Empty: Story = {}

export const WithImage: Story = {
  args: {
    src: "https://picsum.photos/seed/fejd-avatar/200",
  },
}
