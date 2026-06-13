#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

env_value() {
    local key="$1"
    local file="$ROOT/server/.env"
    if [ -f "$file" ]; then
        awk -F= -v key="$key" '$1 == key { print substr($0, index($0, "=") + 1); exit }' "$file"
    fi
}

GRPC_PORT="${GRPC_PORT:-$(env_value GRPC_PORT)}"
GRPC_PORT="${GRPC_PORT:-9444}"
WEB_PORT="${WEB_PORT:-$(env_value WEB_PORT)}"
WEB_PORT="${WEB_PORT:-9080}"
WEB_TLS_PORT="${WEB_TLS_PORT:-$(env_value WEB_TLS_PORT)}"
WEB_TLS_PORT="${WEB_TLS_PORT:-9443}"
DATABASE_URL="${DATABASE_URL:-$(env_value DATABASE_URL)}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[$(date +%H:%M:%S)]${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +%H:%M:%S)]${NC} $1"; }
err()  { echo -e "${RED}[$(date +%H:%M:%S)]${NC} $1"; }

# ── platform detection ───────────────────────────────────────────────
OS_NAME="$(uname -s 2>/dev/null || echo 'unknown')"
IS_MACOS=false
IS_WINDOWS=false
case "$OS_NAME" in
    Darwin) IS_MACOS=true ;;
    MINGW*|MSYS*|CYGWIN*) IS_WINDOWS=true ;;
esac

# ── port helpers ─────────────────────────────────────────────────────
port_pids() {
    local port="$1"
    if [ "$IS_WINDOWS" = true ]; then
        netstat -ano 2>/dev/null | grep ":$port " | grep LISTENING | awk '{print $NF}' | sort -u || true
    elif command -v lsof >/dev/null 2>&1; then
        lsof -ti:"$port" 2>/dev/null || true
    elif command -v fuser >/dev/null 2>&1; then
        fuser "$port"/tcp 2>/dev/null || true
    else
        warn "cannot inspect port $port: neither lsof nor fuser is available"
        return 0
    fi
}

kill_pid() {
    local pid="$1"
    if [ "$IS_WINDOWS" = true ]; then
        taskkill //F //PID "$pid" 2>/dev/null || true
    else
        kill "$pid" 2>/dev/null || true
    fi
}

force_kill_pid() {
    local pid="$1"
    if [ "$IS_WINDOWS" = true ]; then
        taskkill //F //PID "$pid" 2>/dev/null || true
    else
        kill -9 "$pid" 2>/dev/null || true
    fi
}

to_windows_path() {
    if command -v cygpath >/dev/null 2>&1; then
        cygpath -w "$1"
    else
        echo "$1"
    fi
}

ps_quote() {
    printf "'%s'" "$(printf "%s" "$1" | sed "s/'/''/g")"
}

port_listening() {
    local port="$1"
    [ -n "$(port_pids "$port")" ]
}

kill_port() {
    local port="$1"
    local pids
    pids="$(port_pids "$port")"
    if [ -n "$pids" ]; then
        log "killing process(es) on port $port: $pids"
        for pid in $pids; do
            kill_pid "$pid"
        done
        sleep 1
        pids="$(port_pids "$port")"
        if [ -n "$pids" ]; then
            warn "force-killing stale process(es): $pids"
            for pid in $pids; do
                force_kill_pid "$pid"
            done
            sleep 1
        fi
    else
        log "port $port is free"
    fi
}

# ── kill old ─────────────────────────────────────────────────────────
log "clearing ports $GRPC_PORT $WEB_PORT $WEB_TLS_PORT..."
kill_port "$GRPC_PORT"
kill_port "$WEB_PORT"
if [ "$WEB_TLS_PORT" != "$WEB_PORT" ] && [ "$WEB_TLS_PORT" != "$GRPC_PORT" ]; then
    kill_port "$WEB_TLS_PORT"
fi

if [[ "$DATABASE_URL" == *"localhost:5432"* ]] || [[ "$DATABASE_URL" == *"127.0.0.1:5432"* ]]; then
    if command -v nc >/dev/null 2>&1 && ! nc -z localhost 5432; then
        err "database is not reachable on localhost:5432"
        echo "  start it with: docker compose up -d postgres"
        exit 1
    fi
fi

# ── build ────────────────────────────────────────────────────────────
log "building backend..."
cd "$ROOT/server"
if [ "$IS_WINDOWS" = true ]; then
    BIN="${TMP:-/tmp}/ego-server.exe"
else
    BIN="/tmp/ego-server"
fi
go build -o "$BIN" ./cmd/ego/ || { err "BUILD FAILED"; exit 1; }

# ── start ────────────────────────────────────────────────────────────
log "starting backend..."
mkdir -p "$ROOT/server/.tmp/logs/server"
OUT_LOG="$ROOT/server/.tmp/logs/server/restart-backend.out.log"
ERR_LOG="$ROOT/server/.tmp/logs/server/restart-backend.err.log"
if [ "$IS_WINDOWS" = true ]; then
    WIN_BIN="$(to_windows_path "$BIN")"
    WIN_CWD="$(to_windows_path "$ROOT/server")"
    WIN_OUT="$(to_windows_path "$OUT_LOG")"
    WIN_ERR="$(to_windows_path "$ERR_LOG")"
    powershell.exe -NoProfile -Command \
        "Start-Process -FilePath $(ps_quote "$WIN_BIN") -WorkingDirectory $(ps_quote "$WIN_CWD") -RedirectStandardOutput $(ps_quote "$WIN_OUT") -RedirectStandardError $(ps_quote "$WIN_ERR") -WindowStyle Hidden" \
        || { err "backend start failed"; exit 1; }
    BACKEND_PID=""
else
    nohup "$BIN" >> "$OUT_LOG" 2>> "$ERR_LOG" &
    BACKEND_PID=$!
    disown "$BACKEND_PID" 2>/dev/null || true
fi

for i in $(seq 1 20); do
    if port_listening "$GRPC_PORT"; then break; fi
    sleep 0.3
done

if ! port_listening "$GRPC_PORT"; then
    err "backend failed to start on :$GRPC_PORT"
    exit 1
fi

echo ""
echo -e "  ${GREEN}backend ready${NC}   gRPC :$GRPC_PORT  web :$WEB_PORT"
echo "  log file:       $ROOT/server/.tmp/logs/server/server.log"
echo "  tail log:       tail -f $ROOT/server/.tmp/logs/server/server.log"
