#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
GRPC_PORT="${PORT:-9443}"
WEB_PORT="${WEB_PORT:-9080}"
FLUTTER_PORT="${FLUTTER_PORT:-9081}"
REACT_PORT="${REACT_PORT:-5173}"
ES_URL="${ELASTICSEARCH_URL:-http://localhost:9200}"
LOG_DIR="$ROOT/.tmp/logs/start"
STARTUP_COMPLETE=false

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
    local port="$1"
    local pid
    pid="$(port_pid "$port")"
    if [ -n "$pid" ]; then
        log "killing process(es) on port $port: $pid"
        if [ "$IS_WINDOWS" = true ]; then
            for p in $pid; do taskkill //F //PID "$p" 2>/dev/null || true; done
        else
            kill -9 $pid 2>/dev/null || true
        fi
    else
        log "port $port is free"
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
cleanup_started_processes() {
    [ -n "${GO_PID:-}" ] && kill_process "$GO_PID"
    [ -n "${FLUTTER_PID:-}" ] && kill_process "$FLUTTER_PID"
    [ -n "${REACT_PID:-}" ] && kill_process "$REACT_PID"
    [ -n "${GO_PID:-}" ] && wait "$GO_PID" 2>/dev/null || true
    [ -n "${FLUTTER_PID:-}" ] && wait "$FLUTTER_PID" 2>/dev/null || true
    [ -n "${REACT_PID:-}" ] && wait "$REACT_PID" 2>/dev/null || true
}

cleanup_on_exit() {
    if [ "$STARTUP_COMPLETE" = true ]; then
        return
    fi
    log "startup interrupted or failed, shutting down started processes..."
    cleanup_started_processes
    log "done"
}

cleanup_on_signal() {
    STARTUP_COMPLETE=true
    cleanup_started_processes
}

trap cleanup_on_exit EXIT
trap 'cleanup_on_signal; exit 130' INT
trap 'cleanup_on_signal; exit 143' TERM

# ── kill stale ports ────────────────────────────────────────────────
log "clearing ports ${GRPC_PORT} ${WEB_PORT} ${FLUTTER_PORT} ${REACT_PORT}..."
for port in $GRPC_PORT $WEB_PORT $FLUTTER_PORT $REACT_PORT; do
    kill_port "$port"
done
sleep 0.5

# ── docker services ─────────────────────────────────────────────────
cd "$ROOT"
log "starting docker services..."
docker compose up -d

log "waiting for postgres..."
until docker compose exec -T postgres pg_isready -U ego -d ego >/dev/null 2>&1; do
    sleep 0.5
done
log "postgres ready"

log "waiting for elasticsearch..."
until curl -fsS "$ES_URL/_cluster/health?wait_for_status=yellow&timeout=1s" >/dev/null 2>&1; do
    sleep 0.5
done
log "elasticsearch ready"

log "docker services ready"

# ── go backend ──────────────────────────────────────────────────────
log "building backend..."
cd "$ROOT/server"
go build -o /tmp/ego-server ./cmd/ego/

log "starting backend..."
mkdir -p "$ROOT/server/.tmp/logs/server" "$LOG_DIR"
BACKEND_OUT="$ROOT/server/.tmp/logs/server/start-backend.out.log"
BACKEND_ERR="$ROOT/server/.tmp/logs/server/start-backend.err.log"
: > "$BACKEND_OUT"
: > "$BACKEND_ERR"
nohup /tmp/ego-server >> "$BACKEND_OUT" 2>> "$BACKEND_ERR" < /dev/null &
GO_PID=$!

for i in $(seq 1 20); do
    if port_listening "$GRPC_PORT" && port_listening "$WEB_PORT"; then break; fi
    sleep 0.3
done

if ! port_listening "$GRPC_PORT" || ! port_listening "$WEB_PORT"; then
    err "backend failed to start on :$GRPC_PORT/:$WEB_PORT"
    err "backend stderr:"
    tail -n 80 "$BACKEND_ERR" || true
    exit 1
fi

if ! curl -fsS "http://localhost:${WEB_PORT}/health" >/dev/null 2>&1; then
    err "backend health check failed on http://localhost:${WEB_PORT}/health"
    tail -n 80 "$BACKEND_ERR" || true
    exit 1
fi
log "backend ready  gRPC :${GRPC_PORT}  gRPC-web :${WEB_PORT}"

# ── flutter web ─────────────────────────────────────────────────────
log "starting flutter web-server..."
cd "$ROOT/client"
FLUTTER_OUT="$LOG_DIR/flutter.out.log"
FLUTTER_ERR="$LOG_DIR/flutter.err.log"
: > "$FLUTTER_OUT"
: > "$FLUTTER_ERR"
nohup flutter run -d web-server --web-port "$FLUTTER_PORT" --web-hostname 0.0.0.0 >> "$FLUTTER_OUT" 2>> "$FLUTTER_ERR" < /dev/null &
FLUTTER_PID=$!

for i in $(seq 1 30); do
    if curl -s -o /dev/null "http://localhost:${FLUTTER_PORT}" 2>/dev/null; then break; fi
    sleep 1
done

if ! curl -s -o /dev/null "http://localhost:${FLUTTER_PORT}" 2>/dev/null; then
    err "flutter web-server failed to start on :$FLUTTER_PORT"
    err "flutter stderr:"
    tail -n 80 "$FLUTTER_ERR" || true
    exit 1
fi
log "flutter ready  http://localhost:${FLUTTER_PORT}"

# ── react web frontend (optional) ─────────────────────────────────────
if [ -f "$ROOT/web/package.json" ]; then
    log "starting react web frontend..."
    cd "$ROOT/web"
    REACT_OUT="$LOG_DIR/react.out.log"
    REACT_ERR="$LOG_DIR/react.err.log"
    : > "$REACT_OUT"
    : > "$REACT_ERR"
    nohup npx vite --host 0.0.0.0 --port "$REACT_PORT" >> "$REACT_OUT" 2>> "$REACT_ERR" < /dev/null &
    REACT_PID=$!

    for i in $(seq 1 30); do
        if curl -s -o /dev/null "http://localhost:${REACT_PORT}" 2>/dev/null; then break; fi
        sleep 1
    done

    if ! curl -s -o /dev/null "http://localhost:${REACT_PORT}" 2>/dev/null; then
        err "react web frontend failed to start on :$REACT_PORT"
        err "react stderr:"
        tail -n 80 "$REACT_ERR" || true
        exit 1
    fi
    log "react ready  http://localhost:${REACT_PORT}"
else
    log "react skipped (no web/package.json)"
fi

echo ""
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  ${GREEN}  ego dev server running${NC}"
echo ""
if [ -f "$ROOT/web/package.json" ]; then
    echo -e "   React:    http://localhost:${REACT_PORT}"
fi
echo -e "   Flutter:  http://localhost:${FLUTTER_PORT}"
echo -e "   gRPC:     localhost:${GRPC_PORT}"
echo -e "   gRPC-web: localhost:${WEB_PORT}"
echo -e "   Adminer:  http://localhost:10081"
echo ""
echo -e "   Backend logs: $BACKEND_OUT"
echo -e "   Flutter logs: $FLUTTER_OUT"
if [ -f "$ROOT/web/package.json" ]; then
    echo -e "   React logs:   $REACT_OUT"
fi
echo -e "  ${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# ── detach ──────────────────────────────────────────────────────────
STARTUP_COMPLETE=true
disown "$GO_PID" 2>/dev/null || true
disown "$FLUTTER_PID" 2>/dev/null || true
[ -n "${REACT_PID:-}" ] && disown "$REACT_PID" 2>/dev/null || true

log "startup complete; processes are running in the background"
