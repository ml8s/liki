"""natal_chart 契约：快照、裁剪盘、半截盘全部显式报错。"""
from pathlib import Path

import _helpers  # noqa: F401
import app.natal.tools.charts as charts
import pytest
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart, validate_side_chart


def _pan(**changes):
    natal_chart = {
        "standard_time": "1990-05-20T12:00:00+08:00",
        "gregorian_time": "1990-05-20T12:00:00+08:00",
        "gender": "male",
        "chart": {
            **{p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
            "da_yun": {"steps": [], "current_step_index": 0},
        },
        "full": {
            **{p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
            **_helpers.mock_engine_facts(),
            **_helpers.mock_yongshen_fields(),
        },
        "ziwei": _helpers.mock_ziwei(),
        "ziwei_chart": {},
        "ziwei_daxian": _helpers.valid_daxian(),
    }
    natal_chart.update(changes)
    return with_natal_chart_digest(natal_chart)


def test_valid_pan_passes():
    validate_natal_chart(_pan(), action="test")


def test_pan_digest_rejects_tampering():
    natal_chart = _pan()
    natal_chart["gregorian_time"] = "2000-01-01T00:00:00"
    with pytest.raises(ValueError, match="digest mismatch"):
        validate_natal_chart(natal_chart, action="test")


def test_build_natal_chart_has_flat_yongshen_fields():
    source = (Path(__file__).parents[1] / "counsel/app/natal/tools/charts.py").read_text(encoding="utf-8")
    assert '"yongshen":' not in source
    assert '"yong_shen":' not in source
    assert '"full"' in source


@pytest.mark.parametrize("key", ["standard_time", "chart", "full", "ziwei", "ziwei_daxian", "gender"])
def test_missing_required_field_rejected(key):
    natal_chart = _pan()
    natal_chart.pop(key)
    with pytest.raises(ValueError, match=key):
        validate_natal_chart(natal_chart, action="test")


@pytest.mark.parametrize("value", ["", "MALE", "other", None, 1])
def test_invalid_gender_rejected(value):
    with pytest.raises(ValueError, match="gender"):
        validate_natal_chart(_pan(gender=value), action="test")


@pytest.mark.parametrize("pillar", ["nian", "yue", "ri", "shi"])
@pytest.mark.parametrize("mode", ["chart", "full", "gan", "zhi"])
def test_incomplete_pillar_rejected(pillar, mode):
    natal_chart = _pan()
    if mode == "chart":
        natal_chart["chart"].pop(pillar)
    elif mode == "full":
        natal_chart["full"].pop(pillar)
    else:
        del natal_chart["full"][pillar][mode]
    with pytest.raises(ValueError, match="四柱结构"):
        validate_natal_chart(natal_chart, action="test")


def test_da_yun_and_gong_wei_required():
    natal_chart = _pan()
    natal_chart["chart"].pop("da_yun")
    with pytest.raises(ValueError, match="da_yun"):
        validate_natal_chart(natal_chart, action="test")


def test_daxian_must_have_twelve_complete_steps():
    short = _helpers.valid_daxian()[:-1]
    with pytest.raises(ValueError, match="12 个大限段"):
        validate_natal_chart(_pan(ziwei_daxian=short), action="test")

    incomplete = _helpers.valid_daxian()
    incomplete[0].pop("start_year")
    with pytest.raises(ValueError, match="ziwei_daxian\\[0\\].*start_year"):
        validate_natal_chart(_pan(ziwei_daxian=incomplete), action="test")

    invalid_types = _helpers.valid_daxian()
    invalid_types[0]["start_year"] = "1990"
    with pytest.raises(ValueError, match="字段必须为整数.*start_year"):
        validate_natal_chart(_pan(ziwei_daxian=invalid_types), action="test")

    duplicate_palaces = _helpers.valid_daxian()
    duplicate_palaces[1]["gong"] = duplicate_palaces[0]["gong"]
    with pytest.raises(ValueError, match="12 个大限宫位必须唯一"):
        validate_natal_chart(_pan(ziwei_daxian=duplicate_palaces), action="test")
    natal_chart = _pan()
    natal_chart["ziwei"].pop("gong_wei")
    with pytest.raises(ValueError, match="gong_wei"):
        validate_natal_chart(natal_chart, action="test")


def test_ziwei_gong_wei_must_use_engine_palace_closure():
    palaces = [row["name"] for row in _helpers.mock_ziwei()["gong_wei"]]
    palaces[1] = "兄弟宫"
    ziwei = {**_helpers.mock_ziwei(), "gong_wei": [
        {"name": name} for name in palaces
    ]}
    with pytest.raises(ValueError, match="engine 宫位闭集"):
        validate_natal_chart(_pan(ziwei=ziwei), action="test")

    ziwei = _helpers.mock_ziwei()
    ziwei["palace_facts"] = []
    with pytest.raises(ValueError, match="palace_facts 不能为空"):
        validate_natal_chart(_pan(ziwei=ziwei), action="test")


def test_engine_fact_fields_required():
    natal_chart = _pan()
    natal_chart["full"].pop("ten_god_states")
    with pytest.raises(ValueError, match="ten_god_states"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"].pop("relation_groups")
    with pytest.raises(ValueError, match="relation_groups"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["atomic_facts"].pop("day_master_element")
    with pytest.raises(ValueError, match="day_master_element"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["atomic_facts"].pop("month_main_ten_god")
    with pytest.raises(ValueError, match="month_main_ten_god"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["da_yun"]["steps"][0].pop("rooted")
    with pytest.raises(ValueError, match="rooted engine"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["da_yun"]["steps"][0].update({"rooted": True})
    natal_chart["full"]["da_yun"]["steps"][0].pop("root_refs")
    with pytest.raises(ValueError, match="root_refs"):
        validate_natal_chart(natal_chart, action="test")


def test_consumed_engine_atomic_facts_are_required():
    for key in (
        "officer_killing_cleaned", "wealth_tomb_present", "wealth_star_in_tomb",
        "spouse_palace_state", "day_branch_type", "year_officer_killing",
    ):
        natal_chart = _pan()
        natal_chart["full"]["atomic_facts"].pop(key)
        with pytest.raises(ValueError, match=key):
            validate_natal_chart(natal_chart, action="test")


def test_yong_shen_structural_facts_are_required():
    natal_chart = _pan()
    natal_chart["full"]["fu_yi"].pop("basis")
    with pytest.raises(ValueError, match="fu_yi.basis"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["ge_ju"].pop("structure")
    with pytest.raises(ValueError, match="ge_ju.structure"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["tiao_hou"].pop("primary")
    with pytest.raises(ValueError, match="tiao_hou.primary"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["ge_ju"]["structure"]["relation_facts"] = [{"field": "liu_chong"}]
    with pytest.raises(ValueError, match="relation_facts\\[0\\].*group"):
        validate_natal_chart(natal_chart, action="test")


def test_engine_list_fact_item_shapes_are_required():
    natal_chart = _pan()
    natal_chart["full"]["relation_groups"] = [{"field": "", "group": "申子辰"}]
    with pytest.raises(ValueError, match="relation_groups.*field/group"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["ten_god_states"] = [{"shi_shen": "比肩"}]
    with pytest.raises(ValueError, match="ten_god_states.*字符串状态"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["full"]["element_states"][0]["strength"] = ""
    with pytest.raises(ValueError, match="element_states.*字符串状态"):
        validate_natal_chart(natal_chart, action="test")


def test_ziwei_palace_atomic_facts_are_required():
    natal_chart = _pan()
    natal_chart["ziwei"].pop("palace_facts")
    with pytest.raises(ValueError, match="palace_facts"):
        validate_natal_chart(natal_chart, action="test")

    natal_chart = _pan()
    natal_chart["ziwei"]["palace_facts"] = [{"palace": "命宫"}]
    with pytest.raises(ValueError, match="palace/kind/target"):
        validate_natal_chart(natal_chart, action="test")


def test_liunian_and_bond_validate_input():
    with pytest.raises(ValueError, match="liunian"):
        charts.liunian({}, 2026)
    with pytest.raises(ValueError, match="bond pan_a"):
        charts.bond({}, _pan())

def test_tiaohou_secondary_condition_requires_secondary_object():
    natal_chart = _pan()
    natal_chart["full"]["tiao_hou"]["secondary_condition"] = "壬多必取戊制"
    natal_chart["full"]["tiao_hou"].pop("secondary")
    with pytest.raises(ValueError, match="tiao_hou.secondary 缺失"):
        validate_natal_chart(natal_chart, action="test")


def test_tiaohou_secondary_condition_must_be_string():
    natal_chart = _pan()
    natal_chart["full"]["tiao_hou"]["secondary_condition"] = 1
    with pytest.raises(ValueError, match="secondary_condition 不是 string"):
        validate_natal_chart(natal_chart, action="test")


def test_tiaohou_conditional_secondary_passes():
    natal_chart = _pan()
    natal_chart["full"]["tiao_hou"]["secondary_condition"] = "壬多必取戊制"
    validate_natal_chart(with_natal_chart_digest(natal_chart), action="test")


def test_geju_rejects_legacy_yongxi_columns():
    natal_chart = _pan()
    natal_chart["full"]["ge_ju"].update({"yong": "土", "xi": "水", "ji": "火"})
    with pytest.raises(ValueError, match="旧用神字段"):
        validate_natal_chart(natal_chart, action="test")


def test_optional_candidate_values_use_explicit_domain_states_and_reject_empty():
    natal_chart = _pan()
    natal_chart["full"]["fu_yi"].update({
        "yong": "无候选", "xi": "无候选", "ji": "无候选",
    })
    natal_chart["full"]["tiao_hou"]["secondary_wuxing"] = "无辅调候"
    validate_natal_chart(with_natal_chart_digest(natal_chart), action="test")

    natal_chart = _pan()
    natal_chart["full"]["fu_yi"]["yong"] = None
    with pytest.raises(ValueError, match="fu_yi.yong"):
        validate_natal_chart(with_natal_chart_digest(natal_chart), action="test")

    for key in ("yong", "xi", "ji"):
        natal_chart = _pan()
        natal_chart["full"]["fu_yi"][key] = ""
        with pytest.raises(ValueError, match=f"fu_yi.{key}"):
            validate_natal_chart(with_natal_chart_digest(natal_chart), action="test")

    natal_chart = _pan()
    natal_chart["full"]["tiao_hou"]["secondary_wuxing"] = ""
    with pytest.raises(ValueError, match="secondary_wuxing"):
        validate_natal_chart(with_natal_chart_digest(natal_chart), action="test")


def _side_pan(side):
    base = _pan()
    if side == "bazi":
        return {"chart": base["chart"], "full": base["full"], "gender": "male"}
    return {
        "chart": base["chart"],
        "ziwei": base["ziwei"],
        "ziwei_daxian": base["ziwei_daxian"],
        "gender": "male",
    }


def test_side_pan_bazi_valid():
    validate_side_chart(_side_pan("bazi"), "bazi", action="test")


def test_side_pan_bazi_missing_pillar_fails_closed():
    natal_chart = _side_pan("bazi")
    del natal_chart["full"]["ri"]
    with pytest.raises(ValueError, match="四柱结构不完整"):
        validate_side_chart(natal_chart, "bazi", action="test")


def test_side_pan_bazi_bad_gender_fails_closed():
    natal_chart = _side_pan("bazi")
    natal_chart["gender"] = "other"
    with pytest.raises(ValueError, match="gender"):
        validate_side_chart(natal_chart, "bazi", action="test")


def test_side_pan_ziwei_valid():
    validate_side_chart(_side_pan("ziwei"), "ziwei", action="test")


def test_side_pan_ziwei_wrong_gong_count_fails_closed():
    natal_chart = _side_pan("ziwei")
    natal_chart["ziwei"] = dict(natal_chart["ziwei"], gong_wei=natal_chart["ziwei"]["gong_wei"][:11])
    with pytest.raises(ValueError, match="12 宫"):
        validate_side_chart(natal_chart, "ziwei", action="test")
