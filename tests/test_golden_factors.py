"""命理金样例：用正例/反例/边界例锁定因子与流年原语。"""
import json
from pathlib import Path

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
from factors import _LIU_OP_NAMES, _OP_NAMES, _liu_op, _op, evaluate_factors

ROOT = Path(__file__).resolve().parents[1]
CASES_PATH = ROOT / "tests" / "golden_factors" / "cases.json"
MANIFEST = json.loads(CASES_PATH.read_text(encoding="utf-8"))
CASES = MANIFEST["cases"]


def _case_ids():
    return [case["id"] for case in CASES]


def test_golden_manifest_contract() -> None:
    assert MANIFEST["version"] == 1
    assert len(CASES) >= 20
    assert len(_case_ids()) == len(set(_case_ids()))

    allowed_kinds = {"positive", "negative", "boundary"}
    allowed_modes = {"factor_snapshot", "operator"}
    for case in CASES:
        assert case["id"], case
        assert case["rule"], case["id"]
        assert case["basis"], case["id"]
        assert case["kind"] in allowed_kinds, (case["id"], case["kind"])
        assert case["mode"] in allowed_modes, (case["id"], case["mode"])
        if case["mode"] == "factor_snapshot":
            assert case.get("expect_factors"), case["id"]
        else:
            assert case.get("expect") is not None, case["id"]


def test_golden_categories_are_not_app_conclusions() -> None:
    forbidden_words = {"吉", "凶", "好命", "坏命", "富贵命"}
    for case in CASES:
        assert not any(word in case["rule"] for word in forbidden_words), case["id"]
        assert not any(word in case["basis"] for word in forbidden_words), case["id"]


@pytest.mark.parametrize("case_id", _case_ids())
def test_golden_case(case_id: str) -> None:
    case = next(item for item in CASES if item["id"] == case_id)
    input_data = case["input"]

    if case["mode"] == "factor_snapshot":
        snapshot = evaluate_factors(
            input_data["fac"],
            input_data["gender"],
            input_data["chart"],
            shushi=case.get("shushi"),
            current_year=input_data.get("current_year", 0),
        )
        expected = case["expect_factors"]
        assert expected
        for name, value in expected.items():
            if name == "性别" and value == 0:
                assert name not in snapshot, f"{case_id}: 性别不得进入因子快照"
            else:
                assert name in snapshot, f"{case_id}: 缺因子 {name}"
                assert snapshot[name] == value, (
                    f"{case_id}: {name}={snapshot[name]!r}, want {value!r}; {case['basis']}"
                )
        return

    assert case["mode"] == "operator"
    op = input_data["op"]
    args = input_data.get("args", [])
    if op in _OP_NAMES:
        actual = _op(op, args, input_data["fac"], input_data["gender"], input_data["chart"])
    elif op in _LIU_OP_NAMES:
        actual = _liu_op(
            op,
            args,
            input_data["fac"],
            input_data["gender"],
            input_data["chart"],
            input_data.get("ctx", {}),
        )
    else:
        raise AssertionError(f"{case_id}: 未知金样例算子 {op}")

    assert actual == case["expect"], (
        f"{case_id}: {op}={actual!r}, want {case['expect']!r}; {case['basis']}"
    )
