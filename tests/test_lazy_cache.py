"""求值上下文契约：单次求值复用基础聚合，且不写回公共 pan。"""
from unittest import mock

import pytest

import _helpers  # noqa: F401
import factors
import operators_natal


def _mk(ri_gan: str, ri_zhi: str) -> dict:
    return {
        "chart": {"nian": {"gan": "庚", "zhi": "午"}, "yue": {"gan": "壬", "zhi": "午"},
                  "ri": {"gan": ri_gan, "zhi": ri_zhi}, "shi": {"gan": "庚", "zhi": "午"}},
        "full": {"ri": {"gan": ri_gan, "zhi": ri_zhi}, "chang_sheng": []},
        "yongshen": {}, "gender": "male",
    }


def test_evaluate_snap_reuses_base_context_within_single_evaluation():
    pan = _mk("甲", "子")
    with mock.patch.object(
        operators_natal, "_shishen_from_pan", wraps=operators_natal._shishen_from_pan
    ) as aggregate:
        snap = factors.evaluate_snap_from_pan(pan)
        assert snap["八字"]["日主"] == "甲"
        aggregate.assert_called_once()


def test_evaluate_snap_does_not_mutate_pan():
    pan = _mk("甲", "子")
    factors.evaluate_snap_from_pan(pan)
    assert "_ctx" not in pan
    assert "_snap" not in pan


def test_context_isolates_different_pans():
    p1, p2 = _mk("甲", "子"), _mk("丙", "午")
    s1 = factors.evaluate_snap_from_pan(p1)
    s2 = factors.evaluate_snap_from_pan(p2)
    assert s1["八字"]["日主"] == "甲"
    assert s2["八字"]["日主"] == "丙"


def test_natal_context_is_explicit_and_readonly_view():
    pan = _mk("甲", "子")
    context = factors.prepare_natal_context(pan)
    assert isinstance(context, factors.NatalContext)
    assert isinstance(context.evaluation, factors.FactorContext)
    assert context.evaluation.pan is pan
    assert "_ctx" not in pan and "_snap" not in pan
    with pytest.raises(AttributeError):
        context.evaluation = factors._factor_context_from_pan(pan)
