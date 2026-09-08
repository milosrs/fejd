import { useQuery } from "@tanstack/react-query"
import { GET, POST, PUT, DELETE } from "../lib/api"
import type { components } from "../lib/api-types"

const URL_BUSINESS_SERVICES = "/api/business/{slug}/services" as const
const URL_SERVICE_EMPLOYEES = "/api/business/{slug}/services/{serviceID}/employees" as const
const URL_BUSINESS_EMPLOYEES = "/api/business/{slug}/employees" as const
const URL_BUSINESS_SLOTS = "/api/business/{slug}/slots" as const
const URL_MY_APPOINTMENTS = "/api/my/appointments" as const
const URL_ADMIN_EMPLOYEES = "/api/admin/business/{businessID}/employees" as const
const URL_ADMIN_EMPLOYEE_DELETE = "/api/admin/business/{businessID}/employees/{userID}" as const
const URL_ADMIN_EMPLOYEE_IMAGE = "/api/admin/business/{businessID}/employees/{userID}/image" as const
const URL_ADMIN_WORKING_HOURS = "/api/admin/business/{businessID}/employees/{userID}/working-hours" as const
const URL_ADMIN_OVERRIDES = "/api/admin/business/{businessID}/employees/{userID}/overrides" as const
const URL_ADMIN_OVERRIDES_DELETE = "/api/admin/business/{businessID}/employees/{userID}/overrides/{overrideID}" as const
const URL_ADMIN_SERVICES = "/api/admin/business/{businessID}/services" as const
const URL_ADMIN_SERVICES_DELETE = "/api/admin/business/{businessID}/services/{serviceID}" as const
const URL_ADMIN_SERVICE_IMAGE = "/api/admin/business/{businessID}/services/{serviceID}/image" as const
const URL_ADMIN_INVITATIONS = "/api/admin/business/{businessID}/invitations" as const
const URL_APPOINTMENTS = "/api/appointments" as const

type Schemas = components["schemas"]

export type Service = Schemas["dto.Service"]
export type Employee = Schemas["dto.BusinessUser"]
export type TimeSlot = Schemas["dto.TimeSlot"]

export function useServices(slug: string) {
  return useQuery({
    queryKey: ["services", slug],
    queryFn: async () => {
      const { data } = await GET(URL_BUSINESS_SERVICES, {
        params: { path: { slug } },
      })
      return data
    },
    enabled: !!slug,
  })
}

export function useEmployees(slug: string) {
  return useQuery({
    queryKey: ["employees", slug],
    queryFn: async () => {
      const { data } = await GET(URL_BUSINESS_EMPLOYEES, {
        params: { path: { slug } },
      })
      return data
    },
    enabled: !!slug,
  })
}

export function useServiceEmployees(slug: string, serviceId: string) {
  return useQuery({
    queryKey: ["service-employees", slug, serviceId],
    queryFn: async () => {
      const { data } = await GET(URL_SERVICE_EMPLOYEES, {
        params: { path: { slug, serviceID: serviceId } },
      })
      return data
    },
    enabled: !!(slug && serviceId),
  })
}

export function useAvailableSlots(slug: string, serviceId: string, employeeId: string, date: string) {
  return useQuery({
    queryKey: ["slots", slug, serviceId, employeeId, date],
    queryFn: async () => {
      const { data } = await GET(URL_BUSINESS_SLOTS, {
        params: {
          path: { slug },
          query: { service_id: serviceId, employee_id: employeeId, date },
        },
      })
      return data
    },
    enabled: !!(slug && serviceId && employeeId && date),
  })
}

export function useMyAppointments() {
  return useQuery({
    queryKey: ["my-appointments"],
    queryFn: async () => {
      const { data } = await GET(URL_MY_APPOINTMENTS)
      return data
    },
  })
}

export function useAdminEmployees(businessId: string) {
  return useQuery({
    queryKey: ["admin-employees", businessId],
    queryFn: async () => {
      const { data } = await GET(URL_ADMIN_EMPLOYEES, {
        params: { path: { businessID: businessId } },
      })
      return data
    },
    enabled: !!businessId,
  })
}

export function useAdminWorkingHours(businessId: string, userId: string) {
  return useQuery({
    queryKey: ["admin-working-hours", businessId, userId],
    queryFn: async () => {
      const { data } = await GET(
        URL_ADMIN_WORKING_HOURS,
        {
          params: { path: { businessID: businessId, userID: userId } },
        },
      )
      return data
    },
    enabled: !!(businessId && userId),
  })
}

export async function createAppointment(params: {
  business_id: string
  service_id: string
  business_user_id: string
  start_time: string
}) {
  const { data } = await POST(URL_APPOINTMENTS, {
    body: params,
  })
  return data
}

export async function updateWorkingHours(businessId: string, userId: string, workingHours: Schemas["handler.WorkingHoursInput"][]) {
  const { data } = await PUT(
    URL_ADMIN_WORKING_HOURS,
    {
      params: { path: { businessID: businessId, userID: userId } },
      body: { working_hours: workingHours },
    },
  )
  return data
}

export async function createService(businessId: string, service: Schemas["handler.ServiceInput"]) {
  const { data } = await POST(URL_ADMIN_SERVICES, {
    params: { path: { businessID: businessId } },
    body: service,
  })
  return data
}

export async function updateService(businessId: string, serviceId: string, service: Partial<Schemas["handler.ServiceInput"]>) {
  const { data } = await PUT(URL_ADMIN_SERVICES_DELETE, {
    params: { path: { businessID: businessId, serviceID: serviceId } },
    body: service as Schemas["handler.ServiceInput"],
  })
  return data
}

export async function deleteService(businessId: string, serviceId: string) {
  const { data } = await DELETE(URL_ADMIN_SERVICES_DELETE, {
    params: { path: { businessID: businessId, serviceID: serviceId } },
  })
  return data
}

export async function uploadServiceImage(businessId: string, serviceId: string, file: File) {
  const { data } = await POST(URL_ADMIN_SERVICE_IMAGE, {
    params: { path: { businessID: businessId, serviceID: serviceId } },
    body: { file } as any,
  })
  return data
}

export async function addOverride(businessId: string, userId: string, override: Schemas["handler.WorkingHoursOverrideInput"]) {
  const { data } = await POST(
    URL_ADMIN_OVERRIDES,
    {
      params: { path: { businessID: businessId, userID: userId } },
      body: override,
    },
  )
  return data
}

export async function deleteOverride(businessId: string, userId: string, overrideId: string) {
  const { data } = await DELETE(
    URL_ADMIN_OVERRIDES_DELETE,
    {
      params: { path: { businessID: businessId, userID: userId, overrideID: overrideId } },
    },
  )
  return data
}

export async function inviteEmployee(
  businessId: string,
  input: { name: string; email: string; service_ids: string[] },
) {
  const { data } = await POST(URL_ADMIN_EMPLOYEES, {
    params: { path: { businessID: businessId } },
    body: input,
  })
  return data
}

export type Invitation = Schemas["handler.InvitationResponse"]

export async function createInvitation(businessId: string, body?: { expires_in_hours?: number }) {
  const { data } = await POST(URL_ADMIN_INVITATIONS, {
    params: { path: { businessID: businessId } },
    body: body ?? {},
  })
  return data
}

export async function removeEmployee(businessId: string, userId: string) {
  const { data } = await DELETE(URL_ADMIN_EMPLOYEE_DELETE, {
    params: { path: { businessID: businessId, userID: userId } },
  })
  return data
}

export async function uploadEmployeeImage(businessId: string, userId: string, file: File) {
  const { data } = await POST(URL_ADMIN_EMPLOYEE_IMAGE, {
    params: { path: { businessID: businessId, userID: userId } },
    body: { file } as any,
  })
  return data
}
