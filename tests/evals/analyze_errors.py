#!/usr/bin/env python3
"""错题分析：对 grade_iter3.txt 中每道错题，提取 agent 推理段落 + 断语引用，输出根因分类线索。"""
import json, re, os, sys

BASE = os.path.dirname(os.path.abspath(__file__))           # skills/liki/tests/evals
SKILL = os.path.abspath(os.path.join(BASE, "..", ".."))     # skills/liki
TESTS = os.path.join(SKILL, "tests")
WS = os.path.join(SKILL, "..", "liki-workspace", "iteration-3")
GROUPS = json.load(open(os.path.join(TESTS, "groups.json")))
INV = {}
for gid, ftbs in GROUPS.items():
    for i, ftb in enumerate(ftbs, 1):
        INV[ftb] = (gid, i)

# 解析判分输出
errors = []  # (ftb, gid, qnum, cat, pred, truth)
for line in open(os.path.join(BASE, "grade_iter3.txt")):
    m = re.match(r"\s*(ftb_\d+)\s+\[题(\d+)\]\s+\[([^\]]+)\]\s+pred=([A-D])\s+truth=([A-D])", line)
    if m:
        ftb, q, cat, pred, truth = m.groups()
        errors.append((ftb, int(q), cat, pred, truth))

print(f"错题总数: {len(errors)}\n")

def extract_reason(gid, qnum):
    """从 response.md 提取该题推理段（含断语引用统计）。"""
    rsp = os.path.join(WS, gid, "with_skill", "outputs", "response.md")
    if not os.path.exists(rsp):
        return "", [], ""
    txt = open(rsp, encoding="utf-8").read()
    # 找题 N 的段落：从 "题N" 或 "【题N】" 到下一个 "题N+1" / "题N 答案"
    pat = re.compile(rf"(题\s*{qnum}[：: 】].*?)(?=题\s*{qnum+1}[：: 】]|题\s*{qnum}\s*答案|\Z)", re.S)
    m = pat.search(txt)
    seg = m.group(1) if m else ""
    # 断语引用（xx_NNN 或 xxx_xxx_xxx）
    refs = re.findall(r"[a-z]+_\d+", seg)
    # 单选项理由
    opts = re.findall(r'([ABCD])"([^"]{6,60})', seg)
    return seg, refs, opts

# 按类别分组输出
from collections import defaultdict
by_cat = defaultdict(list)
for ftb, q, cat, pred, truth in errors:
    by_cat[cat].append((ftb, q, pred, truth))

for cat in ["性格", "婚姻", "财运", "健康", "事业"]:
    items = by_cat.get(cat, [])
    print(f"########## {cat} ({len(items)} 错) ##########")
    for ftb, q, pred, truth in items[:8]:
        gid, _ = INV.get(ftb, (ftb, 0))
        seg, refs, opts = extract_reason(gid, q)
        # 取推理中最后裁决句
        verdict = ""
        for line in reversed(seg.splitlines()):
            if ("结论" in line or "综合" in line or "裁决" in line or "故选" in line or "答案" in line) and len(line) > 10:
                verdict = line.strip()[:180]
                break
        if not verdict and seg:
            verdict = seg.strip().splitlines()[-1][:180] if seg.strip() else ""
        print(f"  {ftb} ({gid}题{q}) pred={pred} truth={truth} 断语:{refs[:6]}")
        print(f"    裁决: {verdict}")
    print()
