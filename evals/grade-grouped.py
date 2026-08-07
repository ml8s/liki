#!/usr/bin/env python3
"""MingLi-Bench 命盘分组判分：按「题N 答案：X」逐题对比 evals/answers.json。

用法（在 skills/liki 目录下）：
    python3 evals/grade-grouped.py <iteration_dir> [groups.json] [answers.json]

说明：
  - groups.json: case_id -> [ftb_ids 有序列表]（每组 4-6 题）
  - 每题提取「题N 答案：X」；无题号时按答案行出现顺序对应题序
  - 输出：总正确率（按题）+ 分组正确率 + 错题清单
"""
import json, re, glob, os, sys
from collections import Counter, defaultdict

def main():
    it_dir = sys.argv[1]
    base = os.path.dirname(__file__)
    groups = json.load(open(sys.argv[2] if len(sys.argv) > 2 else os.path.join(base, 'groups.json')))
    answers = json.load(open(sys.argv[3] if len(sys.argv) > 3 else os.path.join(base, 'answers.json')))
    try:
        cats = json.load(open(os.path.join(base, 'cats.json')))
    except Exception:
        cats = {}

    res = Counter(); bycat = defaultdict(Counter); fails = []; group_res = []
    for case_id, qids in sorted(groups.items()):
        d = os.path.join(it_dir, case_id)
        rsp_path = os.path.join(d, 'with_skill', 'outputs', 'response.md')
        if not os.path.exists(rsp_path):
            for qid in qids:
                res['ERROR'] += 1; bycat[cats.get(qid, '?')]['ERROR'] += 1
            group_res.append((case_id, 'ERROR', 0, len(qids))); continue
        rsp = open(rsp_path, encoding='utf-8').read()
        # ① 优先按题号提取
        by_no = {}
        for m in re.finditer(r'题\s*([1-6])[：:]\s*答案[：:]\s*([A-D])', rsp):
            by_no[int(m.group(1))] = m.group(2)
        # ② 兜底：取尾部答案块（qwen 常见"答案：B 答案：B..."或"答案：B B B A B"）
        if not by_no:
            tail = rsp[-600:]
            # 尾部一行多字母：答案：D A B C B A
            m1 = re.search(r'答案[：:]\s*([A-D](?:\s+[A-D]){3,5})\s*$', tail)
            # 尾部连续多行：答案：B 答案：B 答案：A
            m2 = re.findall(r'答案[：:]\s*([A-D])', tail)
            if m1:
                seq = m1.group(1).split()
            elif m2:
                seq = m2
            else:
                seq = re.findall(r'答案[：:]\s*([A-D])', rsp)
            by_no = {i+1: v for i, v in enumerate(seq[:len(qids)])}
        g_pass = 0
        for i, qid in enumerate(qids, 1):
            pred = by_no.get(i, '?')
            truth = answers.get(qid, '?')
            if pred == truth:
                res['PASS'] += 1; bycat[cats.get(qid, '?')]['PASS'] += 1; g_pass += 1
            else:
                res['FAIL'] += 1; bycat[cats.get(qid, '?')]['FAIL'] += 1
                fails.append((qid, i, pred, truth, cats.get(qid, '?')))
        group_res.append((case_id, f"{g_pass}/{len(qids)}", g_pass, len(qids)))

    print('=== MingLi-Bench 命盘分组评测结果 ===')
    print(f"总题数: {sum(res.values())} | PASS {res['PASS']} | FAIL {res['FAIL']} | ERROR {res['ERROR']}")
    tot = res['PASS'] + res['FAIL']
    if tot:
        print(f"正确率: {res['PASS']}/{tot} = {res['PASS']/tot*100:.1f}%")
    print()
    print('=== 分组正确率 ===')
    for cid, label, p, n in group_res:
        print(f"  {cid}: {label}")
    print()
    print('=== 分类正确率 ===')
    for c, cnt in sorted(bycat.items()):
        p, f, e = cnt['PASS'], cnt['FAIL'], cnt['ERROR']
        rate = f"{p/(p+f)*100:.0f}%" if p + f else '-'
        print(f"  {c}: PASS {p} FAIL {f} ERR {e} = {rate}")
    print()
    print('=== 错题 ===')
    for qid, no, pred, truth, c in fails:
        print(f"  {qid} [题{no}] [{c}] pred={pred} truth={truth}")

if __name__ == '__main__':
    main()
