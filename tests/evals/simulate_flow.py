#!/usr/bin/env python3
"""模拟 skill 完整使用流程（按 SKILL.md Phase 0-7）——验证流程通畅性。
场景：1986-04-24 晚上9.30 亥时 男（pan11），问「命主在那一年结婚？」选项 2012/2013/2015/2016"""
import sys, json
sys.path.insert(0, "tools")
from paipan import full_paipan, liunian
from duanyu import make_factors, make_liunian_factors, query

print("=" * 60)
print("模拟场景：男命 1986-04-24 晚上9.30亥时 → 问「何年结婚？」")
print("=" * 60)

# ── Phase 0 路由 ──
print("\n[Phase 0] 路由：用户问婚姻 → app/marriage.md → 全流程 Phase 1-7")
print("  ✓ 主域 = 婚姻，题目 = 应期题（年份选项）")

# ── Phase 1 时辰判定 ──
print("\n[Phase 1] 时辰判定：题干「晚上9.30亥时」")
print("  ✓ 题干已定「亥时」→ 路 B → correct=False（不校正，防二次偏移）")

# ── Phase 2 排盘 ──
print("\n[Phase 2] 排盘：full_paipan('1986-04-24T21:30:00+08:00','male',120.0,correct=False)")
pan = full_paipan("1986-04-24T21:30:00+08:00", "male", 120.0, correct=False)
c = pan["chart"]
print("  ✓ 四柱: %s%s %s%s %s%s %s%s" % (c["nian"]["gan"], c["nian"]["zhi"], c["yue"]["gan"], c["yue"]["zhi"],
                                         c["ri"]["gan"], c["ri"]["zhi"], c["shi"]["gan"], c["shi"]["zhi"]))

# ── Phase 3 强弱 ──
ys = pan["yongshen"]
qs = ys.get("fu_yi", {}).get("qiangruo", "?")
print("\n[Phase 3] 身强弱: %s（读 fu_yi.qiangruo，禁止自行数五行）" % qs)

# ── Phase 4 用神 ──
fy = ys.get("fu_yi", {}); th = ys.get("tiao_hou", {}); gj = ys.get("ge_ju", {})
print("\n[Phase 4] 用神: 扶抑=%s 调候=%s 格局=%s" % (fy.get("yong"), th.get("yong"), gj.get("yong")))

# ── Phase 5 领域查表（婚姻卡）──
print("\n[Phase 5] 婚姻卡检查表：配偶星状态 + 大运窗口 + 紫微夫妻宫")
snap = make_factors(pan)
bz = snap["八字"]
print("  配偶星因子:", {k: bz[k] for k in ["配偶星透干", "配偶星藏支", "配偶星得地", "配偶星混杂", "比劫重", "夫妻宫被冲", "夫妻宫被合", "寡宿", "大运配偶星"] if bz.get(k)})

# ── Phase 5.5 规则引擎断语 ──
print("\n[Phase 5.5] 本命 marriage 断语:")
r = query("marriage", snap)
for h in r["八字"]:
    print("   [八字] %s: %s" % (h["id"], h["结论"][:44]))
for h in r["紫微"]:
    print("   [紫微] %s: %s" % (h["id"], h["结论"][:44]))

# ── Phase 6 紫微交叉 ──
zw = pan["ziwei"]
mg = [p for p in zw.get("gong_wei", []) if p.get("index") == "命宫"]
print("\n[Phase 6] 紫微交叉: 命宫=%s 夫妻宫主星=%s" % (
    [x["xing"] for x in mg[0].get("xing_yao", [])] if mg else "?",
    [x["xing"] for x in [p for p in zw.get("gong_wei", []) if p.get("index") == "夫妻"][0].get("xing_yao", [])] if [p for p in zw.get("gong_wei", []) if p.get("index") == "夫妻"] else "?"))

# ── Phase 7 考时（应期题：逐候选年）──
print("\n[Phase 7] 应期候选（何年结婚 → 逐候选年流年分析）:")
cands = []
for y in [2012, 2013, 2015, 2016]:
    ln = liunian(pan, y)
    fl = make_liunian_factors(pan, ln, target="配偶星", year=y)
    ymar = query("yearly_marriage", fl)
    yq = query("yingqi", fl)
    mh = [(h["id"], h["结论"][:18]) for h in ymar["八字"]]
    yh = [(h["id"], h["结论"][:14]) for h in yq["八字"]]
    print("   %d: yearly_marriage=%s" % (y, mh if mh else "空"))
    print("       yingqi=%s" % (yh if yh else "空"))
    if mh:
        cands.append(y)

# 裁决（按准则 0：候选锁婚动 + 混杂不降级）
print("\n[裁决] 候选年=%s（命中婚动断语）" % cands)
print("  ✓ 流程通畅：Phase 0-7 各步衔接完整，工具链（排盘→因子→断语→流年）无断点")
