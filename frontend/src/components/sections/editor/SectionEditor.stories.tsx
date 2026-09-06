import type { Meta, StoryObj } from "@storybook/react"
import { SectionEditor } from "./SectionEditor"
import { SalonFrame, mockSections } from "../../../stories/salon"

const meta: Meta<typeof SectionEditor> = {
  title: "Salon/SectionEditor",
  component: SectionEditor,
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj<typeof SectionEditor>

export const Hero: Story = {
  render: () => (
    <SalonFrame sections={mockSections}>
      <SectionEditor
        section={mockSections[0]}
        onClose={() => {}}
        onSave={() => {}}
      />
    </SalonFrame>
  ),
}

export const About: Story = {
  render: () => (
    <SalonFrame sections={mockSections}>
      <SectionEditor
        section={mockSections[1]}
        onClose={() => {}}
        onSave={() => {}}
      />
    </SalonFrame>
  ),
}

export const Gallery: Story = {
  render: () => (
    <SalonFrame sections={mockSections}>
      <SectionEditor
        section={mockSections[2]}
        onClose={() => {}}
        onSave={() => {}}
      />
    </SalonFrame>
  ),
}

export const Contact: Story = {
  render: () => (
    <SalonFrame sections={mockSections}>
      <SectionEditor
        section={mockSections[3]}
        onClose={() => {}}
        onSave={() => {}}
      />
    </SalonFrame>
  ),
}
