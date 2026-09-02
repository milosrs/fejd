package dto

import "fejd-backend/internal/models"

func BusinessFromModel(m models.Business) Business {
	return Business{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
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
		ID:             m.ID,
		BusinessUserID: m.BusinessUserID,
		StartTime:      m.StartTime,
		EndTime:        m.EndTime,
		Reason:         m.Reason,
	}
}

func BusinessUsersFromModels(ms []models.BusinessUser) []BusinessUser {
	out := make([]BusinessUser, len(ms))
	for i, m := range ms {
		out[i] = BusinessUserFromModel(m)
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

func TimeSlotsFromModels(ms []models.TimeSlot) []TimeSlot {
	out := make([]TimeSlot, len(ms))
	for i, m := range ms {
		out[i] = TimeSlotFromModel(m)
	}
	return out
}
