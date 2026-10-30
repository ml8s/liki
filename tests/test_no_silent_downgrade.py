"""禁止静默降级契约：计算错误必须传播，不得转成因子 0。"""
from unittest import mock

import app.natal.tools.duanyu as duanyu
import app.natal.tools.factors as factors
import pytest
from app.natal.tools.charts import RPCError
from app.natal.tools.duanyu import load_rule_table, query, resolve_current_year


def _fac() -> dict:
    return {
        "ten_god_states": {},
        "wuxing": {"wang_shuai": {}, "count": {}},
        "fu_yi": {}, "tiao_hou": {}, "ge_ju": {},
        "ri_gan": "甲",
    }


def test_evaluate_factors_propagates_operator_errors() -> None:
    with mock.patch.object(factors, "evaluate_operator", side_effect=RuntimeError("boom")):
        with pytest.raises(RuntimeError, match="boom"):
            factors.evaluate_factors("male", _fac(), shushi="bazi")


def test_evaluate_flow_factors_propagates_operator_errors() -> None:
    with mock.patch.object(factors, "evaluate_operator", side_effect=RuntimeError("boom")):
        with pytest.raises(RuntimeError, match="boom"):
            factors.evaluate_liunian_factors(
                "male",
                _fac(),
                {"nian_gan": "甲", "nian_zhi": "子"},
                year=2006,
                shushi="bazi",
            )


def test_query_rejects_snapshot_input() -> None:
    with pytest.raises(ValueError, match="natal_chart 必须是 build_natal_chart"):
        query("十神", {"八字": {}, "紫微": {}})


@pytest.mark.parametrize("natal_chart", [
    {"full": {"ri": {"gan": "甲", "zhi": "子"}}},
    {"chart": {"ri": {"gan": "甲", "zhi": "子"}}},
    {
        "standard_time": "1990-05-20T12:00:00",
        "chart": {"ri": {"gan": "甲", "zhi": "子"}},
        "full": {"ri": {"gan": "甲", "zhi": "子"}},
        "fu_yi": {}, "tiao_hou": {}, "ge_ju": {},
        "ziwei": {},
        "gender": "male",
    },
])
def test_query_rejects_partial_pan(natal_chart) -> None:
    with pytest.raises(ValueError, match="build_natal_chart|完整本命盘|四柱结构|缺失"):
        query("十神", natal_chart)


def test_current_year_does_not_fall_back_to_local_time() -> None:
    with mock.patch("app.natal.tools.charts.call", side_effect=RPCError("time.now failed")):
        with pytest.raises(RPCError, match="time.now failed"):
            resolve_current_year("Asia/Shanghai")


def _reset_year_cache() -> None:
    duanyu.reset_current_year_cache()


def _year_call(cst: str):
    def fake(name, params):
        if name == "time.now":
            return {"data": {
                "standard_time": cst + "T12:00:00+08:00", "offset": 28800,
            }}
        if name == "tianwen.resolve":
            date, clock = params["date"], params["time"]
            return {"data": {"candidates": [{
                "standard_time": f"{date}T{clock}:00+08:00",
                "offset": 28800,
                "reasons": [{"domain": "input", "detail": "输入时间"}],
            }]}}
        if name == "bazi.chart":
            return {"data": {"nian": {"gan": "丙", "zhi": "午"}, "birth_year": 2026}}
        if name == "ziwei.chart":
            return {"data": {"birth_year": 2026}}
        raise AssertionError(f"unexpected tool {name}")
    return fake


def _time_now_calls(call_mock) -> int:
    return sum(1 for c in call_mock.call_args_list if c.args and c.args[0] == "time.now")


def test_current_year_caches_success_within_ttl() -> None:
    _reset_year_cache()
    with mock.patch("app.natal.tools.charts.call", side_effect=_year_call("2026-03-15T12:00:00")) as call_mock:
        assert resolve_current_year("Asia/Shanghai") == (2026, "server")
        assert resolve_current_year("Asia/Shanghai") == (2026, "server")
        # TTL 内第二次调用命中缓存（时间事实 RPC 只发一次）
        assert _time_now_calls(call_mock) == 1
    _reset_year_cache()


def test_current_year_cache_expires_and_refetches() -> None:
    _reset_year_cache()
    with mock.patch("app.natal.tools.charts.call", side_effect=_year_call("2026-12-31T23:00:00")) as call_mock, \
         mock.patch("app.natal.tools.duanyu.time.monotonic", side_effect=[0.0, 30.0, 61.0]):
        resolve_current_year("Asia/Shanghai")          # 首次：发 RPC
        resolve_current_year("Asia/Shanghai")          # +30s：仍命中
        assert _time_now_calls(call_mock) == 1
        resolve_current_year("Asia/Shanghai")          # +61s：过期，重发 RPC
        assert _time_now_calls(call_mock) == 2
    _reset_year_cache()


def test_current_year_failure_not_cached() -> None:
    _reset_year_cache()
    with mock.patch("app.natal.tools.charts.call", side_effect=RPCError("time.now failed")), \
         mock.patch("app.natal.tools.duanyu.time.monotonic", return_value=0.0):
        with pytest.raises(RPCError):
            resolve_current_year("Asia/Shanghai")
        # 失败不写入缓存——下一次仍会再试（不把错误当时间基准缓存）
        assert duanyu._year_cache._data is None
    _reset_year_cache()


def test_current_year_rejects_missing_cst_and_does_not_cache() -> None:
    _reset_year_cache()
    with mock.patch("app.natal.tools.charts.call", return_value={"data": {}}):
        with pytest.raises(ValueError, match="standard_time|local_time"):
            resolve_current_year("Asia/Shanghai")
    assert duanyu._year_cache._data is None


def test_load_rule_table_missing_required_file_raises() -> None:
    assert load_rule_table("bazi_missing_domain", required=False) == []
    with pytest.raises(FileNotFoundError, match="missing_domain.csv"):
        load_rule_table("bazi_missing_domain")


def test_skill_does_not_silently_default_birth_hour() -> None:
    from pathlib import Path

    skill = Path(__file__).resolve().parents[1] / "skills" / "liki" / "references" / "natal-ENTRY.md"
    text = skill.read_text(encoding="utf-8")

    assert "默认午时" not in text
    assert "默认时辰" not in text
    assert "待裁" not in text


def test_calibration_docs_do_not_silently_default_birth_hour() -> None:
    from pathlib import Path

    path = (
        Path(__file__).resolve().parents[1]
        / "skills" / "liki" / "references"
        / "natal-app-calibration.md"
    )
    text = path.read_text(encoding="utf-8")
    assert "直接走默认时辰" not in text
    assert "建议先走默认时辰" not in text


def test_skill_has_no_local_pan_archive_contract() -> None:
    from pathlib import Path

    skill_dir = Path(__file__).resolve().parents[1] / "skills" / "liki" / "references"
    for path in sorted(skill_dir.glob("natal-*.md")):
        text = path.read_text(encoding="utf-8", errors="ignore")
        assert "liki-memory" not in text, path
        assert "$file" not in text, path
        assert "保存命盘" not in text, path
