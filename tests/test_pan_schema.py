"""pan 契约：快照、裁剪盘、半截盘全部显式报错。"""
import pytest

import _helpers  # noqa: F401

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
        "full": {p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
        "yongshen": {},
        "ziwei": {"gong_wei": []},
    }
    pan.update(changes)
    return pan


def test_valid_pan_passes():
    validate_natal_pan(_pan(), action="test")


@pytest.mark.parametrize("key", ["solar", "lunar", "chart", "full", "yongshen", "ziwei", "gender"])
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
    pan = _pan(); pan["ziwei"].pop("gong_wei")
    with pytest.raises(ValueError, match="gong_wei"):
        validate_natal_pan(pan, action="test")


def test_liunian_and_bond_validate_input():
    with pytest.raises(ValueError, match="liunian"):
        paipan.liunian({}, 2026)
    with pytest.raises(ValueError, match="bond pan_a"):
        paipan.bond({}, _pan())
