"""命理规则出处契约：断语必须可溯源到经典，不得按评测选项反向拟合。"""
import csv
import re
from pathlib import Path

TABLE = Path(__file__).resolve().parents[1] / "skills/liki/natal/tools/assertions/assertions.csv"
CANONICAL_SOURCES = {
    "渊海子平", "三命通会", "滴天髓", "子平真诠", "穷通宝鉴",
    "紫微斗数全书", "女命赋", "黄帝内经", "协纪辨方书", "麻衣相法",
}
FORBIDDEN_FITTING_MARKERS = {
    "pan01", "pan20", "mingli", "benchmark", "评测", "选项", "题目",
    "按题指认", "按诉求", "标准答案",
}
# 本轮新增/收紧断语必须至少落到下表经典口径。
CHANGED_ASSERTIONS = {
    "ymar_103": {"渊海子平", "子平真诠"},
    "ymar_108": {"滴天髓"},
    "ymar_112": {"三命通会"},
    "ycai_103": {"渊海子平", "子平真诠", "滴天髓"},
    "ycai_108": {"三命通会", "子平真诠"},
    "ying_h19": {"三命通会", "渊海子平"},
    "ycai_121": {"滴天髓", "子平真诠"},
    "ycai_122": {"滴天髓", "子平真诠"},
    "hun_421": {"渊海子平", "子平真诠"},
    "hun_422": {"三命通会", "滴天髓"},
    "hun_423": {"渊海子平", "三命通会"},
    "shi_115": {"子平真诠"},
    "shi_116": {"子平真诠", "滴天髓"},
    "shi_117": {"滴天髓"},
    "shi_118": {"渊海子平", "子平真诠"},
    "shi_119": {"子平真诠", "三命通会"},
    "shi_120": {"子平真诠"},
}


def _rows():
    with TABLE.open(encoding="utf-8-sig", newline="") as fh:
        return list(csv.DictReader(fh))


def test_every_assertion_has_canonical_source() -> None:
    rows = _rows()
    assert rows
    missing = []
    for row in rows:
        provenance = row["依据"] + row["经典依据"]
        if not any(source in provenance for source in CANONICAL_SOURCES):
            missing.append((row["assertion_id"], provenance))
    assert not missing, f"缺少经典出处: {missing[:20]}"


def test_changed_rules_have_expected_classical_sources() -> None:
    rows = {row["assertion_id"]: row for row in _rows()}
    missing = []
    for aid, expected_sources in CHANGED_ASSERTIONS.items():
        assert aid in rows, f"缺少本轮契约断语: {aid}"
        provenance = rows[aid]["依据"] + rows[aid]["经典依据"]
        if not any(source in provenance for source in expected_sources):
            missing.append((aid, expected_sources, provenance))
    assert not missing, f"新增/收紧断语出处不足: {missing}"


def test_no_assertion_is_fit_to_benchmark_options() -> None:
    for row in _rows():
        text = "|".join(row.values())
        for marker in FORBIDDEN_FITTING_MARKERS:
            assert marker not in text, f"{row['assertion_id']} 含拟合/评测残留: {marker}"
