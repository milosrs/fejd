import { describe, it, expect } from "vitest"
import { parseInviteToken } from "../invite"

describe("parseInviteToken", () => {
  it("extracts the token from an https invite URL", () => {
    expect(parseInviteToken("https://app.example.com/invite/abc123")).toBe("abc123")
  })

  it("extracts the token from a custom-scheme invite URL", () => {
    expect(parseInviteToken("fejd://invite/abc123")).toBe("abc123")
  })

  it("ignores extra path segments after the token", () => {
    expect(parseInviteToken("https://app.example.com/invite/abc123/extra")).toBe("abc123")
  })

  it("returns null for non-invite URLs", () => {
    expect(parseInviteToken("https://app.example.com/salon/fejd")).toBeNull()
    expect(parseInviteToken("fejd://callback?code=xyz")).toBeNull()
  })

  it("returns null for invite URLs without a token", () => {
    expect(parseInviteToken("https://app.example.com/invite/")).toBeNull()
  })

  it("returns null for malformed URLs", () => {
    expect(parseInviteToken("not a url")).toBeNull()
  })
})
