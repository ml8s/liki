#!/usr/bin/env python3
"""排盘正确性检查: 32 盘 agent 四柱 vs 本地重排四柱。"""
import sys, os, re, json
sys.path.insert(0, "tools")
from paipan import full_paipan

CACHE = "/tmp/pan_cache"
os.makedirs(CACHE, exist_ok=True)

def birth_info(gid):
    yaml = open("tests/cases-grouped/%s.yaml" % gid, encoding="utf-8").read()
    m = re.search(r"出生信息[:：]\s*(.*?)(?:\n|$)", yaml)
    s = m.group(1)
    gen = "male" if ("男" in s or "乾" in s) else ("female" if ("女" in s or "坤" in s) else "?")
    dm = re.search(r"(\d{4})[年/.-](\d{1,2})[月/.-](\d{1,2})", s)
    if not dm:
        return None, None, None
    y, mo, d = dm.groups()
    tm = re.search(r"([上下晚凌晨中午]?)(?:午)?\s*(\d{1,2})[时:](\d{1,2})?", s)
    if tm:
        period, hh, mm = tm.group(1), tm.group(2), tm.group(3) or "00"
        hh = int(hh)
        # 晚上/下午 + hh<12 → +12；凌晨/上午/中午 → 不变
        if period in ("晚", "下") and hh < 12:
            hh += 12
        hh = "%02d" % hh
    else:
        hh, mm = "12", "00"
    # 有明确时辰字样(子丑寅卯...)且无具体时:分 → 用时辰中间时刻 + correct=False
    has_shichen = bool(re.search(r"[子丑寅卯辰巳午未申酉戌亥]时", s))
    has_clock = bool(re.search(r"\d{1,2}[时:]\d{1,2}", s))
    tz = "+08:00"
    birth = "%s-%s-%sT%s:%s:00%s" % (y, mo.zfill(2), d.zfill(2), hh, mm, tz)
    return birth, gen, (not has_shichen)

def gz(p):
    c = p["chart"]
    return "%s%s %s%s %s%s %s%s" % (c["nian"]["gan"], c["nian"]["zhi"], c["yue"]["gan"], c["yue"]["zhi"],
                                     c["ri"]["gan"], c["ri"]["zhi"], c["shi"]["gan"], c["shi"]["zhi"])

# 从 agent response.md 提取四柱
def agent_gz(gid):
    rsp = "../liki-workspace/iteration-3/%s/with_skill/outputs/response.md" % gid
    txt = open(rsp, encoding="utf-8").read()
    # 常见格式: 甲寅 戊辰 己亥 丙寅 / 甲寅/戊辰/己亥/丙寅
    m = re.search(r"(年柱|四柱|八字)[：: ]*\s*([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])\s*[/,，\s]\s*([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])\s*[/,，\s]\s*([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])\s*[/,，\s]\s*([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])", txt)
    if m:
        return " ".join(m.groups()[1:])
    # 备用: 形如 "四柱：甲寅 戊辰 己亥 丙寅"
    m2 = re.search(r"(?:四柱|八字)[：:]\s*([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])[ /,，]+([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])[ /,，]+([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])[ /,，]+([甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥])", txt)
    return " ".join(m2.groups()) if m2 else None

print("盘      本地四柱           agent四柱        一致?")
for gid in sorted(os.listdir("../liki-workspace/iteration-3")):
    if not gid.startswith("pan"): continue
    birth, gen, correct = birth_info(gid)
    if not birth or gen == "?":
        print("%s  出生信息解析失败: %s" % (gid, open("tests/cases-grouped/%s.yaml" % gid, encoding="utf-8").read()[:60].replace("\n"," ")))
        continue
    cpath = os.path.join(CACHE, "%s_%s.json" % (gid, "A" if correct else "B"))
    try:
        if os.path.exists(cpath):
            pan = json.load(open(cpath, encoding="utf-8"))
        else:
            pan = full_paipan(birth, gen, 120.0, correct=correct)
            json.dump(pan, open(cpath, "w", encoding="utf-8"))
        local = gz(pan)
    except Exception as e:
        print("%s  本地排盘失败: %s" % (gid, e))
        continue
    ag = agent_gz(gid)
    ok = "✓" if ag and local == ag else ("?" if not ag else "✗ 不同")
    print("%s  %s  %s  %s%s" % (gid, local, ag or "(未提取)", ok, "" if ok != "✗ 不同" else "  <<<"))
