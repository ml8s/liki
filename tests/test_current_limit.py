"""当前限运契约：保存的 pan 在 query 时按引擎当前年重推导，不使用排盘时索引。"""
from unittest import mock

import _helpers  # noqa: F401 —— 注入 tools 路径
import duanyu
import factors


def _saved_pan() -> dict:
    pillars = ("nian", "yue", "ri", "shi")
    return {
        "solar": "1990-05-20T12:00:00",
        "lunar": {"year": 1990, "month": 4, "day": 26},
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
            pillar: {"gan": "甲", "zhi": "子"} for pillar in pillars
        },
        "yongshen": {},
        "ziwei": {"gong_wei": []},
        "ziwei_daxian": _helpers.valid_daxian(),
    }


def test_current_limit_uses_query_year_not_saved_index() -> None:
    pan = _saved_pan()

    snap = factors.evaluate_snap_from_pan(pan, current_year=2015)

    assert snap["八字"]["大运十神类"] == "印星"
    assert snap["八字"]["大运配偶星"] == 0
    assert snap["context"]["性别"] == "male"
    assert snap["context"]["当前年份"] == 2015


def test_query_passes_server_year_for_current_limit_rules() -> None:
    pan = _saved_pan()
    snapshots = {"八字": {}, "紫微": {}, "context": {}}

    with mock.patch.object(duanyu, "_current_year", return_value=(2015, "server")), \
         mock.patch.object(
             duanyu,
             "evaluate_snap_from_pan",
             return_value=snapshots,
         ) as evaluate_snap, \
         mock.patch.object(duanyu, "_match_rule", return_value={"八字": [], "紫微": []}):
        duanyu.query("大运", pan)
        duanyu.query("六亲", pan)

    assert evaluate_snap.call_args_list[0].kwargs["current_year"] == 2015
    assert evaluate_snap.call_args_list[1].kwargs["current_year"] == 0


def test_query_explicit_year_does_not_call_time_now() -> None:
    pan = _saved_pan()
    snapshots = {"八字": {}, "紫微": {}, "合参": [], "context": {}}
    with mock.patch.object(duanyu, "_current_year", side_effect=AssertionError("should not call time.now")), \
         mock.patch.object(duanyu, "evaluate_snap_from_pan", return_value=snapshots) as evaluate_snap, \
         mock.patch.object(
             duanyu, "_match_rule",
             side_effect=lambda *_: {"八字": [], "紫微": [], "合参": []},
         ):
        duanyu.query("大限", pan, year=2005)

    assert evaluate_snap.call_args.kwargs["current_year"] == 2005


def test_query_limit_result_reports_year_source() -> None:
    pan = _saved_pan()
    snapshots = {"八字": {}, "紫微": {}, "合参": [], "context": {}}
    with mock.patch.object(duanyu, "_current_year", return_value=(2015, "server")), \
         mock.patch.object(duanyu, "evaluate_snap_from_pan", return_value=snapshots), \
         mock.patch.object(
             duanyu, "_match_rule",
             side_effect=lambda *_: {"八字": [], "紫微": [], "合参": []},
         ):
        current = duanyu.query("大运", pan)
        specified = duanyu.query("大限", pan, year=2005)

    assert current["current_year"] == 2015
    assert current["current_year_source"] == "server"
    assert specified["current_year"] == 2005
    assert specified["current_year_source"] == "specified"


def test_query_evaluates_only_sides_required_by_rule() -> None:
    pan = _saved_pan()
    snapshots = {"八字": {}, "紫微": {}, "合参": [], "context": {}}
    with mock.patch.object(duanyu, "_current_year", return_value=(2015, "server")), \
         mock.patch.object(
             duanyu, "evaluate_snap_from_pan", return_value=snapshots
         ) as evaluate_snap, \
         mock.patch.object(duanyu, "_match_rule", side_effect=lambda *_: {"八字": [], "紫微": [], "合参": []}):
        duanyu.query("大运", pan)
        duanyu.query("大限", pan)
        duanyu.query("格局", pan)

    assert evaluate_snap.call_args_list[0].kwargs["sides"] == {"bazi"}
    assert evaluate_snap.call_args_list[1].kwargs["sides"] == {"ziwei"}
    assert evaluate_snap.call_args_list[2].kwargs["sides"] == {"bazi", "ziwei"}
    assert evaluate_snap.call_args_list[0].kwargs["factor_names"]
    assert evaluate_snap.call_args_list[1].kwargs["factor_names"]
    assert evaluate_snap.call_args_list[2].kwargs["factor_names"]


def test_factor_closure_preserves_all_natal_rule_matches() -> None:
    pan = _saved_pan()
    full = factors.evaluate_snap_from_pan(pan, current_year=2005)
    for rule in duanyu.NATAL_RULES:
        tables = [
            duanyu.load_table(
                f"bazi_{rule}.csv", required=rule not in duanyu.ZIWEI_ONLY_RULES
            ),
            duanyu.load_table(
                f"ziwei_{rule}.csv", required=rule not in duanyu.BAZI_ONLY_RULES
            ),
            duanyu.load_table(f"common_{rule}.csv", required=False),
        ]
        sides = (
            {"bazi"} if rule in duanyu.BAZI_ONLY_RULES
            else {"ziwei"} if rule in duanyu.ZIWEI_ONLY_RULES
            else {"bazi", "ziwei"}
        )
        pruned = factors.evaluate_snap_from_pan(
            pan,
            current_year=2005,
            sides=sides,
            factor_names=duanyu._required_natal_factors(tables),
        )
        assert duanyu._match_rule(rule, pruned) == duanyu._match_rule(rule, full)


def test_factor_closure_preserves_all_yearly_rule_matches() -> None:
    from factors import evaluate_liunian_snap_from_pan

    pan = _saved_pan()
    liunian_pan = {"bazi": {}, "ziwei": {}}
    natal_context = factors.prepare_natal_context(pan)
    full = evaluate_liunian_snap_from_pan(
        pan, liunian_pan, year=2006, natal_context=natal_context
    )
    for rule in duanyu.YEARLY_RULES:
        tables = [
            duanyu.load_table(
                f"bazi_{rule}.csv", required=rule not in duanyu.ZIWEI_ONLY_RULES
            ),
            duanyu.load_table(
                f"ziwei_{rule}.csv", required=rule not in duanyu.BAZI_ONLY_RULES
            ),
            duanyu.load_table(f"common_{rule}.csv", required=False),
        ]
        pruned = evaluate_liunian_snap_from_pan(
            pan,
            liunian_pan,
            year=2006,
            natal_context=natal_context,
            factor_names=duanyu._required_flow_factors(tables),
        )
        assert duanyu.query_yearly(rule, pruned) == duanyu.query_yearly(rule, full)


def test_current_limit_rule_set_matches_table_consumers() -> None:
    consumers = set()
    for rule in duanyu.NATAL_RULES:
        for table in (
            duanyu.load_table(f"bazi_{rule}.csv", required=rule not in duanyu.ZIWEI_ONLY_RULES),
            duanyu.load_table(f"ziwei_{rule}.csv", required=rule not in duanyu.BAZI_ONLY_RULES),
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

    from factors import evaluate_liunian_snap_from_pan



def test_flow_natal_reference_closure_preserves_matches() -> None:
    from factors import evaluate_liunian_snap_from_pan

    pan = _saved_pan()
    liunian_pan = {"bazi": {}, "ziwei": {}}
    full_context = factors.prepare_natal_context(pan)
    flow_factors = duanyu.flow_factor_names(sorted(duanyu.YEARLY_RULES))
    pruned_context = factors.prepare_natal_context(
        pan, factor_names=duanyu.natal_factors_for_flow(flow_factors)
    )
    full = evaluate_liunian_snap_from_pan(
        pan, liunian_pan, year=2006, natal_context=full_context,
        factor_names=flow_factors,
    )
    pruned = evaluate_liunian_snap_from_pan(
        pan, liunian_pan, year=2006, natal_context=pruned_context,
        factor_names=flow_factors,
    )
    for rule in duanyu.YEARLY_RULES:
        assert duanyu.query_yearly(rule, pruned) == duanyu.query_yearly(rule, full)
