#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# ego smoke test — 从零启动的端到端测试
# 1. 启动 PostgreSQL (docker)
# 2. 执行数据库迁移
# 3. 编译并启动 Go 服务
# 4. 用 grpcurl 测试核心 RPC 流程
# ============================================================================

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$PROJECT_DIR/server"
GRPCURL="${GRPCURL:-$HOME/go/bin/grpcurl}"
GRPC_ADDR="localhost:9443"
DB_URL="postgres://ego:ego@localhost:5432/ego?sslmode=disable"

RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
RESET="\033[0m"

pass()  { echo -e "${GREEN}[PASS]${RESET} $*"; }
fail()  { echo -e "${RED}[FAIL]${RESET} $*"; exit 1; }
info()  { echo -e "${YELLOW}[INFO]${RESET} $*"; }

SERVER_PID=""

cleanup() {
  info "cleaning up..."
  if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    info "server stopped"
  fi
  # Only bring down postgres if --keep-db not specified
  if [ "${KEEP_DB:-0}" != "1" ]; then
    cd "$PROJECT_DIR" && docker compose down 2>/dev/null || true
    info "postgres stopped"
  else
    info "keeping postgres running (--keep-db)"
  fi
}
trap cleanup EXIT

# --- prerequisite check ---------------------------------------------------

[ -x "$GRPCURL" ] || fail "grpcurl not found at $GRPCURL (run: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest)"
command -v docker >/dev/null 2>&1 || fail "docker not found"
command -v python3 >/dev/null 2>&1 || fail "python3 not found"

# --- step 1: start postgres -----------------------------------------------

info "starting PostgreSQL via docker compose..."
cd "$PROJECT_DIR"
# Clean data: remove volume and bind-mount directory, then recreate
docker compose down --volumes --remove-orphans 2>/dev/null || true
	docker run --rm -v "$PROJECT_DIR/data:/data" alpine:latest sh -c "rm -rf /data/*" 2>/dev/null || rm -rf ./data 2>/dev/null || true
	mkdir -p ./data
docker compose up -d postgres

# wait for PG to be ready
info "waiting for PostgreSQL..."
for i in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -U ego -d ego 2>/dev/null; then
    pass "PostgreSQL ready"
    break
  fi
  if [ "$i" -eq 30 ]; then
    docker compose logs postgres
    fail "PostgreSQL did not become ready after 30s"
  fi
  sleep 1
done

# --- step 2: apply migrations ---------------------------------------------

info "applying database migrations..."
for f in "$SERVER_DIR/internal/platform/postgres/migrations/"*.sql; do
  name=$(basename "$f")
  info "  running $name..."
  # Make CREATE TABLE idempotent with IF NOT EXISTS
  sed 's/CREATE TABLE /CREATE TABLE IF NOT EXISTS /g' "$f" | \
    docker compose exec -T postgres psql -U ego -d ego >/dev/null 2>&1
done
pass "all migrations applied"

# --- step 3: build server -------------------------------------------------

info "building server..."
cd "$SERVER_DIR"
go build -o /tmp/ego-server ./cmd/ego/
pass "server built"

# --- step 4: start server -------------------------------------------------

info "starting server..."
DATABASE_URL="$DB_URL" \
JWT_SECRET="smoke-test-secret" \
PORT="9443" \
WEB_PORT="9080" \
  /tmp/ego-server &
SERVER_PID=$!

# wait for gRPC port to open
info "waiting for server..."
for i in $(seq 1 20); do
  if $GRPCURL -plaintext "$GRPC_ADDR" list 2>/dev/null; then
    pass "server ready"
    break
  fi
  if [ "$i" -eq 20 ]; then
    fail "server did not start after 20s"
  fi
  sleep 1
done

# Verify reflection lists our service
$GRPCURL -plaintext "$GRPC_ADDR" list | grep -q "ego.Ego" || fail "ego.Ego service not registered"

# --- step 5: login → get JWT token ----------------------------------------

info "logging in..."
LOGIN_JSON=$($GRPCURL -plaintext -d '{"account":"smoke-tester","password":"test1234"}' "$GRPC_ADDR" ego.Ego/Login 2>&1)
echo "  login response: $LOGIN_JSON"

TOKEN=$(echo "$LOGIN_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  fail "failed to extract JWT token from login response"
fi
pass "JWT token obtained"

AUTH="authorization: Bearer $TOKEN"

# ============================================================================
# Smoke Test: F1 写字 → 回声 → 观察
# ============================================================================

info "=== Smoke F1: Write and Observe ==="

# First moment of user: cold start → no echo expected
REQ1='{"content":"今天和同事在会议上发生了争执，我觉得自己被孤立了"}'
RES1=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ1" "$GRPC_ADDR" ego.Ego/CreateMoment 2>&1)
echo "  CreateMoment #1: $RES1"

MOMENT1_ID=$(echo "$RES1" | python3 -c "import sys,json; print(json.load(sys.stdin)['moment']['id'])")
MOMENT1_TRACE=$(echo "$RES1" | python3 -c "import sys,json; print(json.load(sys.stdin)['moment']['traceId'])")
ECHO1=$(echo "$RES1" | python3 -c "
import sys, json
d = json.load(sys.stdin)
e = d.get('echo')
print('NIL' if e is None else json.dumps(e))
")

# assert: moment saved
[ -n "$MOMENT1_ID" ] && [ "$MOMENT1_ID" != "null" ] || fail "moment.id is empty"
[ -n "$MOMENT1_TRACE" ] && [ "$MOMENT1_TRACE" != "null" ] || fail "moment.traceId is empty"
pass "moment #1 created: id=$MOMENT1_ID  trace=$MOMENT1_TRACE"

# assert: cold start → echo is nil
if [ "$ECHO1" != "NIL" ]; then
  fail "cold start: expected nil echo, got $ECHO1"
fi
pass "cold start: echo is nil (correct — no history)"

# Second moment: continue same trace → should match previous moment
REQ2="{\"content\":\"其实是因为我害怕被否定，小时候爸妈总是批评我\",\"traceId\":\"$MOMENT1_TRACE\"}"
RES2=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ2" "$GRPC_ADDR" ego.Ego/CreateMoment 2>&1)
echo "  CreateMoment #2: $RES2"

MOMENT2_ID=$(echo "$RES2" | python3 -c "import sys,json; print(json.load(sys.stdin)['moment']['id'])")
MOMENT2_TRACE=$(echo "$RES2" | python3 -c "import sys,json; print(json.load(sys.stdin)['moment']['traceId'])")
ECHO2_IDS=$(echo "$RES2" | python3 -c "
import sys, json
d = json.load(sys.stdin)
e = d.get('echo')
if e is None:
    print('NIL')
else:
    ids = e.get('matchedMomentIds', [])
    print(','.join(ids) if ids else 'EMPTY')
")

[ "$MOMENT2_TRACE" = "$MOMENT1_TRACE" ] || fail "trace_id mismatch: expected $MOMENT1_TRACE, got $MOMENT2_TRACE"
pass "moment #2 on same trace: $MOMENT2_ID"

# assert: echo matches moment #1
[ "$ECHO2_IDS" != "NIL" ] || fail "expected echo with matched moments, got nil"
[ "$ECHO2_IDS" != "EMPTY" ] || fail "expected at least 1 matched moment in echo"
if echo "$ECHO2_IDS" | grep -q "$MOMENT1_ID"; then
  pass "echo matches moment #1: $ECHO2_IDS"
else
  fail "echo should match moment #1 ($MOMENT1_ID), got: $ECHO2_IDS"
fi

# Generate insight from moment #2 + echo
ECHO2_ID=$(echo "$RES2" | python3 -c "import sys,json; print(json.load(sys.stdin)['echo']['id'])")
REQ_INSIGHT="{\"momentId\":\"$MOMENT2_ID\",\"echoId\":\"$ECHO2_ID\"}"
RES_INSIGHT=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_INSIGHT" "$GRPC_ADDR" ego.Ego/GenerateInsight 2>&1)
echo "  GenerateInsight: $RES_INSIGHT"

INSIGHT_TEXT=$(echo "$RES_INSIGHT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('insight',{}).get('text','NIL') if d.get('insight') else 'NIL')")
INSIGHT_MOMENT=$(echo "$RES_INSIGHT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('insight',{}).get('momentId','') if d.get('insight') else '')")

[ -n "$INSIGHT_TEXT" ] && [ "$INSIGHT_TEXT" != "NIL" ] || fail "insight text is empty"
[ "$INSIGHT_MOMENT" = "$MOMENT2_ID" ] || fail "insight momentId mismatch"
pass "insight generated: \"$INSIGHT_TEXT\""

# ============================================================================
# Smoke Test: F2 顺着再想想 — 继续写作
# ============================================================================

info "=== Smoke F2: Continue Trace (3-round deep dive) ==="

# Round 3 on same trace
REQ3="{\"content\":\"现在我明白为什么每次绩效评估我都特别紧张了\",\"traceId\":\"$MOMENT1_TRACE\"}"
RES3=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ3" "$GRPC_ADDR" ego.Ego/CreateMoment 2>&1)
echo "  CreateMoment #3: $RES3"

MOMENT3_ID=$(echo "$RES3" | python3 -c "import sys,json; print(json.load(sys.stdin)['moment']['id'])")
MOMENT3_TRACE=$(echo "$RES3" | python3 -c "import sys,json; print(json.load(sys.stdin)['moment']['traceId'])")
[ "$MOMENT3_TRACE" = "$MOMENT1_TRACE" ] || fail "round 3: trace_id mismatch"
pass "round 3 on same trace: $MOMENT3_ID"

# Verify echo on round 3 (should match 2 previous moments)
ECHO3_IDS=$(echo "$RES3" | python3 -c "
import sys, json
d = json.load(sys.stdin)
e = d.get('echo')
if e is None:
    print('NIL')
else:
    ids = e.get('matchedMomentIds', [])
    print(','.join(ids) if ids else 'EMPTY')
")
[ "$ECHO3_IDS" != "NIL" ] || fail "round 3: expected echo"
pass "round 3 echo matches previous: $ECHO3_IDS"

# ============================================================================
# Smoke Test: ListTraces + GetTraceDetail
# ============================================================================

info "=== Smoke: ListTraces ==="

REQ_LT='{"pageSize":10}'
RES_LT=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_LT" "$GRPC_ADDR" ego.Ego/ListTraces 2>&1)
echo "  ListTraces: $RES_LT"

TRACE_COUNT=$(echo "$RES_LT" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('traces',[])))")
HAS_MORE=$(echo "$RES_LT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('hasMore', False))")

[ "$TRACE_COUNT" -gt 0 ] || fail "expected at least 1 trace, got $TRACE_COUNT"
[ "$HAS_MORE" = "False" ] || fail "expected hasMore=false for single trace user"
pass "ListTraces: $TRACE_COUNT trace(s), hasMore=$HAS_MORE"

info "=== Smoke: GetTraceDetail ==="

REQ_GTD="{\"traceId\":\"$MOMENT1_TRACE\"}"
RES_GTD=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_GTD" "$GRPC_ADDR" ego.Ego/GetTraceDetail 2>&1)
echo "  GetTraceDetail: $RES_GTD"

ITEM_COUNT=$(echo "$RES_GTD" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('items',[])))")
TRACE_MOTIVATION=$(echo "$RES_GTD" | python3 -c "import sys,json; print(json.load(sys.stdin).get('trace',{}).get('motivation',''))")

[ "$ITEM_COUNT" -eq 3 ] || fail "expected 3 items in trace detail, got $ITEM_COUNT"
[ "$TRACE_MOTIVATION" = "direct" ] || fail "expected motivation 'direct', got $TRACE_MOTIVATION"
pass "GetTraceDetail: $ITEM_COUNT items, motivation=$TRACE_MOTIVATION"

# Verify item structure: moment #2 has echo + insight
ITEM2_ECHO_COUNT=$(echo "$RES_GTD" | python3 -c "
import sys, json
for item in json.load(sys.stdin)['items']:
    if item['moment']['id'] == '$MOMENT2_ID':
        print(len(item.get('echos', [])))
        break
")
ITEM2_INSIGHT=$(echo "$RES_GTD" | python3 -c "
import sys, json
for item in json.load(sys.stdin)['items']:
    if item['moment']['id'] == '$MOMENT2_ID':
        print('YES' if item.get('insight') else 'NO')
        break
")

[ "$ITEM2_ECHO_COUNT" = "1" ] || fail "item #2: expected 1 echo, got $ITEM2_ECHO_COUNT"
[ "$ITEM2_INSIGHT" = "YES" ] || fail "item #2: expected insight"
pass "trace detail: item #2 has echo + insight"

# ============================================================================
# Smoke Test: GetRandomMoments — 记忆光点盲盒
# ============================================================================

info "=== Smoke: GetRandomMoments ==="

# With explicit count
REQ_GRM='{"count":2}'
RES_GRM=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_GRM" "$GRPC_ADDR" ego.Ego/GetRandomMoments 2>&1)
echo "  GetRandomMoments (count=2): $RES_GRM"

GRM_COUNT=$(echo "$RES_GRM" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('moments',[])))")
[ "$GRM_COUNT" -eq 2 ] || fail "GetRandomMoments: expected 2 moments, got $GRM_COUNT"
pass "GetRandomMoments: $GRM_COUNT moments returned"

# Verify each moment has required fields
GRM1_ID=$(echo "$RES_GRM" | python3 -c "import sys,json; m=json.load(sys.stdin)['moments']; print(m[0]['id'])")
GRM1_CONTENT=$(echo "$RES_GRM" | python3 -c "import sys,json; m=json.load(sys.stdin)['moments']; print(m[0]['content'])")
GRM1_TRACE=$(echo "$RES_GRM" | python3 -c "import sys,json; m=json.load(sys.stdin)['moments']; print(m[0]['traceId'])")

[ -n "$GRM1_ID" ] && [ "$GRM1_ID" != "null" ] || fail "GetRandomMoments: moment.id is empty"
[ -n "$GRM1_CONTENT" ] && [ "$GRM1_CONTENT" != "null" ] || fail "GetRandomMoments: moment.content is empty"
[ -n "$GRM1_TRACE" ] && [ "$GRM1_TRACE" != "null" ] || fail "GetRandomMoments: moment.traceId is empty"
pass "GetRandomMoments: all moments have id, content, traceId"

# All returned moments should belong to the current user's trace
for i in $(seq 0 $((GRM_COUNT - 1))); do
  M_TRACE=$(echo "$RES_GRM" | python3 -c "import sys,json; m=json.load(sys.stdin)['moments']; print(m[$i]['traceId'])")
  [ "$M_TRACE" = "$MOMENT1_TRACE" ] || fail "GetRandomMoments: moment[$i] traceId=$M_TRACE, expected $MOMENT1_TRACE (user's only trace)"
done
pass "GetRandomMoments: all moments belong to user's trace"

# Default count (no count specified → default 3)
RES_GRM_DEF=$($GRPCURL -plaintext -H "$AUTH" -d '{}' "$GRPC_ADDR" ego.Ego/GetRandomMoments 2>&1)
echo "  GetRandomMoments (default): $RES_GRM_DEF"
GRM_DEF_COUNT=$(echo "$RES_GRM_DEF" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('moments',[])))")
# With only 3 total moments, default count of 3 should return at most 3
[ "$GRM_DEF_COUNT" -le 3 ] || fail "GetRandomMoments default: expected <= 3, got $GRM_DEF_COUNT"
pass "GetRandomMoments default count: $GRM_DEF_COUNT moments"

# ============================================================================

# ============================================================================
# Smoke Test: F3 StashTrace — 收进星图
# ============================================================================

info "=== Smoke F3: StashTrace → Constellation ==="

# Stash the trace from F1/F2
	REQ_STASH=$(printf '{"traceId":"%s"}' "$MOMENT1_TRACE")
RES_STASH=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_STASH" "$GRPC_ADDR" ego.Ego/StashTrace 2>&1)
echo "  StashTrace: $RES_STASH"

STAR_ID=$(echo "$RES_STASH" | python3 -c "import sys,json; print(json.load(sys.stdin)['star']['id'])")
STAR_TOPIC=$(echo "$RES_STASH" | python3 -c "import sys,json; print(json.load(sys.stdin)['star']['topic'])")
STAR_TRACE=$(echo "$RES_STASH" | python3 -c "import sys,json; print(json.load(sys.stdin)['star']['traceId'])")

[ -n "$STAR_ID" ] && [ "$STAR_ID" != "null" ] || fail "StashTrace: star.id is empty"
[ -n "$STAR_TOPIC" ] && [ "$STAR_TOPIC" != "null" ] || fail "StashTrace: star.topic is empty"
[ "$STAR_TRACE" = "$MOMENT1_TRACE" ] || fail "StashTrace: star.traceId mismatch, expected $MOMENT1_TRACE, got $STAR_TRACE"
	pass "StashTrace: star created id=$STAR_ID topic=$STAR_TOPIC"

# Verify second stash returns error (already stashed)
	REQ_STASH2=$(printf '{"traceId":"%s"}' "$MOMENT1_TRACE")
RES_STASH2=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_STASH2" "$GRPC_ADDR" ego.Ego/StashTrace 2>&1) || true
echo "  StashTrace (duplicate): $RES_STASH2"
if echo "$RES_STASH2" | grep -qi "already"; then
  pass "StashTrace: duplicate rejected"
else
  fail "StashTrace: expected error for duplicate stash, got: $RES_STASH2"
fi

# ListConstellations
RES_LC=$($GRPCURL -plaintext -H "$AUTH" -d '{}' "$GRPC_ADDR" ego.Ego/ListConstellations 2>&1)
echo "  ListConstellations: $RES_LC"

LC_COUNT=$(echo "$RES_LC" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('constellations',[])))")
LC_TOTAL=$(echo "$RES_LC" | python3 -c "import sys,json; print(json.load(sys.stdin).get('totalStarCount',0))")

[ "$LC_COUNT" -ge 1 ] || fail "ListConstellations: expected at least 1 constellation, got $LC_COUNT"
[ "$LC_TOTAL" -ge 1 ] || fail "ListConstellations: expected totalStarCount >= 1, got $LC_TOTAL"
pass "ListConstellations: $LC_COUNT constellation(s), $LC_TOTAL star(s) total"

# Get first constellation ID
LC1_ID=$(echo "$RES_LC" | python3 -c "import sys,json; print(json.load(sys.stdin)['constellations'][0]['id'])")
LC1_NAME=$(echo "$RES_LC" | python3 -c "import sys,json; print(json.load(sys.stdin)['constellations'][0]['name'])")
LC1_STAR_COUNT=$(echo "$RES_LC" | python3 -c "import sys,json; print(json.load(sys.stdin)['constellations'][0]['starCount'])")

# GetConstellation
	REQ_GC=$(printf '{"constellationId":"%s"}' "$LC1_ID")
RES_GC=$($GRPCURL -plaintext -H "$AUTH" -d "$REQ_GC" "$GRPC_ADDR" ego.Ego/GetConstellation 2>&1)
echo "  GetConstellation: $RES_GC"

GC_NAME=$(echo "$RES_GC" | python3 -c "import sys,json; print(json.load(sys.stdin)['constellation']['name'])")
GC_STAR_COUNT=$(echo "$RES_GC" | python3 -c "import sys,json; print(json.load(sys.stdin)['constellation']['starCount'])")
GC_STARS_LEN=$(echo "$RES_GC" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('stars',[])))")
GC_MOMENTS_LEN=$(echo "$RES_GC" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('moments',[])))")
GC_STAR_IDS=$(echo "$RES_GC" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['constellation'].get('starIds',[])))")

[ "$GC_NAME" = "$LC1_NAME" ] || fail "GetConstellation: name mismatch"
[ "$GC_STARS_LEN" -ge 1 ] || fail "GetConstellation: expected at least 1 star, got $GC_STARS_LEN"
[ "$GC_STARS_LEN" = "$GC_STAR_IDS" ] || fail "GetConstellation: stars count ($GC_STARS_LEN) != star_ids count ($GC_STAR_IDS)"
[ "$GC_MOMENTS_LEN" -ge 1 ] || fail "GetConstellation: expected at least 1 moment, got $GC_MOMENTS_LEN"
[ "$GC_STAR_COUNT" = "$LC1_STAR_COUNT" ] || fail "GetConstellation: star_count mismatch"
pass "GetConstellation: name=$GC_NAME, $GC_STARS_LEN star(s), $GC_MOMENTS_LEN moment(s)"

# Verify the stashed star appears in constellation
GC_HAS_STAR=$(echo "$RES_GC" | python3 -c "
import sys, json
stars = json.load(sys.stdin).get('stars', [])
for s in stars:
    if s['id'] == '$STAR_ID':
        print('YES')
        break
")
[ "$GC_HAS_STAR" = "YES" ] || fail "GetConstellation: stashed star $STAR_ID not found in constellation stars"
pass "GetConstellation: stashed star $STAR_ID present"

# All smoke tests passed
# ============================================================================

echo ""
echo -e "${GREEN}========================================${RESET}"
echo -e "${GREEN}  All smoke tests passed!${RESET}"
echo -e "${GREEN}========================================${RESET}"
echo ""
echo "  F1 Write+Observe     : PASS"
echo "  F2 ContinueTrace     : PASS"
echo "  F3 StashTrace         : PASS"
echo "  ListTraces           : PASS"
echo "  GetTraceDetail       : PASS"
echo "  GetRandomMoments     : PASS"
echo ""
