#!/usr/bin/env bash
# Smoke one locally built service image. These checks intentionally exercise
# image entrypoints and protocols, not rule-engine behavior covered by tests.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE_TAG="${IMAGE_TAG:-local}"
COMPONENT="${1:-all}"
case "$COMPONENT" in
  engine|counsel|liki-mcp|all) ;;
  *) echo "usage: $0 [engine|counsel|liki-mcp|all]" >&2; exit 2 ;;
esac
ENGINE_IMAGE="${ENGINE_IMAGE:-liki-local/liki-engine:$IMAGE_TAG}"
COUNSEL_IMAGE="${COUNSEL_IMAGE:-liki-local/liki-counsel:$IMAGE_TAG}"
LIKI_MCP_IMAGE="${LIKI_MCP_IMAGE:-liki-local/liki-mcp:$IMAGE_TAG}"
ENGINE_MCP_PORT="${ENGINE_MCP_SMOKE_PORT:-18081}"
COUNSEL_PORT="${COUNSEL_SMOKE_PORT:-18086}"
LIKI_MCP_PORT="${LIKI_MCP_SMOKE_PORT:-18087}"

containers=()
cleanup() {
  for container in "${containers[@]:-}"; do
    docker rm -f "$container" >/dev/null 2>&1 || true
  done
}
trap cleanup EXIT

wait_url() {
  local name="$1" url="$2"
  for _ in $(seq 1 30); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    if [ "$(docker inspect -f '{{.State.Running}}' "$name" 2>/dev/null || true)" != true ]; then
      echo "❌ $name exited during smoke" >&2
      docker logs "$name" >&2 || true
      return 1
    fi
    sleep 1
  done
  echo "❌ $name did not become ready: $url" >&2
  docker logs "$name" >&2 || true
  return 1
}

start_container() {
  local name="$1"
  shift
  docker rm -f "$name" >/dev/null 2>&1 || true
  docker run -d --name "$name" "$@" >/dev/null
  containers+=("$name")
}

post_json() {
  local url="$1" body="$2" method="${3:-tools/call}"
  curl -fsS -X POST "$url" \
    -H 'Content-Type: application/json' \
    -H 'MCP-Protocol-Version: 2026-07-28' \
    -H "Mcp-Method: $method" \
    -H 'Accept: application/json, text/event-stream' \
    --data "$body"
}

expect_jsonrpc_error() {
  local url="$1" body="$2" expected_code="$3" label="$4" method="${5:-tools/call}"
  local document
  document="$(post_json "$url" "$body" "$method")"
  EXPECTED_CODE="$expected_code" LABEL="$label" python3 -c '
import json, os, sys
document = json.loads(sys.stdin.read())
code = document.get("error", {}).get("code")
expected = int(os.environ["EXPECTED_CODE"])
if code != expected:
    label = os.environ["LABEL"]
    raise SystemExit(f"{label}: expected JSON-RPC {expected}, got {document}")
' <<<"$document"
}

smoke_engine() {
  echo "── engine MCP smoke ($ENGINE_IMAGE) ──"
  local name="liki-smoke-engine-mcp-$$"
  start_container "$name" \
    --entrypoint /usr/local/bin/engine-mcp \
    -p "$ENGINE_MCP_PORT:8081" \
    "$ENGINE_IMAGE"
  wait_url "$name" "http://127.0.0.1:$ENGINE_MCP_PORT/health"
  python3 "$ROOT/scripts/test_engine_mcp.py" "http://127.0.0.1:$ENGINE_MCP_PORT/mcp"
  echo "✓ engine MCP"
}

smoke_counsel() {
  echo "── counsel MCP smoke ($COUNSEL_IMAGE) ──"
  local name="liki-smoke-counsel-$$"
  start_container "$name" \
    -e LIKI_ENGINE_MCP_URL=http://127.0.0.1:9/mcp \
    -p "$COUNSEL_PORT:8086" \
    "$COUNSEL_IMAGE"
  wait_url "$name" "http://127.0.0.1:$COUNSEL_PORT/healthz"
  body='{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{}}}}'
  post_json "http://127.0.0.1:$COUNSEL_PORT/mcp" "$body" server/discover \
    | grep -q 'io.modelcontextprotocol/serverInfo'
  echo "✓ counsel MCP"
}

smoke_liki_mcp() {
  echo "── LIKI MCP smoke ($LIKI_MCP_IMAGE) ──"
  local name="liki-smoke-liki-mcp-$$"
  local url="http://127.0.0.1:$LIKI_MCP_PORT/mcp"
  start_container "$name" \
    -e LIKI_BIND_HOST=0.0.0.0 \
    -e LIKI_UPSTREAM_ENGINE_URL=http://127.0.0.1:9/mcp \
    -e LIKI_UPSTREAM_COUNSEL_URL=http://127.0.0.1:9/mcp \
    -p "$LIKI_MCP_PORT:8087" \
    "$LIKI_MCP_IMAGE"
  wait_url "$name" "http://127.0.0.1:$LIKI_MCP_PORT/healthz"

  # An unreachable upstream must remain a protocol-level, fail-closed error.
  expect_jsonrpc_error "$url" \
    '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
    -32000 'tools/list unavailable upstream' tools/list

  # Product policy is independent of upstream availability.
  expect_jsonrpc_error "$url" \
    '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"qimen_chart","arguments":{}}}' \
    -32003 'raw qimen_chart forbidden' tools/call
  expect_jsonrpc_error "$url" \
    '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"not_a_liki_tool","arguments":{}}}' \
    -32602 'unknown tool rejected' tools/call
  echo "✓ LIKI MCP"
}

if [ "$COMPONENT" = engine ] || [ "$COMPONENT" = all ]; then smoke_engine; fi
if [ "$COMPONENT" = counsel ] || [ "$COMPONENT" = all ]; then smoke_counsel; fi
if [ "$COMPONENT" = liki-mcp ] || [ "$COMPONENT" = all ]; then smoke_liki_mcp; fi

echo "✅ $COMPONENT image smoke passed"
