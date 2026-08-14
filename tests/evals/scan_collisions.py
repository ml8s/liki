#!/usr/bin/env python3
"""扫描断语表「撞」：同一盘同时命中语义冲突的条目对。
冲突对由命理语义定义（如 婚可成 vs 婚难成 / 外向 vs 内向 / 得财 vs 破财）。"""
import sys, os, re, json
sys.path.insert(0, "tools")
from paipan import full_paipan
from duanyu import make_factors, query

CACHE = "/tmp/pan_cache"
os.makedirs(CACHE, exist_ok=True)

def birth_of(gid):
    yaml = open("tests/cases-grouped/%s.yaml" % gid, encoding="utf-8").read()
    m = re.search(r"出生信息[:：]\s*(.*?)(?:\n|$)", yaml)
    s = m.group(1)
    gen = "male" if ("男" in s or "乾" in s) else ("female" if ("女" in s or "坤" in s) else "?")
    dm = re.search(r"(\d{4})[年/.-](\d{1,2})[月/.-](\d{1,2})", s)
    tm = re.search(r"([上下晚凌晨中午]?)(?:午)?\s*(\d{1,2})[时:](\d{1,2})?", s)
    hh = tm.group(2) if tm else "12"
    mm = tm.group(3) or "00" if tm else "00"
    if tm and tm.group(1) in ("晚", "下") and int(hh) < 12: hh = int(hh) + 12
    return "%s-%s-%sT%s:%s:00+08:00" % (dm.group(1), dm.group(2).zfill(2), dm.group(3).zfill(2), str(hh).zfill(2), mm), gen

def get_snap(gid):
    cpath = os.path.join(CACHE, "collide_%s.json" % gid)
    if os.path.exists(cpath):
        return json.load(open(cpath, encoding="utf-8"))
    birth, gen = birth_of(gid)
    pan = full_paipan(birth, gen, 120.0)
    snap = make_factors(pan)
    json.dump(snap, open(cpath, "w", encoding="utf-8"))
    return snap

# 冲突对定义: (域, [(正面条目关键词), (反面条目关键词)])
CONFLICTS = {
    "marriage": [
        ("婚可成/婚缘佳", ["婚可成", "婚姻正常", "婚可成且稳定", "上等婚姻", "婚姻稳定"], ["婚难成", "婚缘受阻", "独身", "婚而不久", "婚变", "婚缘虚"]),
        ("婚姻正常 vs 婚变", ["婚姻正常", "婚姻稳定"], ["婚变", "婚而不久", "婚缘受损"]),
    ],
    "xingge": [
        ("外向 vs 内向", ["好胜强势", "威严果决", "性急外向", "爱表现", "行动力强"], ["内向敏感", "仁厚保守", "情绪压抑", "性喜独处", "淡泊名利"]),
    ],
    "caiyun": [
        ("得财 vs 破财", ["财源广进", "财运顺畅", "进财", "财丰"], ["破财", "难聚财", "财来财去", "负债", "财被夺"]),
    ],
    "jiankang": [
        ("健康 vs 病灾", ["健康", "无碍", "体健"], ["病灾", "健康凶", "疾病", "受损"]),
    ],
    "shiye": [
        ("事业显达 vs 事业阻", ["显达", "升迁", "事业有成"], ["事业受阻", "挫折", "困顿"]),
    ],
}

def main():
    gids = sorted(os.listdir("tests/cases-grouped/")) if False else None
    # 用评测盘
    g = json.load(open("tests/groups.json"))
    for dom, pairs in CONFLICTS.items():
        print("########## %s 表冲突扫描 ##########" % dom)
        for label, pos_kw, neg_kw in pairs:
            # 收集该域所有条目结论
            import csv
            rows = list(csv.DictReader(open("tools/bazi/%s.csv" % dom, encoding="utf-8")))
            pos_ids = [r["id"] for r in rows if any(k in r["结论"] for k in pos_kw)]
            neg_ids = [r["id"] for r in rows if any(k in r["结论"] for k in neg_kw)]
            hits = []
            for gid in sorted(g.keys()):
                try:
                    snap = get_snap(gid)
                except Exception:
                    continue
                try:
                    r = query(dom, snap)
                    bh = [h["id"] for h in r.get("八字", [])]
                except Exception:
                    continue
                pos_hit = [i for i in pos_ids if i in bh]
                neg_hit = [i for i in neg_ids if i in bh]
                if pos_hit and neg_hit:
                    hits.append((gid, pos_hit, neg_hit))
            if hits:
                print("  [%s vs %s] 撞 %d 盘:" % ("+".join(pos_kw[:2]), "+".join(neg_kw[:2]), len(hits)))
                for gid, p, n in hits:
                    print("    %s: 正面=%s 反面=%s" % (gid, p, n))
            else:
                print("  [%s vs %s] 无撞" % ("+".join(pos_kw[:2]), "+".join(neg_kw[:2])))

if __name__ == "__main__":
    main()
