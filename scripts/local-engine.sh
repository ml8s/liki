#!/usr/bin/env bash
# local-engine.sh —— 本地引擎生命周期（单项目：engine + skills 共用）
#
# 供 make test-all / test-integration / run-qwen.sh 等复用：
#   建立本地引擎，供 skill 工具层（paipan.py，读 LIKI_RPC_URL）与评测连接，
#   完全脱离生产 liki.hk 与 liki-web，本仓独立自测。
#
# 用法：
#   source scripts/local-engine.sh        # 定义 ensure_local_engine / stop_local_engine / LOCAL_RPC
#   ensure_local_engine                   # 起引擎（或复用已在跑的），设置 LOCAL_RPC
#   ... 跑测试 ...
#   stop_local_engine                     # 关掉自己起的引擎（已在跑的保留）
#
# 环境变量：
#   LIKI_RPC_URL      显式指定端点（尊重外部设定，跳过探测/起引擎）
#   LIKI_ENGINE_PORT  引擎端口（默认 8082）
#   LIKI_RPC_MODE     调用方所在位置：local（默认，本机直连 localhost）| docker
#                     （调用方在容器内——skill-up 评测，宿主经 docker bridge 网关访问）
#
# 失败语义（严格）：引擎构建失败 / 启动超时 → ensure_local_engine 返回非 0，
# 调用方（make test-all / run-qwen.sh）应中止——不允许静默跳过造成"假绿"，
# 也不允许回落生产 liki.hk（评测/测试一律脱离生产）。
set -euo pipefail

ENGINE_PORT="${LIKI_ENGINE_PORT:-8082}"

# 节点端点：由调用方位置决定（此前用 pgrep 猜"已运行→localhost / 未运行→网关"，
# 无 docker 的机器上网关 IP 不通，本地集成测试静默跳过——改为显式模式）
if [ -n "${LIKI_RPC_URL:-}" ]; then
  LOCAL_RPC="$LIKI_RPC_URL"
elif [ "${LIKI_RPC_MODE:-local}" = "docker" ]; then
  GATEWAY_IP="$(docker network inspect bridge --format '{{range .IPAM.Config}}{{.Gateway}}{{end}}' 2>/dev/null || true)"
  [ -n "$GATEWAY_IP" ] || GATEWAY_IP="172.17.0.1"
  LOCAL_RPC="http://${GATEWAY_IP}:${ENGINE_PORT}/jsonrpc"
else
  LOCAL_RPC="http://localhost:${ENGINE_PORT}/jsonrpc"
fi
export LOCAL_RPC

# 记录自己起的引擎 PID（仅当不是复用时）
ENGINE_BIN=""

ensure_local_engine() {
  echo "本地引擎: $LOCAL_RPC"
  if curl -sf "http://localhost:${ENGINE_PORT}/health" >/dev/null 2>&1; then
    echo "  ✓ 检测到已运行的本地引擎（:${ENGINE_PORT}），复用"
    return 0
  fi
  echo "  → 启动本地引擎（:${ENGINE_PORT}）..."
  ENGINE_BIN="$(mktemp /tmp/liki-engine.XXXXXX)"
  if ! ( cd engine && go build -o "$ENGINE_BIN" ./cmd/liki/ >/dev/null 2>&1 ); then
    echo "  ❌ 引擎构建失败（go 不可用或编译错误）——中止，不回落生产端点" >&2
    rm -f "$ENGINE_BIN" 2>/dev/null || true
    ENGINE_BIN=""
    return 1
  fi
  "$ENGINE_BIN" -addr ":${ENGINE_PORT}" >/tmp/liki-engine.log 2>&1 &
  for i in $(seq 1 15); do
    curl -sf "http://localhost:${ENGINE_PORT}/health" >/dev/null 2>&1 && { echo "  ✓ 本地引擎就绪"; return 0; }
    sleep 1
  done
  echo "  ❌ 本地引擎启动超时（15s）——中止。引擎日志尾部：" >&2
  tail -5 /tmp/liki-engine.log >&2 2>/dev/null || true
  stop_local_engine
  return 1
}

stop_local_engine() {
  if [ -n "$ENGINE_BIN" ]; then
    kill "$(pgrep -f "^$ENGINE_BIN" 2>/dev/null)" 2>/dev/null || true
    rm -f "$ENGINE_BIN" 2>/dev/null || true
    ENGINE_BIN=""
  fi
}
