"""Schema 门禁的输入不变量：空条件、空因子定义与标量值域。"""
import csv
import json
from pathlib import Path

import _helpers  # noqa: F401
from factor_tables import load_long_rows

ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-bazi" / "tools"
CONSTANTS = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
ENUM_SOURCES = {
    "月令格": "月令格局", "扶抑从格": "扶抑从格", "身强弱": "身强弱状态",
    "调候季节": "调候季节", "日主": "天干", "日主五行": "五行",
    "日主长生状态": "十二长生", "日支神煞类型": "日支神煞",
    "月令本气十神": "十神", "大运十神类": "十神大类",
    "流年日主长生状态": "十二长生",
}


def _assertions() -> list[dict]:
    with (TOOLS / "assertions/assertions.csv").open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def _conditions() -> list[dict]:
    with (TOOLS / "assertions/assertion_conditions.csv").open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def test_assertion_rows_have_conditions() -> None:
    conditioned = {row["assertion_id"] for row in _conditions()}
    for row in _assertions():
        assert row["assertion_id"] in conditioned, f"{row['assertion_id']} 无条件恒命中"


def test_raw_bond_tool_has_no_unused_assertion_tables() -> None:
    """bond 返回引擎原始合盘数据；不消费的断语规则就是死数据。"""
    assert all(row["rule"] != "bond" for row in _assertions())


def test_factor_rows_have_definitions() -> None:
    for filename in ("factors.csv", "factors_liunian.csv"):
        for row in load_long_rows(str(TOOLS / "factors" / filename)):
            assert row.get("直通") or row.get("conds"), f"{filename}:{row['因子']} 空定义恒假"


def test_scalar_constraints_use_constant_closures() -> None:
    assertions = {row["assertion_id"]: row for row in _assertions()}
    for condition in _conditions():
        factor = condition["factor"]
        source = ENUM_SOURCES.get(factor)
        if not source:
            continue
        expected = condition["expected"]
        assert expected in CONSTANTS[source], (
            f"{assertions[condition['assertion_id']]['assertion_id']}: "
            f"{factor}={expected} 不在 constants.{source}"
        )
