generate: generate-backend generate-spec generate-openapi3 generate-frontend

generate-backend:
    cd backend && swag init --parseDependency --parseInternal -g main.go -o docs

generate-spec:
    cp backend/docs/swagger.json frontend/openapi.json

generate-openapi3:
    cd frontend && swagger2openapi openapi.json -o openapi3.json

generate-frontend:
    cd frontend && openapi-typescript openapi3.json -o src/lib/api-types.ts

THEME_ROOT := "./keycloak/themes"
THEME_BUILDER_DIR := "./keycloak/theme-builder"

keycloak-theme-converter theme-name:
    @mkdir -p {{THEME_ROOT}} keycloak/providers
    @[ -d "{{THEME_ROOT}}/{{theme-name}}" ] || { echo "Theme directory not found: {{THEME_ROOT}}/{{theme-name}}"; exit 1; }
    @cd {{THEME_BUILDER_DIR}} && [ -f package.json ] && node ./build.mjs || true
    @echo "Keycloak theme source is ready at: {{THEME_ROOT}}/{{theme-name}}"
    @echo "The theme directory is mounted into the Keycloak container through docker compose."
    @echo "Developers can edit the theme assets in the repository and restart the Keycloak service."

# Start all services in detached mode
up:
    docker compose up -d

# Start all docker services except the frontend (run FE separately: just dev-frontend)
up-backend:
    docker compose up -d backend db keycloak

# Stop all services
down:
    docker compose down

# Restart all services
restart: down build up

# Follow logs for all services (or a specific one: just logs backend)
logs service="":
    docker compose logs -f {{service}}

# Show service status
ps:
    docker compose ps

# Start just the database (useful for local dev)
db:
    docker compose up -d db

# Build all images without starting
build:
    docker compose build

# Wipe all data (destroy volumes)
reset:
    docker compose down -v

# Run backend locally (requires db + keycloak running)
dev-backend:
    cd backend && go run main.go

# Run frontend locally with hot reload
dev-frontend:
    cd frontend && pnpm dev

# Run backend, Keycloak and DB in docker-compose, and the frontend locally with hot reload
dev: up-backend dev-frontend

# Start ngrok tunnel to expose Keycloak for identity provider testing
ngrok-keycloak:
    ngrok http 9090 --domain=stimulate-uniquely-abide.ngrok-free.dev
