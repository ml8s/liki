"""正交化重构基线测试。

记录重构前（排盘+判断一体）的因子/断语/应期输出，重构后新工具
（compute_factors / natal_query / period_query）必须与基线一致——
接口变化，逻辑输出不变。golden 存 tests/golden/orthogonal/baseline.json。
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "natal" / "tools"))

from duanyu import query, yearly_range  # noqa: E402
from factors import evaluate_factors  # noqa: E402
from paipan import full_paipan  # noqa: E402

GOLDEN = Path(__file__).resolve().parents[2] / "tests" / "golden" / "orthogonal" / "baseline.json"

BIRTH = ("1990-05-20T11:49:00+08:00", "male", 116.4, 39.9)


@pytest.fixture(scope="module")
def baseline() -> dict:
    if not GOLDEN.exists():
        pytest.skip("正交化基线不存在")
    return json.loads(GOLDEN.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def pan():
    return full_paipan(*BIRTH)


def test_factors_bazi_matches_baseline(pan, baseline):
    got = evaluate_factors("male", pan, shushi="bazi")
    want = baseline["factors_bazi"]
    assert got == want


def test_factors_ziwei_matches_baseline(pan, baseline):
    got = evaluate_factors("male", pan, shushi="ziwei")
    want = baseline["factors_ziwei"]
    assert got == want


def test_natal_shishen_matches_baseline(pan, baseline):
    got = query("十神", pan)
    want = baseline["natal_十神"]
    assert got == want


def test_natal_mingong_matches_baseline(pan, baseline):
    got = query("命宫", pan)
    want = baseline["natal_命宫"]
    assert got == want


def test_period_2026_matches_baseline(pan, baseline):
    got = yearly_range(pan, 2026, 2026, rules=["年十神"])
    want = baseline["period_2026"]
    assert got == want
