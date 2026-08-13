#!/usr/bin/env python3
"""错题断语引用统计 + 答案行序列检查。"""
import re, os

errors = []
for line in open("tests/evals/grade_iter3.txt"):
    m = re.match(r"\s*(ftb_\d+)\s+\[题(\d+)\]\s+\[([^\]]+)\]\s+pred=([A-D])\s+truth=([A-D])", line)
    if m:
        errors.append(m.groups())

import json
g = json.load(open("tests/groups.json"))
inv = {}
for gid, ftbs in g.items():
    for i, ftb in enumerate(ftbs, 1):
        inv[ftb] = (gid, i)

no_ref = 0
has_ref = 0
no_seg = 0
for ftb, q, cat, pred, truth in errors:
    gid, _ = inv.get(ftb, (None, None))
    if not gid:
        continue
    rsp = os.path.join("..", "liki-workspace", "iteration-3", gid, "with_skill", "outputs", "response.md")
    if not os.path.exists(rsp):
        continue
    txt = open(rsp, encoding="utf-8").read()
    qn = int(q)
    pat = re.compile(r"题\s*%d[：: 】].*?(?=题\s*%d[：: 】]|题\s*%d\s*答案|\Z)" % (qn, qn + 1, qn), re.S)
    seg = ""
    m = pat.search(txt)
    if m:
        seg = m.group(0)
    refs = re.findall(r"[a-z]+_\d+", seg)
    if not seg.strip():
        no_seg += 1
    elif refs:
        has_ref += 1
    else:
        no_ref += 1
print("错题断语引用统计: 有引用=%d 无引用=%d 段落未提取=%d" % (has_ref, no_ref, no_seg))

# pan03 题5 答案行序列（检查 agent 是否中途改答案）
txt = open("../liki-workspace/iteration-3/pan03/with_skill/outputs/response.md", encoding="utf-8").read()
print("\n===pan03 全部「题5」行===")
for i, l in enumerate(txt.splitlines()):
    if "题5" in l:
        print(i, ":", l.strip()[:110])

# grade-grouped 取答案逻辑：看它是否取最后
print("\n===grade-grouped 答案提取逻辑===")
gg = open("tests/grade-grouped.py", encoding="utf-8").read()
for line in gg.splitlines():
    if "finditer" in line or "tail" in line.lower() or "最后" in line or "m1" in line or "m2" in line:
        print(line.strip()[:120])
