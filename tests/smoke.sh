#!/bin/bash
# liki skill smoke tests
set -eo pipefail
PASS=0 FAIL=0
API="https://liki.hk/jsonrpc"

ok()   { PASS=$((PASS+1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ❌ $1"; }

# === API smoke tests ===

echo "=== API 可用性 ==="

# 1. time.now
curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"time.now","params":{},"id":1}' > /dev/null && ok "time.now" || fail "time.now"

# 2. tianwen.time
curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"tianwen.time","params":{"time":"1990-05-20T12:00:00+08:00","longitude":116.4},"id":1}' > /dev/null && ok "tianwen.time" || fail "tianwen.time"

# 3. bazi.chart (minimal)
CHART=$(curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"bazi.chart","params":{"solar_time":"1990-05-20T12:00:00+08:00","gender":"male"},"id":1}')
echo "$CHART" | python3 -c "import sys,json; d=json.load(sys.stdin)['result']['data']; assert 'nian' in d and 'yue' in d and 'ri' in d and 'shi' in d and 'da_yun' in d" && ok "bazi.chart" || fail "bazi.chart"

# 4. bazi.fullchart
CHART_DATA=$(echo "$CHART" | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin)['result']['data'],ensure_ascii=False,separators=(',',':')))")
curl -sf "$API" -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"bazi.fullchart\",\"params\":{\"chart\":$CHART_DATA},\"id\":1}" > /dev/null && ok "bazi.fullchart" || fail "bazi.fullchart"

# 5. bazi.yongshen
curl -sf "$API" -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"bazi.yongshen\",\"params\":{\"chart\":$CHART_DATA},\"id\":1}" > /dev/null && ok "bazi.yongshen" || fail "bazi.yongshen"

# 6. bazi.hehui
curl -sf "$API" -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"bazi.hehui\",\"params\":{\"chart\":$CHART_DATA},\"id\":1}" > /dev/null && ok "bazi.hehui" || fail "bazi.hehui"

# 7. ziwei.chart
curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"ziwei.chart","params":{"solar_time":"1990-05-20T12:00:00+08:00","gender":"male"},"id":1}' > /dev/null && ok "ziwei.chart" || fail "ziwei.chart"

# 8. ziwei.judgment
ZW=$(curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"ziwei.chart","params":{"solar_time":"1990-05-20T12:00:00+08:00","gender":"male"},"id":1}')
ZW_DATA=$(echo "$ZW" | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin)['result']['data'],ensure_ascii=False,separators=(',',':')))")
curl -sf "$API" -H 'Content-Type: application/json' -d "{\"jsonrpc\":\"2.0\",\"method\":\"ziwei.judgment\",\"params\":{\"chart\":$ZW_DATA},\"id\":1}" > /dev/null && ok "ziwei.judgment" || fail "ziwei.judgment"

# 9. liuyao.qigua
curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"liuyao.qigua","params":{"solar":"2026-07-24T10:00:00+08:00"},"id":1}' > /dev/null && ok "liuyao.qigua" || fail "liuyao.qigua"

# 10. qiming.pick
curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"qiming.pick","params":{"wuxing":"木"},"id":1}' > /dev/null && ok "qiming.pick" || fail "qiming.pick"

# 11. rpc.discover
curl -sf "$API" -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","method":"rpc.discover","params":{},"id":1}' > /dev/null && ok "rpc.discover" || fail "rpc.discover"

# === Knowledge file integrity ===

echo ""
echo "=== 知识文件完整性 ==="

KNOWLEDGE_FILES=(
  "bazi/knowledge/yongshen.md"
  "bazi/knowledge/geju.md"
  "bazi/knowledge/tiaohou.md"
  "bazi/knowledge/wangshuai.md"
  "bazi/knowledge/hehui.md"
  "bazi/knowledge/gongwei.md"
  "bazi/knowledge/dayun.md"
  "bazi/knowledge/shishen.md"
  "bazi/knowledge/hepan.md"
  "ziwei/knowledge/cankao.md"
)

for f in "${KNOWLEDGE_FILES[@]}"; do
  [[ -f "$f" ]] && ok "$f" || fail "$f"
done

# === SKILL file existence ===

echo ""
echo "=== Skill 文件完整性 ==="

SKILL_FILES=(
  "SKILL.md"
  "bazi/SKILL.md"
  "bazi/format-chart.md"
  "bazi/format-hepan.md"
  "ziwei/SKILL.md"
  "ziwei/format.md"
  "naming/SKILL.md"
  "naming/format.md"
  "liuyao/SKILL.md"
  "liuyao/format.md"
  "qimen/SKILL.md"
  "qimen/format.md"
  "huangli/SKILL.md"
  "huangli/format.md"
  "bazhai/SKILL.md"
  "bazhai/format.md"
  "xuankong/SKILL.md"
  "xuankong/format.md"
  "reports/mingshu/SKILL.md"
  "reports/mingshu/generate.md"
  "reports/mingshu/review.md"
  "reports/mingshu/revise.md"
  "reports/hepan/SKILL.md"
  "reports/hepan/generate.md"
  "reports/hepan/review.md"
  "reports/hepan/revise.md"
)

for f in "${SKILL_FILES[@]}"; do
  [[ -f "$f" ]] && ok "$f" || fail "$f"
done

# === Summary ===

echo ""
echo "=== 结果 ==="
echo "  通过: $PASS"
echo "  失败: $FAIL"
[[ $FAIL -eq 0 ]] && echo "  ✅ ALL PASS" || echo "  ❌ SOME FAILED"
exit $FAIL
