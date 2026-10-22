"""当前限运契约：保存的 natal_chart 在 query 时按引擎当前年重推导，不使用排盘时索引。"""
from unittest import mock

import _helpers  # noqa: F401 —— 注入 tools 路径
import app.natal.tools.duanyu as duanyu
import app.natal.tools.factors as factors
from app.natal.tools.chart_integrity import with_natal_chart_digest


def _full_mock_rule_snapshot(*rules: str) -> dict:
    """Current-limit orchestration tests need a contract-complete mock snapshot."""
    tables = [
        table
        for rule in rules
        for table in duanyu.load_rule_tables(rule).values()
    ]
    required = duanyu.required_natal_factors(tables)
    return {
        "八字": {factor: 0 for factor in required},
        "紫微": {factor: 0 for factor in required},
        "context": {"性别": "male"},
    }


def _saved_pan() -> dict:
    pillars = ("nian", "yue", "ri", "shi")
    return with_natal_chart_digest({
        "instant": "1990-05-20T04:00:00Z",
        "timezone": "Asia/Shanghai",
        "gregorian_time": "1990-05-20T12:00:00+08:00",
        "gender": "male",
        "chart": {
            **{
                pillar: {"gan": "甲", "zhi": "子"}
                for pillar in ("nian", "yue", "ri", "shi")
            },
            "da_yun": {
                "current_step_index": 0,
                "steps": [
                    {"name": "戊辰", "shi_shen": "正财运", "start_year": 2000, "end_year": 2010},
                    {"name": "己巳", "shi_shen": "偏印运", "start_year": 2011, "end_year": 2020},
                ],
            }
        },
        "full": {
            **{pillar: {"gan": "甲", "zhi": "子"} for pillar in pillars},
            **_helpers.mock_engine_facts(),
            **_helpers.mock_yongshen_fields(),
        },
        "fu_yi": {}, "tiao_hou": {}, "ge_ju": {},
        "ziwei": _helpers.mock_ziwei(),
        "ziwei_chart": {},
        "ziwei_daxian": _helpers.valid_daxian(),
    })


def test_current_limit_uses_query_year_not_saved_index() -> None:
    natal_chart = _saved_pan()

    snap = factors.evaluate_snap_from_natal_chart(natal_chart, current_year=2015)

    assert snap["八字"]["大运十神类"] == "印星"
    assert snap["八字"]["大运配偶星"] == 0
    assert snap["context"]["性别"] == "male"
    assert snap["context"]["当前年份"] == 2015


def test_query_passes_server_year_for_current_limit_rules() -> None:
    natal_chart = _saved_pan()
    snapshots = _full_mock_rule_snapshot("大运", "六亲", "大限")

    with mock.patch.object(duanyu, "resolve_current_years", return_value=(2015, 2015, "server")), \
         mock.patch.object(
             duanyu,
             "evaluate_snap_from_natal_chart",
             return_value=snapshots,
         ) as evaluate_snap, \
         mock.patch.object(duanyu, "match_rule", return_value={"八字": [], "紫微": []}):
        duanyu.query("大运", natal_chart, timezone="Asia/Shanghai")
        duanyu.query("六亲", natal_chart)

    assert evaluate_snap.call_args_list[0].kwargs["current_year"] == 2015
    assert evaluate_snap.call_args_list[1].kwargs["current_year"] == 0


def test_query_explicit_year_does_not_call_time_now() -> None:
    natal_chart = _saved_pan()
    snapshots = _full_mock_rule_snapshot("大限")
    with mock.patch.object(duanyu, "resolve_current_years", side_effect=AssertionError("should not call time.now")), \
         mock.patch.object(duanyu, "evaluate_snap_from_natal_chart", return_value=snapshots) as evaluate_snap, \
         mock.patch.object(
             duanyu, "match_rule",
             side_effect=lambda *_: {"八字": [], "紫微": []},
         ):
        duanyu.query("大限", natal_chart, year=2005)

    assert evaluate_snap.call_args.kwargs["current_year"] == 2005


def test_query_limit_result_reports_year_source() -> None:
    natal_chart = _saved_pan()
    snapshots = _full_mock_rule_snapshot("大运", "六亲", "大限")
    with mock.patch.object(duanyu, "resolve_current_years", return_value=(2015, 2015, "server")), \
         mock.patch.object(duanyu, "evaluate_snap_from_natal_chart", return_value=snapshots), \
         mock.patch.object(
             duanyu, "match_rule",
             side_effect=lambda *_: {"八字": [], "紫微": []},
         ):
        current = duanyu.query("大运", natal_chart, timezone="Asia/Shanghai")
        specified = duanyu.query("大限", natal_chart, year=2005)

    assert current["current_year"] == 2015
    assert current["current_year_source"] == "server"
    assert specified["current_year"] == 2005
    assert specified["current_year_source"] == "specified"
    assert current["current_dayun"] == {
        "index": 1, "name": "己巳", "start_year": 2011, "end_year": 2020,
        "next_step_in_years": 5,
    }
    assert 2015 >= current["current_daxian"]["start_year"]
    assert 2015 <= current["current_daxian"]["end_year"]


def test_query_reports_lunar_year_when_different_from_bazi() -> None:
    natal_chart = _saved_pan()
    snapshots = _full_mock_rule_snapshot("大运")
    with mock.patch.object(duanyu, "resolve_current_years", return_value=(2025, 2026, "server")), \
         mock.patch.object(duanyu, "evaluate_snap_from_natal_chart", return_value=snapshots), \
         mock.patch.object(duanyu, "match_rule", return_value={"八字": [], "紫微": []}):
        result = duanyu.query("大运", natal_chart, timezone="Asia/Shanghai")

    assert result["current_year"] == 2025
    assert result["current_lunar_year"] == 2026
    assert result["current_year_source"] == "server"
    assert "current_dayun" not in result
    assert 2026 >= result["current_daxian"]["start_year"]
    assert 2026 <= result["current_daxian"]["end_year"]


def test_snap_sides_use_respective_current_years() -> None:
    natal_chart = _saved_pan()
    snap = factors.evaluate_snap_from_natal_chart(
        natal_chart, current_year=2015, current_year_ziwei=2016,
    )
    assert snap["context"]["当前年份"] == 2015
    assert snap["context"]["当前农历年"] == 2016


def test_query_evaluates_only_sides_required_by_rule() -> None:
    natal_chart = _saved_pan()
    with mock.patch.object(duanyu, "resolve_current_years", return_value=(2015, 2015, "server")), \
         mock.patch.object(
             duanyu,
             "evaluate_snap_from_natal_chart",
             side_effect=lambda *_args, **_kwargs: _full_mock_rule_snapshot(
                 "大运", "大限", "格局"
             ),
         ) as evaluate_snap, \
         mock.patch.object(duanyu, "match_rule", side_effect=lambda *_: {"八字": [], "紫微": []}):
        duanyu.query("大运", natal_chart, timezone="Asia/Shanghai")
        duanyu.query("大限", natal_chart, timezone="Asia/Shanghai")
        duanyu.query("格局", natal_chart)

    assert evaluate_snap.call_args_list[0].kwargs["sides"] == {"bazi"}
    assert evaluate_snap.call_args_list[1].kwargs["sides"] == {"ziwei"}
    assert evaluate_snap.call_args_list[2].kwargs["sides"] == {"bazi", "ziwei"}
    assert evaluate_snap.call_args_list[0].kwargs["factor_names"]
    assert evaluate_snap.call_args_list[1].kwargs["factor_names"]
    assert evaluate_snap.call_args_list[2].kwargs["factor_names"]


def test_factor_closure_preserves_all_natal_rule_matches() -> None:
    natal_chart = _saved_pan()
    full = factors.evaluate_snap_from_natal_chart(natal_chart, current_year=2005)
    for rule in duanyu.NATAL_RULES:
        tables = [
            duanyu.load_rule_table(
                f"bazi_{rule}.csv", required=rule not in duanyu.ZIWEI_ONLY_RULES
            ),
            duanyu.load_rule_table(
                f"ziwei_{rule}.csv", required=rule not in duanyu.BAZI_ONLY_RULES
            ),
        ]
        sides = (
            {"bazi"} if rule in duanyu.BAZI_ONLY_RULES
            else {"ziwei"} if rule in duanyu.ZIWEI_ONLY_RULES
            else {"bazi", "ziwei"}
        )
        pruned = factors.evaluate_snap_from_natal_chart(
            natal_chart,
            current_year=2005,
            sides=sides,
            factor_names=duanyu.required_natal_factors(tables),
        )
        assert duanyu.match_rule(rule, pruned) == duanyu.match_rule(rule, full)


def test_factor_closure_preserves_all_yearly_rule_matches() -> None:
    from app.natal.tools.factors import evaluate_liunian_snap_from_natal_chart

    natal_chart = _saved_pan()
    liunian_pan = {"bazi": {}, "ziwei": {}}
    natal_context = factors.prepare_natal_context(natal_chart)
    full = evaluate_liunian_snap_from_natal_chart(
        natal_chart, liunian_pan, year=2006, natal_context=natal_context
    )
    for rule in duanyu.YEARLY_RULES:
        tables = [
            duanyu.load_rule_table(
                f"bazi_{rule}.csv", required=rule not in duanyu.ZIWEI_ONLY_RULES
            ),
            duanyu.load_rule_table(
                f"ziwei_{rule}.csv", required=rule not in duanyu.BAZI_ONLY_RULES
            ),
        ]
        pruned = evaluate_liunian_snap_from_natal_chart(
            natal_chart,
            liunian_pan,
            year=2006,
            natal_context=natal_context,
            factor_names=duanyu.required_flow_factors(tables),
        )
        assert duanyu.query_yearly(rule, pruned) == duanyu.query_yearly(rule, full)


def test_current_limit_rule_set_matches_table_consumers() -> None:
    consumers = set()
    for rule in duanyu.NATAL_RULES:
        for table in (
            duanyu.load_rule_table(f"bazi_{rule}.csv", required=rule not in duanyu.ZIWEI_ONLY_RULES),
            duanyu.load_rule_table(f"ziwei_{rule}.csv", required=rule not in duanyu.BAZI_ONLY_RULES),
        ):
            DAYUN_FACTORS = {"大运十神类", "大运配偶星", "大运印星运", "大运官杀运", "大运财星运", "大运食伤运", "大运比劫运", "当前大限宫"}
            if table and any(
                k in DAYUN_FACTORS
                for row in table
                for conditions in row.get("约束组", [])
                for k in conditions
            ):
                consumers.add(rule)

    assert consumers == duanyu.CURRENT_LIMIT_RULES




def test_flow_natal_reference_closure_preserves_matches() -> None:
    from app.natal.tools.factors import evaluate_liunian_snap_from_natal_chart

    natal_chart = _saved_pan()
    liunian_pan = {"bazi": {}, "ziwei": {}}
    full_context = factors.prepare_natal_context(natal_chart)
    flow_factors = duanyu.flow_factor_names(sorted(duanyu.YEARLY_RULES))
    pruned_context = factors.prepare_natal_context(
        natal_chart, factor_names=duanyu.natal_factors_for_flow(flow_factors)
    )
    full = evaluate_liunian_snap_from_natal_chart(
        natal_chart, liunian_pan, year=2006, natal_context=full_context,
        factor_names=flow_factors,
    )
    pruned = evaluate_liunian_snap_from_natal_chart(
        natal_chart, liunian_pan, year=2006, natal_context=pruned_context,
        factor_names=flow_factors,
    )
    for rule in duanyu.YEARLY_RULES:
        assert duanyu.query_yearly(rule, pruned) == duanyu.query_yearly(rule, full)
