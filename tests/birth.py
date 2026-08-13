"""出生信息解析（单一来源）：题干 → (solar_time, gender, longitude, correct)。

被 predict.py 与全部 eval_*.py 共用，避免多份拷贝漂移。

规则（对应 SKILL.md 路 A/路 B）：
- 路 A：题干给了具体钟表时刻（无"X时"命理时辰标注）→ 真太阳时校正（correct=True）
- 路 B：题干已定时辰（"X时/時/裸地支"）→ 不校正（correct=False），防二次偏移；
  时刻取值：题干有具体时刻（如"23:34子时"）→ 用该时刻；否则用 X 时标准时刻
- 12 小时制：下午/晚上/夜/PM → +12；中午/正午 → 12；凌晨/早上/上午/AM → 不变
- 容错：繁体"時"→"时"；"已时"→"巳时"（错别字）；"02:00丑"（丑无"时"字）
"""
import re

_PLACES = {
    "台湾": 121.5, "usa": -75.0, "malaysia": 101.7, "马来": 101.7,
    "香港": 114.2, "北京": 116.4, "广东": 113.3, "广州": 113.3,
    "日本": 139.7, "宫崎": 131.4, "澳门": 113.5, "上海": 121.5,
    "新加坡": 103.8,
}

_SHICHEN = {"子": "00:00:00", "丑": "02:00:00", "寅": "04:00:00", "卯": "06:00:00",
            "辰": "08:00:00", "巳": "10:00:00", "午": "12:00:00", "未": "14:00:00",
            "申": "16:00:00", "酉": "18:00:00", "戌": "20:00:00", "亥": "22:00:00"}


def _to_24h(hh: int, mm: int, s: str) -> tuple[int, int]:
    """12 小时制 → 24 小时制（下午/晚上/夜/PM +12；中午/正午→12；凌晨/早上/上午/AM 不变）。"""
    p = re.search(r"凌晨|早上|上午|中午|正午|下午|晚上|夜|(?:PM|pm|AM|am)", s)
    if not p:
        return hh, mm
    w = p.group(0)
    if w in ("下午", "晚上", "夜", "PM", "pm"):
        if hh < 12:
            hh += 12
    elif w in ("中午", "正午"):
        if hh < 12:
            hh = 12
    elif hh == 12:
        hh = 0  # 凌晨/早上/上午 12 点 → 00 点（12 小时制午夜）
    return hh, mm


def parse_birth(s: str):
    """解析出生信息 → (solar_time_str, gender, longitude, correct)。

    solar_time_str 形如 "1981-08-26T00:15:00+08:00"（东八区钟表时间，correct 决定是否校正）。
    """
    gender = "male" if "男" in s else "female"
    lonv = None
    for k, v in _PLACES.items():
        if k in s:
            lonv = v
            break
    # 优先公历（公元/公历/西历/西元/阳历），有括号阳历则取括号内（农历括号标注阳历）
    m = re.search(r"(?:公元|公历|西历|西元|阳历)\s*(\d{4})年(\d{1,2})月(\d{1,2})日", s)
    if not m:
        m = re.search(r"（阳历(\d{4})年(\d{1,2})月(\d{1,2})日）", s)
    if not m:
        m = re.search(r"(\d{4})年(\d{1,2})月(\d{1,2})日", s)
    if not m:
        m = re.search(r"(?:公元|公历|西历|西元|阳历)\s*(\d{4})-(\d{1,2})-(\d{1,2})", s)
    if not m:
        m = re.search(r"(\d{4})-(\d{1,2})-(\d{1,2})", s)
    if not m:
        return None, gender, lonv, True
    y, mo, d = int(m.group(1)), int(m.group(2)), int(m.group(3))
    t = "12:00:00"
    # 先剥离日期整段（含点号日期"1990.8.26"），避免其数字被误读为时刻
    s_tm = re.sub(r"(?:公元|公历|西历|西元|阳历)?\s*\d{4}[年.\-]\d{1,2}[月.\-]\d{1,2}日?", "", s)
    tm = re.search(r"(\d{1,2})(?:[:：点时]|\.)(\d{1,2})?分?", s_tm)
    if tm:
        hh, mm = int(tm.group(1)), int(tm.group(2)) if tm.group(2) else 0
        hh, mm = _to_24h(hh, mm, s)
        t = f"{hh:02d}:{mm:02d}:00"
    # 路 B：题干已定时辰（X时/時/裸地支）→ 不校正（correct=False），防二次偏移
    s2 = s.replace("已时", "巳时").replace("時", "时")
    sm = re.search(r"（?([子丑寅卯辰巳午未申酉戌亥])(?:时|$)", s2)
    if sm:
        has_exact = re.search(r"\d{1,2}(?:[:：点时]|\.)", s) is not None
        if not has_exact:
            t = _SHICHEN[sm.group(1)]
        return f"{y:04d}-{mo:02d}-{d:02d}T{t}+08:00", gender, lonv, False
    # 无 X 时标注 → 具体时刻 → 路 A 真太阳时校正
    return f"{y:04d}-{mo:02d}-{d:02d}T{t}+08:00", gender, lonv, True
