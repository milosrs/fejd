# Fejd backup and restore

Run these commands from the production repository root with the production
Compose file and `.env.production` available. Store backup output outside the
repository, encrypt it, and restrict access.

## PostgreSQL

PostgreSQL contains both Fejd application data and the Keycloak schema. A
logical dump is the primary backup:

```sh
mkdir -p backups/postgres

docker compose --env-file .env.production -f docker-compose.production.yml \
  exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom' \
  > backups/postgres/fejd-$(date -u +%Y%m%dT%H%M%SZ).dump
```

Restore into a stopped or isolated target database, not over an active
production database:

```sh
docker compose --env-file .env.production -f docker-compose.production.yml \
  exec -T postgres sh -c 'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner' \
  < backups/postgres/fejd-<timestamp>.dump
```

For a clean restore, create an empty target database and restore there first;
validate application and Keycloak login before switching traffic.

## SeaweedFS objects

SeaweedFS stores uploaded images through its private S3-compatible endpoint.
Use the pinned AWS CLI image to mirror the `fejd-images` bucket to encrypted
backup storage:

```sh
mkdir -p backups/seaweedfs

docker run --rm --entrypoint /bin/sh \
  --network fejd \
  -v "$PWD/backups/seaweedfs:/backup" \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION \
  amazon/aws-cli:2.27.41 \
  -c 'aws s3 sync s3://fejd-images /backup/fejd-images --endpoint-url http://seaweedfs:8333'
```

The command expects `SEAWEEDFS_S3_ACCESS_KEY` and `SEAWEEDFS_S3_SECRET_KEY` to be exported
from the protected environment without printing them:

```sh
set -a; . ./.env.production; set +a
export AWS_ACCESS_KEY_ID="$SEAWEEDFS_S3_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$SEAWEEDFS_S3_SECRET_KEY"
export AWS_DEFAULT_REGION="${S3_REGION:-us-east-1}"
```

Restore objects to an empty or explicitly selected target bucket:

```sh
docker run --rm --entrypoint /bin/sh \
  --network fejd \
  -v "$PWD/backups/seaweedfs:/backup:ro" \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION \
  amazon/aws-cli:2.27.41 \
  -c 'aws s3 sync /backup/fejd-images s3://fejd-images --endpoint-url http://seaweedfs:8333'
```

## Caddy and SMTP volumes

Caddy's `caddy_data` contains certificates and should be included in host
snapshots or volume backups. `caddy_config` preserves Caddy state. `smtp_data`
only contains local mail spool data when the SMTP relay falls back to local
mail storage. Back up these volumes with the VPS backup system or a controlled
Docker volume archive; do not use `docker compose down -v`.

Before restoring any volume archive, stop the affected service and verify the
volume name with `docker volume ls`. Test restores on a separate VPS before
using them in production.
