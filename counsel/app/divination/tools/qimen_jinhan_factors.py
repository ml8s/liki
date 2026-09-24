"""金函玉镜因子投影；不与标准奇门因子混用。"""
from __future__ import annotations


SCHOOL = "jinhan_yujing"
SCOPE = "day"


def _palace(row: dict) -> dict:
    identity = row.get("gong") or {}
    if not isinstance(identity, dict):
        raise ValueError("jinhan palace identity must be an object")
    gong = identity.get("name")
    luoshu = identity.get("luoshu")
    if not isinstance(gong, str) or not gong:
        raise ValueError("jinhan palace lacks gong")
    if isinstance(luoshu, bool) or not isinstance(luoshu, int) or not 1 <= luoshu <= 9:
        raise ValueError("jinhan palace lacks valid luoshu")
    men = row.get("men")
    return {
        "gong": gong,
        "luoshu": luoshu,
        "xing": row.get("xing"),
        "men": men if row.get("men_present") else None,
        "men_present": bool(row.get("men_present")),
    }


def _day_spirit(row: dict) -> dict:
    zhi = row.get("zhi")
    shen = row.get("shen")
    if not isinstance(zhi, str) or not zhi:
        raise ValueError("jinhan day spirit lacks zhi")
    if not isinstance(shen, str) or not shen:
        raise ValueError("jinhan day spirit lacks shen")
    return {"zhi": zhi, "shen": shen}


def project(chart: dict) -> dict:
    """把金函玉镜独立盘型投影为稳定 factors，不混入标准奇门因子。"""
    if not isinstance(chart, dict) or chart.get("method", {}).get("school") != SCHOOL:
        raise ValueError("chart must be jinhan_yujing")
    pan = chart.get("pan")
    if not isinstance(pan, dict):
        raise ValueError("jinhan chart lacks pan")

    palaces = [_palace(row) for row in pan.get("gong_wei", [])]
    if len(palaces) != 9:
        raise ValueError("jinhan pan must contain exactly 9 palaces")
    day_spirits = [_day_spirit(row) for row in pan.get("day_spirits", [])]
    if len(day_spirits) != 12:
        raise ValueError("jinhan pan must contain exactly 12 day spirits")

    return {
        "kind": "jinhan",
        "school": SCHOOL,
        "scope": SCOPE,
        "yin_dun": bool(pan.get("yin_dun")),
        "ri_gan": pan.get("ri_gan"),
        "ri_zhi": pan.get("ri_zhi"),
        "palaces": palaces,
        "day_spirits": day_spirits,
    }


def validate(factors: dict) -> None:
    expected = {
        "kind", "school", "scope", "yin_dun", "ri_gan", "ri_zhi",
        "palaces", "day_spirits",
    }
    if not isinstance(factors, dict):
        raise ValueError("jinhan factors must be an object")
    if set(factors) != expected:
        missing = sorted(expected - set(factors))
        extra = sorted(set(factors) - expected)
        raise ValueError(f"jinhan factors invalid: missing={missing}, extra={extra}")
    if factors.get("kind") != "jinhan" or factors.get("school") != SCHOOL:
        raise ValueError("factors must be jinhan_yujing")
    if factors.get("scope") != SCOPE:
        raise ValueError("jinhan factors require scope=day")
    if not isinstance(factors.get("palaces"), list) or len(factors["palaces"]) != 9:
        raise ValueError("jinhan factors require 9 palaces")
    if not isinstance(factors.get("day_spirits"), list) or len(factors["day_spirits"]) != 12:
        raise ValueError("jinhan factors require 12 day spirits")
