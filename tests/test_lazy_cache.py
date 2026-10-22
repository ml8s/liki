"""求值上下文契约：单次求值复用基础聚合，且不写回公共 natal_chart。"""
from unittest import mock

import app.natal.tools.factors as factors
import app.natal.tools.operators_natal as operators_natal
import pytest


def _mk(ri_gan: str, ri_zhi: str) -> dict:
    return {
        "chart": {"nian": {"gan": "庚", "zhi": "午"}, "yue": {"gan": "壬", "zhi": "午"},
                  "ri": {"gan": ri_gan, "zhi": ri_zhi}, "shi": {"gan": "庚", "zhi": "午"}},
        "full": {
            "ri": {"gan": ri_gan, "zhi": ri_zhi},
            "chang_sheng": [],
            "atomic_facts": {"day_master_stem": ri_gan, "day_branch": ri_zhi},
        },
        "fu_yi": {}, "gender": "male",
    }


def test_evaluate_snap_reuses_base_context_within_single_evaluation():
    natal_chart = _mk("甲", "子")
    with mock.patch.object(
        operators_natal, "_ten_god_states_from_natal_chart", wraps=operators_natal._ten_god_states_from_natal_chart
    ) as aggregate:
        snap = factors.evaluate_snap_from_natal_chart(natal_chart)
        assert snap["八字"]["日主"] == "甲"
        aggregate.assert_called_once()


def test_evaluate_snap_does_not_mutate_pan():
    natal_chart = _mk("甲", "子")
    factors.evaluate_snap_from_natal_chart(natal_chart)
    assert "_ctx" not in natal_chart
    assert "_snap" not in natal_chart


def test_context_isolates_different_pans():
    p1, p2 = _mk("甲", "子"), _mk("丙", "午")
    s1 = factors.evaluate_snap_from_natal_chart(p1)
    s2 = factors.evaluate_snap_from_natal_chart(p2)
    assert s1["八字"]["日主"] == "甲"
    assert s2["八字"]["日主"] == "丙"


def test_natal_context_is_explicit_and_readonly_view():
    natal_chart = _mk("甲", "子")
    context = factors.prepare_natal_context(natal_chart)
    assert isinstance(context, factors.NatalContext)
    assert isinstance(context.evaluation, factors.FactorContext)
    assert context.evaluation.natal_chart is natal_chart
    assert "_ctx" not in natal_chart and "_snap" not in natal_chart
    with pytest.raises(AttributeError):
        context.evaluation = factors._factor_context_from_natal_chart(natal_chart)
