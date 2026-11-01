#!/usr/bin/env bash
# Full browser-BFF -> private agent -> HRIS API -> PostgreSQL test.
# Uses synthetic local demo data and the mock model provider, never a real AI key.
set -euo pipefail
export OPENAI_API_KEY="ci-mock-model-key-not-a-secret"
COMPOSE=(docker compose -f docker-compose.yml -f docker-compose.agentic.yml -f docker-compose.ci-agentic.yml)
COOKIE_JAR=$(mktemp)
TMP_BODY=$(mktemp)
cleanup() {
  code=$?
  if [ "$code" -ne 0 ]; then
    "${COMPOSE[@]}" ps || true
    "${COMPOSE[@]}" logs --tail=90 api web agent mock-model migrate || true
  fi
  "${COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1 || true
  rm -f "$COOKIE_JAR" "$TMP_BODY"
  exit "$code"
}
trap cleanup EXIT

# Validate the separate production Compose model using only placeholder credentials.
MIGRATION_DATABASE_URL='postgres://migrator:ci@db:5432/hris' \
DATABASE_URL='postgres://app:ci@db:5432/hris' \
AGENT_DATABASE_URL='postgres://agent:ci@db:5432/hris' \
JWT_SECRET='ci-placeholder-signing-key-not-for-production-0123456789' \
WEB_ORIGIN='https://ci.example.invalid' \
OPENAI_API_KEY='ci-mock-model-key-not-a-secret' \
docker compose -f docker-compose.production.yml -f docker-compose.production.agentic.yml config --format json \
  | python3 scripts/check-production-compose.py

"${COMPOSE[@]}" config --quiet
"${COMPOSE[@]}" up -d --build --wait
curl --fail --silent http://localhost:8080/health >/dev/null
curl --fail --silent http://localhost:3000/docs >/dev/null

curl --fail --silent --show-error -c "$COOKIE_JAR" -X POST http://localhost:3000/api/session/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ava@northstar.local","password":"local-demo-password-2026","organization_slug":"northstar"}' \
  | python3 -c 'import json,sys; assert json.load(sys.stdin)["authenticated"] is True'

curl --fail --silent -b "$COOKIE_JAR" http://localhost:3000/api/agent/actions?limit=10 \
  | python3 -c 'import json,sys; assert json.load(sys.stdin)["data"] == []'

curl --fail --silent --show-error -b "$COOKIE_JAR" -X POST http://localhost:3000/api/agent/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Check my annual leave balance","history":[]}' \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); assert d["reply"] and not d["pending_actions"]'

# Use a weekday future range so the test remains valid after 2026.
START=$(date -u -d "next monday + 14 days" +%F)
END=$(date -u -d "next monday + 15 days" +%F)
MESSAGE="Prepare annual leave from $START to $END"

prepare() {
  python3 -c 'import json,sys; print(json.dumps({"message": sys.argv[1], "history": []}))' "$MESSAGE" |
    curl --fail --silent --show-error -b "$COOKIE_JAR" -X POST http://localhost:3000/api/agent/chat \
      -H 'Content-Type: application/json' --data-binary @-
}
FIRST=$(prepare)
FIRST_ID=$(printf '%s' "$FIRST" | python3 -c 'import json,sys; d=json.load(sys.stdin); assert len(d["pending_actions"])==1; print(d["pending_actions"][0]["id"])')
curl --fail --silent --show-error -b "$COOKIE_JAR" -X POST "http://localhost:3000/api/agent/actions/$FIRST_ID/dismiss" \
  -H 'Content-Type: application/json' -d '{}' \
  | python3 -c 'import json,sys; assert json.load(sys.stdin)["status"] == "dismissed"'

CODE=$(curl -s -o "$TMP_BODY" -w '%{http_code}' -b "$COOKIE_JAR" -X POST http://localhost:3000/api/agent/confirm \
  -H 'Content-Type: application/json' -d "{\"action_id\":\"$FIRST_ID\"}")
test "$CODE" = "409"

SECOND=$(prepare)
SECOND_ID=$(printf '%s' "$SECOND" | python3 -c 'import json,sys; d=json.load(sys.stdin); assert len(d["pending_actions"])==1; print(d["pending_actions"][0]["id"])')
curl --fail --silent --show-error -b "$COOKIE_JAR" -X POST http://localhost:3000/api/agent/confirm \
  -H 'Content-Type: application/json' -d "{\"action_id\":\"$SECOND_ID\"}" \
  | python3 -c 'import json,sys; assert json.load(sys.stdin)["status"] == "completed"'

curl --fail --silent --show-error -b "$COOKIE_JAR" http://localhost:3000/api/agent/actions?limit=10 \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); statuses={x["status"] for x in d["data"]}; assert {"dismissed","completed"}.issubset(statuses), statuses'

echo "PASS: BFF session, model tool calls, read scopes, staged mutation, one-time dismiss, confirmation and history."
