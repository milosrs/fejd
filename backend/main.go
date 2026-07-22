// @title           Fejd API
// @version         1.0
// @description     Booking and appointment management API.
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fejd-backend/auth"
	_ "fejd-backend/docs"
	"fejd-backend/internal/db"
	customMiddleware "fejd-backend/internal/middleware"
	"fejd-backend/internal/handler"
	"fejd-backend/internal/service"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/store"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	keycloakURL := getEnv("KEYCLOAK_URL", "http://localhost:9090")
	realm := getEnv("KEYCLOAK_REALM", "fejd")
	clientID := getEnv("KEYCLOAK_CLIENT_ID", "fejd-backend")

	config := auth.KeycloakConfig{
		RealmURL: fmt.Sprintf("%s/realms/%s", keycloakURL, realm),
		ClientID: clientID,
	}

	authMiddleware, err := auth.NewMiddleware(config)
	if err != nil {
		log.Fatalf("Failed to initialize authentication middleware: %v", err)
	}

	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	appointmentStore := store.NewAppointmentStore(pool)
	workingHoursStore := store.NewWorkingHoursStore(pool)
	overrideStore := store.NewWorkingHoursOverrideStore(pool)

	hub := sse.NewHub()

	slotService := service.NewSlotService(
		appointmentStore, workingHoursStore, overrideStore,
		serviceStore, buStore, hub,
	)

	workingHoursService := service.NewWorkingHoursService(
		workingHoursStore, overrideStore, buStore,
		slotService, businessStore,
	)

	businessHandler := handler.NewBusinessHandler(
		businessStore, buStore, serviceStore, slotService,
	)

	appointmentHandler := handler.NewAppointmentHandler(
		appointmentStore, serviceStore, businessStore, buStore, slotService,
	)

	adminHandler := handler.NewAdminHandler(
		businessStore, buStore, serviceStore, workingHoursService, appointmentStore,
	)

	sseHandler := handler.NewSSEHandler(hub)

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

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		r.Route("/business/{slug}", func(r chi.Router) {
			r.Get("/", businessHandler.GetBusiness)
			r.Get("/services", businessHandler.GetServices)
			r.Get("/employees", businessHandler.GetEmployees)
			r.Get("/slots", businessHandler.GetAvailableSlots)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/appointments", appointmentHandler.Create)
			r.Get("/my/appointments", appointmentHandler.ListMyAppointments)
			r.Delete("/my/appointments/{appointmentID}", appointmentHandler.Cancel)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Route("/admin/business/{businessID}", func(r chi.Router) {
				r.Use(customMiddleware.RequireBusinessAdmin(buStore))

				r.Get("/employees", adminHandler.GetEmployees)
				r.Get("/employees/{userID}/working-hours", adminHandler.GetWorkingHours)
				r.Put("/employees/{userID}/working-hours", adminHandler.SetWorkingHours)
				r.Post("/employees/{userID}/overrides", adminHandler.AddOverride)
				r.Delete("/employees/{userID}/overrides/{overrideID}", adminHandler.DeleteOverride)
				r.Post("/services", adminHandler.CreateService)
				r.Put("/services/{serviceID}", adminHandler.UpdateService)
				r.Delete("/services/{serviceID}", adminHandler.DeleteService)
			})
		})

		r.Get("/sse/business/{slug}/slots", sseHandler.StreamSlots)
	})

	port := getEnv("PORT", "8080")
	fmt.Printf("backend listening on :%s\n", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
