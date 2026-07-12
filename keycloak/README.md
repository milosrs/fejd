# Keycloak theme automation

This folder is mounted directly into the Keycloak container at `/opt/keycloak/themes` and `/opt/keycloak/providers`.

- Add new theme folders under `themes/<theme-name>/` and they will be visible to Keycloak automatically.
- Drop `.jar` provider files into `providers/` and they will be picked up on container restart.
