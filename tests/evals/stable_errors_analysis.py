#!/usr/bin/env python3
"""稳定错题命理分析：排盘 → 本命断语 → 流年断语命中，输出根因线索。
用法: python3 stable_errors_analysis.py <类别>  (如 婚姻 / 性格 / 事业 ...)"""
import sys, os, re, json
sys.path.insert(0, "tools")
from paipan import full_paipan, liunian
from duanyu import make_factors, make_liunian_factors, query

CACHE = "/tmp/pan_cache"
os.makedirs(CACHE, exist_ok=True)

def birth_of(gid):
    yaml = open("tests/evals/cases/%s.yaml" % gid, encoding="utf-8").read()
    m = re.search(r"出生信息[:：]\s*(.*?)(?:\n|$)", yaml)
    s = m.group(1)
    gen = "male" if ("男" in s or "乾" in s) else ("female" if ("女" in s or "坤" in s) else "?")
    dm = re.search(r"(\d{4})[年/.-](\d{1,2})[月/.-](\d{1,2})", s)
    tm = re.search(r"([上下晚凌晨中午]?)(?:午)?\s*(\d{1,2})[时:](\d{1,2})?", s)
    hh = tm.group(2) if tm else "12"
    mm = tm.group(3) or "00" if tm else "00"
    if tm and tm.group(1) in ("晚", "下") and int(hh) < 12: hh = int(hh) + 12
    birth = "%s-%s-%sT%s:%s:00+08:00" % (dm.group(1), dm.group(2).zfill(2), dm.group(3).zfill(2), str(hh).zfill(2), mm)
    return birth, gen

def get_pan(gid):
    cpath = os.path.join(CACHE, "sa_%s.json" % gid)
    if os.path.exists(cpath):
        return json.load(open(cpath, encoding="utf-8"))
    birth, gen = birth_of(gid)
    pan = full_paipan(birth, gen, 120.0)
    json.dump(pan, open(cpath, "w", encoding="utf-8"))
    return pan

def main(cat):
    # 稳定错: 三轮回都错的题
    def parse(fn):
        d = {}
        for line in open(fn, encoding="utf-8"):
            m = re.match(r"\s*(ftb_\d+)\s+\[题(\d+)\]\s+\[([^\]]+)\]\s+pred=([A-D])\s+truth=([A-D])", line)
            if m: d[m.group(1)] = (m.group(2), m.group(3), m.group(4), m.group(5))
        return d
    e4, e5, e6 = parse("tests/evals/grade_iter4.txt"), parse("tests/evals/grade_iter5.txt"), parse("tests/evals/grade_iter6.txt")
    stable = [ftb for ftb in e4 if ftb in e5 and ftb in e6 and (e4[ftb][1] == cat or (e4[ftb][2] != e4[ftb][3]))]
    stable = [ftb for ftb in e4 if ftb in e5 and ftb in e6 and e4[ftb][1] == cat]
    g = json.load(open("tests/groups.json"))
    inv = {}
    for gid, ftbs in g.items():
        for i, ftb in enumerate(ftbs, 1): inv[ftb] = (gid, i)
    for ftb in sorted(stable, key=lambda x: int(x.split("_")[1])):
        q, cat2, pred, truth = e4[ftb]
        gid, _ = inv[ftb]
        yaml = open("tests/evals/cases/%s.yaml" % gid, encoding="utf-8").read()
        m2 = re.search(r"【题%s】问题：(.*?)(?=【题%s】|\Z)" % (q, int(q)+1), yaml, re.S)
        prob = m2.group(1).strip()[:40] if m2 else "?"
        try:
            pan = get_pan(gid)
        except Exception as ex:
            print("## %s(%s题%s) %s 排盘失败 %s" % (ftb, gid, q, cat2, ex)); continue
        snap = make_factors(pan)
        # 主域本命断语
        dom = {"婚姻": "marriage", "性格": "xingge", "事业": "shiye", "财运": "caiyun",
               "家庭": "liuqin", "学业": "xueye", "灾劫": "jiankang", "运势": "dayun",
               "子女": "zinv", "外貌": "waimao", "官非": "yingqi", "健康": "jiankang"}[cat2]
        try:
            r = query(dom, snap)
            bh = [(h["id"], h["结论"][:26]) for h in r.get("八字", [])]
            zh = [(h["id"], h["结论"][:26]) for h in r.get("紫微", [])]
        except Exception:
            bh, zh = [], []
        print("## %s(%s题%s) [%s] pred=%s truth=%s | %s" % (ftb, gid, q, cat2, pred, truth, prob))
        print("   本命%s断语: 八字=%s" % (dom, bh if bh else "空"))
        print("             紫微=%s" % (zh if zh else "空"))
        print()

if __name__ == "__main__":
    main(sys.argv[1])
