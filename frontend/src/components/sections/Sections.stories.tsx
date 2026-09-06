import type { Meta, StoryObj } from "@storybook/react"
import { HeroSection } from "./HeroSection"
import { AboutSection } from "./AboutSection"
import { GallerySection } from "./GallerySection"
import { ContactSection } from "./ContactSection"
import { SalonFrame } from "../../stories/salon"

const meta: Meta = {
  title: "Salon/Sections",
  parameters: {
    layout: "fullscreen",
  },
}

export default meta
type Story = StoryObj

export const Hero: Story = {
  render: () => (
    <SalonFrame>
      <HeroSection
        content={{
          headline: "Sharp cuts, done right",
          subheadline: "Book your next appointment in seconds.",
          cta_text: "Book now",
        }}
        slug="fejd"
      />
    </SalonFrame>
  ),
}

export const About: Story = {
  render: () => (
    <SalonFrame>
      <AboutSection
        content={{
          heading: "About us",
          body: "A neighbourhood barbershop run by people who care about the craft.",
        }}
      />
    </SalonFrame>
  ),
}

export const Gallery: Story = {
  render: () => (
    <SalonFrame>
      <GallerySection
        content={{
          heading: "Our work",
          image_urls: [
            "https://picsum.photos/seed/g1/400",
            "https://picsum.photos/seed/g2/400",
            "https://picsum.photos/seed/g3/400",
          ],
        }}
      />
    </SalonFrame>
  ),
}

export const Contact: Story = {
  render: () => (
    <SalonFrame>
      <ContactSection
        content={{
          heading: "Visit us",
          phone: "+46 8 123 45 67",
          email: "hello@fejd.example",
          address: "Main Street 1, Stockholm",
        }}
      />
    </SalonFrame>
  ),
}
