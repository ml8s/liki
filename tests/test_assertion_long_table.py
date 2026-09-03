"""断语长表契约：元数据、条件、唯一性和 loader 聚合。"""
import csv
from pathlib import Path

import pytest

import _helpers  # noqa: F401
from duanyu import load_table

ROOT = Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools/assertions"


def _rows(name):
    with (ROOT / name).open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def test_assertion_long_table_counts_and_unique_ids():
    assertions = _rows("assertions.csv")
    conditions = _rows("assertion_conditions.csv")
    ids = [row["assertion_id"] for row in assertions]
    assert len(assertions) == 722
    assert len(conditions) == 1013
    assert len(ids) == len(set(ids)) == 722
    assert all(row["side"] in {"bazi", "ziwei"} for row in assertions)
    assert all(row["rule"] for row in assertions)


def test_all_conditions_reference_known_assertions_and_have_expected():
    assertions = {row["assertion_id"] for row in _rows("assertions.csv")}
    conditions = _rows("assertion_conditions.csv")
    for row in conditions:
        assert row["assertion_id"] in assertions
        assert row["factor"]
        assert row["expected"]


def test_loader_returns_metadata_and_typed_constraints():
    rows = load_table("bazi_十神.csv")
    assert rows
    assert all({"id", "事件", "约束", "结论", "依据", "经典原文"} <= set(row) for row in rows)
    assert any(row["约束"] for row in rows)
    typed = [value for row in rows for value in row["约束"].values()]
    assert any(value == 1 or value == 0 for value in typed)
    assert any(isinstance(value, str) and value not in ("0", "1") for value in typed)


def test_loader_optional_missing_table_and_required_missing_table():
    assert load_table("bazi_missing_domain", required=False) == []
    with pytest.raises(FileNotFoundError):
        load_table("bazi_missing_domain")


def test_rule_table_groups_all_sides_and_rules():
    rows = load_table("bazi_格局.csv")
    assert rows
    assert all(row["约束"] for row in rows)
    assert {row["id"] for row in rows}
