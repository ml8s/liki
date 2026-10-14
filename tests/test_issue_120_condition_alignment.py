import csv
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CONDITIONS = ROOT / "counsel/app/natal/tools/assertions/assertion_conditions.csv"


def _factors() -> dict[str, set[str]]:
    result: dict[str, set[str]] = {}
    with CONDITIONS.open(encoding="utf-8", newline="") as fh:
        for row in csv.DictReader(fh):
            result.setdefault(row["assertion_id"], set()).add(row["factor"])
    return result


def test_health_internal_conflict_assertions_use_internal_conflict_factors():
    factors = _factors()
    expected = {
        "jk_311": {"金木交战"},
        "jk_312": {"水火相战"},
        "jk_313": {"土壅木郁"},
        "jk_314": {"心肝俱弱"},
        "jk_315": {"肾心不交"},
    }
    for assertion_id, wanted in expected.items():
        assert factors.get(assertion_id, set()) == wanted, assertion_id


def test_health_structural_assertions_use_matching_branch_structures():
    factors = _factors()
    expected = {
        "jk_320": {"地支三合寅午戌"},
        "jk_321": {"地支三合巳酉丑"},
        "jk_322": {"地支三合申子辰"},
        "jk_323": {"地支三合亥卯未"},
        "jk_324": {"地支三会寅卯辰"},
    }
    for assertion_id, wanted in expected.items():
        assert factors.get(assertion_id, set()) == wanted, assertion_id
