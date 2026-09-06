import { describe, it, expect } from "vitest"
import { resolveOnboardingState } from "../onboarding"

describe("resolveOnboardingState", () => {
  it("routes pending users to the pending gate regardless of salon", () => {
    expect(resolveOnboardingState("pending", false)).toBe("pending")
    expect(resolveOnboardingState("pending", true)).toBe("pending")
  })

  it("routes approved users without a salon to setup", () => {
    expect(resolveOnboardingState("approved", false)).toBe("setup")
  })

  it("routes approved users with a salon to ready", () => {
    expect(resolveOnboardingState("approved", true)).toBe("ready")
  })

  it("routes everything else to ready", () => {
    expect(resolveOnboardingState("rejected", false)).toBe("ready")
    expect(resolveOnboardingState(null, false)).toBe("ready")
    expect(resolveOnboardingState(undefined, false)).toBe("ready")
    expect(resolveOnboardingState("", false)).toBe("ready")
  })
})
