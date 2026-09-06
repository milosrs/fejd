import { describe, it, expect } from "vitest"
import { isOwnerOfBusiness } from "../ownership"
import type { Me } from "../../hooks/useMe"

function me(businesses: Me["businesses"]): Me {
  return { approval_status: "approved", has_salon: true, businesses }
}

describe("isOwnerOfBusiness", () => {
  it("is true when the caller is an admin of the salon", () => {
    expect(
      isOwnerOfBusiness(
        me([{ id: "1", name: "Salon", slug: "my-salon", role: "admin" }]),
        "my-salon",
      ),
    ).toBe(true)
  })

  it("is false when the caller is only an employee", () => {
    expect(
      isOwnerOfBusiness(
        me([{ id: "1", name: "Salon", slug: "my-salon", role: "employee" }]),
        "my-salon",
      ),
    ).toBe(false)
  })

  it("is false for a different salon", () => {
    expect(
      isOwnerOfBusiness(
        me([{ id: "1", name: "Salon", slug: "other-salon", role: "admin" }]),
        "my-salon",
      ),
    ).toBe(false)
  })

  it("is false without a /api/me response", () => {
    expect(isOwnerOfBusiness(undefined, "my-salon")).toBe(false)
  })
})
