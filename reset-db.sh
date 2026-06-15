#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
DB_SERVICE="${DB_SERVICE:-postgres}"
DB_USER="${POSTGRES_USER:-ego}"
DB_NAME="${POSTGRES_DB:-ego}"
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +%H:%M:%S)]${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +%H:%M:%S)]${NC} $1"; }

cd "$ROOT"

if ! docker compose ps "$DB_SERVICE" --status running >/dev/null 2>&1; then
    warn "postgres is not running; starting docker compose service '$DB_SERVICE'..."
    docker compose up -d "$DB_SERVICE"
fi

log "waiting for postgres..."
until docker compose exec -T "$DB_SERVICE" pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; do
    sleep 0.5
done

log "truncating all public tables in database '$DB_NAME'..."
docker compose exec -T "$DB_SERVICE" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
DECLARE
  table_list text;
BEGIN
  SELECT string_agg(format('%I.%I', schemaname, tablename), ', ')
  INTO table_list
  FROM pg_tables
  WHERE schemaname = 'public';

  IF table_list IS NULL THEN
    RAISE NOTICE 'no public tables found';
  ELSE
    EXECUTE 'TRUNCATE TABLE ' || table_list || ' RESTART IDENTITY CASCADE';
  END IF;
END $$;
SQL

log "done - all public tables cleared"

if ! command -v curl >/dev/null 2>&1; then
    warn "curl is not available; skipped Elasticsearch cleanup"
    exit 0
fi

log "done - database cleared"
