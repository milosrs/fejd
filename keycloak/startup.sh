#!/bin/sh
set -eu
mkdir -p /opt/keycloak/themes/fejd
cp -R /opt/keycloak/data/import/themes/fejd /opt/keycloak/themes/
exec /opt/keycloak/bin/kc.sh start-dev --http-port=8080 --db=postgres --db-url=jdbc:postgresql://db:5432/fejd --db-username=postgres --db-password=postgres --hostname-strict=false --hostname-debug=true
