"""禁止静默降级契约：计算错误必须传播，不得转成因子 0。"""
from unittest import mock

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
import factors
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
            factors.evaluate_factors(_fac(), "male", {}, shushi="bazi")


def test_evaluate_flow_factors_propagates_operator_errors() -> None:
    with mock.patch.object(factors, "_atomic", side_effect=RuntimeError("boom")):
        with pytest.raises(RuntimeError, match="boom"):
            factors.evaluate_liunian_factors(
                _fac(),
                "male",
                {},
                {"nian_gan": "甲", "nian_zhi": "子"},
                year=2006,
                shushi="bazi",
            )


def test_query_rejects_snapshot_input() -> None:
    with pytest.raises(ValueError, match="pan 必须是 full_paipan"):
        query("十神", {"八字": {}, "紫微": {}})


def test_current_year_does_not_fall_back_to_local_time() -> None:
    with mock.patch("paipan.call", side_effect=RPCError("time.now failed")):
        with pytest.raises(RPCError, match="time.now failed"):
            _current_year()


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
