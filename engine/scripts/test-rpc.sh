#!/bin/bash
# Liki — JSON-RPC engine smoke test
# Usage: scripts/test-rpc.sh [rpc_url]
# Default rpc_url: http://localhost:8080/jsonrpc
set -uo pipefail

RPC_URL="${1:-http://localhost:8080/jsonrpc}"
if [[ "$RPC_URL" != */jsonrpc ]]; then
  RPC_URL="$RPC_URL/jsonrpc"
fi
PASS=0; FAIL=0

RED=""; GREEN=""; BOLD=""; NC=""
[ -t 1 ] && { RED='\033[31m'; GREEN='\033[32m'; BOLD='\033[1m'; NC='\033[0m'; }

TMP=$(mktemp -d /tmp/test-rpc-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

# ── helpers ──────────────────────────────────────────────────────

body() { cat "$TMP/body" 2>/dev/null || true; }

json_val() { echo "$1" | jq -r "$2" 2>/dev/null || true; }

check() {
  local desc="$1" expected="$2" actual="${3:-}"
  if [ "$actual" = "$expected" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}\xe2\x9c\x97${NC} $desc (expected '$expected', got '$actual')"
    FAIL=$((FAIL + 1))
  fi
}

# ── RPC helpers ───────────────────────────────────────────────────

RPC_ID=0
RPC_BODY=""

# rpc <method> <params_json>
rpc() {
  local method="$1"
  local params='{}'
  [ $# -ge 2 ] && params="$2"
  RPC_ID=$((RPC_ID + 1))
  local payload
  payload=$(jq -nc --arg m "$method" --argjson p "$params" --argjson id "$RPC_ID" \
    '{jsonrpc:"2.0", id:$id, method:$m, params:$p}')
  local f="$TMP/rpc_body"
  curl -s -w '%{http_code}' -o "$f" -X POST "$RPC_URL" \
    -H 'Content-Type: application/json' \
    -d "$payload" || echo "000"
  RPC_BODY=$(cat "$f")
  cp "$f" "$TMP/body"
}

check_rpc() {
  local desc="$1" filter="$2" expected="$3"
  check "$desc" "$expected" "$(json_val "$RPC_BODY" "$filter")"
}

check_rpc_ok() {
  local has_err
  has_err=$(json_val "$RPC_BODY" '.error != null')
  if [ "$has_err" = "false" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $1"
    PASS=$((PASS + 1))
  else
    local emsg
    emsg=$(json_val "$RPC_BODY" '.error.message')
    echo -e "  ${RED}\xe2\x9c\x97${NC} $1 (RPC error: $emsg)"
    FAIL=$((FAIL + 1))
  fi
}

check_rpc_err() {
  local desc="$1" expected_code="$2"
  local has_err actual_code
  has_err=$(json_val "$RPC_BODY" '.error != null')
  actual_code=$(json_val "$RPC_BODY" '.error.code')
  if [ "$has_err" = "true" ] && [ "$actual_code" = "$expected_code" ]; then
    echo -e "  ${GREEN}\xe2\x9c\x93${NC} $desc"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}\xe2\x9c\x97${NC} $desc (expected error $expected_code, has_err=$has_err code=$actual_code)"
    FAIL=$((FAIL + 1))
  fi
}

# ── test data ─────────────────────────────────────────────────────

BT='"1984-02-04T18:30:00+08:00"'
BT_A='"1990-03-20T10:30:00+08:00"'
BT_B='"1992-07-08T14:30:00+08:00"'
ST=$BT
ST_A=$BT_A
ST_B=$BT_B
BR="{\"solar_time\":$ST,\"gender\":\"male\"}"
BR_A="{\"solar_time\":$ST_A,\"gender\":\"male\"}"
BR_B="{\"solar_time\":$ST_B,\"gender\":\"female\"}"

# ============================================================================
echo "${BOLD}Liki — JSON-RPC Engine Smoke Test${NC}"
echo "Target: $RPC_URL"
command -v jq &>/dev/null || { echo "jq is required"; exit 1; }
echo ""

# ============================================================================
# RPC Protocol
# ============================================================================
echo "${BOLD}── RPC Protocol ──${NC}"

rpc '""' '{}'
check_rpc_err "rpc: invalid method" "-32601"

RPC_ID=$((RPC_ID + 1))
curl -s -o "$TMP/body" -X POST "$RPC_URL" -H 'Content-Type: application/json' \
  -d '{"id":1,"method":"bazi.chart","params":{}}'
RPC_BODY=$(body)
check_rpc_err "rpc: missing jsonrpc error" "-32600"

rpc rpc.discover '{}'
check_rpc_ok "rpc.discover"
check_rpc "  openrpc version" '.result.openrpc' '1.4.1'

# ============================================================================
# BaZi
# ============================================================================
echo ""
echo "${BOLD}── BaZi ──${NC}"

rpc bazi.chart "$BR"
check_rpc_ok "bazi.chart"
check_rpc "  has nian" '.result.data.nian.gan != null' 'true'
check_rpc "  has da_yun" '.result.data.da_yun != null' 'true'
check_rpc "  has gender" '.result.data.gender != null' 'true'

BAZI_CORE=$(json_val "$RPC_BODY" '.result.data')

rpc bazi.chart '{"solar_time":""}'
check_rpc_err "bazi.chart (missing gender)" "-32602"


# bazi.fullchart -- 扩展命盘
rpc bazi.fullchart '{"chart":null}'
check_rpc_err "bazi.fullchart (null chart)" "-32602"

rpc bazi.chart "$BR_A"
BAZI_CORE_A=$(json_val "$RPC_BODY" .result.data)
rpc bazi.chart "$BR_B"
BAZI_CORE_B=$(json_val "$RPC_BODY" .result.data)
rpc bazi.bond "{\"a\":{\"chart\":$BAZI_CORE_A},\"b\":{\"chart\":$BAZI_CORE_B}}"
check_rpc_ok "bazi.bond"
check_rpc "  has zhu_cross" '.result.data.zhu_cross != null' 'true'

rpc bazi.liunian "{\"year\":2026,\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.liunian"

rpc bazi.liuyue "{\"year\":2026,\"month\":6,\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.liuyue"

rpc bazi.liuri "{\"year\":2026,\"month\":6,\"day\":15,\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.liuri"

rpc bazi.liushi "{\"year\":2026,\"month\":6,\"day\":15,\"hour\":12,\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.liushi"

rpc bazi.xiaoyun "{\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.xiaoyun"

# bazi.chart 纯排盘（2.6.14 起用神三派归 fullchart——chart 不含 yong_shen）
rpc bazi.chart "$BR"
check_rpc_ok "bazi.chart"
check_rpc "  不含 yong_shen" '.result.data | has("yong_shen")' 'false'

# 合会冲刑 + 三元/长生/纳音 在 bazi.fullchart（原 hehui/chart_extra 已并入）
rpc bazi.fullchart "{\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.fullchart"
check_rpc "  has gan_he key" '.result.data | has("gan_he")' 'true'
check_rpc "  has liu_chong key" '.result.data | has("liu_chong")' 'true'
check_rpc "  has san_yuan" '.result.data.san_yuan.tai_yuan != null' 'true'
check_rpc "  has chang_sheng" '.result.data.chang_sheng[0].name != null' 'true'
check_rpc "  has nayin_rel" '.result.data.nayin_rel != null' 'true'
check_rpc "  has yong_shen" '.result.data.yong_shen.fu_yi.yong != null' 'true'

# ============================================================================
# ZiWei
# ============================================================================
echo ""
echo "${BOLD}── ZiWei ──${NC}"

# ziwei.chart 参数为 lunar{year,month,day,shichen}+gender——先经 tianwen.time 取 lunar
rpc tianwen.time '{"time":'"$BT"',"longitude":116.4}'
ZW_LUNAR=$(json_val "$RPC_BODY" '.result.data.lunar | {year,month,day,shichen}')
rpc ziwei.chart "{\"lunar\":$ZW_LUNAR,\"gender\":\"male\"}"
check_rpc_ok "ziwei.chart"
check_rpc "  has gong_wei" '.result.data.gong_wei != null' 'true'

ZW_CHART=$(json_val "$RPC_BODY" '.result.data')

rpc ziwei.daxian "{\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.daxian"

rpc ziwei.liunian "{\"lunar_year\":2026,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liunian"

rpc ziwei.liuyue "{\"lunar_year\":2026,\"lunar_month\":5,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liuyue"

rpc ziwei.liuri "{\"lunar_year\":2026,\"lunar_month\":5,\"lunar_day\":10,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liuri"

rpc ziwei.liushi "{\"lunar_year\":2026,\"lunar_month\":5,\"lunar_day\":10,\"shi_zhi\":\"午\",\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liushi"

rpc ziwei.fullchart "{\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.fullchart"


rpc ziwei.chart "{\"lunar\":$ZW_LUNAR,\"gender\":\"male\"}"
CHART_A=$(json_val "$RPC_BODY" '.result.data')
rpc ziwei.chart "{\"lunar\":$ZW_LUNAR,\"gender\":\"female\"}"
CHART_B=$(json_val "$RPC_BODY" '.result.data')
rpc ziwei.bond "{\"a\":$CHART_A,\"b\":$CHART_B}"
check_rpc_ok "ziwei.bond"

rpc ziwei.chart '{"lunar":{}}'
check_rpc_err "ziwei.chart (bad lunar)" "-32602"

# ============================================================================
# QiMen
# ============================================================================
echo ""
echo "${BOLD}── QiMen ──${NC}"

rpc qimen.chart "{\"solar_time\":$ST,\"kind\":\"shi\"}"
check_rpc_ok "qimen.chart (shi)"

rpc qimen.chart "{\"solar_time\":$ST,\"kind\":\"invalid\"}"
check_rpc_err "qimen.chart (bad kind)" "-32602"

# ============================================================================
# QiMing
# ============================================================================
echo ""
echo "${BOLD}── QiMing ──${NC}"

rpc qiming.pick '{"wuxing1":"水","wuxing2":"金","count":2}'
check_rpc_ok "qiming.pick"
check_rpc "  has pools" '.result.data.pools != null' 'true'

PICK_FIRST=$(json_val "$RPC_BODY" '.result.data.pools[] | select(.slot == "first") | .chars[0:3]')
PICK_SECOND=$(json_val "$RPC_BODY" '.result.data.pools[] | select(.slot == "second") | .chars[0:3]')
rpc qiming.compose "{\"first\":$PICK_FIRST,\"second\":$PICK_SECOND,\"max_names\":10}"
check_rpc_ok "qiming.compose"
check_rpc "  has names" '.result.data.names | length > 0' 'true'

BUILD_NAMES=$(json_val "$RPC_BODY" '.result.data.names[0:5]')
rpc qiming.check "{\"given_names\":$BUILD_NAMES,\"yongshen\":\"水\",\"xishen\":[\"金\"],\"jishen\":[\"土\"]}"
check_rpc_ok "qiming.check"
check_rpc "  has wuxing.yong" '.result.data[0].wuxing.yong != null' 'true'

rpc qiming.check '{"given_names":["沐泽"]}'
check_rpc_ok "qiming.check (basic)"
check_rpc "  has valid" '.result.data[0].valid != null' 'true'

rpc qiming.check '{}'
check_rpc_err "qiming.check (missing given_names)" "-32602"

rpc qiming.char '{"char":"林"}'
check_rpc_ok "qiming.char (林)"
check_rpc "  has wuxing" '.result.data.wuxing != null' 'true'
check_rpc "  has stroke" '.result.data.stroke > 0' 'true'

# ============================================================================
# Bazhai
# ============================================================================
echo ""
echo "${BOLD}── Bazhai ──${NC}"


rpc bazhai.chart "$BR"
check_rpc_ok "bazhai.chart"
check_rpc "  has ming_gua" '.result.data.ming_gua != null' 'true'
check_rpc "  has ba_zhai_dirs" '.result.data.ba_zhai_dirs != null' 'true'

BH_CHART=$(json_val "$RPC_BODY" '.result.data')
rpc bazhai.layout "{\"chart\":$BH_CHART,\"door_gua\":\"坎\",\"master_gua\":\"震\",\"stove_gua\":\"离\"}"
check_rpc_ok "bazhai.layout"
rpc bazhai.layout "{\"chart\":$BH_CHART}"
check_rpc_err "bazhai.layout (missing gua)" "-32602"

# ============================================================================
# XuanKong
# ============================================================================
echo ""
echo "${BOLD}── XuanKong ──${NC}"

rpc xuankong.liunian '{"year":2024}'
check_rpc_ok "xuankong.liunian (2024)"
rpc xuankong.liunian '{"year":1800}'
check_rpc_err "xuankong.liunian (bad year)" "-32000"

rpc xuankong.chart "{\"solar_time\":$ST,\"zuo_shan\":0,\"xiang_shan\":11}"
check_rpc_ok "xuankong.chart"

rpc xuankong.chart "{\"solar_time\":$ST}"
check_rpc_err "xuankong.chart (missing mountains)" "-32602"

# ============================================================================
# LiuYao
# ============================================================================
echo ""
echo "${BOLD}── LiuYao ──${NC}"

rpc liuyao.qigua '{}'
check_rpc_ok "liuyao.qigua"

YAOS=$(json_val "$RPC_BODY" '.result.data.yaos')
YAOS_JSON=$(echo "$YAOS" | jq -c '.')

rpc liuyao.chart "{\"solar_time\":$ST,\"yaos\":$YAOS_JSON}"
check_rpc_ok "liuyao.chart"

rpc liuyao.chart "{\"solar_time\":$ST,\"yaos\":[6,7,8,9,6,7],\"yong_shen\":\"妻财\"}"
check_rpc_ok "liuyao.chart (with yong_shen)"

rpc liuyao.chart "{\"solar_time\":$ST,\"yaos\":[7,8,7,6,9,7]}"
check_rpc_ok "liuyao.chart (mixed yaos)"

# ============================================================================
# Huangli
# ============================================================================
echo ""
echo "${BOLD}── Huangli ──${NC}"

rpc huangli.days '{"start_date":"2026-06-19","count":3}'
check_rpc_ok "huangli.days"
check_rpc "  has jian_chu" '.result.data[0].jian_chu != null' 'true'
check_rpc "  has huangdao" '.result.data[0].huangdao.name != null' 'true'

rpc huangli.days '{}'
check_rpc_err "huangli.days (missing start_date)" "-32602"

# ============================================================================
# Infra
# ============================================================================
echo ""
echo "${BOLD}── Infra ──${NC}"

rpc time.now '{}'
check_rpc_ok "time.now"
b=$(body)
check "  has utc" "false" "$(json_val "$b" '.result.data.utc' | grep -q . && echo false || echo true)"
check "  has cst" "false" "$(json_val "$b" '.result.data.cst' | grep -q . && echo false || echo true)"

rpc tianwen.time '{"time":"2000-06-15T12:00:00+08:00","longitude":116.4}'
check_rpc_ok "tianwen.time"
check_rpc "  has solar" '.result.data.solar != null' 'true'
check_rpc "  has gregorian" '.result.data.gregorian != null' 'true'
check_rpc "  has lunar" '.result.data.lunar != null' 'true'

rpc city.coords '{}'
check_rpc_err "city.coords (missing city)" "-32602"




# ============================================================================
echo ""
echo "${BOLD}────────────────────────────────────────${NC}"


TOTAL=$((PASS + FAIL))
echo "Total: $TOTAL  ${GREEN}PASS: $PASS${NC}  ${RED}FAIL: $FAIL${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo "${RED}Some tests failed.${NC}"
  exit 1
else
  echo "${GREEN}All RPC tests passed.${NC}"
  exit 0
fi
