#!/usr/bin/env bash
set -eo pipefail

# pre-check.sh — 推送前本地 CI 检查
# 用法: make pre-check 或 scripts/pre-check.sh

cd "$(dirname "$0")/.."

PASS=0; FAIL=0
GREEN='\033[32m'; RED='\033[31m'; BOLD='\033[1m'; NC='\033[0m'

step_ok() { PASS=$((PASS+1)); echo -e "  ${GREEN}✓${NC} $1"; }
step_fail() { FAIL=$((FAIL+1)); echo -e "  ${RED}✗${NC} $1"; }

echo "${BOLD}[推送前检查]${NC}"

# ── 1. 静态检查 ──────────────────────────────────────────────
echo "--- golangci-lint ---"
if golangci-lint run ./... 2>&1; then
  step_ok "golangci-lint"
else
  step_fail "golangci-lint"
fi

echo "--- go vet ---"
if go vet ./... 2>&1; then
  step_ok "go vet"
else
  step_fail "go vet"
fi

# ── 2. 单元测试 ──────────────────────────────────────────────
echo "--- 单元测试 ---"
if go test ./internal/... -count=1 2>&1; then
  step_ok "单元测试"
else
  step_fail "单元测试"
fi

# ── 3. 方法名变更检查 ────────────────────────────────────────
echo "--- 方法名变更检查 ---"
# 检查 RPC 注册表与 test-rpc.sh 是否同步
REGISTERED=$(grep -oP 'Name:\s*"[^"]*"' internal/agent/*_methods*.go internal/agent/*_other*.go internal/agent/*_bazi*.go internal/agent/*_ziwei*.go internal/agent/*_qiming*.go 2>/dev/null | grep -oP '"[^"]*"' | tr -d '"' | sort -u)
TESTED=$(grep -oP 'rpc\s+[a-z]+\.[a-z]+' scripts/test-rpc.sh 2>/dev/null | awk '{print $2}' | sort -u)

# 检查 test-rpc.sh 中是否有未注册的方法（排除 rpc.discover，它是框架内置的）
UNREGISTERED=$(comm -23 <(echo "$TESTED" | grep -v "^rpc.discover$") <(echo "$REGISTERED"))
if [ -n "$UNREGISTERED" ]; then
  step_fail "test-rpc.sh 引用了未注册的方法: $UNREGISTERED"
else
  step_ok "方法名同步（test-rpc.sh ↔ 注册表）"
fi

# 检查方法计数是否与测试文件同步
METHOD_COUNT=$(echo "$REGISTERED" | wc -l)
# 搜索测试文件中的期望计数
TEST_COUNT=$(grep -oP 'expected \d+ methods|want \d+ \(|got \d+ methods' internal/agent/rpc_registry_test.go internal/http/rpc_test.go 2>/dev/null | grep -oP '\d+' | head -1)
if [ -n "$TEST_COUNT" ] && [ "$METHOD_COUNT" != "$TEST_COUNT" ]; then
  step_fail "方法计数不匹配: 注册表=$METHOD_COUNT, 测试期望=$TEST_COUNT"
else
  step_ok "方法计数同步（$METHOD_COUNT 个方法）"
fi

# ── 4. RPC 冒烟测试（可选，需引擎运行）─────────────────────
if curl -sf -o /dev/null http://localhost:8082/health 2>/dev/null; then
  echo "--- RPC 冒烟测试 ---"
  if scripts/test-rpc.sh http://localhost:8082 2>&1; then
    step_ok "RPC 冒烟测试"
  else
    step_fail "RPC 冒烟测试"
  fi
else
  echo "  ⏭ 跳过 RPC 冒烟测试（引擎未运行）"
fi

# ── 汇总 ─────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $FAIL -gt 0 ]; then
  echo -e "${RED}✗ 检查失败 ($FAIL/$((PASS+FAIL)))${NC}"
  exit 1
else
  echo -e "${GREEN}✓ 全部通过 ($PASS/$((PASS+FAIL)))${NC}"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
