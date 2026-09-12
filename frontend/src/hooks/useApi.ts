import { useQuery } from "@tanstack/react-query"
import { GET, POST, PUT, DELETE } from "../lib/api"
import type { components } from "../lib/api-types"

const URL_BUSINESS_SERVICES = "/api/business/{slug}/services" as const
const URL_SERVICE_EMPLOYEES = "/api/business/{slug}/services/{serviceID}/employees" as const
const URL_BUSINESS_EMPLOYEES = "/api/business/{slug}/employees" as const
const URL_BUSINESS_SLOTS = "/api/business/{slug}/slots" as const
const URL_MY_APPOINTMENTS = "/api/my/appointments" as const
const URL_MY_APPOINTMENTS_DELETE = "/api/my/appointments/{appointmentID}" as const
const URL_SALON_POLICY = "/api/admin/business/{businessID}/policy" as const
const URL_ADMIN_EMPLOYEES = "/api/admin/business/{businessID}/employees" as const
const URL_ADMIN_EMPLOYEE_DELETE = "/api/admin/business/{businessID}/employees/{userID}" as const
const URL_ADMIN_EMPLOYEE_IMAGE = "/api/admin/business/{businessID}/employees/{userID}/image" as const
const URL_ADMIN_BUSINESS_IMAGES = "/api/admin/business/{businessID}/images" as const
const URL_ADMIN_WORKING_HOURS = "/api/admin/business/{businessID}/employees/{userID}/working-hours" as const
const URL_ADMIN_OVERRIDES = "/api/admin/business/{businessID}/employees/{userID}/overrides" as const
const URL_ADMIN_OVERRIDES_DELETE = "/api/admin/business/{businessID}/employees/{userID}/overrides/{overrideID}" as const
const URL_ADMIN_SERVICES = "/api/admin/business/{businessID}/services" as const
const URL_ADMIN_SERVICES_DELETE = "/api/admin/business/{businessID}/services/{serviceID}" as const
const URL_ADMIN_SERVICE_IMAGE = "/api/admin/business/{businessID}/services/{serviceID}/image" as const
const URL_ADMIN_SERVICE_EMPLOYEES = "/api/admin/business/{businessID}/services/{serviceID}/employees" as const
const URL_ADMIN_INVITATIONS = "/api/admin/business/{businessID}/invitations" as const
const URL_MY_UNAVAILABILITY = "/api/admin/business/{businessID}/me/unavailability" as const
const URL_MY_UNAVAILABILITY_DELETE = "/api/admin/business/{businessID}/me/unavailability/{unavailabilityID}" as const
const URL_MY_RESERVATIONS = "/api/admin/business/{businessID}/me/appointments" as const
const URL_MY_RESERVATIONS_DELETE = "/api/admin/business/{businessID}/me/appointments/{appointmentID}" as const
const URL_MY_RESERVATIONS_NO_SHOW = "/api/admin/business/{businessID}/me/appointments/{appointmentID}/no-show" as const
const URL_MY_SERVICES = "/api/admin/business/{businessID}/me/services" as const
const URL_CUSTOMERS = "/api/admin/business/{businessID}/customers" as const
const URL_INVITATION = "/api/invitations/{token}" as const
const URL_INVITATION_ACCEPT = "/api/invitations/{token}/accept" as const
const URL_APPOINTMENTS = "/api/appointments" as const
const URL_ME_AVATAR = "/api/me/avatar" as const
const URL_BUSINESSES = "/api/businesses" as const
const URL_ADMIN_BUSINESS_NAME = "/api/admin/business/{businessID}/name" as const
const URL_ADMIN_APPOINTMENTS = "/api/admin/business/{businessID}/appointments" as const
const URL_ADMIN_APPOINTMENTS_ACCEPT = "/api/admin/business/{businessID}/appointments/{appointmentID}/accept" as const
const URL_ADMIN_APPOINTMENTS_REJECT = "/api/admin/business/{businessID}/appointments/{appointmentID}/reject" as const
const URL_ADMIN_UNAVAILABILITY = "/api/admin/business/{businessID}/unavailability" as const
const URL_ADMIN_UNAVAILABILITY_ACCEPT = "/api/admin/business/{businessID}/unavailability/{unavailabilityID}/accept" as const
const URL_ADMIN_UNAVAILABILITY_REJECT = "/api/admin/business/{businessID}/unavailability/{unavailabilityID}/reject" as const

type Schemas = components["schemas"]

export type Service = Schemas["dto.Service"]
export type Employee = Schemas["dto.BusinessUser"]
export type TimeSlot = Schemas["dto.TimeSlot"]
export type DirectoryBusiness = Schemas["dto.DirectoryBusiness"]

export function useBusinesses() {
  return useQuery({
    queryKey: ["businesses"],
    queryFn: async () => {
      const { data } = await GET(URL_BUSINESSES)
      return data
    },
  })
}

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

export async function cancelAppointment(appointmentId: string, reason: string) {
  const { data } = await DELETE(URL_MY_APPOINTMENTS_DELETE, {
    params: { path: { appointmentID: appointmentId } },
    body: { cancellation_reason: reason },
  })
  return data
}

export async function getSalonPolicy(businessId: string) {
  const { data } = await GET(URL_SALON_POLICY, {
    params: { path: { businessID: businessId } },
  })
  return data
}

export function useSalonPolicy(businessId: string) {
  return useQuery({
    queryKey: ["salon-policy", businessId],
    queryFn: () => getSalonPolicy(businessId),
    enabled: !!businessId,
  })
}

export type SalonPolicyInput = {
  cancellation_lead_hours: number
  no_show_after_hours: number
  slot_interval_minutes: number
  timezone?: string
  working_hours: components["schemas"]["handler.BusinessHoursInput"][]
}

export async function updateSalonPolicy(businessId: string, policy: SalonPolicyInput) {
  const { data } = await PUT(URL_SALON_POLICY, {
    params: { path: { businessID: businessId } },
    body: policy,
  })
  return data
}

export async function renameBusiness(businessId: string, name: string) {
  const { data } = await PUT(URL_ADMIN_BUSINESS_NAME, {
    params: { path: { businessID: businessId } },
    body: { name },
  })
  return data
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
  const formData = new FormData()
  formData.append("file", file)
  const { data } = await POST(URL_ADMIN_SERVICE_IMAGE, {
    params: { path: { businessID: businessId, serviceID: serviceId } },
    body: formData as any,
  })
  return data
}

export function useServiceEmployeesAdmin(businessId: string, serviceId: string) {
  return useQuery({
    queryKey: ["admin-service-employees", businessId, serviceId],
    queryFn: async () => {
      const { data } = await GET(URL_ADMIN_SERVICE_EMPLOYEES, {
        params: { path: { businessID: businessId, serviceID: serviceId } },
      })
      return data ?? []
    },
    enabled: !!(businessId && serviceId),
  })
}

export async function setServiceEmployees(
  businessId: string,
  serviceId: string,
  businessUserIds: string[],
) {
  const { data } = await PUT(URL_ADMIN_SERVICE_EMPLOYEES, {
    params: { path: { businessID: businessId, serviceID: serviceId } },
    body: { business_user_ids: businessUserIds },
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
export type PublicInvitation = Schemas["handler.PublicInvitationResponse"]

export async function createInvitation(businessId: string, body?: { expires_in_hours?: number }) {
  const { data } = await POST(URL_ADMIN_INVITATIONS, {
    params: { path: { businessID: businessId } },
    body: body ?? {},
  })
  return data
}

export async function getInvitation(token: string) {
  const { data } = await GET(URL_INVITATION, {
    params: { path: { token } },
  })
  return data
}

export async function acceptInvitation(token: string) {
  const { data } = await POST(URL_INVITATION_ACCEPT, {
    params: { path: { token } },
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
  const formData = new FormData()
  formData.append("file", file)
  const { data } = await POST(URL_ADMIN_EMPLOYEE_IMAGE, {
    params: { path: { businessID: businessId, userID: userId } },
    body: formData as any,
  })
  return data
}

export async function uploadAvatar(file: File) {
  const formData = new FormData()
  formData.append("file", file)
  const { data } = await POST(URL_ME_AVATAR, {
    body: formData as any,
  })
  return data
}

export type BusinessImagePurpose = "hero" | "logo" | "background" | "gallery"

export async function uploadBusinessImage(
  businessId: string,
  file: File,
  purpose: BusinessImagePurpose,
) {
  const formData = new FormData()
  formData.append("purpose", purpose)
  formData.append("file", file)
  const { data } = await POST(URL_ADMIN_BUSINESS_IMAGES, {
    params: { path: { businessID: businessId } },
    body: formData as any,
  })
  return data
}

export type EmployeeUnavailability = Schemas["dto.EmployeeUnavailability"]

export function useMyUnavailability(businessId: string) {
  return useQuery({
    queryKey: ["my-unavailability", businessId],
    queryFn: async () => {
      const { data } = await GET(URL_MY_UNAVAILABILITY, {
        params: { path: { businessID: businessId } },
      })
      return data
    },
    enabled: !!businessId,
  })
}

export async function reserveOwnSlot(
  businessId: string,
  input: Schemas["handler.CreateUnavailabilityRequest"],
) {
  const { data } = await POST(URL_MY_UNAVAILABILITY, {
    params: { path: { businessID: businessId } },
    body: input,
  })
  return data
}

export async function deleteOwnSlot(businessId: string, unavailabilityId: string) {
  const { data } = await DELETE(URL_MY_UNAVAILABILITY_DELETE, {
    params: { path: { businessID: businessId, unavailabilityID: unavailabilityId } },
  })
  return data
}

export function useBusinessAppointments(businessId: string) {
  return useQuery({
    queryKey: ["business-appointments", businessId],
    queryFn: async () => {
      const { data } = await GET(URL_ADMIN_APPOINTMENTS, {
        params: { path: { businessID: businessId } },
      })
      return data
    },
    enabled: !!businessId,
  })
}

export function useBusinessUnavailability(businessId: string) {
  return useQuery({
    queryKey: ["business-unavailability", businessId],
    queryFn: async () => {
      const { data } = await GET(URL_ADMIN_UNAVAILABILITY, {
        params: { path: { businessID: businessId } },
      })
      return data
    },
    enabled: !!businessId,
  })
}

export async function acceptAppointment(businessId: string, appointmentId: string) {
  const { data } = await POST(URL_ADMIN_APPOINTMENTS_ACCEPT, {
    params: { path: { businessID: businessId, appointmentID: appointmentId } },
  })
  return data
}

export async function rejectAppointment(businessId: string, appointmentId: string, reason: string) {
  const { data } = await POST(URL_ADMIN_APPOINTMENTS_REJECT, {
    params: { path: { businessID: businessId, appointmentID: appointmentId } },
    body: { reason },
  })
  return data
}

export async function acceptUnavailability(businessId: string, unavailabilityId: string) {
  const { data } = await POST(URL_ADMIN_UNAVAILABILITY_ACCEPT, {
    params: { path: { businessID: businessId, unavailabilityID: unavailabilityId } },
  })
  return data
}

export async function rejectUnavailability(businessId: string, unavailabilityId: string, reason: string) {
  const { data } = await POST(URL_ADMIN_UNAVAILABILITY_REJECT, {
    params: { path: { businessID: businessId, unavailabilityID: unavailabilityId } },
    body: { reason },
  })
  return data
}

export type Appointment = Schemas["dto.Appointment"]
export type Customer = Schemas["dto.Customer"]

export function useMyReservations(businessId: string, date: string) {
  return useQuery({
    queryKey: ["my-reservations", businessId, date],
    queryFn: async () => {
      const { data } = await GET(URL_MY_RESERVATIONS, {
        params: { path: { businessID: businessId }, query: { date } },
      })
      return data
    },
    enabled: !!(businessId && date),
  })
}

export async function cancelReservation(businessId: string, appointmentId: string, reason: string) {
  const { data } = await DELETE(URL_MY_RESERVATIONS_DELETE, {
    params: { path: { businessID: businessId, appointmentID: appointmentId } },
    body: { cancellation_reason: reason },
  })
  return data
}

export async function markNoShow(businessId: string, appointmentId: string) {
  const { data } = await POST(URL_MY_RESERVATIONS_NO_SHOW, {
    params: { path: { businessID: businessId, appointmentID: appointmentId } },
  })
  return data
}

export function useMyServices(businessId: string) {
  return useQuery({
    queryKey: ["my-services", businessId],
    queryFn: async () => {
      const { data } = await GET(URL_MY_SERVICES, {
        params: { path: { businessID: businessId } },
      })
      return data
    },
    enabled: !!businessId,
  })
}

export function useCustomers(businessId: string) {
  return useQuery({
    queryKey: ["customers", businessId],
    queryFn: async () => {
      const { data } = await GET(URL_CUSTOMERS, {
        params: { path: { businessID: businessId } },
      })
      return data
    },
    enabled: !!businessId,
  })
}

export async function bookOwnAppointment(
  businessId: string,
  input: Schemas["handler.CreateOwnAppointmentRequest"],
) {
  const { data } = await POST(URL_MY_RESERVATIONS, {
    params: { path: { businessID: businessId } },
    body: input,
  })
  return data
}
