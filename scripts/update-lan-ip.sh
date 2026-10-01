#!/usr/bin/env bash
set -euo pipefail

# Reconfigure the app's LAN-facing URLs after the machine's IP changes:
#   1. frontend/.env                 (VITE_KEYCLOAK_URL)
#   2. keycloak/import/fejd-realm.json (fejd-frontend redirectUris + webOrigins)
#   3. the running Keycloak, via the Admin API (best-effort)
#
# Usage:
#   just ip-update
#   ./scripts/update-lan-ip.sh
#
# Overrides:
#   FEJD_DEV_PORT            frontend dev port         (default 5173)
#   KEYCLOAK_DEV_PORT        keycloak host port        (default 9090)
#   KEYCLOAK_URL             keycloak base URL         (default http://localhost:$KEYCLOAK_DEV_PORT)
#   KEYCLOAK_ADMIN_USERNAME  admin user                (default admin)
#   KEYCLOAK_ADMIN_PASSWORD  admin password            (default admin)
#   REALM                    keycloak realm            (default fejd)

cd "$(dirname "$0")/.." # repo root

FRONTEND_PORT="${FEJD_DEV_PORT:-5173}"
KEYCLOAK_PORT="${KEYCLOAK_DEV_PORT:-9090}"
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:$KEYCLOAK_PORT}"
KEYCLOAK_ADMIN_USERNAME="${KEYCLOAK_ADMIN_USERNAME:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-admin}"
REALM="${REALM:-fejd}"

ENV_FILE="frontend/.env"
REALM_FILE="keycloak/import/fejd-realm.json"

# --- detect LAN IP ----------------------------------------------------------
# The source IP used to reach the internet is the machine's primary address.
IP="$(ip route get 1.1.1.1 2>/dev/null | sed -n 's/.* src \([0-9.]*\).*/\1/p')"
if [ -z "$IP" ]; then
  IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
fi
if [ -z "$IP" ]; then
  echo "error: could not detect LAN IP" >&2
  exit 1
fi
echo "LAN IP: $IP"

# --- frontend/.env ----------------------------------------------------------
if [ -f "$ENV_FILE" ] && grep -q '^VITE_KEYCLOAK_URL=' "$ENV_FILE"; then
  sed -i "s|^VITE_KEYCLOAK_URL=.*|VITE_KEYCLOAK_URL=http://$IP:$KEYCLOAK_PORT|" "$ENV_FILE"
else
  printf '\n# Keycloak base URL (LAN). Managed by `just ip-update`.\nVITE_KEYCLOAK_URL=http://%s:%s\n' \
    "$IP" "$KEYCLOAK_PORT" >> "$ENV_FILE"
fi
echo "  updated $ENV_FILE -> VITE_KEYCLOAK_URL=http://$IP:$KEYCLOAK_PORT"

# --- keycloak realm (fejd-frontend redirectUris + webOrigins) ---------------
IP="$IP" PORT="$FRONTEND_PORT" FILE="$REALM_FILE" node -e '
  const fs = require("fs")
  const { IP, PORT, FILE } = process.env
  const data = JSON.parse(fs.readFileSync(FILE, "utf8"))
  for (const client of data.clients || []) {
    if (client.clientId === "fejd-frontend") {
      const keep = (u) => /fejd|localhost/.test(u)
      client.redirectUris = (client.redirectUris || []).filter(keep).concat([`http://${IP}:${PORT}/*`])
      client.webOrigins = (client.webOrigins || []).filter(keep).concat([`http://${IP}:${PORT}`])
    }
  }
  fs.writeFileSync(FILE, JSON.stringify(data, null, 2) + "\n")
'
echo "  updated $REALM_FILE"

# --- apply to the running Keycloak (best-effort) ----------------------------
IP="$IP" PORT="$FRONTEND_PORT" KEYCLOAK_URL="$KEYCLOAK_URL" \
KEYCLOAK_ADMIN_USERNAME="$KEYCLOAK_ADMIN_USERNAME" KEYCLOAK_ADMIN_PASSWORD="$KEYCLOAK_ADMIN_PASSWORD" \
REALM="$REALM" node --input-type=module <<'NODE'
const { IP, PORT, KEYCLOAK_URL, KEYCLOAK_ADMIN_USERNAME, KEYCLOAK_ADMIN_PASSWORD, REALM } = process.env
const keep = (u) => /fejd|localhost/.test(u)

async function main() {
  const reachable = await fetch(`${KEYCLOAK_URL}/realms/${REALM}`).catch(() => null)
  if (!reachable) {
    console.log(`  Keycloak not reachable at ${KEYCLOAK_URL} — skipped (applies on next realm import)`)
    return
  }

  const tokenRes = await fetch(`${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: "admin-cli",
      username: KEYCLOAK_ADMIN_USERNAME,
      password: KEYCLOAK_ADMIN_PASSWORD,
      grant_type: "password",
    }),
  })
  if (!tokenRes.ok) throw new Error(`auth failed (${tokenRes.status})`)
  const { access_token } = await tokenRes.json()
  const auth = { Authorization: `Bearer ${access_token}` }

  const listRes = await fetch(`${KEYCLOAK_URL}/admin/realms/${REALM}/clients?clientId=fejd-frontend`, { headers: auth })
  if (!listRes.ok) throw new Error(`list clients failed (${listRes.status})`)
  const [client] = await listRes.json()
  if (!client?.id) throw new Error("fejd-frontend client not found")

  const getRes = await fetch(`${KEYCLOAK_URL}/admin/realms/${REALM}/clients/${client.id}`, { headers: auth })
  if (!getRes.ok) throw new Error(`get client failed (${getRes.status})`)
  const rep = await getRes.json()

  rep.redirectUris = (rep.redirectUris || []).filter(keep).concat([`http://${IP}:${PORT}/*`])
  rep.webOrigins = (rep.webOrigins || []).filter(keep).concat([`http://${IP}:${PORT}`])

  const putRes = await fetch(`${KEYCLOAK_URL}/admin/realms/${REALM}/clients/${client.id}`, {
    method: "PUT",
    headers: { ...auth, "Content-Type": "application/json" },
    body: JSON.stringify(rep),
  })
  if (!putRes.ok) throw new Error(`update client failed (${putRes.status})`)
  console.log(`  updated running Keycloak client -> http://${IP}:${PORT}`)
}

main().catch((e) => console.error(`  warning: keycloak update failed — ${e.message}`))
NODE

echo
echo "Done. Restart the frontend dev server to pick up the change:  just dev-frontend"
echo "Access from other devices at: http://$IP:$FRONTEND_PORT"
