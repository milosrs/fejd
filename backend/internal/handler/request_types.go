package handler

import "fejd-backend/internal/models"

type CreateAppointmentRequest struct {
	BusinessID     string `json:"business_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceID      string `json:"service_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BusinessUserID string `json:"business_user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartTime      string `json:"start_time" validate:"required" example:"2024-01-01T09:00:00Z"`
}

type SetWorkingHoursRequest struct {
	WorkingHours []models.WorkingHours `json:"working_hours" validate:"required"`
}

type ErrorResponse struct {
	Error string `json:"error" validate:"required" example:"error message"`
}

type MessageResponse struct {
	Message string `json:"message" validate:"required" example:"operation complete"`
}

type BusinessResponse struct {
	Business  models.Business      `json:"business" validate:"required"`
	Services  []models.Service     `json:"services" validate:"required"`
	Employees []models.BusinessUser `json:"employees" validate:"required"`
}

type SlotsResponse struct {
	Slots []models.TimeSlot `json:"slots" validate:"required"`
	Date  string            `json:"date" validate:"required" example:"2024-01-01"`
}

type WorkingHoursResponse struct {
	WorkingHours []models.WorkingHours         `json:"working_hours" validate:"required"`
	Overrides    []models.WorkingHoursOverride `json:"overrides" validate:"required"`
}
