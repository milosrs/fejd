# syntax=docker/dockerfile:1
FROM node:26.0.0-alpine3.22 AS build
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm@11.15.1
RUN --mount=type=cache,id=fejd-pnpm-store,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --reporter=append-only \
    --network-concurrency=4 --fetch-retries=8 --fetch-retry-factor=2 \
    --fetch-retry-mintimeout=10000 --fetch-retry-maxtimeout=120000 \
    --fetch-timeout=120000
COPY . .
ARG VITE_API_URL
ARG VITE_KEYCLOAK_URL
ARG VITE_KEYCLOAK_REALM
ARG VITE_KEYCLOAK_CLIENT_ID
ARG VITE_GOOGLE_MAPS_API_KEY
ARG VITE_APP_DOMAIN
ARG VITE_APP_HOST
ARG VITE_APP_STORE_URL
ARG VITE_PLAY_STORE_URL
ENV VITE_API_URL=$VITE_API_URL \
    VITE_KEYCLOAK_URL=$VITE_KEYCLOAK_URL \
    VITE_KEYCLOAK_REALM=$VITE_KEYCLOAK_REALM \
    VITE_KEYCLOAK_CLIENT_ID=$VITE_KEYCLOAK_CLIENT_ID \
    VITE_GOOGLE_MAPS_API_KEY=$VITE_GOOGLE_MAPS_API_KEY \
    VITE_APP_DOMAIN=$VITE_APP_DOMAIN \
    VITE_APP_HOST=$VITE_APP_HOST \
    VITE_APP_STORE_URL=$VITE_APP_STORE_URL \
    VITE_PLAY_STORE_URL=$VITE_PLAY_STORE_URL
RUN pnpm build

FROM nginx:1.30.4-alpine-perl
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]