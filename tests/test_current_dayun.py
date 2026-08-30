"""当前大运契约：保存的 pan 在 query 时按引擎当前年重推导，不使用排盘时索引。"""
from unittest import mock

import _helpers  # noqa: F401 —— 注入 tools 路径
import duanyu
import factors
from _helpers import mock_factors


def _saved_pan() -> dict:
    fac = mock_factors()
    fac["dayun_steps"] = [
        {
            "name": "戊辰",
            "shi_shen": "正财运",
            "start_year": 2000,
            "end_year": 2010,
        },
        {
            "name": "己巳",
            "shi_shen": "偏印运",
            "start_year": 2011,
            "end_year": 2020,
        },
    ]
    return {
        "fac": fac,
        "gender": "male",
        "chart": {
            "da_yun": {
                "current_step_index": 0,
                "steps": [{"shi_shen": "正财运"}],
            }
        },
        "full": {},
        "ziwei": {},
    }


def test_current_dayun_uses_query_year_not_saved_index() -> None:
    pan = _saved_pan()

    snap = factors.make_factors(pan, current_year=2015)

    assert snap["八字"]["大运十神类"] == "印星"
    assert snap["八字"]["大运配偶星"] == 0
    assert snap["context"] == {"性别": "male", "当前年份": 2015}


def test_query_passes_server_year_for_current_dayun_rules() -> None:
    pan = _saved_pan()
    snapshots = {"八字": {}, "紫微": {}, "context": {}}

    with mock.patch.object(duanyu, "_current_year", return_value=(2015, "server")), \
         mock.patch.object(
             duanyu,
             "make_factors",
             return_value=snapshots,
         ) as make_factors, \
         mock.patch.object(duanyu, "_match_rule", return_value={"八字": [], "紫微": []}):
        duanyu.query("大运", pan)
        duanyu.query("六亲", pan)

    assert make_factors.call_args_list[0].kwargs["current_year"] == 2015
    assert make_factors.call_args_list[1].kwargs["current_year"] == 0


def test_current_dayun_rule_set_matches_table_consumers() -> None:
    consumers = set()
    for rule in duanyu._NATAL_RULES:
        for table in (
            duanyu.load_table(f"bazi_{rule}.csv", required=rule not in duanyu._ZIWEI_ONLY_RULES),
            duanyu.load_table(f"ziwei_{rule}.csv", required=rule not in duanyu._BAZI_ONLY_RULES),
        ):
            if table and {"大运十神类", "大运配偶星"} & set(table[0]["约束"]):
                consumers.add(rule)

    assert consumers == duanyu._CURRENT_DAYUN_RULES
