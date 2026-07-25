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

rpc bazi.xiaoxian '{"gender":"male"}'
check_rpc_ok "bazi.xiaoxian"

rpc bazi.yongshen "{\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.yongshen"
check_rpc "  has fu_yi" '.result.data.fu_yi.yong != null' 'true'
check_rpc "  has tiao_hou" '.result.data.tiao_hou.yong != null' 'true'
check_rpc "  has ge_ju" '.result.data.ge_ju.yong != null' 'true'

rpc bazi.hehui "{\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.hehui"
check_rpc "  has gan_he key" '.result.data | has("gan_he")' 'true'
check_rpc "  has zhi_liu_he key" '.result.data | has("zhi_liu_he")' 'true'
check_rpc "  has san_he key" '.result.data | has("san_he")' 'true'
check_rpc "  has san_hui key" '.result.data | has("san_hui")' 'true'
check_rpc "  has liu_chong key" '.result.data | has("liu_chong")' 'true'
check_rpc "  has liu_hai key" '.result.data | has("liu_hai")' 'true'
check_rpc "  has liu_xing key" '.result.data | has("liu_xing")' 'true'

rpc bazi.chart_extra "{\"chart\":$BAZI_CORE}"
check_rpc_ok "bazi.chart_extra"
check_rpc "  has san_yuan" '.result.data.san_yuan.tai_yuan != null' 'true'
check_rpc "  has chang_sheng" '.result.data.chang_sheng[0].name != null' 'true'
check_rpc "  has nayin_rel" '.result.data.nayin_rel != null' 'true'

# ============================================================================
# ZiWei
# ============================================================================
echo ""
echo "${BOLD}── ZiWei ──${NC}"

rpc ziwei.chart "$BR"
check_rpc_ok "ziwei.chart"
check_rpc "  has palaces" '.result.data.palaces != null' 'true'

ZW_CHART=$(json_val "$RPC_BODY" '.result.data')

rpc ziwei.daxian "{\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.daxian"

rpc ziwei.liunian "{\"liu_year\":2026,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liunian"

rpc ziwei.liuyue "{\"liu_year\":2026,\"lunar_month\":5,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liuyue"

rpc ziwei.liuri "{\"liu_year\":2026,\"lunar_month\":5,\"lunar_day\":10,\"chart\":$ZW_CHART}"
check_rpc_ok "ziwei.liuri"


# ziwei.judgment -- 综合盘论断
rpc ziwei.judgment '{"chart":null}'
check_rpc_err "ziwei.judgment (null chart)" "-32602"
rpc ziwei.chart "$BR_A"
CHART_A=$(json_val "$RPC_BODY" '.result.data')
rpc ziwei.chart "$BR_B"
CHART_B=$(json_val "$RPC_BODY" '.result.data')
rpc ziwei.bond "{\"a\":$CHART_A,\"b\":$CHART_B}"
check_rpc_ok "ziwei.bond"

rpc ziwei.chart '{"solar_time":""}'
check_rpc_err "ziwei.chart (missing gender)" "-32602"

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

rpc qiming.pick '{"surname":"李","wuxing":"水"}'
check_rpc_ok "qiming.pick"
check_rpc "  has chars" '.result.data.chars != null' 'true'

PICK_CHARS=$(json_val "$RPC_BODY" '.result.data.chars')
rpc qiming.build "{\"surname\":\"李\",\"chars1\":$PICK_CHARS}"
check_rpc_ok "qiming.build"
check_rpc "  has names" '.result.data.names | length > 0' 'true'

BUILD_NAMES=$(json_val "$RPC_BODY" '.result.data.names')
rpc qiming.check "{\"surname\":\"李\",\"names\":$BUILD_NAMES,\"yongshen\":\"水\",\"xishen\":[\"金\"],\"jishen\":[\"土\"]}"
check_rpc_ok "qiming.check"
check_rpc "  has wuxing_match" '.result.data[0].wuxing_match != null' 'true'
check_rpc "  has wuxing.yong" '.result.data[0].wuxing.yong != null' 'true'

rpc qiming.check '{"surname":"李","names":["沐泽"]}'
check_rpc_ok "qiming.check (basic)"
check_rpc "  has wuge" '.result.data[0].wuge != null' 'true'

rpc qiming.check '{}'
check_rpc_err "qiming.check (missing surname)" "-32602"

rpc qiming.wuge '{"surname":"李","count":2}'
check_rpc_ok "qiming.wuge"
check_rpc "  has surname_stroke" '.result.data.surname_stroke > 0' 'true'
# pairs 用 jq -e 检查
check "  has pairs" "true" "$(echo "$RPC_BODY" | jq -e '.result.data.pairs | length > 0' > /dev/null 2>&1 && echo true || echo false)"

rpc qiming.char '{"char":"林"}'
check_rpc_ok "qiming.char (林)"
check_rpc "  has wuxing" '.result.data.wuxing != null' 'true'
check_rpc "  has stroke" '.result.data.stroke > 0' 'true'

# check_rpc "  has pairs" skipped (intermittent)

# ============================================================================
# Bazhai
# ============================================================================
echo ""
echo "${BOLD}── Bazhai ──${NC}"


# bazhai.judgment -- 门主灶论断
rpc bazhai.judgment '{"chart":null,"door_gua":1,"master_gua":2,"stove_gua":3}'
check_rpc_err "bazhai.judgment (null chart)" "-32602"

rpc bazhai.minggua '{"gender":"male","birth_year":1984}'
check_rpc_ok "bazhai.minggua"

rpc bazhai.chart "$BR"
check_rpc_ok "bazhai.chart"

rpc bazhai.minggua '{"gender":"other","birth_year":1984}'
check_rpc_err "bazhai.minggua (bad gender)" "-32602"

rpc bazhai.minggua '{"gender":"male"}'
check_rpc_err "bazhai.minggua (missing year)" "-32602"

# ============================================================================
# XuanKong
# ============================================================================
echo ""
echo "${BOLD}── XuanKong ──${NC}"

rpc xuankong.sanyuan '{"year":2026}'
check_rpc_ok "xuankong.sanyuan"


# xuankong.annual -- 流年飞星
rpc xuankong.annual '{"year":2024}'
check_rpc_ok "xuankong.annual (2024)"
check_rpc "  ru_zhong=3" '.result.data.ru_zhong == 3' "true"
rpc xuankong.annual '{"year":1800}'
check_rpc_err "xuankong.annual (bad year)" "-32000"

rpc xuankong.chart "{\"solar_time\":$ST,\"sit_mountain\":0,\"face_mountain\":11}"
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

rpc huangli.date '{"date":"2026-06-19","event":"嫁娶"}'
check_rpc_ok "huangli.date"

rpc huangli.month '{"month":"2026-06","event":"嫁娶"}'
check_rpc_ok "huangli.month"

HL_BOND='{"solar_time":'"$BT"',"event_type":"嫁娶","date":"2026-06-19"}'
rpc huangli.bond.date "$HL_BOND"
check_rpc_ok "huangli.bond.date"

rpc huangli.bond.month '{"solar_time":'"$BT"',"event_type":"嫁娶","month":"2026-06"}'
check_rpc_ok "huangli.bond.month"

rpc huangli.date '{}'
check_rpc_err "huangli.date (missing params)" "-32602"

HL_BAD_BIRTH='{"solar_time":"","event_type":"嫁娶","date":"2026-06-19"}'
rpc huangli.bond.date "$HL_BAD_BIRTH"
check_rpc_err "huangli.bond.date (empty birth)" "-32000"

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

rpc city '{}'
check_rpc_err "city (missing city)" "-32602"




# ============================================================================
echo ""
echo "${BOLD}────────────────────────────────────────${NC}"

rpc qimen.judgment '{"chart":{"pan":{"ri_gan":"丙","ri_zhi":"子","drive_gan":"壬","drive_zhi":"申"},"patterns":[],"stem_interactions":[],"door_interactions":[],"star_interactions":[],"ying_qi":{"ma_xing":"","kong_wang":"","duty_move":"","summary":""}},"event":"general"}'
check_rpc_ok "qimen.judgment"
check_rpc "  has rating" '.result.data.rating != null' 'true'

rpc qimen.select '{"start_date":"2026-07-20","end_date":"2026-07-21","event":"travel"}'
check_rpc_ok "qimen.select"
check_rpc "  has slots" '.result.data | length > 0' 'true'

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
