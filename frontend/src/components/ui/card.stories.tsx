import type { Meta, StoryObj } from "@storybook/react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./card"
import { Button } from "./button"

const meta: Meta<typeof Card> = {
  title: "UI/Card",
  component: Card,
}

export default meta
type Story = StoryObj<typeof Card>

export const Default: Story = {
  render: () => (
    <Card className="w-80">
      <CardHeader>
        <CardTitle>Card Title</CardTitle>
        <CardDescription>This is a description of the card content.</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">
          Card content goes here. Use this for any rich content like forms, text, or nested components.
        </p>
      </CardContent>
    </Card>
  ),
}

export const ServiceCard: Story = {
  render: () => (
    <Card className="w-72 cursor-pointer hover:shadow-md hover:border-primary transition-all">
      <CardHeader>
        <CardTitle>Haircut</CardTitle>
        <CardDescription>30 min · $25.00</CardDescription>
      </CardHeader>
      <CardContent>
        <Button variant="outline" className="w-full">Book</Button>
      </CardContent>
    </Card>
  ),
}

export const MultipleCards: Story = {
  render: () => (
    <div className="grid grid-cols-3 gap-4">
      {["Haircut", "Beard Trim", "Coloring"].map((name) => (
        <Card key={name} className="cursor-pointer hover:shadow-md transition-all">
          <CardHeader>
            <CardTitle>{name}</CardTitle>
            <CardDescription>30 min · $25.00</CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" className="w-full">Book</Button>
          </CardContent>
        </Card>
      ))}
    </div>
  ),
}
