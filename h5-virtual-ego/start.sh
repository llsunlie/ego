#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

FRONTEND_PORT="${FRONTEND_PORT:-8080}"
PROXY_PORT="${PROXY_PORT:-8090}"

echo "=== ego H5 Virtual Ego ==="
echo "[start] proxy:    http://localhost:$PROXY_PORT/chat"
echo "[start] frontend: http://localhost:$FRONTEND_PORT"

python3 proxy/server.py &
PROXY_PID=$!

python3 -m http.server "$FRONTEND_PORT" --bind 0.0.0.0 &
FRONTEND_PID=$!

trap "kill $PROXY_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM
wait
