#!/usr/bin/env bash
# Restart ego backend: kill old process, build, start new.
# Usage: bash restart-backend.sh

ROOT="$(cd "$(dirname "$0")" && pwd)"
GRPC_PORT="${PORT:-9443}"

echo "=== killing old backend (port $GRPC_PORT) ==="
pids=$(netstat -ano 2>/dev/null | grep ":$GRPC_PORT " | grep LISTENING | awk '{print $NF}' | sort -u || true)
if [ -n "$pids" ]; then
    for pid in $pids; do
        taskkill //F //PID "$pid" 2>/dev/null && echo "killed PID $pid" || true
    done
    sleep 1
else
    echo "no backend process found"
fi

echo "=== building backend ==="
cd "$ROOT/server"
go build -o /tmp/ego-server ./cmd/ego/ || { echo "BUILD FAILED"; exit 1; }

echo "=== starting backend ==="
cd "$ROOT/server" && /tmp/ego-server &
BACKEND_PID=$!

for i in $(seq 1 20); do
    if netstat -ano 2>/dev/null | grep -q ":$GRPC_PORT .*LISTENING"; then
        echo ""
        echo "  backend ready   gRPC :$GRPC_PORT  web :${WEB_PORT:-9080}"
        echo "  log file:       $ROOT/server/.tmp/logs/server/server.log"
        echo "  tail log:       tail -f $ROOT/server/.tmp/logs/server/server.log"
        exit 0
    fi
    sleep 0.3
done

echo "ERROR: backend failed to start (PID $BACKEND_PID)"
exit 1
