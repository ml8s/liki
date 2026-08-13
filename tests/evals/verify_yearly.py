#!/usr/bin/env python3
"""验证 yearly_* 流年域: 流年快照命中、本命快照不命中(隔离)、本命原域回归正常。"""
import sys
sys.path.insert(0, "tools")
from paipan import full_paipan, liunian
from duanyu import make_factors, make_liunian_factors, query

pan = full_paipan("1981-08-26T23:00:00+08:00", "male", 130.7, correct=True)
snap = make_factors(pan)
print("== 本命快照: 查 yearly_* 应不命中(隔离) ==")
for rule in ["yearly_marriage", "yearly_liuqin", "yearly_caiyun"]:
    r = query(rule, snap)
    n = len(r.get("八字", [])) + len(r.get("紫微", []))
    print("  query(%s, 本命快照) = %d 条 %s" % (rule, n, "✓ 隔离正确" if n == 0 else "!! 误命中"))

print("\n== 本命原域回归(应有输出; yingqi 为流年专用,本命不查) ==")
for rule in ["marriage", "liuqin", "caiyun", "yingqi"]:
    r = query(rule, snap)
    n = len(r.get("八字", [])) + len(r.get("紫微", []))
    if rule == "yingqi":
        print("  query(%s, 本命快照) = %d 条 (预期 0——流年专用表)" % (rule, n))
    else:
        print("  query(%s, 本命快照) = %d 条 %s" % (rule, n, "✓" if n > 0 else "!! 空"))

print("\n== 流年快照: 查 yearly_* 应命中 + 原域不命中(隔离) ==")
for y in [2006, 2018, 2021]:
    ln = liunian(pan, y)
    fl = make_liunian_factors(pan, ln, target="配偶星", year=y)
    print("  --- 流年 %d ---" % y)
    for rule in ["yearly_marriage", "marriage", "yingqi"]:
        r = query(rule, fl)
        hits = [h["id"] for h in r.get("八字", [])]
        print("    query(%s) = %s" % (rule, hits if hits else "空"))
