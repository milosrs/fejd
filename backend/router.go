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
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// newRouter builds the full HTTP router. Auth middlewares are injected as
// functions so routing can be tested without a live Keycloak JWKS.
func newRouter(
	cfg *config.Config,
	authenticate, optionalAuthenticate, requireApproved func(http.Handler) http.Handler,
	businessHandler *handler.BusinessHandler,
	appointmentHandler *handler.AppointmentHandler,
	adminHandler *handler.AdminHandler,
	sseHandler *handler.SSEHandler,
	imageHandler *handler.ImageHandler,
	meHandler *handler.MeHandler,
	i18nHandler *handler.I18nHandler,
	buStore *store.BusinessUserStore,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
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
			r.Get("/employees", businessHandler.GetEmployees)
			r.Get("/sections", businessHandler.GetSections)
			r.Get("/slots", businessHandler.GetAvailableSlots)
		})

		r.Get("/i18n/{locale}", i18nHandler.GetTranslations)

		// GET /api/me is exempt from approval so a pending user can read their
		// own status; POST /api/me/business is gated.
		r.Route("/me", func(r chi.Router) {
			r.Use(authenticate)

			r.Get("/", meHandler.GetMe)

			r.Group(func(r chi.Router) {
				r.Use(requireApproved)
				r.Post("/business", meHandler.CreateBusiness)
			})
		})

		// Everything else protected: Authenticate + RequireApproved.
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.Use(requireApproved)

			r.Post("/appointments", appointmentHandler.Create)
			r.Get("/my/appointments", appointmentHandler.ListMyAppointments)
			r.Delete("/my/appointments/{appointmentID}", appointmentHandler.Cancel)

			r.Route("/admin/business/{businessID}", func(r chi.Router) {
				r.Use(customMiddleware.RequireBusinessAdmin(buStore))

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
				r.Post("/images", imageHandler.UploadBusinessImage)
				r.Post("/services/{serviceID}/image", imageHandler.UploadServiceImage)
			})

			r.Post("/admin/business/{businessID}/employees/{userID}/image", imageHandler.UploadEmployeeImage)
			r.Delete("/admin/business/{businessID}/images/{imageID}", imageHandler.DeleteImage)
		})

		r.With(optionalAuthenticate).Get("/images/{imageID}", imageHandler.GetImage)

		r.Get("/sse/business/{slug}/slots", sseHandler.StreamSlots)
	})

	return r
}
