"""正交化重构基线测试。

记录重构前（排盘+判断一体）的因子/断语/应期输出，重构后新工具
（compute_factors / natal_query / period_query）必须与基线一致——
接口变化，逻辑输出不变。golden 存 tests/golden/orthogonal/baseline.json。
"""

import json
from pathlib import Path

import pytest
from app.natal.tools.charts import build_natal_chart  # noqa: E402
from app.natal.tools.duanyu import query, yearly_range  # noqa: E402
from app.natal.tools.factors import evaluate_factors  # noqa: E402

GOLDEN = Path(__file__).resolve().parents[2] / "tests" / "golden" / "orthogonal" / "baseline.json"

BIRTH = ("1990-05-20T03:49:00Z", "Asia/Shanghai", "male", 116.4)


@pytest.fixture(scope="module")
def baseline() -> dict:
    if not GOLDEN.exists():
        pytest.skip("正交化基线不存在")
    return json.loads(GOLDEN.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def natal_chart():
    return build_natal_chart(*BIRTH)


def test_factors_bazi_matches_baseline(natal_chart, baseline):
    from app.natal.tools.yongshen import resolve_yongshen

    evaluated = {**natal_chart, "_selected_yongshen": resolve_yongshen(natal_chart["full"])}
    got = evaluate_factors("male", evaluated, shushi="bazi")
    want = baseline["factors_bazi"]
    assert got == want


def test_factors_ziwei_matches_baseline(natal_chart, baseline):
    got = evaluate_factors("male", natal_chart, shushi="ziwei")
    want = baseline["factors_ziwei"]
    assert got == want


def test_natal_shishen_matches_baseline(natal_chart, baseline):
    got = query("十神", natal_chart)
    want = baseline["natal_十神"]
    assert got == want


def test_natal_mingong_matches_baseline(natal_chart, baseline):
    got = query("命宫", natal_chart)
    want = baseline["natal_命宫"]
    assert got == want


def test_period_2026_matches_baseline(natal_chart, baseline):
    got = yearly_range(natal_chart, 2026, 2026, rules=["年十神"])
    want = baseline["period_2026"]
    assert got == want
