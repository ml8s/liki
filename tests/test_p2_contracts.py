"""P2 数据流契约：算子正交、求值复用、快照隔离与表结构闭包。"""
import copy

from unittest import mock

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
import duanyu
import factors
import operators_liunian
import operators_natal
from factor_tables import load_factor_rows, load_liunian_rows


def test_natal_and_liunian_operator_registries_are_orthogonal():
    assert not (operators_natal._OP_NAMES & operators_liunian._LIU_OP_NAMES)


def test_truth_table_evaluates_repeated_atomic_expression_once():
    rows = [
        {"因子": "印现", "术数": "bazi", "直通": "", "conds": {"现[印星]": 1}, "依据": ""},
        {"因子": "印旺", "术数": "bazi", "直通": "", "conds": {"现[印星]": 1}, "依据": ""},
    ]
    calls = []

    def atomic(expression):
        calls.append(expression)
        return 1

    result = factors._evaluate_truth_table(rows, atomic)
    assert result == {"印现": 1, "印旺": 1}
    assert calls == ["现[印星]"]


def test_production_factor_tables_keep_sides_disjoint():
    natal = load_factor_rows()
    flow = load_liunian_rows()
    for rows in (natal, flow):
        sides = {row["术数"] for row in rows}
        assert sides <= {"bazi", "ziwei"}
        bazi_ids = {row["因子"] for row in rows if row["术数"] == "bazi"}
        ziwei_ids = {row["因子"] for row in rows if row["术数"] == "ziwei"}
        assert not bazi_ids & ziwei_ids


def _pan() -> dict:
    pillars = ("nian", "yue", "ri", "shi")
    return {
        "solar": "1990-05-20T12:00:00",
        "lunar": {"year": 1990, "month": 4, "day": 26},
        "gender": "male",
        "chart": {
            **{pillar: {"gan": "甲", "zhi": "子"} for pillar in pillars},
            "da_yun": {"steps": [], "current_step_index": -1},
        },
        "full": {
            pillar: {
                "gan": "甲", "zhi": "子",
                "shi_shens": [{"shi_shen": "比肩", "gan": "甲", "source": "gan"}],
            } for pillar in pillars
        },
        "yongshen": {},
        "ziwei": {"gong_wei": []},
    }


def test_flow_evaluation_reuses_prepared_natal_base():
    pan = _pan()
    natal = factors.prepare_natal_context(pan)
    with mock.patch.object(
        operators_natal, "_base_ctx_from_pan", wraps=operators_natal._base_ctx_from_pan
    ) as build_base:
        for side in ("bazi", "ziwei"):
            result = factors.evaluate_liunian_factors(
                pan["gender"], natal.evaluation, {}, year=2026, shushi=side,
                natal_snapshot=natal.snapshot,
            )
            assert result
    build_base.assert_not_called()


def test_flow_snapshot_is_readonly_and_typed():
    pan = _pan()
    liunian_pan = {"bazi": {"nian_gan": "丙"}, "ziwei": {"nian_gan": "丙"}}
    pan_copy = copy.deepcopy(pan)
    flow_copy = copy.deepcopy(liunian_pan)
    snap = factors.evaluate_liunian_snap_from_pan(pan, liunian_pan, year=2026)
    assert snap["_snapshot_type"] == "liunian"
    assert pan == pan_copy
    assert liunian_pan == flow_copy


def test_natal_snapshot_has_no_flow_runtime_marker():
    pan = _pan()
    snap = factors.evaluate_snap_from_pan(pan)
    assert "_snapshot_type" not in snap
    assert set(snap) == {"八字", "紫微", "context"}


def test_rule_matching_does_not_mutate_snapshot_with_context():
    snapshot = {"八字": {"日主": "甲"}, "紫微": {}, "context": {"性别": "male"}}
    result = duanyu._match_rule("十神", snapshot)
    assert snapshot["八字"].keys() == {"日主"}
    assert all("性别" not in row for row in result["八字"])
