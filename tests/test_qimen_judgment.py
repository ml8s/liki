"""qimen 通用吉凶判断（qimen_judgment）单元测试：不依赖外部服务。"""
from __future__ import annotations

import sys
from pathlib import Path

TOOLS = Path(__file__).resolve().parents[1] / "counsel/app/divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import qimen_judgment  # noqa: E402


def _factors(*, duty_door="死门", duty_gong="离", door_ji="大凶",
             kong_gongs=("离",), star_gongs=(("离", "天冲", "平"),),
             patterns=()):
    return {
        "duty_door": duty_door,
        "zhi_shi_men_gong": duty_gong,
        "doors": [{"door": duty_door, "gong": duty_gong, "ji_xiong": door_ji}],
        "kong_wang": [
            {"symbol": "值使", "role": "pillar", "branch": "午", "gong": gong}
            for gong in kong_gongs
        ],
        "stars": [
            {"star": star, "gong": gong, "ji_xiong": ji}
            for gong, star, ji in star_gongs
        ],
        "patterns": [{"name": name, "gong": duty_gong} for name in patterns],
    }


def test_big_failure_hits_when_dead_door_duty_and_void_and_bad_star():
    result = qimen_judgment.evaluate(_factors())
    item = next(rule for rule in result["results"] if rule["id"] == "big-failure-rout")
    assert item["applies"] is True
    assert "溃败级" in item["conclusion"]
    assert item["forbidden"]


def test_big_failure_hits_when_pattern_present():
    result = qimen_judgment.evaluate(_factors(patterns=("太白入荧",)))
    item = next(rule for rule in result["results"] if rule["id"] == "big-failure-rout")
    assert item["applies"] is True


def test_big_failure_not_hit_when_duty_gong_not_void():
    result = qimen_judgment.evaluate(_factors(kong_gongs=("乾",)))
    item = next(rule for rule in result["results"] if rule["id"] == "big-failure-rout")
    assert item["applies"] is False


def test_big_failure_not_hit_when_duty_door_not_dead():
    factors = _factors(duty_door="开门", door_ji="大吉",
                       star_gongs=(("离", "天心", "吉"),))
    result = qimen_judgment.evaluate(factors)
    item = next(rule for rule in result["results"] if rule["id"] == "big-failure-rout")
    assert item["applies"] is False


def test_big_failure_not_hit_without_bad_star_or_pattern():
    result = qimen_judgment.evaluate(_factors(star_gongs=(("离", "天心", "吉"),)))
    item = next(rule for rule in result["results"] if rule["id"] == "big-failure-rout")
    assert item["applies"] is False


def test_rules_all_are_candidates_not_verdicts():
    result = qimen_judgment.evaluate(_factors())
    assert result["policy"]["candidate_not_absolute_verdict"] is True
    assert all(
        rule["conclusion_scope"] == "candidate_not_absolute_verdict"
        for rule in result["results"]
    )


def test_empty_result_shape():
    result = qimen_judgment.empty()
    assert result["results"] == []
    assert result["schema_version"] == qimen_judgment.load_rules()["schema_version"]