package main

import (
	"fmt"
	"net/http"

	"fejd-backend/internal/config"
	"fejd-backend/internal/handler"
	customMiddleware "fejd-backend/internal/middleware"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// newRouter builds the full HTTP router. Auth middlewares are injected as
// functions so routing can be tested without a live Keycloak JWKS.
func newRouter(
	cfg *config.Config,
	authenticate, optionalAuthenticate, requireApproved, requireOwner func(http.Handler) http.Handler,
	businessHandler *handler.BusinessHandler,
	appointmentHandler *handler.AppointmentHandler,
	adminHandler *handler.AdminHandler,
	sseHandler *handler.SSEHandler,
	imageHandler *handler.ImageHandler,
	meHandler *handler.MeHandler,
	i18nHandler *handler.I18nHandler,
	invitationHandler *handler.InvitationHandler,
	buStore *store.BusinessUserStore,
	userStore *store.UserStore,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(customMiddleware.CORS(cfg.CORS.AllowedOrigins, cfg.CORS.AllowedSuffix))
	r.Use(customMiddleware.MaxBodyBytes(cfg.ImageStorage.MaxUploadBytes))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		// Public.
		r.Route("/business/{slug}", func(r chi.Router) {
			r.Get("/", businessHandler.GetBusiness)
			r.Get("/services", businessHandler.GetServices)
			r.Get("/services/{serviceID}/employees", businessHandler.GetServiceEmployees)
			r.Get("/employees", businessHandler.GetEmployees)
			r.Get("/sections", businessHandler.GetSections)
			r.Get("/slots", businessHandler.GetAvailableSlots)
		})

		r.Get("/i18n/{locale}", i18nHandler.GetTranslations)

		// Public resolve of an invite token (no auth): the landing page greets
		// the invitee before they register.
		r.Get("/invitations/{token}", invitationHandler.GetInvitation)

		// GET /api/me is exempt from approval so a pending user can read their
		// own status; POST /api/me/business is gated.
		r.Route("/me", func(r chi.Router) {
			r.Use(authenticate)
			r.Use(customMiddleware.SyncUser(userStore))

			r.Get("/", meHandler.GetMe)

			r.Group(func(r chi.Router) {
				r.Use(requireApproved)
				r.Post("/business", meHandler.CreateBusiness)
			})
		})

		// Customer-facing booking routes: authenticated but NOT approval-gated,
		// so a self-registered customer can book right after registering (the
		// "QR -> register -> book" flow). approval_status gates only the
		// owner-facing onboarding and management below.
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.Use(customMiddleware.SyncUser(userStore))

			r.Post("/appointments", appointmentHandler.Create)
			r.Get("/my/appointments", appointmentHandler.ListMyAppointments)
			r.Delete("/my/appointments/{appointmentID}", appointmentHandler.Cancel)

			// Accept is authenticated but NOT approval-gated, so a brand-new or
			// still-pending user can redeem an invite right after registering.
			r.Post("/invitations/{token}/accept", invitationHandler.AcceptInvitation)
		})

		// Owner-facing: Authenticate + RequireApproved.
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.Use(requireApproved)
			r.Use(customMiddleware.SyncUser(userStore))

			r.Route("/admin/business/{businessID}", func(r chi.Router) {
				r.Use(requireOwner)
				r.Use(customMiddleware.RequireBusinessAdmin(buStore))

				r.Post("/employees", adminHandler.CreateEmployee)
				r.Get("/employees", adminHandler.GetEmployees)
				r.Delete("/employees/{userID}", adminHandler.RemoveEmployee)
				r.Get("/employees/{userID}/working-hours", adminHandler.GetWorkingHours)
				r.Put("/employees/{userID}/working-hours", adminHandler.SetWorkingHours)
				r.Post("/employees/{userID}/overrides", adminHandler.AddOverride)
				r.Delete("/employees/{userID}/overrides/{overrideID}", adminHandler.DeleteOverride)
				r.Put("/employees/{userID}/services", adminHandler.SetEmployeeServices)
				r.Post("/employees/{userID}/unavailability", adminHandler.AddUnavailability)
				r.Delete("/employees/{userID}/unavailability/{unavailabilityID}", adminHandler.DeleteUnavailability)
				r.Post("/services", adminHandler.CreateService)
				r.Put("/services/{serviceID}", adminHandler.UpdateService)
				r.Delete("/services/{serviceID}", adminHandler.DeleteService)
				r.Put("/sections/reorder", adminHandler.ReorderSections)
				r.Post("/sections", adminHandler.CreateSection)
				r.Put("/sections/{sectionID}", adminHandler.UpdateSection)
				r.Delete("/sections/{sectionID}", adminHandler.DeleteSection)
				r.Post("/images", imageHandler.UploadBusinessImage)
				r.Post("/services/{serviceID}/image", imageHandler.UploadServiceImage)
				r.Get("/policy", adminHandler.GetSalonPolicy)
				r.Put("/policy", adminHandler.UpdateSalonPolicy)
			})

			r.Post("/admin/business/{businessID}/employees/{userID}/image", imageHandler.UploadEmployeeImage)
			r.Delete("/admin/business/{businessID}/images/{imageID}", imageHandler.DeleteImage)

			r.With(customMiddleware.RequireBusinessMember(buStore)).
				Post("/admin/business/{businessID}/invitations", invitationHandler.CreateInvitation)

			// Self-service blocked time: any active member (owner or employee)
			// may reserve their own slots so they are hidden from customers.
			r.With(customMiddleware.RequireBusinessMember(buStore)).
				Route("/admin/business/{businessID}/me/unavailability", func(r chi.Router) {
					r.Get("/", adminHandler.ListMyUnavailability)
					r.Post("/", adminHandler.AddMyUnavailability)
					r.Delete("/{unavailabilityID}", adminHandler.DeleteMyUnavailability)
				})

			// Self-service reservations: any active member can list their own
			// appointments for a date and cancel them with a reason.
			r.With(customMiddleware.RequireBusinessMember(buStore)).
				Route("/admin/business/{businessID}/me/appointments", func(r chi.Router) {
					r.Get("/", adminHandler.ListMyReservations)
					r.Post("/", adminHandler.BookOwnAppointment)
					r.Delete("/{appointmentID}", adminHandler.CancelMyReservation)
					r.Post("/{appointmentID}/no-show", adminHandler.MarkNoShow)
				})

			r.With(customMiddleware.RequireBusinessMember(buStore)).
				Get("/admin/business/{businessID}/me/services", adminHandler.ListMyServices)

			r.With(customMiddleware.RequireBusinessMember(buStore)).
				Get("/admin/business/{businessID}/customers", adminHandler.ListCustomers)
		})

		r.With(optionalAuthenticate).Get("/images/{imageID}", imageHandler.GetImage)

		r.Get("/sse/business/{slug}/slots", sseHandler.StreamSlots)
	})

	return r
}
