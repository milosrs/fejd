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
go run ./cmd/fejd-admin invite --email owner@example.com --name "Salon Owner" --role owner
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
	role := fs.String("role", "", "platform role to grant: customer or owner")
	fs.Parse(args)

	if *email == "" {
		fmt.Fprintln(os.Stderr, "error: --email is required")
		usage()
		os.Exit(2)
	}

	roleName, approved, err := platformRole(*role)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		usage()
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	client := keycloak.NewClient(cfg.Keycloak)
	ctx := context.Background()

	attributes := map[string][]string{}
	if approved {
		attributes["approval_status"] = []string{"approved"}
	}

	userID, err := client.CreateUser(ctx, keycloak.CreateUserInput{
		Username:        *email,
		Email:           *email,
		FirstName:       *name,
		EmailVerified:   true,
		Enabled:         true,
		RequiredActions: []string{"UPDATE_PASSWORD"},
		Attributes:      attributes,
	})
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	if err := client.AddRealmRole(ctx, userID, roleName); err != nil {
		log.Fatalf("failed to assign %q role: %v", roleName, err)
	}

	lifespan := cfg.Jobs.InviteExpiryHours * 3600
	if err := client.ExecuteActionsEmail(ctx, userID, []string{"UPDATE_PASSWORD"}, "salon-mobile", cfg.Jobs.InviteRedirectURI, lifespan); err != nil {
		log.Fatalf("failed to send invite email: %v", err)
	}

	fmt.Printf("invited %s as %s (user id %s)\n", *email, roleName, userID)
}

// platformRole resolves a CLI role flag to the corresponding Keycloak realm
// role name and whether the invited user should be pre-approved. Only the two
// platform roles (customer, owner) are valid here: employees are salon-scoped
// and are invited through the salon admin API, not by the realm admin.
func platformRole(role string) (roleName string, approved bool, err error) {
	switch role {
	case "customer":
		return "Customer", false, nil
	case "owner":
		return "Owner", true, nil
	default:
		return "", false, fmt.Errorf("--role must be 'customer' or 'owner'")
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fejd-admin invite --email <addr> --role <customer|owner> [--name <n>]")
}
