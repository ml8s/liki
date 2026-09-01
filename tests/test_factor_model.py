"""因子实现、constants 闭集与 FACTOR_MODEL.md 的一致性测试。"""
from __future__ import annotations

import csv
import json
import re
from collections import defaultdict, deque
from pathlib import Path

import _helpers  # noqa: F401
from factor_tables import load_long_rows

ROOT = Path(__file__).resolve().parents[1]
DOC = ROOT / "docs" / "FACTOR_MODEL.md"
TOOLS = ROOT / "skills" / "liki-bazi" / "tools"
META = {"因子", "术数", "原语直通", "依据"}
D = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
CLASSES = set(D["十神大类"])
ROLES = set(D["六亲角色"])
ATOM_TEN_GODS = set(D["十神"])
STAR_GROUPS = {k for k, v in D.items() if isinstance(v, list) and k.startswith("紫微")}
COMPLEX_DIRECT = {"财库现[]", "财星入墓[]", "官杀取清[]"}


def read_groups(path: Path) -> dict[str, list[dict[str, str]]]:
    grouped: dict[str, list[dict]] = defaultdict(list)
    for row in load_long_rows(str(path)):
        grouped[row["因子"]].append(row)
    return grouped


def natal_groups() -> dict[str, list[dict]]:
    return read_groups(TOOLS / "factors" / "factors.csv")


def flow_groups() -> dict[str, list[dict]]:
    return read_groups(TOOLS / "factors" / "factors_liunian.csv")


def conditions(rows: list[dict]) -> list[list[tuple[str, str]]]:
    return [[(k, str(v)) for k, v in row["conds"].items()] for row in rows]


def factor_kind(rows: list[dict]) -> str:
    direct = (rows[0].get("直通") or "").strip()
    if direct:
        return "复合" if direct in COMPLEX_DIRECT else "直通原子"
    condition_rows = conditions(rows)
    if len(rows) != 1 or len(condition_rows[0]) != 1:
        return "复合"
    key = condition_rows[0][0][0]
    if "[" not in key:
        return "复合"
    match = re.match(r"^([^\[]+)\[(.*)\]$", key)
    if not match:
        return "复合"
    op, args = match.group(1), match.group(2).split(",")
    if op in {"现", "透", "藏", "得令", "有根", "克", "生", "为用", "为忌"}:
        return "复合" if any(x in CLASSES | ROLES | STAR_GROUPS for x in args) else "提取原子"
    if op == "数量至少":
        return "复合" if any(x in CLASSES | ROLES for x in args[1:]) else "提取原子"
    if op == "宫含":
        return "复合" if args[1] in CLASSES | ROLES | STAR_GROUPS | {"煞星", "无主星", "任意"} else "提取原子"
    if op == "大运十神":
        return "复合" if args[1] in CLASSES | ROLES else "提取原子"
    if op in {"流年透", "流年值", "流年合", "流年冲", "流年克", "大运窗口流年", "换运流年"}:
        return "复合" if args[0] in CLASSES | ROLES else "提取原子"
    if op == "引用本命":
        return "复合"
    return "提取原子"


def factor_value(rows: list[dict]) -> str:
    direct = (rows[0].get("直通") or "").strip()
    return "string" if direct and "任意" in direct else "0/1"


def factor_definition(rows: list[dict]) -> str:
    direct = (rows[0].get("直通") or "").strip()
    if direct:
        return direct
    variants = []
    for row_conditions in conditions(rows):
        variants.append(" AND ".join(f"{k}={v}" for k, v in row_conditions) or "TRUE")
    return " OR ".join(f"({variant})" for variant in variants)



def doc_rows(title: str, heading_level: int = 3) -> list[list[str]]:
    text = DOC.read_text(encoding="utf-8")
    prefix = "#" * heading_level
    start = text.index(f"{prefix} {title}\n")
    following = re.search(r"^#{2,3} ", text[start + 10:], re.M)
    stop = start + 10 + (following.start() if following else len(text[start + 10:]))
    rows: list[list[str]] = []
    for line in text[start:stop].splitlines():
        if not line.startswith("| ") or line.startswith("| #") or line.startswith("|---"):
            continue
        rows.append([cell.strip() for cell in line.strip("|").split("|")])
    return rows


def test_constant_closures_are_partitioned_and_complete() -> None:
    assert len(D["十神"]) == len(set(D["十神"])) == 10
    members = [x for group in D["十神大类"].values() for x in group]
    assert len(members) == len(set(members))
    assert set(members) == ATOM_TEN_GODS
    assert set(D["五行"]) == {"木", "火", "土", "金", "水"}
    assert set(D["天干"]) == set(D["天干五行"])
    assert set(D["地支"]) == set(D["地支五行"])
    assert len(D["十二长生"]) == 12
    assert set(D["旺衰状态"]) == {"旺", "相", "休", "囚", "死"}
    assert len(D["紫微主星"]) == len(set(D["紫微主星"])) == 14
    assert set(D["紫微六吉星"]) == {"左辅", "右弼", "文昌", "文曲", "天魁", "天钺"}
    assert set(D["紫微文星"]) <= set(D["紫微六吉星"])


def test_relation_closures_are_complete() -> None:
    relation_key_counts = {
        "天干五合": 10, "六合": 12, "三合": 12, "三会": 12,
        "六冲": 12, "六害": 12,
    }
    assert all(len(D[name]) == count for name, count in relation_key_counts.items())
    assert len(D["旬空"]) == 6
    # 关系映射必须是对称闭集。
    for name in ("六合", "六冲", "六害", "天干五合"):
        for a, b in D[name].items():
            assert D[name][b] == a


def test_documented_natal_inventory_matches_implementation() -> None:
    groups = natal_groups()
    atoms = doc_rows("1. 原子因子（334 个）")
    compounds = doc_rows("2. 复合因子（110 个）")
    documented = [row[1] for row in atoms + compounds]
    assert len(groups) == 444
    assert len(atoms) == 334
    assert len(compounds) == 110
    assert len(documented) == len(set(documented))
    assert set(documented) == set(groups)
    art = {"common": "共同", "bazi": "八字", "ziwei": "紫微"}
    for row in atoms + compounds:
        name = row[1]
        assert row[2] == art[groups[name][0]["术数"]]
        assert row[3] == factor_kind(groups[name])
        assert row[4] == factor_value(groups[name])
        assert row[5] == factor_definition(groups[name])
    text = DOC.read_text(encoding="utf-8")
    assert "| 本命因子 | 444 |" in text
    assert "| 本命原子因子 | 334 |" in text
    assert "| 本命复合因子 | 110 |" in text


def test_documented_flow_inventory_matches_implementation() -> None:
    groups = flow_groups()
    rows = doc_rows("二、流年因子（70 个）", heading_level=2)
    assert len(groups) == 70
    assert len(rows) == 70
    assert {row[1] for row in rows} == set(groups)
    art = {"common": "共同", "bazi": "八字", "ziwei": "紫微"}
    for row in rows:
        name = row[1]
        assert row[2] == art[groups[name][0]["术数"]]
        assert row[3] == factor_kind(groups[name])
        assert row[4] == factor_value(groups[name])
        assert row[5] == factor_definition(groups[name])
    assert "| 流年因子 | 70 |" in DOC.read_text(encoding="utf-8")


def test_context_is_not_factor_and_flow_targets_are_explicit() -> None:
    assert "性别" not in natal_groups()
    forbidden = {
        "性别", "大运窗口", "本命婚凶", "食伤克官", "比劫伤官格", "财星受克",
        "从杀格", "从财格", "从儿格", "从杀成格", "建禄格", "羊刃格",
        "流年目标星透", "流年值宫", "流年合会", "流年冲", "流年克目标星",
    }
    assert not forbidden & set(natal_groups())
    assert not forbidden & set(flow_groups())
    for name in flow_groups():
        if name.startswith("流年配偶星") or name.startswith("流年财星") or name.startswith("流年母星") or name.startswith("流年子女星"):
            assert name.split("流年", 1)[1]
    assert "流年官杀换运首年" in flow_groups()


def test_stable_factor_names_use_consistent_entities() -> None:
    forbidden = {
        "宫破",
        "本命宫破",
        "配偶星透干",
        "配偶星藏支",
        "财透",
        "财透有根",
        "财旺",
        "财弱",
        "财得令",
        "财为用",
        "财为忌",
        "印透",
    }
    assert not forbidden & set(natal_groups())
    assert not forbidden & set(flow_groups())

    required = {
        "夫妻宫破",
        "配偶星透",
        "配偶星藏",
        "财星透",
        "财星透根",
        "财星旺",
        "财星弱",
        "财星得令",
        "财星为用",
        "财星为忌",
        "印星透",
    }
    assert required <= set(natal_groups())
    assert {"本命夫妻宫破"} <= set(flow_groups())


def test_factor_long_table_schema_and_unique_signatures() -> None:
    expected_fields = [
        "factor_id", "shushi", "group_id", "term_index", "kind",
        "expression", "expected", "basis",
    ]
    for filename in ("factors.csv", "factors_liunian.csv"):
        path = TOOLS / "factors" / filename
        with path.open(encoding="utf-8", newline="") as f:
            reader = csv.DictReader(f)
            fields = list(reader.fieldnames or [])
            rows = list(reader)
        assert fields == expected_fields
        assert all(row["factor_id"].strip() for row in rows)
        assert all(row["kind"] in {"direct", "condition", "factor_ref"} for row in rows)
        grouped = read_groups(path)
        signatures: dict[tuple, str] = {}
        for name, group in grouped.items():
            if any((row.get("直通") or "").strip() for row in group):
                continue
            signature = tuple(sorted(tuple(sorted(item["conds"].items())) for item in group))
            assert signature not in signatures, f"{name} duplicates {signatures[signature]}"
            signatures[signature] = name


def test_unreferenced_factors_are_only_complete_use_family() -> None:
    groups = natal_groups()
    names = set(groups)
    refs = set()
    with (TOOLS / "assertions" / "assertion_conditions.csv").open(encoding="utf-8", newline="") as f:
        for row in csv.DictReader(f):
            if row.get("factor") in names:
                refs.add(row["factor"])
    for row in flow_groups().values():
        for item in row:
            for key in item["conds"]:
                if key.startswith("引用本命["):
                    name = key[len("引用本命["):-1]
                    if name in names:
                        refs.add(name)
    reach = set(refs)
    queue = deque(refs)
    while queue:
        name = queue.popleft()
        for item in groups[name]:
            for key in item["conds"]:
                if "[" not in key and key in names and key not in reach:
                    reach.add(key)
                    queue.append(key)
    assert set(names - reach) == set()
