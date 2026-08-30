"""第二批因子正确性回归：柱位、取清前提与流年目标五行推导。"""
from unittest import mock

from _helpers import mock_factors
from factors import _target_stars, load_constants
import factors
from factors import _liu_op, _op


def _chart(pillars: dict) -> dict:
    return {
        "chart": {
            name: {"gan": value.get("gan", ""), "zhi": value.get("zhi", "")}
            for name, value in pillars.items()
        },
        "full": pillars,
    }


def test_wealth_in_tomb_requires_wealth_star_on_tomb_pillar() -> None:
    fac = mock_factors()
    fac["ri_gan"] = "甲"

    # 甲以土为财，土库在辰。年支虽见辰，但年柱无财星；财星在月支午。
    unrelated_tomb = _chart({
        "nian": {
            "gan": "甲",
            "zhi": "辰",
            "shi_shens": [
                {"shi_shen": "比肩", "gan": "甲", "source": "gan"}
            ],
        },
        "yue": {
            "gan": "己",
            "zhi": "午",
            "shi_shens": [
                {"shi_shen": "正财", "gan": "己", "source": "gan"}
            ],
        },
    })
    assert _op("财星入墓", [], fac, "male", unrelated_tomb) == 0

    # 月支辰中土气为甲之财，财星与墓库同柱，才构成财星入墓。
    in_tomb = _chart({
        "yue": {
            "gan": "甲",
            "zhi": "辰",
            "shi_shens": [
                {"shi_shen": "偏财", "gan": "戊", "source": "main_qi"}
            ],
        },
    })
    assert _op("财星入墓", [], fac, "male", in_tomb) == 1


def test_guansha_purification_requires_mixed_guansha() -> None:
    fac = mock_factors()
    fac["ri_gan"] = "甲"  # 甲以金为官杀；庚为七杀，辛为正官。

    one_guan_only = _chart({
        "nian": {
            "gan": "庚",
            "zhi": "申",
            "shi_shens": [
                {"shi_shen": "七杀", "gan": "庚", "source": "gan"}
            ],
        },
        "ri": {"gan": "甲", "zhi": "子", "shi_shens": []},
    })
    one_guan_only["full"]["zhi_liu_he"] = [
        {"pillar_a": 0, "pillar_b": 2, "zhi_a": "申", "zhi_b": "巳"}
    ]
    assert _op("官杀取清", [], fac, "male", one_guan_only) == 0

    mixed_guansha = _chart({
        "nian": {
            "gan": "庚",
            "zhi": "申",
            "shi_shens": [
                {"shi_shen": "七杀", "gan": "庚", "source": "gan"}
            ],
        },
        "yue": {
            "gan": "辛",
            "zhi": "酉",
            "shi_shens": [
                {"shi_shen": "正官", "gan": "辛", "source": "gan"}
            ],
        },
        "ri": {"gan": "甲", "zhi": "子", "shi_shens": []},
    })
    mixed_guansha["full"]["zhi_liu_he"] = [
        {"pillar_a": 0, "pillar_b": 2, "zhi_a": "申", "zhi_b": "巳"}
    ]
    assert _op("官杀取清", [], fac, "male", mixed_guansha) == 1

    # 官杀混杂后，取清方向是合杀/冲杀留官；仅合正官不构成「合杀留官」。
    combining_correct_officer = _chart({
        "nian": {
            "gan": "庚",
            "zhi": "申",
            "shi_shens": [
                {"shi_shen": "七杀", "gan": "庚", "source": "gan"}
            ],
        },
        "yue": {
            "gan": "辛",
            "zhi": "酉",
            "shi_shens": [
                {"shi_shen": "正官", "gan": "辛", "source": "gan"}
            ],
        },
        "ri": {"gan": "甲", "zhi": "子", "shi_shens": []},
        "shi": {"gan": "丙", "zhi": "巳", "shi_shens": []},
    })
    combining_correct_officer["full"]["zhi_liu_he"] = [
        {"pillar_a": 1, "pillar_b": 3, "zhi_a": "酉", "zhi_b": "辰"}
    ]
    assert _op("官杀取清", [], fac, "male", combining_correct_officer) == 0


def test_flow克_derives_target_wuxing_from_day_master() -> None:
    fac = mock_factors()
    fac["ri_gan"] = "甲"  # 甲木以土为财；庚/申金克土。

    ctx = {"liunian": {"nian_gan": "甲", "nian_zhi": "寅"}}
    assert _liu_op("流年克", ["财星"], fac, "male", {}, ctx) == 1
    assert _liu_op("流年克", ["配偶星"], fac, "male", {}, ctx) == 1

    ctx_water = {"liunian": {"nian_gan": "壬", "nian_zhi": "子"}}
    assert _liu_op("流年克", ["财星"], fac, "male", {}, ctx_water) == 0


def test_flow_truth_table_consumes_common_factors() -> None:
    rows = [{
        "因子": "测试共同因子",
        "术数": "common",
        "直通": "直读[gender,male]",
        "conds": {},
    }]
    with mock.patch.object(factors, "load_liunian_rows", return_value=rows):
        snap = factors.evaluate_liunian_factors(
            mock_factors(),
            "male",
            {},
            {},
            year=2026,
            shushi="bazi",
        )

    assert snap["测试共同因子"] == 1


def test_parent_palace_is_year_pillar() -> None:
    fac = mock_factors()
    fac["ri_gan"] = "甲"
    chart = {
        "chart": {
            "nian": {"gan": "壬", "zhi": "戌"},
            "yue": {"gan": "庚", "zhi": "申"},
            "ri": {"gan": "甲", "zhi": "子"},
            "shi": {"gan": "甲", "zhi": "子"},
        },
    }
    ctx = {"liunian": {"nian_gan": "壬", "nian_zhi": "戌"}}

    # 父宫/母宫在年支；流年值年支戌必须命中，而不是错读月支申。
    assert _liu_op("流年值", ["父星"], fac, "male", chart, ctx) == 1
    assert _liu_op("流年值", ["母星"], fac, "male", chart, ctx) == 1


def test_mother_star_is_direct_seal_not_the_broad_seal_class() -> None:
    const = load_constants()
    assert const["六亲角色"]["母星"] == "正印"
    assert _target_stars("母星", "male", const) == ("正印",)


def test_flow_sanhe_requires_all_three_branches() -> None:
    fac = mock_factors()
    fac["ri_gan"] = "甲"
    fac["palace_ri"] = {"zhi": "子"}

    def chart(has_third: bool):
        return {
            "chart": {
                "nian": {"gan": "甲", "zhi": "辰" if has_third else "午"},
                "yue": {"gan": "丙", "zhi": "午"},
                "ri": {"gan": "甲", "zhi": "子"},
                "shi": {"gan": "甲", "zhi": "子"},
            },
        }

    ctx = {
        "liunian": {
            "nian_zhi": "申",
            "natal_interactions": [
                {
                    "zhi_rels": [
                        {"zhi_a": "申", "zhi_b": "子", "type": "三合"}
                    ]
                }
            ],
        }
    }

    # 申子辰三方齐备才成局。
    assert _liu_op("流年合", ["配偶星"], fac, "male", chart(True), ctx) == 1
    # 申子两支只是半合，不作为完整三合合会。
    assert _liu_op("流年合", ["配偶星"], fac, "male", chart(False), ctx) == 0
