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
	_ "time/tzdata" // embed the IANA tz database so LoadLocation works without host tzdata

	"fejd-backend/auth"
	_ "fejd-backend/docs"
	"fejd-backend/internal/config"
	"fejd-backend/internal/db"
	"fejd-backend/internal/email"
	"fejd-backend/internal/handler"
	"fejd-backend/internal/jobs"
	"fejd-backend/internal/keycloak"
	"fejd-backend/internal/push"
	"fejd-backend/internal/service"
	"fejd-backend/internal/sse"
	"fejd-backend/internal/storage"
	"fejd-backend/internal/store"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	logStartupConfig(cfg)

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	keycloakConfig := auth.KeycloakConfig{
		RealmURL:  fmt.Sprintf("%s/realms/%s", cfg.Keycloak.AdminURL, cfg.Keycloak.Realm),
		IssuerURL: cfg.Keycloak.IssuerURL,
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
	businessClosureStore := store.NewBusinessClosureStore(pool)
	employeeServiceStore := store.NewEmployeeServiceStore(pool)
	unavailabilityStore := store.NewEmployeeUnavailabilityStore(pool)
	imageStore := store.NewImageStore(pool)
	imageLinkStore := store.NewImageLinkStore(pool)
	sectionStore := store.NewSectionStore(pool)
	pageStore := store.NewPageStore(pool)
	translationStore := store.NewTranslationStore(pool)
	pushTokenStore := store.NewPushTokenStore(pool)

	var imageStorage storage.ImageStorage
	if cfg.ImageStorage.Backend != config.BackendPostgres {
		imageStorage, err = storage.NewFromConfig(ctx, cfg.ImageStorage)
		if err != nil {
			log.Fatalf("Failed to initialize image storage: %v", err)
		}
	}

	imageService := service.NewImageService(cfg.ImageStorage, imageStorage, imageStore, imageLinkStore, buStore, userStore, pool)

	hub := sse.NewHub()

	slotService := service.NewSlotService(
		appointmentStore, workingHoursStore, businessHoursStore, overrideStore,
		businessClosureStore, serviceStore, businessStore, buStore, employeeServiceStore, unavailabilityStore, hub, pool,
	)

	workingHoursService := service.NewWorkingHoursService(
		workingHoursStore, overrideStore, buStore,
		slotService, businessStore,
	)

	businessHandler := handler.NewBusinessHandler(
		businessStore, buStore, userStore, serviceStore, pageStore, sectionStore, imageLinkStore, employeeServiceStore, businessClosureStore, businessHoursStore, slotService,
	)

	appointmentHandler := handler.NewAppointmentHandler(
		appointmentStore, serviceStore, businessStore, buStore, slotService, imageLinkStore,
	)

	keycloakAdmin := keycloak.NewClient(cfg.Keycloak)

	employeeService := service.NewEmployeeService(
		keycloakAdmin, buStore, employeeServiceStore, pool, cfg.Jobs.InviteRedirectURI, cfg.Jobs.InviteExpiryHours*3600,
	)

	invitationStore := store.NewInvitationStore(pool)
	invitationService := service.NewInvitationService(invitationStore, businessStore, buStore, userStore, keycloakAdmin, pool, cfg.Jobs.InviteBaseURL)
	invitationHandler := handler.NewInvitationHandler(invitationService, time.Duration(cfg.Jobs.InviteExpiryHours)*time.Hour)

	adminHandler := handler.NewAdminHandler(
		businessStore, buStore, serviceStore, pageStore, sectionStore, businessHoursStore, businessClosureStore, workingHoursService, appointmentStore, slotService, imageService, employeeService, pool,
	)

	sseHandler := handler.NewSSEHandler(hub, businessStore)

	imageHandler := handler.NewImageHandler(imageService, serviceStore, buStore)

	salonDomainService := service.NewSalonDomainService(keycloakAdmin, cfg.AppDomain)

	meHandler := handler.NewMeHandler(businessStore, buStore, userStore, businessHoursStore, service.NewRegistrationService(keycloakAdmin), salonDomainService, pool)

	i18nHandler := handler.NewI18nHandler(translationStore)

	pushHandler := handler.NewPushHandler(pushTokenStore)

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	if cfg.Jobs.RunJobs {
		emailSender := email.NewSender(cfg.Email)
		pendingNotifier := jobs.NewPendingNotifier(keycloakAdmin, emailSender, cfg.Email.SuperadminNotifyEmail)
		inviteCleanup := jobs.NewInviteCleanup(keycloakAdmin, cfg.Jobs.InviteExpiryHours)
		scheduler := jobs.NewScheduler(pendingNotifier, inviteCleanup, cfg.Jobs.PendingScanInterval, cfg.Jobs.CleanupInterval)
		go scheduler.Run(jobCtx)
	}

	// Validate Firebase push credentials at startup so a misconfigured service
	// account fails fast. The sender is reused by the notification triggers.
	if cfg.Firebase.Enabled() {
		sa, err := cfg.Firebase.ServiceAccount()
		if err != nil {
			log.Fatalf("Failed to load Firebase service account: %v", err)
		}
		if _, err := push.NewSender(sa); err != nil {
			log.Fatalf("Failed to initialize Firebase push sender: %v", err)
		}
		log.Printf("Firebase push notifications enabled")
	}

	r := newRouter(
		cfg,
		authMiddleware.Authenticate,
		authMiddleware.AuthenticateUnverified,
		authMiddleware.OptionalAuthenticate,
		authMiddleware.RequireApproved,
		authMiddleware.RequireRole(auth.RoleOwner),
		authMiddleware.RequireRealmAdmin,
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
		pushHandler,
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

func logStartupConfig(cfg *config.Config) {
	log.Printf("Starting backend with configuration: database=configured keycloak_url=%s keycloak_realm=%s keycloak_audiences=%v image_storage_backend=%s image_storage_bucket=%s image_storage_endpoint=%s port=%s run_jobs=%t app_domain=%s",
		getEnv("KEYCLOAK_URL", "http://localhost:9090"), cfg.Keycloak.Realm, cfg.Keycloak.Audiences,
		cfg.ImageStorage.Backend, storageBucket(cfg), storageEndpoint(cfg), getEnv("PORT", "8080"), cfg.Jobs.RunJobs, cfg.AppDomain)
}

func storageBucket(cfg *config.Config) string {
	if cfg.ImageStorage.Backend == config.BackendS3 {
		return cfg.ImageStorage.S3.Bucket
	}
	return cfg.ImageStorage.Seaweedfs.Bucket
}

func storageEndpoint(cfg *config.Config) string {
	if cfg.ImageStorage.Backend == config.BackendS3 {
		return cfg.ImageStorage.S3.Endpoint
	}
	return cfg.ImageStorage.Seaweedfs.Endpoint
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
