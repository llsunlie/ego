#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
DB_SERVICE="${DB_SERVICE:-postgres}"
DB_USER="${POSTGRES_USER:-ego}"
DB_NAME="${POSTGRES_DB:-ego}"
ES_SERVICE="${ES_SERVICE:-elasticsearch}"
ES_URL="${ELASTICSEARCH_URL:-http://localhost:9200}"

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

if ! docker compose ps "$ES_SERVICE" --status running >/dev/null 2>&1; then
    warn "elasticsearch is not running; starting docker compose service '$ES_SERVICE'..."
    docker compose up -d "$ES_SERVICE"
fi

log "waiting for elasticsearch..."
until curl -fsS "$ES_URL/_cluster/health?wait_for_status=yellow&timeout=1s" >/dev/null 2>&1; do
    sleep 0.5
done

log "deleting all non-system Elasticsearch indices at '$ES_URL'..."
indices="$(curl -fsS "$ES_URL/_cat/indices?h=index" | awk 'NF && $1 !~ /^\./ { print $1 }')"
if [[ -z "$indices" ]]; then
    warn "no non-system Elasticsearch indices found"
else
    while IFS= read -r index; do
        [[ -z "$index" ]] && continue
        curl -fsS -X DELETE "$ES_URL/$index" >/dev/null
        log "  deleted ES index $index"
    done <<< "$indices"
fi

log "done - Elasticsearch indices cleared"
