package dto

import (
	"encoding/json"

	"fejd-backend/internal/models"

	"github.com/google/uuid"
)

func BusinessFromModel(m models.Business) Business {
	return Business{
		ID:                    m.ID,
		Name:                  m.Name,
		Slug:                  m.Slug,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
		CancellationLeadHours: m.CancellationLeadHours,
		NoShowAfterHours:      m.NoShowAfterHours,
		SlotIntervalMinutes:   m.SlotIntervalMinutes,
	}
}

func MeBusinessFromModel(m models.BusinessMembership) MeBusiness {
	return MeBusiness{
		ID:   m.BusinessID,
		Name: m.Name,
		Slug: m.Slug,
		Role: m.Role,
	}
}

func MeBusinessesFromModels(ms []models.BusinessMembership) []MeBusiness {
	out := make([]MeBusiness, len(ms))
	for i, m := range ms {
		out[i] = MeBusinessFromModel(m)
	}
	return out
}

func DirectoryBusinessFromModel(m models.Business) DirectoryBusiness {
	return DirectoryBusiness{
		ID:   m.ID,
		Name: m.Name,
		Slug: m.Slug,
	}
}

func DirectoryBusinessesFromModels(ms []models.Business) []DirectoryBusiness {
	out := make([]DirectoryBusiness, len(ms))
	for i, m := range ms {
		out[i] = DirectoryBusinessFromModel(m)
	}
	return out
}

func BusinessUserFromModel(m models.BusinessUser) BusinessUser {
	return BusinessUser{
		ID:          m.ID,
		BusinessID:  m.BusinessID,
		UserID:      m.UserID,
		Role:        m.Role,
		DisplayName: m.DisplayName,
		Active:      m.Active,
	}
}

func BusinessHoursFromModel(m models.BusinessHours) BusinessHours {
	return BusinessHours{
		DayOfWeek: m.DayOfWeek,
		StartTime: m.StartTime.Format("15:04"),
		EndTime:   m.EndTime.Format("15:04"),
	}
}

func BusinessHoursFromModels(ms []models.BusinessHours) []BusinessHours {
	out := make([]BusinessHours, len(ms))
	for i, m := range ms {
		out[i] = BusinessHoursFromModel(m)
	}
	return out
}

func SectionFromModel(m models.Section) Section {
	content := json.RawMessage(m.Content)
	if len(content) == 0 {
		content = json.RawMessage(`{}`)
	}
	return Section{
		ID:       m.ID,
		PageID:   m.PageID,
		Type:     m.Type,
		Content:  content,
		Position: m.Position,
	}
}

func SectionsFromModels(ms []models.Section) []Section {
	out := make([]Section, len(ms))
	for i, m := range ms {
		out[i] = SectionFromModel(m)
	}
	return out
}

func ServiceFromModel(m models.Service) Service {
	return Service{
		ID:              m.ID,
		BusinessID:      m.BusinessID,
		Name:            m.Name,
		DurationMinutes: m.DurationMinutes,
		Price:           m.Price,
		Active:          m.Active,
		Description:     m.Description,
		PictureID:       m.PictureID,
		CreatedAt:       m.CreatedAt,
	}
}

func WorkingHoursFromModel(m models.WorkingHours) WorkingHours {
	return WorkingHours{
		ID:             m.ID,
		BusinessUserID: m.BusinessUserID,
		DayOfWeek:      m.DayOfWeek,
		StartTime:      m.StartTime,
		EndTime:        m.EndTime,
	}
}

func WorkingHoursOverrideFromModel(m models.WorkingHoursOverride) WorkingHoursOverride {
	return WorkingHoursOverride{
		ID:             m.ID,
		BusinessUserID: m.BusinessUserID,
		OverrideDate:   m.OverrideDate,
		StartTime:      m.StartTime,
		EndTime:        m.EndTime,
		IsOff:          m.IsOff,
		Reason:         m.Reason,
	}
}

func AppointmentFromModel(m models.Appointment) Appointment {
	return Appointment{
		ID:                 m.ID,
		BusinessID:         m.BusinessID,
		ServiceID:          m.ServiceID,
		BusinessUserID:     m.BusinessUserID,
		CustomerUserID:     m.CustomerUserID,
		StartTime:          m.StartTime,
		EndTime:            m.EndTime,
		Status:             string(m.Status),
		CreatedBy:          m.CreatedBy,
		CancellationReason: m.CancellationReason,
		CreatedAt:          m.CreatedAt,
	}
}

func TimeSlotFromModel(m models.TimeSlot) TimeSlot {
	return TimeSlot{
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
	}
}

func EmployeeUnavailabilityFromModel(m models.EmployeeUnavailability) EmployeeUnavailability {
	return EmployeeUnavailability{
		ID:              m.ID,
		BusinessUserID:  m.BusinessUserID,
		StartTime:       m.StartTime,
		EndTime:         m.EndTime,
		Reason:          m.Reason,
		Status:          string(m.Status),
		RejectionReason: m.RejectionReason,
	}
}

func ImageFromModel(m models.Image) Image {
	return Image{
		ID:          m.ID,
		URL:         "/api/images/" + m.ID.String(),
		ContentType: m.ContentType,
		CreatedAt:   m.CreatedAt,
	}
}

func BusinessUsersFromModels(ms []models.BusinessUser) []BusinessUser {
	out := make([]BusinessUser, len(ms))
	for i, m := range ms {
		out[i] = BusinessUserFromModel(m)
	}
	return out
}

func CustomerFromModel(m models.Customer) Customer {
	return Customer{
		UserID:      m.UserID,
		DisplayName: m.DisplayName,
	}
}

func CustomersFromModels(ms []models.Customer) []Customer {
	out := make([]Customer, len(ms))
	for i, m := range ms {
		out[i] = CustomerFromModel(m)
	}
	return out
}

func ServicesFromModels(ms []models.Service) []Service {
	out := make([]Service, len(ms))
	for i, m := range ms {
		out[i] = ServiceFromModel(m)
	}
	return out
}

func WorkingHoursFromModels(ms []models.WorkingHours) []WorkingHours {
	out := make([]WorkingHours, len(ms))
	for i, m := range ms {
		out[i] = WorkingHoursFromModel(m)
	}
	return out
}

func WorkingHoursOverridesFromModels(ms []models.WorkingHoursOverride) []WorkingHoursOverride {
	out := make([]WorkingHoursOverride, len(ms))
	for i, m := range ms {
		out[i] = WorkingHoursOverrideFromModel(m)
	}
	return out
}

func AppointmentsFromModels(ms []models.Appointment) []Appointment {
	out := make([]Appointment, len(ms))
	for i, m := range ms {
		out[i] = AppointmentFromModel(m)
	}
	return out
}

// StaffAppointmentsFromModels enriches appointments with their service name and
// the salon's no-show grace period for staff-facing reservation lists.
func StaffAppointmentsFromModels(ms []models.Appointment, serviceNames map[uuid.UUID]string, noShowAfterHours int) []Appointment {
	out := make([]Appointment, len(ms))
	for i, m := range ms {
		out[i] = AppointmentFromModel(m)
		out[i].ServiceName = serviceNames[m.ServiceID]
		out[i].NoShowAfterHours = noShowAfterHours
	}
	return out
}

// CustomerAppointmentsFromModels enriches customer appointments with each
// salon's cancellation notice window so the UI can surface the deadline.
func CustomerAppointmentsFromModels(ms []models.Appointment, leadHoursByBusiness map[uuid.UUID]int) []Appointment {
	out := make([]Appointment, len(ms))
	for i, m := range ms {
		out[i] = AppointmentFromModel(m)
		out[i].CancellationLeadHours = leadHoursByBusiness[m.BusinessID]
	}
	return out
}

func EmployeeUnavailabilitysFromModels(ms []models.EmployeeUnavailability) []EmployeeUnavailability {
	out := make([]EmployeeUnavailability, len(ms))
	for i, m := range ms {
		out[i] = EmployeeUnavailabilityFromModel(m)
	}
	return out
}

func TimeSlotsFromModels(ms []models.TimeSlot) []TimeSlot {
	out := make([]TimeSlot, len(ms))
	for i, m := range ms {
		out[i] = TimeSlotFromModel(m)
	}
	return out
}
