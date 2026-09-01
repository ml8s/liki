"""禁止静默降级契约：计算错误必须传播，不得转成因子 0。"""
from unittest import mock

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
import factors
import duanyu
from duanyu import _current_year, query
from duanyu import load_table
from paipan import RPCError


def _fac() -> dict:
    return {
        "shishen": {},
        "wuxing": {"wang_shuai": {}, "count": {}},
        "yongshen": {},
        "ri_gan": "甲",
    }


def test_evaluate_factors_propagates_operator_errors() -> None:
    with mock.patch.object(factors, "_atomic", side_effect=RuntimeError("boom")):
        with pytest.raises(RuntimeError, match="boom"):
            factors.evaluate_factors("male", _fac(), shushi="bazi")


def test_evaluate_flow_factors_propagates_operator_errors() -> None:
    with mock.patch.object(factors, "_atomic", side_effect=RuntimeError("boom")):
        with pytest.raises(RuntimeError, match="boom"):
            factors.evaluate_liunian_factors(
                "male",
                _fac(),
                {"nian_gan": "甲", "nian_zhi": "子"},
                year=2006,
                shushi="bazi",
            )


def test_query_rejects_snapshot_input() -> None:
    with pytest.raises(ValueError, match="pan 必须是 full_paipan"):
        query("十神", {"八字": {}, "紫微": {}})


@pytest.mark.parametrize("pan", [
    {"full": {"ri": {"gan": "甲", "zhi": "子"}}},
    {"chart": {"ri": {"gan": "甲", "zhi": "子"}}},
    {
        "solar": "1990-05-20T12:00:00",
        "lunar": {"year": 1990, "month": 4, "day": 26},
        "chart": {"ri": {"gan": "甲", "zhi": "子"}},
        "full": {"ri": {"gan": "甲", "zhi": "子"}},
        "yongshen": {},
        "ziwei": {},
        "gender": "male",
    },
])
def test_query_rejects_partial_pan(pan) -> None:
    with pytest.raises(ValueError, match="full_paipan|完整本命盘|四柱结构|缺失"):
        query("十神", pan)


def test_current_year_does_not_fall_back_to_local_time() -> None:
    with mock.patch("paipan.call", side_effect=RPCError("time.now failed")):
        with pytest.raises(RPCError, match="time.now failed"):
            _current_year()


def _reset_year_cache() -> None:
    duanyu._reset_current_year_cache()


def test_current_year_caches_success_within_ttl() -> None:
    _reset_year_cache()
    payload = {"data": {"cst": "2026-03-15T12:00:00"}}
    with mock.patch("paipan.call", return_value=payload) as call_mock:
        assert _current_year() == (2026, "server")
        assert _current_year() == (2026, "server")
        # TTL 内第二次调用命中缓存
        call_mock.assert_called_once_with("time.now", {})
    _reset_year_cache()


def test_current_year_cache_expires_and_refetches() -> None:
    _reset_year_cache()
    payload = {"data": {"cst": "2026-12-31T23:00:00"}}
    with mock.patch("paipan.call", return_value=payload) as call_mock, \
         mock.patch("duanyu.time.monotonic", side_effect=[0.0, 30.0, 61.0]):
        _current_year()          # 首次：发 RPC
        _current_year()          # +30s：仍命中
        assert call_mock.call_count == 1
        _current_year()          # +61s：过期，重发 RPC
        assert call_mock.call_count == 2
    _reset_year_cache()


def test_current_year_failure_not_cached() -> None:
    _reset_year_cache()
    with mock.patch("paipan.call", side_effect=RPCError("time.now failed")), \
         mock.patch("duanyu.time.monotonic", return_value=0.0):
        with pytest.raises(RPCError):
            _current_year()
        # 失败不写入缓存——下一次仍会再试（不把错误当时间基准缓存）
        assert duanyu._current_year_cached is None
    _reset_year_cache()


def test_current_year_rejects_missing_cst_and_does_not_cache() -> None:
    _reset_year_cache()
    with mock.patch("paipan.call", return_value={"data": {}}):
        with pytest.raises(ValueError, match="cst"):
            _current_year()
    assert duanyu._current_year_cached is None


def test_load_table_missing_required_file_raises() -> None:
    assert load_table("bazi_missing_domain", required=False) == []
    with pytest.raises(FileNotFoundError, match="missing_domain.csv"):
        load_table("bazi_missing_domain")


def test_skill_does_not_silently_default_birth_hour() -> None:
    from pathlib import Path

    skill = Path(__file__).resolve().parents[1] / "skills" / "liki-bazi" / "SKILL.md"
    text = skill.read_text(encoding="utf-8")

    assert "默认午时" not in text
    assert "默认时辰" not in text
    assert "待裁" not in text


def test_calibration_docs_do_not_silently_default_birth_hour() -> None:
    from pathlib import Path

    path = (
        Path(__file__).resolve().parents[1]
        / "skills" / "liki-bazi" / "domains" / "bazi" / "calibration.md"
    )
    text = path.read_text(encoding="utf-8")
    assert "直接走默认时辰" not in text
    assert "建议先走默认时辰" not in text


def test_skill_has_no_local_pan_archive_contract() -> None:
    from pathlib import Path

    skill_dir = Path(__file__).resolve().parents[1] / "skills" / "liki-bazi"
    for path in skill_dir.rglob("*"):
        if not path.is_file() or "__pycache__" in path.parts:
            continue
        text = path.read_text(encoding="utf-8", errors="ignore")
        assert "liki-memory" not in text, path
        assert "$file" not in text, path
        assert "保存命盘" not in text, path
