#!/usr/bin/env bash
set -euo pipefail

# Apply the fejd realm configuration changes to an EXISTING Keycloak without
# dropping the schema or touching users. Uses the Admin REST API (idempotent).
#
# What it updates (safe, attribute-level, no users deleted):
#   - internationalizationEnabled, supportedLocales, defaultLocale
#   - loginTheme = fejd
#   - rememberMe = true
#   - registration_role protocol mappers on fejd-frontend + salon-mobile
#   - user profile unmanagedAttributePolicy = ENABLED
#
# Usage (override as needed):
#   KEYCLOAK_URL=https://auth.fejd.fyi \
#   ADMIN_USER="$KEYCLOAK_ADMIN_USERNAME" ADMIN_PASSWORD="$KEYCLOAK_ADMIN_PASSWORD" \
#   ./apply-realm-config.sh

KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:9090}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin}"
REALM="${REALM:-fejd}"

for bin in curl jq; do
  command -v "$bin" >/dev/null || { echo "missing dependency: $bin" >&2; exit 1; }
done

BASE="$KEYCLOAK_URL"

echo "==> Authenticating to Keycloak Admin API ($KEYCLOAK_URL)"
TOKEN="$(curl -sf -X POST "$BASE/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=admin-cli" \
  -d "username=$ADMIN_USER" \
  -d "password=$ADMIN_PASSWORD" \
  -d "grant_type=password" | jq -r '.access_token')"

AUTH=(-H "Authorization: Bearer $TOKEN")

echo "==> Updating realm config (i18n / theme / remember me)"
curl -sf "$BASE/admin/realms/$REALM" "${AUTH[@]}" > /tmp/fejd-realm.json
jq '
  .internationalizationEnabled = true
  | .supportedLocales = ["en", "sr"]
  | .defaultLocale = "en"
  | .loginTheme = "fejd"
  | .emailTheme = "fejd"
  | .rememberMe = true
' /tmp/fejd-realm.json > /tmp/fejd-realm-updated.json
curl -sf -X PUT "$BASE/admin/realms/$REALM" "${AUTH[@]}" \
  -H "Content-Type: application/json" \
  --data-binary @/tmp/fejd-realm-updated.json > /dev/null

echo "==> Ensuring registration_role protocol mappers"
for CLIENT in fejd-frontend salon-mobile; do
  CLIENT_ID="$(curl -sf "$BASE/admin/realms/$REALM/clients?clientId=$CLIENT" "${AUTH[@]}" | jq -r '.[0].id')"
  [ "$CLIENT_ID" != "null" ] || { echo "client $CLIENT not found" >&2; exit 1; }
  EXISTS="$(curl -sf "$BASE/admin/realms/$REALM/clients/$CLIENT_ID/protocol-mappers/models" "${AUTH[@]}" \
    | jq '[.[] | select(.name == "registration_role")] | length')"
  if [ "$EXISTS" = "0" ]; then
    curl -sf -X POST "$BASE/admin/realms/$REALM/clients/$CLIENT_ID/protocol-mappers/models" \
      "${AUTH[@]}" -H "Content-Type: application/json" \
      -d '{
        "name": "registration_role",
        "protocol": "openid-connect",
        "protocolMapper": "oidc-usermodel-attribute-mapper",
        "config": {
          "user.attribute": "registration_role",
          "id.token.claim": "false",
          "access.token.claim": "true",
          "claim.name": "registration_role",
          "jsonType.label": "String",
          "userinfo.token.claim": "false"
        }
      }' > /dev/null
    echo "    + $CLIENT registration_role mapper"
  else
    echo "    $CLIENT already has registration_role mapper"
  fi
done

echo "==> Enabling user profile unmanaged attributes"
UP_COMPONENT="$(curl -sf "$BASE/admin/realms/$REALM/components?type=org.keycloak.userprofile.UserProfileProvider" "${AUTH[@]}")"
UP_ID="$(echo "$UP_COMPONENT" | jq -r '.[] | select(.providerId == "declarative-user-profile") | .id' | head -n1)"
if [ -n "$UP_ID" ] && [ "$UP_ID" != "null" ]; then
  curl -sf "$BASE/admin/realms/$REALM/components/$UP_ID" "${AUTH[@]}" \
    | jq '.config["kc.user.profile.config"] = [ (.config["kc.user.profile.config"][0] | fromjson | .unmanagedAttributePolicy = "ENABLED" | tojson) ]' \
    > /tmp/fejd-user-profile.json
  curl -sf -X PUT "$BASE/admin/realms/$REALM/components/$UP_ID" "${AUTH[@]}" \
    -H "Content-Type: application/json" \
    --data-binary @/tmp/fejd-user-profile.json > /dev/null
  echo "    user profile unmanagedAttributePolicy = ENABLED"
else
  echo "    ! declarative-user-profile component not found (skipping)"
fi

echo "==> Done. Realm '$REALM' updated in place."
