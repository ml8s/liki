"""Counsel's internal Bazi factor projection.

Canonical Bazi facts live in si-zhu, chang-sheng, relations, and named
conclusions. These helpers build only the short-lived operator index; the
projection is never published as another chart contract.
"""


def _stem_ten_god(pillar: dict) -> str:
    for item in pillar.get("shi_shens") or []:
        if item.get("source") == "gan":
            return item.get("shi_shen") or ""
    return ""


def _main_hidden_ten_god(pillar: dict) -> str:
    for item in pillar.get("shi_shens") or []:
        if item.get("source") == "main_qi":
            return item.get("shi_shen") or ""
    return ""


def _chang_sheng_lookup(bazi: dict) -> dict[str, str]:
    return {
        item.get("index") or "": item.get("stage") or item.get("name") or ""
        for item in bazi.get("chang_sheng") or []
    }


def _pillar_punishments(bazi: dict) -> dict[str, bool]:
    result = {"nian": False, "yue": False, "ri": False, "shi": False}
    for relation in bazi.get("liu_xing") or []:
        for index in (relation.get("pillar_a"), relation.get("pillar_b")):
            pillar = {0: "nian", 1: "yue", 2: "ri", 3: "shi"}.get(index)
            if pillar:
                result[pillar] = True
    return result


def _spouse_palace_state(bazi: dict) -> str:
    groups = [
        ("冲", bazi.get("liu_chong") or []),
        ("合", bazi.get("zhi_liu_he") or []),
        ("刑", bazi.get("liu_xing") or []),
        ("害", bazi.get("liu_hai") or []),
    ]
    for state, relations in groups:
        for relation in relations:
            if 2 in (relation.get("pillar_a"), relation.get("pillar_b")):
                return state
    return "静"


def _day_branch_type(bazi: dict) -> str:
    names = {
        item.get("name") or item.get("xing") or ""
        for item in (bazi.get("ri") or {}).get("shen_sha") or []
    }
    if "桃花" in names:
        return "桃花"
    if "驿马" in names:
        return "驿马"
    if (bazi.get("ri") or {}).get("zhi") in {"辰", "戌", "丑", "未"}:
        return "墓库"
    return "无"


def build_atomic_facts(bazi: dict) -> dict:
    from app.natal.tools.factor_constants import load_constants

    constants = load_constants()
    gan_wuxing = constants.get("天干五行") or {}
    ri = bazi.get("ri") or {}
    nian = bazi.get("nian") or {}
    yue = bazi.get("yue") or {}
    shi = bazi.get("shi") or {}
    chang_sheng = _chang_sheng_lookup(bazi)
    ge_ju = bazi.get("ge_ju") or {}
    guan_sha = bazi.get("guan_sha") or {}
    cai_ku = bazi.get("cai_ku") or {}
    return {
        "day_master_element": gan_wuxing.get(ri.get("gan") or "", ""),
        "day_master_stem": ri.get("gan") or "",
        "day_branch": ri.get("zhi") or "",
        "month_longevity": chang_sheng.get(yue.get("zhi") or "", ""),
        "year_stem_ten_god": _stem_ten_god(nian),
        "month_main_ten_god": _main_hidden_ten_god(yue),
        "hour_stem_ten_god": _stem_ten_god(shi),
        "pattern_god_transparent": bool(ge_ju.get("ge_shen_tou")),
        "pillar_punishments": _pillar_punishments(bazi),
        "officer_killing_cleaned": bool(guan_sha.get("qu_qing")),
        "wealth_tomb_present": bool(cai_ku.get("xian")),
        "wealth_star_in_tomb": bool(cai_ku.get("xing_ru_ku")),
        "spouse_palace_state": _spouse_palace_state(bazi),
        "day_branch_type": _day_branch_type(bazi),
        "year_officer_killing": _stem_ten_god(nian) in {"正官", "七杀"},
    }
