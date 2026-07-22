generate: generate-backend generate-spec generate-openapi3 generate-frontend

generate-backend:
    cd backend && swag init --parseDependency --parseInternal -g main.go -o docs

generate-spec:
    cp backend/docs/swagger.json frontend/openapi.json

generate-openapi3:
    cd frontend && swagger2openapi openapi.json -o openapi3.json

generate-frontend:
    cd frontend && openapi-typescript openapi3.json -o src/lib/api-types.ts

THEME_ROOT := keycloak/themes
THEME_BUILDER_DIR := keycloak/theme-builder

keycloak-theme-converter theme-name:
    @mkdir -p {{THEME_ROOT}} keycloak/providers
    @[ -d "{{THEME_ROOT}}/{{theme-name}}" ] || { echo "Theme directory not found: {{THEME_ROOT}}/{{theme-name}}"; exit 1; }
    @cd {{THEME_BUILDER_DIR}} && [ -f package.json ] && node ./build.mjs || true
    @echo "Keycloak theme source is ready at: {{THEME_ROOT}}/{{theme-name}}"
    @echo "The theme directory is mounted into the Keycloak container through docker compose."
    @echo "Developers can edit the theme assets in the repository and restart the Keycloak service."
