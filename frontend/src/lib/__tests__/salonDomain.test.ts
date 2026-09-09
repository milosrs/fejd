import { describe, it, expect } from "vitest"
import { resolveSubdomainSlug } from "../salonDomain"

describe("resolveSubdomainSlug", () => {
  it("extracts the slug from a single-label subdomain", () => {
    expect(resolveSubdomainSlug("dragicevic.fejd.com", "fejd.com")).toBe("dragicevic")
    expect(resolveSubdomainSlug("my-salon.fejd.com", "fejd.com")).toBe("my-salon")
  })

  it("returns null for the apex", () => {
    expect(resolveSubdomainSlug("fejd.com", "fejd.com")).toBeNull()
  })

  it("returns null for the app host", () => {
    expect(resolveSubdomainSlug("www.fejd.com", "fejd.com", "www.fejd.com")).toBeNull()
  })

  it("returns null for multi-label hosts", () => {
    expect(resolveSubdomainSlug("a.b.fejd.com", "fejd.com")).toBeNull()
  })

  it("returns null for lookalike domains", () => {
    expect(resolveSubdomainSlug("evilfejd.com", "fejd.com")).toBeNull()
    expect(resolveSubdomainSlug("fejd.com.evil.com", "fejd.com")).toBeNull()
  })

  it("returns null when no base domain is configured", () => {
    expect(resolveSubdomainSlug("dragicevic.fejd.com", "")).toBeNull()
  })

  it("is case-insensitive and trims whitespace", () => {
    expect(resolveSubdomainSlug("  Dragicevic.FEJD.COM  ", "fejd.com")).toBe("dragicevic")
  })
})
