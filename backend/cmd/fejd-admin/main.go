package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"fejd-backend/internal/config"
	"fejd-backend/internal/keycloak"
)

/*
How to run:

KEYCLOAK_URL=http://localhost:9090 KEYCLOAK_REALM=fejd \
KEYCLOAK_ADMIN_CLIENT_ID=fejd-admin \
KEYCLOAK_ADMIN_CLIENT_SECRET=H6vKp9sQ2wXrT4yL8mN1cB3dV5fG7jZ0 \
INVITE_REDIRECT_URI=fejd://callback \
go run ./cmd/fejd-admin invite --email owner@example.com --name "Salon Owner"
*/

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "invite":
		invite(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func invite(args []string) {
	fs := flag.NewFlagSet("invite", flag.ExitOnError)
	email := fs.String("email", "", "email address to invite")
	name := fs.String("name", "", "display name (optional)")
	fs.Parse(args)

	if *email == "" {
		fmt.Fprintln(os.Stderr, "error: --email is required")
		usage()
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	client := keycloak.NewClient(cfg.Keycloak)
	ctx := context.Background()

	userID, err := client.CreateUser(ctx, keycloak.CreateUserInput{
		Username:        *email,
		Email:           *email,
		FirstName:       *name,
		EmailVerified:   true,
		Enabled:         true,
		RequiredActions: []string{"UPDATE_PASSWORD"},
		Attributes:      map[string][]string{"approval_status": {"approved"}},
	})
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	lifespan := cfg.Jobs.InviteExpiryHours * 3600
	if err := client.ExecuteActionsEmail(ctx, userID, []string{"UPDATE_PASSWORD"}, "salon-mobile", cfg.Jobs.InviteRedirectURI, lifespan); err != nil {
		log.Fatalf("failed to send invite email: %v", err)
	}

	fmt.Printf("invited %s (user id %s)\n", *email, userID)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fejd-admin invite --email <addr> [--name <n>]")
}
