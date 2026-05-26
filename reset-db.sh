#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +%H:%M:%S)]${NC} $1"; }

log "truncating all tables in ego database..."
docker exec ego-postgres-1 psql -U ego -d ego -c \
    "TRUNCATE TABLE chat_messages, chat_sessions, constellations, echos, insights, moments, stars, traces, users CASCADE;"

log "done — all tables cleared"
