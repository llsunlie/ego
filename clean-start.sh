#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
GRPC_PORT="${PORT:-9443}"
WEB_PORT="${WEB_PORT:-9080}"
FLUTTER_PORT="${FLUTTER_PORT:-9081}"
DATA_DIR="$ROOT/data"
MIGRATIONS_DIR="$ROOT/server/internal/platform/postgres/migrations"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[$(date +%H:%M:%S)]${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +%H:%M:%S)]${NC} $1"; }
err()  { echo -e "${RED}[$(date +%H:%M:%S)]${NC} $1"; }

cleanup() {
    log "shutting down..."
    kill ${GO_PID:-} 2>/dev/null || true
    kill ${FLUTTER_PID:-} 2>/dev/null || true
    wait ${GO_PID:-} 2>/dev/null || true
    wait ${FLUTTER_PID:-} 2>/dev/null || true
    log "done"
}
trap cleanup EXIT INT TERM

# ---------- kill stale ports ----------
log "clearing ports ${GRPC_PORT} ${WEB_PORT} ${FLUTTER_PORT}..."
for port in $GRPC_PORT $WEB_PORT $FLUTTER_PORT; do
    lsof -ti:"$port" 2>/dev/null | xargs -r kill -9 || true
done
sleep 0.5

# ---------- clean postgres data ----------
log "stopping postgres..."
docker compose down 2>/dev/null || true

log "removing postgres data..."
docker run --rm -v "$DATA_DIR:/data" alpine sh -c 'rm -rf /data/*' 2>/dev/null || true

# ---------- postgres ----------
log "starting postgres..."
docker compose up -d postgres

log "waiting for postgres..."
until docker exec ego-postgres-1 pg_isready -U ego >/dev/null 2>&1; do
    sleep 0.5
done
log "postgres ready"
sleep 1

# ---------- run migrations ----------
log "running migrations..."
for f in "$MIGRATIONS_DIR"/*.sql; do
    name=$(basename "$f")
    docker exec -i ego-postgres-1 psql -h localhost -U ego -d ego < "$f" >/dev/null 2>&1
    log "  applied $name"
done
log "all migrations applied"

# ---------- seed test user ----------
log "seeding test user..."
SEED_PHONE="18861622557"
SEED_PASSWORD_HASH='$2b$12$PKkQDhmSYlgRqsB1QsYAkuhou.usAXhV.vFUKD5xKQgtmu28ANQdS'  # bcrypt("llccllcc")
SEED_UUID="7be63c93-0cd1-4408-b1a4-e1add4e29649"
SEED_TS="2025-01-15T00:00:00Z"

docker compose exec -T postgres psql -U ego -d ego >/dev/null 2>&1 <<SQL
INSERT INTO users (id, phone, password_hash, created_at)
VALUES ('$SEED_UUID', '$SEED_PHONE', '$SEED_PASSWORD_HASH', '$SEED_TS')
ON CONFLICT DO NOTHING;
SQL
log "test user seeded: phone=$SEED_PHONE"

# ---------- go backend ----------
log "building & starting backend..."
cd "$ROOT/server"
go build -o /tmp/ego-server ./cmd/ego/
/tmp/ego-server &
GO_PID=$!

for i in $(seq 1 20); do
    if lsof -ti:"$GRPC_PORT" >/dev/null 2>&1; then break; fi
    sleep 0.3
done

if ! lsof -ti:"$GRPC_PORT" >/dev/null 2>&1; then
    err "backend failed to start on :$GRPC_PORT"
    exit 1
fi
log "backend ready  gRPC :${GRPC_PORT}  gRPC-web :${WEB_PORT}"

# ---------- flutter web ----------
log "syncing app version from git tag..."
cd "$ROOT"
make version

log "building flutter web (release)..."
cd "$ROOT/client"
flutter build web --release -O4 --no-source-maps --base-href /

log "starting web server..."
cd "$ROOT/client/build/web"
python3 -m http.server "$FLUTTER_PORT" --bind 0.0.0.0 &
FLUTTER_PID=$!

for i in $(seq 1 10); do
    if curl -s -o /dev/null "http://localhost:${FLUTTER_PORT}" 2>/dev/null; then break; fi
    sleep 1
done

if ! curl -s -o /dev/null "http://localhost:${FLUTTER_PORT}" 2>/dev/null; then
    err "web server failed to start on :$FLUTTER_PORT"
    exit 1
fi
log "flutter ready  http://localhost:${FLUTTER_PORT}"

echo ""
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  ${GREEN}  ego clean dev server running${NC}"
echo ""
echo -e "   Web UI: http://localhost:${FLUTTER_PORT}"
echo -e "   gRPC:   localhost:${GRPC_PORT}"
echo -e "   WebRPC: localhost:${WEB_PORT}"
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

wait
