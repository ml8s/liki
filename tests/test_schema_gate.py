"""Schema 门禁的输入不变量：空条件、空因子定义与标量值域。"""
import csv
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills" / "liki-bazi" / "tools"
CONSTANTS = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
META = {"因子", "术数", "原语直通", "依据"}
ENUM_SOURCES = {
    "月令格": "月令格局",
    "扶抑从格": "扶抑从格",
    "身强弱": "身强弱状态",
    "调候季节": "调候季节",
    "日主": "天干",
    "日主五行": "五行",
    "日主长生状态": "十二长生",
    "日支神煞类型": "日支神煞",
    "月令本气十神": "十神",
    "大运十神类": "十神大类",
    "流年日主长生状态": "十二长生",
}


def test_assertion_rows_have_conditions() -> None:
    for path in (*TOOLS.joinpath("bazi").glob("*.csv"), *TOOLS.joinpath("ziwei").glob("*.csv")):
        with path.open(encoding="utf-8", newline="") as f:
            for row in csv.DictReader(f):
                conditions = {
                    key: value for key, value in row.items()
                    if key not in {"id", "事件", "结论", "依据", "经典原文"}
                    and (value or "").strip()
                }
                assert conditions, f"{path}:{row.get('id')} 无条件恒命中"


def test_raw_bond_tool_has_no_unused_assertion_tables() -> None:
    """bond 返回引擎原始合盘数据；不消费的断语表就是死数据。"""
    assert not (TOOLS / "bazi" / "bond.csv").exists()
    assert not (TOOLS / "ziwei" / "bond.csv").exists()


def test_factor_rows_have_definitions() -> None:
    for filename in ("factors.csv", "factors_liunian.csv"):
        path = TOOLS / "factors" / filename
        with path.open(encoding="utf-8", newline="") as f:
            for row in csv.DictReader(f):
                direct = (row.get("原语直通") or "").strip()
                conditions = any(
                    (value or "").strip() for key, value in row.items() if key not in META
                )
                assert direct or conditions, f"{filename}:{row['因子']} 空定义恒假"


def test_scalar_constraints_use_constant_closures() -> None:
    for path in (*TOOLS.joinpath("bazi").glob("*.csv"), *TOOLS.joinpath("ziwei").glob("*.csv")):
        with path.open(encoding="utf-8", newline="") as f:
            for row in csv.DictReader(f):
                for column, source in ENUM_SOURCES.items():
                    value = (row.get(column) or "").strip()
                    if not value:
                        continue
                    allowed = set(CONSTANTS[source]) if source else {"0", "1"}
                    assert value in allowed, f"{path}:{row.get('id')} {column}={value} 不在闭集"
