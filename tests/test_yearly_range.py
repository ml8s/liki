"""yearly_range 编排契约：同一年份的流年快照只生成一次。"""
from unittest import mock

import pytest

from pathlib import Path
import _helpers  # noqa: F401 —— 注入 tools 路径
import duanyu
from yearly_eval import query_year_rules
from pan_integrity import with_natal_digest


def _valid_mock_pan() -> dict:
    return with_natal_digest({
        "solar": "1990-05-20T12:00:00",
        "lunar": {"year": 1990, "month": 4, "day": 26},
        "gender": "male",
        "chart": {
            **{
                pillar: {"gan": "甲", "zhi": "子"}
                for pillar in ("nian", "yue", "ri", "shi")
            },
            "da_yun": {"steps": [], "current_step_index": -1},
        },
        "full": {
            **{pillar: {"gan": "甲", "zhi": "子"} for pillar in ("nian", "yue", "ri", "shi")},
            **_helpers.mock_engine_facts(),
            **_helpers.mock_yongshen_fields(),
        },
        "fu_yi": {}, "tiao_hou": {}, "ge_ju": {},
        "ziwei": _helpers.mock_ziwei(),
        "ziwei_daxian": _helpers.valid_daxian(),
    })


def test_yearly_range_builds_one_snapshot_per_year() -> None:
    flow_snapshot = {
        "_snapshot_type": "liunian",
        "八字": {},
        "紫微": {},
        "context": {},
    }
    liunian_pan = {"bazi": {"nian_gan": "丙"}, "ziwei": {}}

    with mock.patch.object(duanyu, "resolve_current_year", return_value=(2026, "server")), \
         mock.patch("paipan.liunian", return_value=liunian_pan) as liunian_mock, \
         mock.patch.object(
             duanyu,
             "evaluate_liunian_snap_from_pan",
             return_value=flow_snapshot,
         ) as make_snapshot, \
         mock.patch.object(
             duanyu,
             "query_yearly",
            return_value={"八字": [], "紫微": [], "合参": []},
         ):
        result = duanyu.yearly_range(
            _valid_mock_pan(),
            2026,
            2026,
            rules=["年十神", "年合会"],
        )

    assert result["years"]["2026"]["年十神"] == {"八字": [], "紫微": [], "合参": []}
    liunian_mock.assert_called_once()
    make_snapshot.assert_called_once()
    assert make_snapshot.call_args.kwargs["factor_names"]


def test_detail_year_rules_do_not_restore_snapshot_evidence() -> None:
    """evidence 旁路退役；解释链只允许经由断语 trace 承载。"""
    detail_row = {
        "id": "ying_h09",
        "结论": "三刑成立",
        "trace": [{"condition_group": 1, "factors": {
            "流年地支相刑寅巳申": {"expected": 1, "actual": 1},
        }}],
    }
    snapshot = {
        "_snapshot_type": "liunian",
        "evidence": {"三刑流年": {"group": "寅巳申"}},
        "八字": {},
        "紫微": {},
    }

    result = query_year_rules(
        snapshot,
        ["年合会"],
        detail=True,
        query_yearly=lambda _rule, _snapshot: {"八字": [detail_row], "紫微": [], "合参": []},
        brief=lambda rows: rows,
    )

    assert result["年合会"]["八字"][0]["trace"][0]["factors"] == {
        "流年地支相刑寅巳申": {"expected": 1, "actual": 1},
    }
    assert "evidence" not in result["年合会"]


def test_yearly_range_rejects_empty_or_reversed_range() -> None:
    with mock.patch.object(duanyu, "resolve_current_year", return_value=(2026, "server")):
        with mock.patch.object(duanyu, "evaluate_liunian_snap_from_pan") as make_snapshot:
            with mock.patch.object(duanyu, "query_yearly"):
                with pytest.raises(ValueError, match="rules 不能为空"):
                    duanyu.yearly_range(_valid_mock_pan(), 2026, 2026, rules=None)
                with pytest.raises(ValueError, match="rules 不能为空"):
                    duanyu.yearly_range(_valid_mock_pan(), 2026, 2026, rules=[])
                with pytest.raises(ValueError, match="rules 不能有重复域"):
                    duanyu.yearly_range(
                        _valid_mock_pan(), 2026, 2026,
                        rules=["yearly_career", "yearly_career"],
                    )
                with pytest.raises(ValueError, match="start 不能大于 end"):
                    duanyu.yearly_range(_valid_mock_pan(), 2027, 2026, rules=["yingqi"])

    make_snapshot.assert_not_called()


def test_yearly_range_rejects_oversized_span() -> None:
    with mock.patch.object(duanyu, "resolve_current_year", return_value=(2026, "server")), \
         mock.patch.object(duanyu, "evaluate_liunian_snap_from_pan") as make_snapshot, \
         mock.patch.object(duanyu, "query_yearly"):
        with pytest.raises(ValueError, match="单次最多 120 年"):
            duanyu.yearly_range(_valid_mock_pan(), 1900, 2100, rules=["yingqi"])

    make_snapshot.assert_not_called()


def test_yearly_range_rejects_incomplete_pan() -> None:
    with mock.patch.object(duanyu, "resolve_current_year", return_value=(2026, "server")), \
         mock.patch.object(duanyu, "evaluate_liunian_snap_from_pan") as make_snapshot, \
         mock.patch.object(duanyu, "query_yearly"):
        with pytest.raises(ValueError, match="完整本命盘"):
            duanyu.yearly_range({"gender": "male"}, 2026, 2026, rules=["yingqi"])

    make_snapshot.assert_not_called()


def test_yearly_range_prepares_natal_context_once() -> None:
    pan = _valid_mock_pan()
    natal_context = {"evaluation": object(), "snapshot": {}}
    flow_snapshot = {"_snapshot_type": "liunian", "八字": {}, "紫微": {}, "context": {}}
    liunian_pan = {"bazi": {}, "ziwei": {}}

    with mock.patch.object(duanyu, "resolve_current_year", return_value=(2026, "server")), \
         mock.patch.object(duanyu, "prepare_natal_context", return_value=natal_context) as prepare, \
         mock.patch("paipan.liunian", return_value=liunian_pan), \
         mock.patch.object(
             duanyu, "evaluate_liunian_snap_from_pan", return_value=flow_snapshot
         ) as make_snapshot, \
         mock.patch.object(duanyu, "query_yearly", return_value={"八字": [], "紫微": []}):
        duanyu.yearly_range(pan, 2026, 2027, rules=["yingqi"])

    prepare.assert_called_once()
    assert prepare.call_args.args == (pan,)
    assert prepare.call_args.kwargs["factor_names"]
    assert make_snapshot.call_args_list[0].kwargs["natal_context"] is natal_context
    assert make_snapshot.call_args_list[1].kwargs["natal_context"] is natal_context


def test_eval_hybrid_builds_one_flow_snapshot_per_pan() -> None:
    import sys as _sys, os as _os
    _sys.path.insert(0, _os.path.join(_os.path.dirname(__file__), '..', 'scripts'))
    import eval_hybrid

    flow_snapshot = {
        "_snapshot_type": "liunian",
        "八字": {},
        "紫微": {},
        "context": {},
    }
    natal_snapshot = {"八字": {}, "紫微": {}, "context": {}}

    with mock.patch.object(eval_hybrid, "evaluate_snap_from_pan", return_value=natal_snapshot), \
         mock.patch.object(eval_hybrid, "match_rule", return_value={"八字": [], "紫微": []}), \
         mock.patch.object(eval_hybrid, "resolve_current_year", return_value=(2026, "server")), \
         mock.patch.object(eval_hybrid, "liunian", return_value={"bazi": {}, "ziwei": {}}), \
         mock.patch.object(
             eval_hybrid,
             "evaluate_liunian_snap_from_pan",
             return_value=flow_snapshot,
         ) as make_flow, \
         mock.patch.object(eval_hybrid, "query_yearly", return_value={"八字": [], "紫微": []}):
        eval_hybrid.query_all({"base": {}, "gender": "male"})

    make_flow.assert_called_once()
