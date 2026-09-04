#!/bin/sh
set -eu

exec /opt/keycloak/bin/kc.sh start-dev \
  --import-realm \
  --http-port=8080 \
  --db=postgres \
  --db-url=jdbc:postgresql://db:5432/fejd \
  --db-schema=keycloak \
  --db-username=postgres \
  --db-password=postgres \
  --hostname-strict=false \
  --hostname-debug=true
