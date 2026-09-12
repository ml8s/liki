"""因子实现、constants 闭集与 BAZI_MODEL.md 的一致性测试。"""
from __future__ import annotations

import csv
import json
from collections import defaultdict, deque
from pathlib import Path

import _helpers  # noqa: F401
from factor_tables import load_long_rows

ROOT = Path(__file__).resolve().parents[1]
DOC = ROOT / "docs" / "BAZI_MODEL.md"
TOOLS = ROOT / "skills" / "liki-bazi" / "tools"
D = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
ATOM_TEN_GODS = set(D["十神"])


def read_groups(path: Path) -> dict[str, list[dict[str, str]]]:
    grouped: dict[str, list[dict]] = defaultdict(list)
    for row in load_long_rows(str(path)):
        grouped[row["因子"]].append(row)
    return grouped


def natal_groups() -> dict[str, list[dict]]:
    return read_groups(TOOLS / "factors" / "factors.csv")


def flow_groups() -> dict[str, list[dict]]:
    return read_groups(TOOLS / "factors" / "factors_liunian.csv")


def test_constant_closures_are_partitioned_and_complete() -> None:
    assert len(D["十神"]) == len(set(D["十神"])) == 10
    members = [x for group in D["十神大类"].values() for x in group]
    assert len(members) == len(set(members))
    assert set(members) == ATOM_TEN_GODS
    assert set(D["五行"]) == {"木", "火", "土", "金", "水"}
    assert {"天干五行", "地支五行", "五行生克", "得令状态", "十神旺弱规则", "旺衰"}.isdisjoint(D)
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
    assert set(D["三合半合"]) == {"子", "午", "卯", "酉"}
    for imperial_branch, partners in D["三合半合"].items():
        assert len(partners) == 2
        assert set(partners) == set(D["三合"][imperial_branch])
    assert len(D["旬空"]) == 6
    # 关系映射必须是对称闭集。
    for name in ("六合", "六冲", "六害", "天干五合"):
        for a, b in D[name].items():
            assert D[name][b] == a


def test_factor_inventory_has_single_source_of_truth() -> None:
    groups = natal_groups()
    flows = flow_groups()
    raw_groups: dict[str, list[dict[str, str]]] = defaultdict(list)
    raw_flows: dict[str, list[dict[str, str]]] = defaultdict(list)
    for filename, target in (
        ("factors.csv", raw_groups),
        ("factors_liunian.csv", raw_flows),
    ):
        with (TOOLS / "factors" / filename).open(encoding="utf-8", newline="") as source:
            for row in csv.DictReader(source):
                target[row["factor_id"]].append(row)

    def category(rows: list[dict[str, str]]) -> str:
        classes = set(D["十神大类"])
        roles = set(D["六亲角色"])
        ziwei_groups = {
            key for key, value in D.items()
            if isinstance(value, list) and key.startswith("紫微")
        }
        compound_terms = classes | roles | ziwei_groups
        complex_direct = {"财库现", "财星入墓", "官杀取清"}
        if rows[0]["kind"] == "direct":
            return "factor_ref" if rows[0]["factor_id"] in complex_direct else "direct"
        kinds = {row["kind"] for row in rows}
        if "factor_ref" in kinds or len(rows) > 1 or any(
            len(group) > 1
            for group in (
                [row for row in rows if row["group_id"] == group_id]
                for group_id in {row["group_id"] for row in rows}
            )
        ):
            return "factor_ref"
        expression = rows[0]["expression"]
        if "[" not in expression:
            return "factor_ref"
        operator, raw_args = expression[:-1].split("[", 1)
        args = raw_args.split(",") if raw_args else []
        if operator in {"现", "透", "藏", "得令", "有根", "克", "生", "为用", "为忌"}:
            return "factor_ref" if any(arg in compound_terms for arg in args) else "condition"
        if operator == "数量至少":
            return "factor_ref" if any(arg in classes | roles for arg in args[1:]) else "condition"
        if operator == "宫含":
            compound_targets = compound_terms | {"煞星", "无主星", "任意"}
            return "factor_ref" if args[1] in compound_targets else "condition"
        if operator == "大运十神":
            return "factor_ref" if args[1] in classes | roles else "condition"
        if operator in {"流年透", "流年值", "流年合", "流年冲", "流年克", "大运窗口流年", "换运流年"}:
            return "factor_ref" if args[0] in classes | roles else "condition"
        if operator == "引用本命":
            return "factor_ref"
        return "condition"

    def side(rows: list[dict]) -> str:
        return next(iter({row["shushi"] for row in rows}))

    natal_categories = defaultdict(int)
    natal_sides = defaultdict(int)
    for rows in raw_groups.values():
        natal_categories[category(rows)] += 1
        natal_sides[side(rows)] += 1
    flow_categories = defaultdict(int)
    flow_sides = defaultdict(int)
    for rows in raw_flows.values():
        flow_categories[category(rows)] += 1
        flow_sides[side(rows)] += 1

    assert len(groups) == 475
    assert len(flows) == 101

    text = DOC.read_text(encoding="utf-8")
    assert "tools/factors/factors.csv" in text
    assert "tools/factors/factors_liunian.csv" in text
    assert "CSV 是因子清单唯一事实源" in text
    assert f"| 本命因子 | {len(groups)} |" in text
    assert f"| 本命八字因子 | {natal_sides['bazi']} |" in text
    assert f"| 本命紫微因子 | {natal_sides['ziwei']} |" in text
    assert f"| 本命定义组 | {sum(map(len, groups.values()))} |" in text
    assert f"| 本命数据行 | {sum(map(len, raw_groups.values()))} |" in text
    assert f"| 本命直通原子 | {natal_categories['direct']} |" in text
    assert f"| 本命提取原子 | {natal_categories['condition']} |" in text
    assert f"| 本命复合因子 | {natal_categories['factor_ref']} |" in text
    assert "| 流年因子 | 101 |" in text
    assert f"| 流年八字因子 | {flow_sides['bazi']} |" in text
    assert f"| 流年紫微因子 | {flow_sides['ziwei']} |" in text
    assert f"| 流年直通原子 | {flow_categories['direct']} |" in text
    assert f"| 流年提取原子 | {flow_categories['condition']} |" in text
    assert f"| 流年复合因子 | {flow_categories['factor_ref']} |" in text


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


def test_yong_shen_guidance_requires_effective_support_not_team_labels() -> None:
    text = (ROOT / "skills/liki-bazi/domains/bazi/yongshen.md").read_text(encoding="utf-8")
    assert "数量不等于有效力量" in text
    assert "禁止用 `wuxing_count` 做加总评分" in text
    assert "喜神不是“同党标签”" in text
    for required in ("生克方向", "力量反转", "通关条件", "合冲牵制", "调候辅证"):
        assert required in text
    assert "若候选五行克用神，不得直接作喜神" in text

    model = DOC.read_text(encoding="utf-8")
    assert "`duanyu.query(rule=用神)` 除断语外返回 `yong_shen_context`" in model


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
    flow = flow_groups()
    names = set(groups)
    natal_refs = set()
    flow_refs = set()
    with (TOOLS / "assertions" / "assertion_conditions.csv").open(encoding="utf-8", newline="") as f:
        for row in csv.DictReader(f):
            factor = row["factor"]
            if factor in names:
                natal_refs.add(factor)
            if factor in flow:
                flow_refs.add(factor)
    for row in flow_groups().values():
        for item in row:
            for key in item["conds"]:
                if key.startswith("引用本命["):
                    name = key[len("引用本命["):-1]
                    if name in names:
                        natal_refs.add(name)
                    if name in flow:
                        flow_refs.add(name)

    def closure(table: dict[str, list[dict]], seeds: set[str]) -> set[str]:
        reach = set(seeds)
        queue = deque(seeds)
        while queue:
            name = queue.popleft()
            for item in table[name]:
                for key in item["conds"]:
                    if "[" not in key and key in table and key not in reach:
                        reach.add(key)
                        queue.append(key)
        return reach

    natal_reach = closure(groups, natal_refs)
    flow_reach = closure(flow, flow_refs)
    assert set(names - natal_reach) == set()
    assert set(flow) == flow_reach
