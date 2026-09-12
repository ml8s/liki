"""pan 契约：快照、裁剪盘、半截盘全部显式报错。"""
from pathlib import Path

import pytest

import _helpers  # noqa: F401

from pan_integrity import with_natal_digest
from pan_schema import validate_natal_pan
import paipan


def _pan(**changes):
    pan = {
        "solar": "1990-05-20T12:00:00",
        "lunar": {"year": 1990, "month": 4, "day": 26},
        "gender": "male",
        "chart": {
            **{p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
            "da_yun": {"steps": [], "current_step_index": 0},
        },
        "full": {
            **{p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
            **_helpers.mock_engine_facts(),
            "yong_shen": _helpers.mock_yong_shen(),
        },
        "ziwei": _helpers.mock_ziwei(),
        "ziwei_daxian": _helpers.valid_daxian(),
    }
    pan.update(changes)
    return with_natal_digest(pan)


def test_valid_pan_passes():
    validate_natal_pan(_pan(), action="test")


def test_pan_digest_rejects_tampering():
    pan = _pan()
    pan["solar"] = "2000-01-01T00:00:00"
    with pytest.raises(ValueError, match="digest mismatch"):
        validate_natal_pan(pan, action="test")


def test_full_paipan_has_single_yong_shen_path():
    source = (Path(__file__).parents[1] / "skills/liki-bazi/tools/paipan.py").read_text(encoding="utf-8")
    assert '"yongshen":' not in source
    assert '"full"' in source


@pytest.mark.parametrize("key", ["solar", "lunar", "chart", "full", "ziwei", "ziwei_daxian", "gender"])
def test_missing_required_field_rejected(key):
    pan = _pan(); pan.pop(key)
    with pytest.raises(ValueError, match=key):
        validate_natal_pan(pan, action="test")


@pytest.mark.parametrize("value", ["", "MALE", "other", None, 1])
def test_invalid_gender_rejected(value):
    with pytest.raises(ValueError, match="gender"):
        validate_natal_pan(_pan(gender=value), action="test")


@pytest.mark.parametrize("pillar", ["nian", "yue", "ri", "shi"])
@pytest.mark.parametrize("mode", ["chart", "full", "gan", "zhi"])
def test_incomplete_pillar_rejected(pillar, mode):
    pan = _pan()
    if mode == "chart":
        pan["chart"].pop(pillar)
    elif mode == "full":
        pan["full"].pop(pillar)
    else:
        del pan["full"][pillar][mode]
    with pytest.raises(ValueError, match="四柱结构"):
        validate_natal_pan(pan, action="test")


def test_da_yun_and_gong_wei_required():
    pan = _pan(); pan["chart"].pop("da_yun")
    with pytest.raises(ValueError, match="da_yun"):
        validate_natal_pan(pan, action="test")


def test_daxian_must_have_twelve_complete_steps():
    short = _helpers.valid_daxian()[:-1]
    with pytest.raises(ValueError, match="12 个大限段"):
        validate_natal_pan(_pan(ziwei_daxian=short), action="test")

    incomplete = _helpers.valid_daxian()
    incomplete[0].pop("start_year")
    with pytest.raises(ValueError, match="ziwei_daxian\\[0\\].*start_year"):
        validate_natal_pan(_pan(ziwei_daxian=incomplete), action="test")

    invalid_types = _helpers.valid_daxian()
    invalid_types[0]["start_year"] = "1990"
    with pytest.raises(ValueError, match="字段必须为整数.*start_year"):
        validate_natal_pan(_pan(ziwei_daxian=invalid_types), action="test")

    duplicate_palaces = _helpers.valid_daxian()
    duplicate_palaces[1]["gong"] = duplicate_palaces[0]["gong"]
    with pytest.raises(ValueError, match="12 个大限宫位必须唯一"):
        validate_natal_pan(_pan(ziwei_daxian=duplicate_palaces), action="test")
    pan = _pan(); pan["ziwei"].pop("gong_wei")
    with pytest.raises(ValueError, match="gong_wei"):
        validate_natal_pan(pan, action="test")


def test_ziwei_gong_wei_must_use_engine_palace_closure():
    palaces = [row["name"] for row in _helpers.mock_ziwei()["gong_wei"]]
    palaces[1] = "兄弟宫"
    ziwei = {**_helpers.mock_ziwei(), "gong_wei": [
        {"name": name} for name in palaces
    ]}
    with pytest.raises(ValueError, match="engine 宫位闭集"):
        validate_natal_pan(_pan(ziwei=ziwei), action="test")

    ziwei = _helpers.mock_ziwei()
    ziwei["palace_facts"] = []
    with pytest.raises(ValueError, match="palace_facts 不能为空"):
        validate_natal_pan(_pan(ziwei=ziwei), action="test")


def test_engine_fact_fields_required():
    pan = _pan(); pan["full"].pop("ten_god_states")
    with pytest.raises(ValueError, match="ten_god_states"):
        validate_natal_pan(pan, action="test")

    pan = _pan(); pan["full"].pop("relation_groups")
    with pytest.raises(ValueError, match="relation_groups"):
        validate_natal_pan(pan, action="test")

    pan = _pan(); pan["full"]["atomic_facts"].pop("day_master_element")
    with pytest.raises(ValueError, match="day_master_element"):
        validate_natal_pan(pan, action="test")

    pan = _pan(); pan["full"]["atomic_facts"].pop("month_main_ten_god")
    with pytest.raises(ValueError, match="month_main_ten_god"):
        validate_natal_pan(pan, action="test")

    pan = _pan(); pan["full"]["da_yun"]["steps"][0].pop("rooted")
    with pytest.raises(ValueError, match="rooted engine"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["full"]["da_yun"]["steps"][0].update({"rooted": True})
    pan["full"]["da_yun"]["steps"][0].pop("root_refs")
    with pytest.raises(ValueError, match="root_refs"):
        validate_natal_pan(pan, action="test")


def test_consumed_engine_atomic_facts_are_required():
    for key in (
        "officer_killing_cleaned", "wealth_tomb_present", "wealth_star_in_tomb",
        "spouse_palace_state", "day_branch_type", "year_officer_killing",
    ):
        pan = _pan()
        pan["full"]["atomic_facts"].pop(key)
        with pytest.raises(ValueError, match=key):
            validate_natal_pan(pan, action="test")


def test_yong_shen_structural_facts_are_required():
    pan = _pan()
    pan["full"]["yong_shen"]["fu_yi"].pop("basis")
    with pytest.raises(ValueError, match="fu_yi.basis"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["full"]["yong_shen"]["ge_ju"].pop("structure")
    with pytest.raises(ValueError, match="ge_ju.structure"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["full"]["yong_shen"]["tiao_hou"].pop("primary")
    with pytest.raises(ValueError, match="tiao_hou.primary"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["full"]["yong_shen"]["ge_ju"]["structure"]["relation_facts"] = [{"field": "liu_chong"}]
    with pytest.raises(ValueError, match="relation_facts\\[0\\].*group"):
        validate_natal_pan(pan, action="test")


def test_engine_list_fact_item_shapes_are_required():
    pan = _pan()
    pan["full"]["relation_groups"] = [{"field": "", "group": "申子辰"}]
    with pytest.raises(ValueError, match="relation_groups.*field/group"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["full"]["ten_god_states"] = [{"shi_shen": "比肩"}]
    with pytest.raises(ValueError, match="ten_god_states.*字符串状态"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["full"]["element_states"][0]["strength"] = ""
    with pytest.raises(ValueError, match="element_states.*字符串状态"):
        validate_natal_pan(pan, action="test")


def test_ziwei_palace_atomic_facts_are_required():
    pan = _pan()
    pan["ziwei"].pop("palace_facts")
    with pytest.raises(ValueError, match="palace_facts"):
        validate_natal_pan(pan, action="test")

    pan = _pan()
    pan["ziwei"]["palace_facts"] = [{"palace": "命宫"}]
    with pytest.raises(ValueError, match="palace/kind/target"):
        validate_natal_pan(pan, action="test")


def test_liunian_and_bond_validate_input():
    with pytest.raises(ValueError, match="liunian"):
        paipan.liunian({}, 2026)
    with pytest.raises(ValueError, match="bond pan_a"):
        paipan.bond({}, _pan())
