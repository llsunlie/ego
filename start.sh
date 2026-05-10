#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
GRPC_PORT="${PORT:-9443}"
WEB_PORT="${WEB_PORT:-9080}"
FLUTTER_PORT="${FLUTTER_PORT:-9081}"

# Terminal colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[$(date +%H:%M:%S)]${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +%H:%M:%S)]${NC} $1"; }
err()  { echo -e "${RED}[$(date +%H:%M:%S)]${NC} $1"; }

# ── platform detection ──────────────────────────────────────────────
IS_WINDOWS=false
case "$(uname -s 2>/dev/null || echo 'unknown')" in
    MINGW*|MSYS*|CYGWIN*) IS_WINDOWS=true ;;
esac

# ── port helpers (cross-platform) ───────────────────────────────────
port_pid() {
    # prints PIDs listening on $1, one per line
    if [ "$IS_WINDOWS" = true ]; then
        netstat -ano 2>/dev/null | grep ":$1 " | grep LISTENING | awk '{print $NF}' || true
    else
        lsof -ti:"$1" 2>/dev/null || true
    fi
}

kill_port() {
    local pid
    pid="$(port_pid "$1")"
    if [ -n "$pid" ]; then
        if [ "$IS_WINDOWS" = true ]; then
            for p in $pid; do taskkill //F //PID "$p" 2>/dev/null || true; done
        else
            kill -9 $pid 2>/dev/null || true
        fi
    fi
}

port_listening() {
    # returns 0 if something is listening on $1
    [ -n "$(port_pid "$1")" ]
}

kill_process() {
    local pid="$1"
    if [ "$IS_WINDOWS" = true ]; then
        taskkill //F //PID "$pid" 2>/dev/null || true
    else
        kill "$pid" 2>/dev/null || true
    fi
}

# ── cleanup ─────────────────────────────────────────────────────────
cleanup() {
    log "shutting down..."
    [ -n "${GO_PID:-}" ] && kill_process "$GO_PID"
    [ -n "${FLUTTER_PID:-}" ] && kill_process "$FLUTTER_PID"
    wait $GO_PID 2>/dev/null || true
    wait $FLUTTER_PID 2>/dev/null || true
    log "done"
}
trap cleanup EXIT INT TERM

# ── kill stale ports ────────────────────────────────────────────────
log "clearing ports ${GRPC_PORT} ${WEB_PORT} ${FLUTTER_PORT}..."
for port in $GRPC_PORT $WEB_PORT $FLUTTER_PORT; do
    kill_port "$port"
done
sleep 0.5

# ── postgres ────────────────────────────────────────────────────────
cd "$ROOT"
if ! docker compose ps postgres 2>/dev/null | grep -q "Up"; then
    log "starting postgres..."
    docker compose up -d postgres
    sleep 2
fi
log "postgres ready"

# ── go backend ──────────────────────────────────────────────────────
log "building backend..."
cd "$ROOT/server"
go build -o /tmp/ego-server ./cmd/ego/

log "starting backend..."
/tmp/ego-server &
GO_PID=$!

for i in $(seq 1 20); do
    if port_listening "$GRPC_PORT"; then break; fi
    sleep 0.3
done

if ! port_listening "$GRPC_PORT"; then
    err "backend failed to start on :$GRPC_PORT"
    exit 1
fi
log "backend ready  gRPC :${GRPC_PORT}  gRPC-web :${WEB_PORT}"

# ── flutter web ─────────────────────────────────────────────────────
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
echo -e "  ${GREEN}  ego dev server running${NC}"
echo ""
echo -e "   Web UI: http://localhost:${FLUTTER_PORT}"
echo -e "   gRPC:   localhost:${GRPC_PORT}"
echo -e "   WebRPC: localhost:${WEB_PORT}"
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# ── wait ────────────────────────────────────────────────────────────
wait
