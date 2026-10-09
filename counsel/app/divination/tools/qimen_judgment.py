"""奇门通用吉凶判断：凶象叠加等组合判断由表驱动，不留在 skill 文档。

条件语义（qimen_judgment_rules.json）：
- conditions.all：全部满足才命中
- conditions.any：任一满足即可（在 all 已满足的前提下）
条件类型：
- door_ji_xiong_duty：指定门（大凶）为值使门，且门吉凶命中
- duty_door_gong_kong_wang：值使门落宫为空亡宫
- gong_has_xing_ji_xiong：值使门落宫内有指定吉凶的星
- gong_has_pattern：值使门落宫命中指定格局
"""
from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path

PATH = Path(__file__).with_name("qimen_judgment_rules.json")


@lru_cache
def load_rules() -> dict:
    return json.loads(PATH.read_text(encoding="utf-8"))


def _doors(factors: dict) -> list[dict]:
    return factors.get("doors") or []


def _stars(factors: dict) -> list[dict]:
    return factors.get("stars") or []


def _kong_wang(factors: dict) -> list[dict]:
    return factors.get("kong_wang") or []


def _patterns(factors: dict) -> list[dict]:
    return factors.get("patterns") or []


def _duty_door_gong(factors: dict) -> str:
    return factors.get("zhi_shi_men_gong") or ""


def _match_condition(factors: dict, cond: dict) -> bool:
    kind = cond.get("kind")
    if kind == "door_ji_xiong_duty":
        door = cond.get("door")
        want_ji = cond.get("ji_xiong")
        if factors.get("duty_door") != door:
            return False
        return any(
            item.get("door") == door and item.get("ji_xiong") == want_ji
            for item in _doors(factors)
        )
    if kind == "duty_door_gong_kong_wang":
        gong = _duty_door_gong(factors)
        if not gong:
            return False
        return any(item.get("gong") == gong for item in _kong_wang(factors))
    if kind == "gong_has_xing_ji_xiong":
        gong = _duty_door_gong(factors)
        want = set(cond.get("ji_xiong_in") or [])
        return any(
            item.get("gong") == gong and item.get("ji_xiong") in want
            for item in _stars(factors)
        )
    if kind == "gong_has_pattern":
        gong = _duty_door_gong(factors)
        want = set(cond.get("pattern_in") or [])
        return any(
            item.get("name") in want
            for item in _patterns(factors)
            if not gong or item.get("gong") == gong
        )
    return False


def empty() -> dict:
    """无命中断语的空结果（金函玉镜等无通用判断的流派）。"""
    return {
        "schema_version": load_rules()["schema_version"],
        "results": [],
        "policy": {
            "candidate_not_absolute_verdict": True,
            "forbid_downscaling_obvious_big_failure": True,
        },
    }



def evaluate(factors: dict) -> dict:
    """输入 qimen 标准 factors，输出命中的通用判断候选。"""
    results = []
    for rule in load_rules().get("rules", []):
        conditions = rule.get("conditions", {})
        all_conds = conditions.get("all", [])
        any_conds = conditions.get("any", [])
        all_ok = all(_match_condition(factors, cond) for cond in all_conds)
        any_ok = any(_match_condition(factors, cond) for cond in any_conds) if any_conds else True
        applies = all_ok and any_ok
        results.append({
            "id": rule["id"],
            "name": rule["name"],
            "statement": rule.get("statement", ""),
            "conclusion": rule.get("conclusion", "") if applies else "",
            "basis": rule.get("basis", ""),
            "applies": applies,
            "forbidden": rule.get("forbidden", []),
            "conclusion_scope": load_rules().get("conclusion_scope", ""),
        })
    return {
        "schema_version": load_rules()["schema_version"],
        "results": results,
        "policy": {
            "candidate_not_absolute_verdict": True,
            "forbid_downscaling_obvious_big_failure": True,
        },
    }
