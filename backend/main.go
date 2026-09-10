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
	"fejd-backend/internal/config"
	"fejd-backend/internal/db"
	"fejd-backend/internal/email"
	"fejd-backend/internal/handler"
	"fejd-backend/internal/jobs"
	"fejd-backend/internal/keycloak"
	"fejd-backend/internal/service"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/storage"
	"fejd-backend/internal/store"
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

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	keycloakConfig := auth.KeycloakConfig{
		RealmURL:  fmt.Sprintf("%s/realms/%s", cfg.Keycloak.AdminURL, cfg.Keycloak.Realm),
		Audiences: cfg.Keycloak.Audiences,
	}

	authMiddleware, err := auth.NewMiddleware(keycloakConfig)
	if err != nil {
		log.Fatalf("Failed to initialize authentication middleware: %v", err)
	}

	businessStore := store.NewBusinessStore(pool)
	buStore := store.NewBusinessUserStore(pool)
	userStore := store.NewUserStore(pool)
	serviceStore := store.NewServiceStore(pool)
	appointmentStore := store.NewAppointmentStore(pool)
	workingHoursStore := store.NewWorkingHoursStore(pool)
	overrideStore := store.NewWorkingHoursOverrideStore(pool)
	businessHoursStore := store.NewBusinessHoursStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)
	unavailabilityStore := store.NewEmployeeUnavailabilityStore(pool)
	imageStore := store.NewImageStore(pool)
	imageLinkStore := store.NewImageLinkStore(pool)
	sectionStore := store.NewSectionStore(pool)
	pageStore := store.NewPageStore(pool)
	translationStore := store.NewTranslationStore(pool)

	var imageStorage storage.ImageStorage
	if cfg.ImageStorage.Backend != config.BackendPostgres {
		imageStorage, err = storage.NewFromConfig(cfg.ImageStorage)
		if err != nil {
			log.Fatalf("Failed to initialize image storage: %v", err)
		}
	}

	imageService := service.NewImageService(cfg.ImageStorage, imageStorage, imageStore, imageLinkStore, buStore, userStore, pool)

	hub := sse.NewHub()

	slotService := service.NewSlotService(
		appointmentStore, workingHoursStore, businessHoursStore, overrideStore,
		serviceStore, businessStore, buStore, employeeServiceStore, unavailabilityStore, hub, pool,
	)

	workingHoursService := service.NewWorkingHoursService(
		workingHoursStore, overrideStore, buStore,
		slotService, businessStore,
	)

	businessHandler := handler.NewBusinessHandler(
		businessStore, buStore, userStore, serviceStore, pageStore, sectionStore, imageLinkStore, employeeServiceStore, slotService,
	)

	appointmentHandler := handler.NewAppointmentHandler(
		appointmentStore, serviceStore, businessStore, buStore, slotService,
	)

	keycloakAdmin := keycloak.NewClient(cfg.Keycloak)

	employeeService := service.NewEmployeeService(
		keycloakAdmin, buStore, employeeServiceStore, pool, cfg.Jobs.InviteRedirectURI, cfg.Jobs.InviteExpiryHours*3600,
	)

	invitationStore := store.NewInvitationStore(pool)
	invitationService := service.NewInvitationService(invitationStore, businessStore, buStore, userStore, keycloakAdmin, pool, cfg.Jobs.InviteBaseURL)
	invitationHandler := handler.NewInvitationHandler(invitationService, time.Duration(cfg.Jobs.InviteExpiryHours)*time.Hour)

	adminHandler := handler.NewAdminHandler(
		businessStore, buStore, serviceStore, pageStore, sectionStore, businessHoursStore, workingHoursService, appointmentStore, slotService, imageService, employeeService, pool,
	)

	sseHandler := handler.NewSSEHandler(hub, businessStore)

	imageHandler := handler.NewImageHandler(imageService, serviceStore, buStore)

	meHandler := handler.NewMeHandler(businessStore, buStore, userStore, businessHoursStore, pool)

	i18nHandler := handler.NewI18nHandler(translationStore)

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	if cfg.Jobs.RunJobs {
		emailSender := email.NewSender(cfg.Email)
		pendingNotifier := jobs.NewPendingNotifier(keycloakAdmin, emailSender, cfg.Email.SuperadminNotifyEmail)
		inviteCleanup := jobs.NewInviteCleanup(keycloakAdmin, cfg.Jobs.InviteExpiryHours)
		scheduler := jobs.NewScheduler(pendingNotifier, inviteCleanup, cfg.Jobs.PendingScanInterval, cfg.Jobs.CleanupInterval)
		go scheduler.Run(jobCtx)
	}

	r := newRouter(
		cfg,
		authMiddleware.Authenticate,
		authMiddleware.OptionalAuthenticate,
		authMiddleware.RequireApproved,
		authMiddleware.RequireRole(auth.RoleOwner),
		businessHandler,
		appointmentHandler,
		adminHandler,
		sseHandler,
		imageHandler,
		meHandler,
		i18nHandler,
		invitationHandler,
		buStore,
		userStore,
	)

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
		jobCancel()

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
