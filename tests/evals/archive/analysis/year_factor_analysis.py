#!/usr/bin/env python3
"""错题年份因子分析：对婚姻/六亲/财运年份错题，逐候选年算流年因子 + yingqi 断语，找区分规则。"""
import sys, json
sys.path.insert(0, "tools")
from paipan import full_paipan, liunian
from duanyu import make_liunian_factors, query

CASES = [
    # (出生, 性别, 经度, 题目域, target, 候选年, 答案年, 说明)
    ("1986-04-24T21:30:00+08:00", "male", 120.0, "婚姻", "配偶星", [2012, 2013, 2015, 2016], 2016, "pan11题1 何年结婚"),
    ("1977-10-26T11:10:00+08:00", "female", 101.7, "婚姻", "配偶星", [2010, 2013, 2016, 2018], 2013, "pan04题3 何年离婚"),
    ("1983-11-01T21:00:00+08:00", "male", 121.0, "财运", "财星", [2014, 2018, 2019, 2021], 2021, "pan06题3 何时被骗"),
    ("1974-04-28T16:40:00+08:00", "male", -75.0, "六亲", "父星", [2009, 2013, 2018, 2020], 2018, "pan01题4 父亡之年(对照:答对)"),
]

for birth, gender, lon, domain, target, years, truth, note in CASES:
    try:
        pan = full_paipan(birth, gender, lon)
    except Exception as e:
        print("!! 排盘失败 %s: %s" % (note, e))
        continue
    print("\n########## %s (truth=%d) ##########" % (note, truth))
    for y in years:
        ln = liunian(pan, y)
        fl = make_liunian_factors(pan, ln, target=target, year=y)
        bz = fl["八字"]
        on = [k for k, v in bz.items() if v]
        yq = query("yingqi", fl)
        yh = [h["结论"][:22] for h in yq.get("八字", [])]
        mark = "  <<< 答案年" if y == truth else ""
        print("  %d: 因子=%s" % (y, on))
        print("      yingqi=%s%s" % (yh, mark))
