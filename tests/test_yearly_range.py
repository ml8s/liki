"""yearly_range 编排契约：同一年份的流年快照只生成一次。"""
from unittest import mock

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
import duanyu


def test_yearly_range_builds_one_snapshot_per_year() -> None:
    flow_snapshot = {
        "_snapshot_type": "liunian",
        "八字": {},
        "紫微": {},
        "context": {},
    }
    liunian_pan = {"bazi": {"nian_gan": "丙"}, "ziwei": {}}

    with mock.patch.object(duanyu, "_current_year", return_value=(2026, "server")), \
         mock.patch("paipan.liunian", return_value=liunian_pan) as liunian_mock, \
         mock.patch.object(
             duanyu,
             "make_liunian_factors",
             return_value=flow_snapshot,
         ) as make_snapshot, \
         mock.patch.object(
             duanyu,
             "query_yearly",
             return_value={"八字": [], "紫微": []},
         ):
        result = duanyu.yearly_range(
            {"fac": {}, "gender": "male"},
            2026,
            2026,
            rules=["年十神", "年合会"],
        )

    assert result["years"]["2026"]["年十神"] == {"八字": [], "紫微": []}
    liunian_mock.assert_called_once()
    make_snapshot.assert_called_once()


def test_yearly_range_rejects_empty_or_reversed_range() -> None:
    with mock.patch.object(duanyu, "_current_year", return_value=(2026, "server")):
        with mock.patch.object(duanyu, "make_liunian_factors") as make_snapshot:
            with mock.patch.object(duanyu, "query_yearly"):
                with pytest.raises(ValueError, match="rules 不能为空"):
                    duanyu.yearly_range({"fac": {}}, 2026, 2026, rules=[])
                with pytest.raises(ValueError, match="rules 不能有重复域"):
                    duanyu.yearly_range(
                        {"fac": {}}, 2026, 2026,
                        rules=["yearly_career", "yearly_career"],
                    )
                with pytest.raises(ValueError, match="start 不能大于 end"):
                    duanyu.yearly_range({"fac": {}}, 2027, 2026)

    make_snapshot.assert_not_called()


def test_eval_hybrid_builds_one_flow_snapshot_per_pan() -> None:
    import eval_hybrid

    flow_snapshot = {
        "_snapshot_type": "liunian",
        "八字": {},
        "紫微": {},
        "context": {},
    }
    natal_snapshot = {"八字": {}, "紫微": {}, "context": {}}

    with mock.patch.object(eval_hybrid, "make_factors", return_value=natal_snapshot), \
         mock.patch.object(eval_hybrid, "_match_rule", return_value={"八字": [], "紫微": []}), \
         mock.patch.object(eval_hybrid, "_current_year", return_value=(2026, "server")), \
         mock.patch.object(eval_hybrid, "liunian", return_value={"bazi": {}, "ziwei": {}}), \
         mock.patch.object(
             eval_hybrid,
             "make_liunian_factors",
             return_value=flow_snapshot,
         ) as make_flow, \
         mock.patch.object(eval_hybrid, "query_yearly", return_value={"八字": [], "紫微": []}):
        eval_hybrid.query_all({"fac": {}, "gender": "male"})

    make_flow.assert_called_once()
