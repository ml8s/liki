#!/usr/bin/env python3
"""重建 32 盘本命快照缓存（最新因子表）+ 完整零撞回归扫描。
用法: python3 zero_collision_check.py"""
import sys, os, re, json, csv
sys.path.insert(0, "tools")
from paipan import full_paipan
from duanyu import make_factors, query

CACHE = "/tmp/pan_cache"
os.makedirs(CACHE, exist_ok=True)

def birth_of(gid):
    yaml = open("tests/evals/cases/%s.yaml" % gid, encoding="utf-8").read()
    m = re.search(r"出生信息[:：]\s*(.*?)(?:\n|$)", yaml)
    s = m.group(1)
    gen = "male" if ("男" in s or "乾" in s) else ("female" if ("女" in s or "坤" in s) else "?")
    dm = re.search(r"(\d{4})[年/.-](\d{1,2})[月/.-](\d{1,2})", s)
    if not dm:
        return None, None
    tm = re.search(r"([上下晚凌晨中午]?)(?:午)?\s*(\d{1,2})[时:](\d{1,2})?", s)
    hh = tm.group(2) if tm else "12"
    mm = tm.group(3) or "00" if tm else "00"
    if tm and tm.group(1) in ("晚", "下") and int(hh) < 12: hh = int(hh) + 12
    birth = "%s-%s-%sT%s:%s:00+08:00" % (dm.group(1), dm.group(2).zfill(2), dm.group(3).zfill(2), str(hh).zfill(2), mm)
    return birth, gen

def main():
    g = json.load(open("tests/groups.json"))
    # 1. 重建快照缓存
    snaps = {}
    for gid in sorted(g.keys()):
        cpath = os.path.join(CACHE, "zc_%s.json" % gid)
        if os.path.exists(cpath):
            snaps[gid] = json.load(open(cpath, encoding="utf-8"))
            continue
        birth, gen = birth_of(gid)
        if not birth:
            print("!! %s 出生解析失败" % gid, flush=True); continue
        try:
            pan = full_paipan(birth, gen, 120.0)
        except Exception as e:
            print("!! %s 排盘失败 %s" % (gid, e), flush=True); continue
        snap = make_factors(pan)
        json.dump(snap, open(cpath, "w", encoding="utf-8"))
        snaps[gid] = snap
        print(".", end="", flush=True)
    print("\n缓存重建完成: %d 盘" % len(snaps))

    # 2. 完整零撞扫描(八字侧)
    CONFLICTS = {
        "marriage": [("婚可成vs难成", ["婚可成", "婚姻正常", "婚可成且稳定"], ["婚难成", "婚缘受阻", "独身", "婚而不久"])],
        "xingge": [("外向vs内向", ["好胜强势", "威严果决", "性急外向"], ["内向敏感", "情绪压抑", "性喜独处"])],
        "caiyun": [("得财vs破财", ["财源广进", "财运顺畅"], ["破财", "难聚财"])],
        "liuqin": [("父旺vs父损", ["父星旺而得地"], ["父寿不永"])],
        "shiye": [("成vs阻", ["显达", "事业有成"], ["事业受阻", "困顿"])],
        "xueye": [("学成vs阻", ["成学有望"], ["财坏印", "学历受阻"])],
        "zinv": [("有子vs损", ["生育", "子女多"], ["无子", "子女缘薄"])],
        "jiankang": [("健vs病", ["体健", "健康"], ["病弱", "多病"])],
        "waimao": [("貌吉vs貌劣", ["貌美", "俊"], ["貌陋", "丑"])],
    }
    total = 0
    for dom, pairs in CONFLICTS.items():
        rows = list(csv.DictReader(open("tools/bazi/%s.csv" % dom, encoding="utf-8")))
        for label, pos_kw, neg_kw in pairs:
            pos = [r["id"] for r in rows if any(k in r["结论"] for k in pos_kw)]
            neg = [r["id"] for r in rows if any(k in r["结论"] for k in neg_kw)]
            if not pos or not neg:
                continue
            for gid, snap in snaps.items():
                try:
                    r = query(dom, snap)
                    bh = [h["id"] for h in r.get("八字", [])]
                except Exception:
                    continue
                p = [i for i in pos if i in bh]; n = [i for i in neg if i in bh]
                if p and n:
                    total += 1
                    print("撞: %s[%s] %s 正=%s 负=%s" % (dom, label, gid, p, n), flush=True)
    print("八字侧零撞扫描: 撞 %d 处" % total)

    # 3. xingge 月令主面唯一性
    uniq = multi = missing = 0
    for gid, snap in snaps.items():
        r = query("xingge", snap)
        mains = [h["id"] for h in r.get("八字", []) if h["id"].startswith("xg_m")]
        if len(mains) == 1: uniq += 1
        elif len(mains) > 1: multi += 1
        else: missing += 1
    print("月令主面: 唯一 %d / 多 %d / 缺 %d (共 %d 盘)" % (uniq, multi, missing, len(snaps)))

if __name__ == "__main__":
    main()
