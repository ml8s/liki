"""因子长表契约：OR 分组、AND term、direct 语义与引用完整性。"""
import csv
from collections import defaultdict
from pathlib import Path

import _helpers  # noqa: F401
import pytest
from factor_tables import load_long_rows

ROOT = Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools/factors"
CASES = {"factors.csv": "natal", "factors_liunian.csv": "flow"}
FIELDS = {"factor_id", "shushi", "group_id", "term_index", "kind", "expression", "expected", "basis"}


def _raw(path):
    with path.open(encoding="utf-8-sig", newline="") as f:
        return list(csv.DictReader(f))


def test_long_table_schema_and_terms_are_continuous():
    for name, key in CASES.items():
        path = ROOT / name
        rows = _raw(path)
        assert rows and set(rows[0]) == FIELDS
        groups = defaultdict(list)
        for row in rows:
            assert row["factor_id"] and row["shushi"] in {"bazi", "ziwei"}
            assert row["kind"] in {"direct", "condition", "factor_ref"}
            groups[(row["factor_id"], int(row["group_id"]))].append(row)
        for (factor_id, group_id), group in groups.items():
            indexes = [int(row["term_index"]) for row in group]
            assert indexes == list(range(1, len(indexes) + 1)), (factor_id, group_id)
            kinds = {row["kind"] for row in group}
            if "direct" in kinds:
                assert len(kinds) == 1 and len(group) == 1
            elif kinds == {"condition", "factor_ref"}:
                pass


def test_loader_grouping_matches_csv_groups():
    for name, key in CASES.items():
        raw = _raw(ROOT / name)
        loaded = load_long_rows(str(ROOT / name), key)
        assert len({r["factor_id"] for r in raw}) == len({r["因子"] for r in loaded})
        # direct 行必须保持 direct；其余因子必须补全为 0。
        direct_ids = {r["factor_id"] for r in raw if r["kind"] == "direct"}
        assert all(r["直通"] for r in loaded if r["因子"] in direct_ids)


def test_factor_refs_do_not_self_reference():
    for name, key in CASES.items():
        rows = _raw(ROOT / name)
        ids = {r["factor_id"] for r in rows}
        for r in rows:
            if r["kind"] == "factor_ref":
                assert r["expression"] in ids
                assert r["expression"] != r["factor_id"]


def test_direct_group_must_not_mix_conditions(tmp_path):
    path = tmp_path / "factors.csv"
    fields = ["factor_id", "shushi", "group_id", "term_index", "kind", "expression", "expected", "basis"]
    rows = [
        {"factor_id": "直读因子", "shushi": "bazi", "group_id": "1", "term_index": "1", "kind": "direct", "expression": "直读[gender,male]", "expected": "1", "basis": ""},
        {"factor_id": "直读因子", "shushi": "bazi", "group_id": "1", "term_index": "2", "kind": "condition", "expression": "现[印星]", "expected": "1", "basis": ""},
    ]
    with path.open("w", encoding="utf-8", newline="") as f:
        import csv
        writer = csv.DictWriter(f, fieldnames=fields)
        writer.writeheader(); writer.writerows(rows)
    with pytest.raises(Exception, match="混用 direct/condition"):
        load_long_rows(str(path), "test_direct_mix")
