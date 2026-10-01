# syntax=docker/dockerfile:1
# React Router v8 framework-mode (SSR) runtime. Serves the web app with
# server-rendered public pages + SEO metadata.

FROM node:26.0.0-alpine3.22 AS build
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN npm install -g pnpm@v12.8.1
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
ARG VITE_FIREBASE_API_KEY
ARG VITE_FIREBASE_APP_ID
ARG VITE_FIREBASE_VAPID_KEY
ENV VITE_API_URL=$VITE_API_URL \
    VITE_KEYCLOAK_URL=$VITE_KEYCLOAK_URL \
    VITE_KEYCLOAK_REALM=$VITE_KEYCLOAK_REALM \
    VITE_KEYCLOAK_CLIENT_ID=$VITE_KEYCLOAK_CLIENT_ID \
    VITE_GOOGLE_MAPS_API_KEY=$VITE_GOOGLE_MAPS_API_KEY \
    VITE_APP_DOMAIN=$VITE_APP_DOMAIN \
    VITE_APP_HOST=$VITE_APP_HOST \
    VITE_APP_STORE_URL=$VITE_APP_STORE_URL \
    VITE_PLAY_STORE_URL=$VITE_PLAY_STORE_URL \
    VITE_FIREBASE_API_KEY=$VITE_FIREBASE_API_KEY \
    VITE_FIREBASE_APP_ID=$VITE_FIREBASE_APP_ID \
    VITE_FIREBASE_VAPID_KEY=$VITE_FIREBASE_VAPID_KEY
# Framework-mode build produces build/client (static assets) + build/server (SSR).
RUN pnpm build

FROM node:26.0.0-alpine3.22 AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    PORT=80 \
    FEJD_DOMAIN=fejd.fyi \
    APP_HOST=fejd.fyi \
    API_ORIGIN=http://backend:8080
RUN npm install -g pnpm@v12.8.1
COPY --from=build /app/package.json /app/pnpm-lock.yaml /app/pnpm-workspace.yaml ./
RUN pnpm install --prod --frozen-lockfile --reporter=append-only \
    --network-concurrency=4 --fetch-retries=8 --fetch-retry-factor=2 \
    --fetch-retry-mintimeout=10000 --fetch-retry-maxtimeout=120000 \
    --fetch-timeout=120000
COPY --from=build /app/build ./build
COPY --from=build /app/server.js ./server.js
COPY --from=build /app/app ./app
EXPOSE 80
# Serve the SSR app via the custom Node server (provides the query-client load context).
CMD ["node", "server.js"]
