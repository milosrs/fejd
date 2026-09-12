import { describe, it, expect } from "vitest"
import { instagramUrl, instagramHandle } from "../instagram"

describe("instagramUrl", () => {
  it("passes through a full URL", () => {
    expect(instagramUrl("https://www.instagram.com/fejd/")).toBe(
      "https://www.instagram.com/fejd/",
    )
  })

  it("normalizes a bare handle", () => {
    expect(instagramUrl("fejd")).toBe("https://www.instagram.com/fejd")
  })

  it("strips a leading @", () => {
    expect(instagramUrl("@fejd")).toBe("https://www.instagram.com/fejd")
  })

  it("adds the scheme to a schemeless instagram URL", () => {
    expect(instagramUrl("instagram.com/fejd")).toBe("https://instagram.com/fejd")
    expect(instagramUrl("www.instagram.com/fejd")).toBe(
      "https://www.instagram.com/fejd",
    )
  })

  it("returns empty for blank input", () => {
    expect(instagramUrl("")).toBe("")
    expect(instagramUrl("   ")).toBe("")
  })
})

describe("instagramHandle", () => {
  it("extracts the handle from a full URL", () => {
    expect(instagramHandle("https://www.instagram.com/fejdbarbershop")).toBe(
      "@fejdbarbershop",
    )
  })

  it("ignores trailing slashes and query params", () => {
    expect(instagramHandle("https://www.instagram.com/fejd/")).toBe("@fejd")
    expect(instagramHandle("https://www.instagram.com/fejd/?igsh=abc")).toBe(
      "@fejd",
    )
  })

  it("extracts the handle from a bare handle", () => {
    expect(instagramHandle("fejd")).toBe("@fejd")
    expect(instagramHandle("@fejd")).toBe("@fejd")
  })

  it("returns empty for blank input", () => {
    expect(instagramHandle("")).toBe("")
    expect(instagramHandle("   ")).toBe("")
  })
})
