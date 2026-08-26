#!/usr/bin/env python3
"""复现验证：流年查询对 yingqi 命中、对域表(marriage/caiyun/liuqin)全空。"""
import sys, json
sys.path.insert(0, "tools")
from paipan import full_paipan, liunian
from duanyu import make_factors, make_liunian_factors, query

pan = full_paipan("1974-04-28T16:40:00+08:00", "male", -75.0)
print("本命盘 OK: %s %s" % (pan["chart"]["nian"]["gan"] + pan["chart"]["nian"]["zhi"],
                            pan["chart"]["ri"]["gan"] + pan["chart"]["ri"]["zhi"]))

# 本命快照
snap = make_factors(pan)
print("\n本命因子数: 八字=%d 紫微=%d" % (len(snap["八字"]), len(snap["紫微"])))

# 本命域表查询（应该命中）
for rule in ["marriage", "liuqin"]:
    r = query(rule, snap)
    print("本命 query(%s): 八字命中 %d / 紫微命中 %d" % (rule, len(r["八字"]), len(r["紫微"])))

# 流年快照（1996 丙子——pan01 题1 年份）
ln = liunian(pan, 1996)
fl = make_liunian_factors(pan, ln, target="配偶星", year=1996)
print("\n流年(1996)因子数: 八字=%d 紫微=%d" % (len(fl["八字"]), len(fl["紫微"])))
print("流年八字因子样例:", [k for k in fl["八字"] if fl["八字"][k]][:15])

# 流年查询各域
for rule in ["yingqi", "marriage", "caiyun", "liuqin", "jiankang", "xueye"]:
    r = query(rule, fl)
    print("流年 query(%s): 八字命中 %d / 紫微命中 %d" % (rule, len(r["八字"]), len(r["紫微"])),
          "" if r["八字"] else "<-- 空!")
