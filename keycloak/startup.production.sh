#!/bin/sh
set -eu

import_dir="/opt/keycloak/data/import"
mkdir -p "$import_dir"

sed \
  -e '/"http:\/\/localhost:5173\/\*"/d' \
  -e '/"http:\/\/localhost:5173"/d' \
  -e '/"https:\/\/fejd\.fyi\/\*",/s/,$//' \
  -e '/"https:\/\/fejd\.fyi",/s/,$//' \
  -e "s|\${FEJD_ADMIN_CLIENT_SECRET}|${FEJD_ADMIN_CLIENT_SECRET}|g" \
  /opt/keycloak/realm-template.json > "${import_dir}/fejd-realm.json"

exec /opt/keycloak/bin/kc.sh start \
  --import-realm \
  --http-port=8080 \
  --http-enabled=true \
  --proxy-headers=xforwarded \
  --hostname="${KC_HOSTNAME:-https://auth.fejd.fyi}"