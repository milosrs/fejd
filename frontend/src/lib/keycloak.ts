import Keycloak from "keycloak-js"

const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL || "http://localhost:9090",
  realm: import.meta.env.VITE_KEYCLOAK_REALM || "fejd",
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || "fejd-frontend",
})

export default keycloak
