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
	serviceCombinationStore := store.NewServiceCombinationStore(pool)
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
		businessClosureStore, serviceStore, businessStore, buStore, employeeServiceStore, unavailabilityStore, serviceCombinationStore, hub, pool,
	)

	workingHoursService := service.NewWorkingHoursService(
		workingHoursStore, overrideStore, buStore,
		slotService, businessStore,
	)

	businessHandler := handler.NewBusinessHandler(
		businessStore, buStore, userStore, serviceStore, pageStore, sectionStore, imageLinkStore, employeeServiceStore, businessClosureStore, businessHoursStore, slotService, cfg.AppDomain,
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

	// Build the push sender from the Firebase service account. Enabled by
	// default; disable with FIREBASE_ENABLED=false. A missing sender makes all
	// notifications no-ops, so the app still runs without credentials.
	var pushSender service.PushSender
	if cfg.Firebase.Enabled {
		if !cfg.Firebase.Configured() {
			log.Printf("Firebase push enabled but no service account configured; push notifications are disabled")
		} else {
			sa, err := cfg.Firebase.ServiceAccount()
			if err != nil {
				log.Fatalf("Failed to load Firebase service account: %v", err)
			}
			sender, err := push.NewSender(sa)
			if err != nil {
				log.Fatalf("Failed to initialize Firebase push sender: %v", err)
			}
			pushSender = sender
			log.Printf("Firebase push notifications enabled")
		}
	}

	notificationService := service.NewNotificationService(pushTokenStore, buStore, businessStore, userStore, imageLinkStore, pushSender, cfg.PublicURL, cfg.Jobs.InviteBaseURL)
	slotService.SetNotifier(notificationService)

	jobCtx, jobCancel := context.WithCancel(context.Background())
	defer jobCancel()

	if cfg.Jobs.RunJobs {
		emailSender := email.NewSender(cfg.Email)
		pendingNotifier := jobs.NewPendingNotifier(keycloakAdmin, emailSender, cfg.Email.SuperadminNotifyEmail)
		inviteCleanup := jobs.NewInviteCleanup(keycloakAdmin, cfg.Jobs.InviteExpiryHours)
		reminderNotifier := jobs.NewReminderNotifier(businessStore, appointmentStore, notificationService)
		scheduler := jobs.NewScheduler(pendingNotifier, inviteCleanup, reminderNotifier, cfg.Jobs.PendingScanInterval, cfg.Jobs.CleanupInterval, cfg.Jobs.ReminderScanInterval)
		go scheduler.Run(jobCtx)
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

	// One-time, best-effort geocoding backfill for salons that already had a
	// free-text address before structured location existed. Runs in the
	// background and is idempotent.
	backfillLocationGeo(context.Background(), businessStore)

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

// backfillLocationGeo geocodes salons that already have an address but are
// missing a city or geo coordinates. It runs in the background, is rate-limited
// to respect the Nominatim usage policy, and is idempotent (salons that already
// have city+geo are skipped). Best-effort: failures are logged, not fatal.
func backfillLocationGeo(ctx context.Context, businessStore *store.BusinessStore) {
	go func() {
		businesses, err := businessStore.ListNeedingGeocode(ctx)
		if err != nil {
			log.Printf("location geocode backfill: list failed: %v", err)
			return
		}
		if len(businesses) == 0 {
			return
		}
		log.Printf("location geocode backfill: geocoding %d salon(s)", len(businesses))

		for _, b := range businesses {
			query := b.AddressLine
			if b.City != "" {
				query = b.AddressLine + ", " + b.City
			}
			if query == "" {
				continue
			}

			g, ok := service.GeocodeNominatim(ctx, query)
			if !ok {
				continue
			}

			city := b.City
			if city == "" {
				city = g.City
			}
			postal := b.PostalCode
			if postal == "" {
				postal = g.Postcode
			}
			country := b.Country
			if country == "" {
				country = g.Country
			}
			lat := b.Latitude
			lon := b.Longitude
			if lat == nil {
				lat = &g.Latitude
			}
			if lon == nil {
				lon = &g.Longitude
			}

			if err := businessStore.UpdateLocation(ctx, b.ID, b.AddressLine, city, postal, country, b.Phone, lat, lon); err != nil {
				log.Printf("location geocode backfill: update %s failed: %v", b.Slug, err)
			}

			// Nominatim policy: at most ~1 request per second.
			time.Sleep(1100 * time.Millisecond)
		}
	}()
}
