#!/usr/bin/env bash
# Smoke the locally built runtime images. These checks intentionally exercise
# image entrypoints and protocols, not rule-engine behavior covered by tests.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE_TAG="${IMAGE_TAG:-dev}"
ENGINE_IMAGE="${ENGINE_IMAGE:-ghcr.io/ml8s/liki-engine:$IMAGE_TAG}"
COUNSEL_IMAGE="${COUNSEL_IMAGE:-ghcr.io/ml8s/liki-counsel:$IMAGE_TAG}"
ENGINE_MCP_PORT="${ENGINE_MCP_SMOKE_PORT:-18081}"
ENGINE_RPC_PORT="${ENGINE_RPC_SMOKE_PORT:-18083}"
COUNSEL_PORT="${COUNSEL_SMOKE_PORT:-18086}"

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

echo "── engine MCP smoke ($ENGINE_IMAGE) ──"
name="liki-smoke-engine-mcp-$$"
start_container "$name" \
  --entrypoint /usr/local/bin/engine-mcp \
  -p "$ENGINE_MCP_PORT:8081" \
  "$ENGINE_IMAGE"
wait_url "$name" "http://127.0.0.1:$ENGINE_MCP_PORT/health"
python3 "$ROOT/scripts/test_engine_mcp.py" "http://127.0.0.1:$ENGINE_MCP_PORT/mcp"
echo "✓ engine MCP"

echo "── engine RPC smoke ($ENGINE_IMAGE) ──"
name="liki-smoke-engine-rpc-$$"
start_container "$name" \
  --entrypoint /usr/local/bin/engine-rpc \
  -p "$ENGINE_RPC_PORT:8080" \
  "$ENGINE_IMAGE"
wait_url "$name" "http://127.0.0.1:$ENGINE_RPC_PORT/health"
curl -fsS -X POST "http://127.0.0.1:$ENGINE_RPC_PORT/jsonrpc" \
  -H 'Content-Type: application/json' \
  --data '{"jsonrpc":"2.0","id":1,"method":"rpc.discover","params":{}}' \
  | grep -q 'openrpc'
echo "✓ engine RPC"

echo "── counsel MCP smoke ($COUNSEL_IMAGE) ──"
name="liki-smoke-counsel-$$"
start_container "$name" -p "$COUNSEL_PORT:8086" "$COUNSEL_IMAGE"
wait_url "$name" "http://127.0.0.1:$COUNSEL_PORT/healthz"
body='{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{}}}}'
curl -fsS -X POST "http://127.0.0.1:$COUNSEL_PORT/mcp" \
  -H 'Content-Type: application/json' \
  -H 'MCP-Protocol-Version: 2026-07-28' \
  -H 'Mcp-Method: server/discover' \
  -H 'Accept: application/json, text/event-stream' \
  --data "$body" | grep -q 'io.modelcontextprotocol/serverInfo'
echo "✓ counsel MCP"

echo "✅ liki image smoke passed"
