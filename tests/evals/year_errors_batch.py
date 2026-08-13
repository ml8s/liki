#!/usr/bin/env python3
"""年份错题批量分析 v2: 排盘缓存到 /tmp/pan_cache, 分块处理。用法: python3 year_errors_batch.py <start> <end>"""
import sys, os, json, re, time
sys.path.insert(0, "tools")
from paipan import full_paipan, liunian
from duanyu import make_liunian_factors, query

CACHE = "/tmp/pan_cache"
os.makedirs(CACHE, exist_ok=True)

def birth_info(gid):
    yaml = open("tests/cases-grouped/%s.yaml" % gid, encoding="utf-8").read()
    m = re.search(r"出生信息：(.*?)(?:\n|$)", yaml)
    s = m.group(1)
    gen = "male" if "男" in s else ("female" if "女" in s else "?")
    dm = re.search(r"(\d{4})[年/.-](\d{1,2})[月/.-](\d{1,2})", s)
    if not dm:
        return None, None
    y, mo, d = dm.groups()
    tm = re.search(r"(\d{1,2})[时:](\d{1,2})?", s)
    hh = tm.group(1) if tm else "12"
    mm = tm.group(2) if tm and tm.group(2) else "00"
    tz = "+08:00"
    return "%s-%s-%sT%s:%s:00%s" % (y, mo.zfill(2), d.zfill(2), hh.zfill(2), mm, tz), gen

CASES = [
    ("ftb_0007", "pan02", 2, "婚姻", "2005", "2006", ["2005", "2006", "2012", "2009"], "配偶星"),
    ("ftb_0018", "pan04", 3, "婚姻", "2018", "2013", ["2010", "2013", "2016", "2018"], "配偶星"),
    ("ftb_0028", "pan06", 3, "财运", "2019", "2021", ["2014", "2018", "2019", "2021"], "财星"),
    ("ftb_0032", "pan07", 2, "六亲", "1997", "2003", ["1995", "1997", "2003", "2007"], "父星"),
    ("ftb_0051", "pan11", 1, "婚姻", "2013", "2016", ["2012", "2013", "2015", "2016"], "配偶星"),
    ("ftb_0063", "pan13", 3, "婚姻", "2009", "2008", ["2006", "2007", "2008", "2009"], "配偶星"),
    ("ftb_0067", "pan14", 2, "财运", "1989", "1987", ["1987", "1989", "1990", "1992"], "财星"),
    ("ftb_0076", "pan16", 1, "婚姻", "2000", "2005", ["2000", "2001", "2005", "2006"], "配偶星"),
    ("ftb_0077", "pan16", 2, "婚姻", "2011", "2013", ["2011", "2012", "2013", "2015"], "配偶星"),
    ("ftb_0089", "pan18", 4, "六亲", "2021", "2011", ["1989", "1990", "2011", "2021"], "母星"),
    ("ftb_0097", "pan20", 1, "婚姻", "2003", "2005", ["2003", "2005", "2007", "2009"], "配偶星"),
    ("ftb_0099", "pan20", 3, "婚姻", "2015", "2017", ["2011", "2013", "2015", "2017"], "配偶星"),
    ("ftb_0101", "pan21", 1, "婚姻", "2002", "1999", ["1998", "1999", "2002", "2003"], "配偶星"),
    ("ftb_0143", "pan29", 3, "婚姻", "2021", "2016", ["2021", "2017", "2022", "2016"], "配偶星"),
]

start, end = int(sys.argv[1]), int(sys.argv[2])
for ftb, gid, q, cat, pred_y, truth_y, years, target in CASES[start:end]:
    birth, gen = birth_info(gid)
    cpath = os.path.join(CACHE, "%s.json" % gid)
    if os.path.exists(cpath):
        pan = json.load(open(cpath, encoding="utf-8"))
    else:
        try:
            pan = full_paipan(birth, gen, 120.0)
            json.dump(pan, open(cpath, "w", encoding="utf-8"))
        except Exception as e:
            print("## %s 排盘失败 %s" % (ftb, e), flush=True)
            continue
    print("## %s (%s题%d) %s truth=%s pred=%s" % (ftb, gid, q, cat, truth_y, pred_y), flush=True)
    for y in years:
        ln = liunian(pan, int(y))
        fl = make_liunian_factors(pan, ln, target=target, year=int(y))
        on = [k for k, v in fl["八字"].items() if v]
        try:
            yq = query("yingqi", fl)
            yh = [h["结论"][:14] for h in yq.get("八字", [])]
        except Exception as e:
            yh = ["ERR"]
        mark = " <<<<<<< TRUTH" if y == truth_y else (" <-- PRED" if y == pred_y else "")
        print("  %s: %s | %s%s" % (y, on, yh, mark), flush=True)
