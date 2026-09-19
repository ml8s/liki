"""逐题错题审查台账契约：80 道错题必须都有可追踪的命理条件链。"""
import csv
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LEDGER = ROOT / "tests/benchmark/mingli160/evals/error_audit_iteration1.csv"


def test_error_ledger_covers_exact_80_errors_with_complete_logic() -> None:
    with LEDGER.open(encoding="utf-8-sig", newline="") as fh:
        rows = list(csv.DictReader(fh))
    assert len(rows) == 80
    ids = [row["error_id"] for row in rows]
    assert len(ids) == len(set(ids)) == 80
    assert all(row["error_id"].startswith("ftb_") for row in rows)
    assert all(row["case_id"].startswith("pan") for row in rows)
    assert all(row["standard_answer"] in set("ABCD") for row in rows)
    assert all(row["model_answer"] in set("ABCD") for row in rows)
    assert all(row["mingli_object"] for row in rows)
    assert all(row["chart_layers"] for row in rows)
    assert all("+" in row["mingli_condition_chain"] for row in rows)
    assert all(("候选" in row["mingli_condition_chain"] or "闭环" in row["mingli_condition_chain"]) for row in rows)
    assert all(row["failure_hypothesis"] for row in rows)
    assert all(row["table_optimization"].startswith("table: ") for row in rows)


def test_error_ledger_has_major_cluster_coverage() -> None:
    with LEDGER.open(encoding="utf-8-sig", newline="") as fh:
        rows = list(csv.DictReader(fh))
    domains = Counter(row["domain"] for row in rows)
    assert domains["婚姻"] == 26
    assert domains["事业"] == 15
    assert domains["家庭"] == 9
    assert domains["财运"] == 8
    assert domains["健康"] == 8
    tasks = Counter(row["task_type"] for row in rows)
    assert tasks["状态取象"] == 28
    assert tasks["应期"] == 18
    assert tasks["事件细则"] == 19
