# Fejd production deployment

This deployment builds the Fejd frontend, backend, and SMTP image on the VPS
and runs them with Docker Compose. Caddy is the only service publishing host
ports. Cloudflare should proxy `fejd.fyi` and `auth.fejd.fyi` to the VPS.

## Prerequisites

- Ubuntu VPS with Docker Engine and the Compose plugin.
- DNS records for `fejd.fyi` and `auth.fejd.fyi` pointing to Cloudflare.
- Cloudflare proxy enabled and HTTP/HTTPS allowed to the VPS.
- External bridge network `fejd`, normally created by `fejd-infra` Ansible.
- A checkout of the desired Fejd commit.

Check Docker and the network:

```sh
docker version
docker compose version
docker network inspect fejd
```

If Ansible has not created it yet, create the expected network once:

```sh
docker network create --driver bridge fejd
```

Do not create a second application network.

## Secrets and environment

Copy `.env.production.example` to `.env.production` on the VPS and replace
all placeholders. Prefer the secret-management procedure from `fejd-infra`.
If using a local file, protect it:

```sh
cp .env.production.example .env.production
chmod 600 .env.production
```

Required secret values are documented in `.secrets/README.md`. Never commit
`.env.production`, print it in CI logs, or pass its contents as a command-line
argument. The file must remain on the VPS.

## Deploy or update

Run from the repository root at the selected commit:

```sh
git fetch --tags origin
git checkout <commit-or-tag>
docker compose --env-file .env.production -f docker-compose.production.yml config --quiet
docker compose --env-file .env.production -f docker-compose.production.yml build --pull
docker compose --env-file .env.production -f docker-compose.production.yml up -d --remove-orphans
```

The first start creates the named volumes and runs database migrations from
the backend. The SeaweedFS S3 initializer creates `fejd-images` idempotently
and does not make it public.

The frontend production image is a static Nginx bundle. Its web API base is
empty, so browser requests use same-origin `/api`. Keycloak is built with the
public URL `https://auth.fejd.fyi`.

## Verify

Inspect service state and health:

```sh
docker compose --env-file .env.production -f docker-compose.production.yml ps
docker inspect --format '{{.Name}} {{.State.Health.Status}}' \
  fejd-frontend fejd-backend fejd-postgres fejd-keycloak fejd-caddy
```

Check the public endpoints through Caddy/Cloudflare:

```sh
curl -fsS https://fejd.fyi/ >/dev/null
curl -fsS https://fejd.fyi/health
curl -fsS https://auth.fejd.fyi/realms/fejd/.well-known/openid-configuration >/dev/null
```

Validate Caddy inside the container:

```sh
docker compose --env-file .env.production -f docker-compose.production.yml \
  exec caddy caddy validate --config /etc/caddy/Caddyfile
```

Check that only Caddy publishes host ports:

```sh
docker compose -f docker-compose.production.yml config | rg -n 'published:|ports:'
```

The only published ports should be Caddy's `80` and `443`.

## Logs and operations

```sh
docker compose --env-file .env.production -f docker-compose.production.yml logs -f caddy

docker compose --env-file .env.production -f docker-compose.production.yml logs -f backend keycloak

docker compose --env-file .env.production -f docker-compose.production.yml restart backend

docker compose --env-file .env.production -f docker-compose.production.yml restart
```

A full update should use a reviewed commit, rebuild with `--pull`, and then
run `up -d --remove-orphans`. Do not use `down -v` during normal maintenance;
it deletes persistent data volumes.

## Persistent data and backups

Named volumes contain production state:

- PostgreSQL: `postgres_data` contains the application and Keycloak database.
- SeaweedFS: `seaweedfs_data` contains uploaded image objects.
- Caddy: `caddy_data` contains ACME certificates and TLS state; `caddy_config`
  contains Caddy config state.
- SMTP: `smtp_data` contains the local mail spool when Resend is unavailable.

The exact Docker volume names include the Compose project name. List them
before backup:

```sh
docker volume ls | rg 'postgres_data|seaweedfs_data|caddy_data|caddy_config|smtp_data'
```

Use [BACKUP.md](BACKUP.md) for database and object-storage backup/restore
commands. Backups must be encrypted, access-controlled, and tested by a
restore into a separate environment.

## Important assumptions

- `fejd-infra` owns VPS provisioning, Docker installation, the external
  `fejd` network, and secret delivery.
- Cloudflare terminates or passes HTTPS to Caddy; Caddy still manages its own
  certificates and requires outbound ACME access unless Cloudflare Origin
  Certificates are configured separately.
- The imported realm is filtered at production startup to remove the local
  `http://localhost:5173` frontend redirect origins. Existing Keycloak data is
  not overwritten on every restart; realm changes require an explicit migration
  or controlled re-import procedure.
- The current repository has no GitHub Actions deployment or GHCR workflow.
