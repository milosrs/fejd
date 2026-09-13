#!/bin/sh
set -eu

data_dir="${KC_DATA_DIR:-/tmp/keycloak-data}"
export KC_DATA_DIR="$data_dir"
mkdir -p "${data_dir}/import"
sed \
  -e '/"http:\/\/localhost:5173\/\*"/d' \
  -e '/"http:\/\/localhost:5173"/d' \
  -e '/"https:\/\/fejd\.fyi\/\*",/s/,$//' \
  -e '/"https:\/\/fejd\.fyi",/s/,$//' \
  /opt/keycloak/realm-template.json > "${data_dir}/import/fejd-realm.json"

exec /opt/keycloak/bin/kc.sh start \
  --import-realm \
  --http-port=8080 \
  --http-enabled=true \
  --proxy-headers=xforwarded \
  --hostname="${KC_HOSTNAME:-https://auth.fejd.fyi}"