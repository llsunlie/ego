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

# ---------- run migrations ----------
log "running migrations..."
for f in "$MIGRATIONS_DIR"/*.sql; do
    name=$(basename "$f")
    docker exec -i ego-postgres-1 psql -U ego -d ego < "$f" >/dev/null 2>&1
    log "  applied $name"
done
log "all migrations applied"

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
log "starting flutter web-server..."
cd "$ROOT/client"
flutter run -d web-server --web-port "$FLUTTER_PORT" --web-hostname 0.0.0.0 &
FLUTTER_PID=$!

for i in $(seq 1 30); do
    if curl -s -o /dev/null "http://localhost:${FLUTTER_PORT}" 2>/dev/null; then break; fi
    sleep 1
done

if ! curl -s -o /dev/null "http://localhost:${FLUTTER_PORT}" 2>/dev/null; then
    err "flutter web-server failed to start on :$FLUTTER_PORT"
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
