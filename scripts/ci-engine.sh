#!/usr/bin/env bash
set -eo pipefail

# ci-engine.sh — 全量测试（等价于 CI 所有 job 之和）
# 用法: make test-all 或 scripts/ci-engine.sh

cd "$(dirname "$0")/.."

# ── Phases ────────────────────────────────────────────────────────
# Phase 1: 静态检查 + 单元测试（无需引擎）
# Phase 2: 集成测试（需引擎）

PASS=0; FAIL=0
GREEN='\033[32m'; RED='\033[31m'; BOLD='\033[1m'; NC='\033[0m'

step_ok() { PASS=$((PASS+1)); echo -e "  ${GREEN}✓${NC} $1"; }
step_fail() { FAIL=$((FAIL+1)); echo -e "  ${RED}✗${NC} $1"; }

# ==================================================================
# Phase 1: 静态检查 + 单元测试
# ==================================================================
echo "${BOLD}[Phase 1/2] 静态检查 + 单元测试${NC}"

echo "--- golangci-lint ---"
if golangci-lint run ./... 2>&1; then
  step_ok "golangci-lint"
else
  step_fail "golangci-lint"
  exit 1
fi

echo "--- go vet ---"
if go vet ./... 2>&1; then
  step_ok "go vet"
else
  step_fail "go vet"
  exit 1
fi

echo "--- go test (race + short) ---"
if go test -race -count=1 -short ./... 2>&1; then
  step_ok "go test -race -short"
else
  step_fail "go test -race -short"
  exit 1
fi

# ==================================================================
# Phase 2: 集成测试 + RPC 冒烟（需要引擎）
# ==================================================================
echo ""
echo "${BOLD}[Phase 2/2] 集成测试 + RPC 冒烟${NC}"

echo "--- 集成测试 ---"
if go test -tags integration -count=1 -timeout 60s ./internal/agent/ ./internal/http/ ./internal/engine/bazi/ 2>&1; then
  step_ok "集成测试"
else
  step_fail "集成测试"
  exit 1
fi

# 编译引擎
echo "--- 编译引擎 ---"
go build -o /tmp/liki-engine ./cmd/liki/
step_ok "编译引擎"

# 清理旧引擎进程
pkill -f "liki-engine -addr :8082" 2>/dev/null || true
sleep 1

# 启动引擎
/tmp/liki-engine -addr :8082 &
ENGINE_PID=$!

cleanup() {
  kill $ENGINE_PID 2>/dev/null || true
  wait $ENGINE_PID 2>/dev/null || true
  rm -f /tmp/liki-engine
}
trap cleanup EXIT

echo -n "--- 等待引擎就绪"
for i in $(seq 1 15); do
  if curl -sf -o /dev/null http://localhost:8082/health 2>/dev/null; then
    echo " ✓"
    break
  fi
  echo -n .
  sleep 1
  if [ $i -eq 15 ]; then
    echo " ✗ 引擎启动超时"
    exit 1
  fi
done

echo "--- RPC 冒烟测试 ---"
if scripts/test-rpc.sh http://localhost:8082; then
  step_ok "RPC 冒烟测试"
else
  step_fail "RPC 冒烟测试"
  exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✓ 全量测试通过  (lint + vet + unit + integration + smoke)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
