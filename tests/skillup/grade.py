#!/usr/bin/env python3
"""Generic behavior judge for Liki domain skill-up smoke cases."""

import json
import os
import re
import sys


RULES = {
    "bazi-chart": {
        "required_all": ["smoke_id: bazi-chart", "四柱", "十神"],
        "rpc_required": ["full_paipan", "query"],
        "required_any": ["大运", "命宫"],
        "forbidden_any": [],
    },
    "bazi-yongshen": {
        "required_all": ["smoke_id: bazi-yongshen", "用神", "喜神", "忌神"],
        "rpc_required": ["full_paipan", "query"],
        "required_any": ["扶抑", "格局", "调候"],
        "forbidden_any": [],
    },
    "bazi-no-birth": {
        "required_all": ["smoke_id: bazi-no-birth", "出生时间"],
        "required_any": ["无法排盘", "不能排盘", "请提供", "需要提供"],
        "forbidden_any": [],
    },
    "bazi-ziwei-fusion": {
        "required_all": ["smoke_id: bazi-ziwei-fusion", "八字", "紫微"],
        "rpc_required": ["full_paipan", "query"],
        "required_any": ["命宫", "大运"],
        "forbidden_any": [],
    },
    "liuyao-career": {
        "required_all": ["smoke_id: liuyao-career", "casting", "用神", "世应"],
        "required_any": ["liuyao_snapshot", "liuyao.qigua", "liuyao_qigua", "liuyao.chart", "liuyao_chart"],
        "forbidden_any": [],
    },
    "liuyao-lost-item": {
        "required_all": ["smoke_id: liuyao-lost-item", "casting", "用神"],
        "required_any": ["liuyao_snapshot", "liuyao.qigua", "liuyao_qigua", "liuyao.chart", "liuyao_chart"],
        "forbidden_any": [],
    },
    "liuyao-health": {
        "required_all": ["smoke_id: liuyao-health", "casting", "官鬼" ],
        "required_any": ["liuyao_snapshot", "liuyao.qigua", "liuyao_qigua", "liuyao.chart", "liuyao_chart"],
        "forbidden_any": [],
    },
    "qimen-direction": {
        "required_all": ["smoke_id: qimen-direction", "盘局", "方位"],
        "rpc_required": ["qimen_snapshot"],
        "required_any": ["用神", "符使"],
        "forbidden_any": [],
    },
    "qimen-negotiation": {
        "required_all": ["smoke_id: qimen-negotiation", "时家", "盘局"],
        "rpc_required": ["qimen_snapshot"],
        "required_any": ["开门", "值符", "用神"],
        "forbidden_any": [],
    },
    "qimen-late-zi": {
        "required_all": ["smoke_id: qimen-late-zi", "晚子时", "日柱"],
        "rpc_required": ["qimen_snapshot"],
        "required_any": ["时家", "盘局"],
        "forbidden_any": [],
    },
    "bazhai-minggua": {
        "required_all": ["smoke_id: bazhai-minggua", "命卦"],
        "rpc_required": ["bazhai.chart"],
        "required_any": ["东四命", "西四命"],
        "forbidden_any": [],
    },
    "bazhai-layout": {
        "required_all": ["smoke_id: bazhai-layout", "门主灶", "伏位"],
        "rpc_required": ["bazhai.layout"],
        "required_any": ["生气", "延年", "天医", "绝命", "五鬼", "祸害", "六煞"],
        "forbidden_any": [],
    },
    "xuankong-chart": {
        "required_all": ["smoke_id: xuankong-chart", "山星", "向星"],
        "rpc_required": ["xuankong.chart"],
        "required_any": ["旺山旺向", "上山下水", "七星打劫", "飞星"],
        "forbidden_any": [],
    },
    "xuankong-liunian": {
        "required_all": ["smoke_id: xuankong-liunian", "2026", "流年"],
        "rpc_required": ["xuankong.chart", "xuankong.liunian"],
        "required_any": ["山星", "向星", "飞星"],
        "forbidden_any": [],
    },
    "fengshui-incomplete": {
        "required_all": ["smoke_id: fengshui-incomplete", "缺少"],
        "required_any": ["坐向", "山向", "出生年份", "性别"],
        "forbidden_any": [],
    },
    "naming-lee": {
        "required_all": ["smoke_id: naming-lee", "李", "romanization_exact", "lǐ"],
        "rpc_required": ["qiming.surname"],
        "required_any": ["baijiaxing_index", "传统《百家姓》"],
        "forbidden_any": [],
    },
    "naming-wong": {
        "required_all": ["smoke_id: naming-wong", "王", "黄", "romanization_exact"],
        "required_any": ["候选", "并列"],
        "rpc_required": ["qiming.surname"],
        "forbidden_any": [],
    },
    "naming-fallback": {
        "required_all": ["smoke_id: naming-fallback", "baijiaxing_fallback", "没有可靠音近姓", "传统《百家姓》顺序"],
        "required_any": ["qiming.surname"],
        "forbidden_any": [],
    },
    "naming-no-birth": {
        "required_all": ["smoke_id: naming-no-birth", "未评估用神", "五行"],
        "rpc_required": ["qiming.pick"],
        "required_any": ["qiming.compose", "候选名"],
        "forbidden_any": [],
    },
    "naming-selfcheck": {
        "required_all": ["smoke_id: naming-selfcheck", "字库", "拼音"],
        "rpc_required": ["qiming.check"],
        "required_any": ["五行", "未评估用神"],
        "forbidden_any": [],
    },
    "naming-non-latin": {
        "required_all": ["smoke_id: naming-non-latin", "罗马字", "官方或惯用"],
        "required_any": ["不自行转写", "无法直接转写", "请提供"],
        "forbidden_any": [],
    },
}


def read_transcript(path):
    if not path or not os.path.exists(path):
        return []
    try:
        with open(path, encoding="utf-8") as fh:
            data = json.load(fh)
        return data if isinstance(data, list) else []
    except Exception:
        return []


def extract_rpc_receipt(text):
    """Collect the RPC_CALLS field and its continuation list.

    qwen_code may render the receipt as a plain block, a Markdown list, or a
    fenced block. Required RPC names can therefore appear on continuation
    lines instead of the RPC_CALLS line itself.
    """
    lines = text.splitlines()
    starts = [
        index for index, line in enumerate(lines)
        if re.match(r"^\s*(?:[-*]\s*)?\**SMOKE_ID\**\s*[:：]", line, re.IGNORECASE)
    ]
    for start in reversed(starts):
        rpc_index = None
        for index, line in enumerate(lines[start:], start):
            if "rpc_calls" not in line.casefold():
                continue
            rpc_index = index
            break
        if rpc_index is None:
            continue
        block = [line]
        for following in lines[rpc_index + 1:]:
            stripped = following.strip()
            if not stripped:
                break
            if re.match(r"^(?:-\s*)?\**(?:EVIDENCE|BOUNDARY|SMOKE_ID)\**\s*[:：]", stripped, re.IGNORECASE):
                break
            block.append(following)
        return "\n".join(block)
    return ""


def normalize_contract(text):
    text = text.casefold().replace("**", "").replace("`", "")
    return re.sub(r"\s*:\s*", ":", text)


def main():
    final = os.environ.get("EVAL_FINAL_MESSAGE", "")
    transcript = read_transcript(os.environ.get("EVAL_TRANSCRIPT_PATH", ""))
    prompt = ""
    for message in transcript:
        if message.get("role") == "user":
            prompt = str(message.get("content") or "")
            break
    for message in transcript:
        if message.get("role") == "assistant":
            final += "\n" + str(message.get("content") or "")

    match = re.search(r"^SMOKE_ID:\s*([a-z0-9-]+)\s*$", prompt, re.MULTILINE)
    if not match:
        print("ERROR: prompt does not contain SMOKE_ID")
        return 2
    case_id = match.group(1)
    rule = RULES.get(case_id)
    if not rule:
        print(f"ERROR: no smoke rule for {case_id}")
        return 2

    text = normalize_contract(final)
    missing = [
        token for token in rule.get("required_all", [])
        if normalize_contract(token) not in text
    ]
    any_terms = rule.get("required_any", [])
    if any_terms and not any(normalize_contract(token) in text for token in any_terms):
        missing.append("one of: " + " / ".join(any_terms))
    hits = [token for token in rule.get("forbidden_any", []) if normalize_contract(token) in text]

    rpc_line = extract_rpc_receipt(final)
    for rpc in rule.get("rpc_required", []):
        if rpc.casefold() not in rpc_line.casefold():
            missing.append(f"RPC_CALLS missing {rpc}")
    if rule.get("rpc_none"):
        receipt = rpc_line.casefold()
        if "rpc_calls" not in receipt or "none" not in receipt:
            missing.append("RPC_CALLS must be none")

    if missing or hits:
        print(f"FAIL {case_id}")
        if missing:
            print("MISSING:")
            for item in missing:
                print(f"- {item}")
        if hits:
            print("FORBIDDEN:")
            for item in hits:
                print(f"- {item}")
        return 1
    print(f"PASS {case_id}: behavior contract satisfied")
    return 0


if __name__ == "__main__":
    sys.exit(main())
