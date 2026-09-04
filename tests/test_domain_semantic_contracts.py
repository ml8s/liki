"""当前命理数据与机械算子的领域契约。"""
import ast
import csv
from pathlib import Path

import _helpers  # noqa: F401
import duanyu
import operators_liunian
import operators_natal
from domain_snapshot import load_contract
import factors
from factor_tables import load_factor_rows


def test_financial_star_damages_resource_requires_real_target():
    rows = [r for r in load_factor_rows() if r["因子"] == "财坏印"]
    assert rows
    assert [r["conds"] for r in rows] == [{
        "透[财星]": 1,
        "现[印星]": 1,
        "克[财,印]": 1,
    }]


def test_factor_groups_do_not_duplicate_definitions():
    root = Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools/factors"
    for name in ("factors.csv", "factors_liunian.csv"):
        with (root / name).open(encoding="utf-8", newline="") as source:
            rows = list(csv.DictReader(source))
        groups: dict[tuple[str, str], list[dict]] = {}
        for row in rows:
            groups.setdefault((row["factor_id"], row["group_id"]), []).append(row)
        signatures: dict[tuple, list[tuple[str, str]]] = {}
        for key, terms in groups.items():
            signature = tuple(sorted(
                (row["kind"], row["expression"], row["expected"])
                for row in terms
            ))
            signatures.setdefault(signature, []).append(key)
        duplicates = [keys for keys in signatures.values() if len(keys) > 1]
        assert not duplicates, f"{name}: {duplicates}"


def test_factor_model_preserves_assertion_unconsumed_domain_fact():
    root = Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools"
    with (root / "factors/factors.csv").open(encoding="utf-8", newline="") as source:
        factor_ids = {row["factor_id"] for row in csv.DictReader(source)}
    with (root / "assertions/assertion_conditions.csv").open(
        encoding="utf-8", newline=""
    ) as source:
        asserted = {row["factor"] for row in csv.DictReader(source)}
    assert "身弱" in factor_ids
    assert "身弱" not in asserted


def test_snapshot_side_request_is_rejected_for_assertion_only_side():
    import pytest
    from errors import FactorEvaluateError

    with pytest.raises(FactorEvaluateError, match="sides"):
        factors.evaluate_snap_from_pan({"gender": "male"}, sides={"common"})


def test_factor_rows_require_declared_side(tmp_path):
    import pytest
    from factor_tables import FactorTableError, load_long_rows

    path = tmp_path / "factors.csv"
    path.write_text(
        "factor_id,shushi,group_id,term_index,kind,expression,expected,basis\n"
        "test,,1,1,condition,现[印星],1,\n",
        encoding="utf-8",
    )
    with pytest.raises(FactorTableError, match="shushi 不能为空"):
        load_long_rows(path)


def test_half_sanhe_requires_imperial_branch():
    ctx = {"liunian": {}, "zw_liunian": {}, "year": 2026}
    assert operators_liunian._liu_op("半合", ["子", "辰"], "male", {}, ctx) == 1
    assert operators_liunian._liu_op("半合", ["亥", "未"], "male", {}, ctx) == 0
    source_ctx = {
        "liunian": {"nian_zhi": "辰"},
        "zw_liunian": {},
        "year": 2026,
        "chart": {"chart": {"ri": {"zhi": "子"}}},
    }
    assert operators_liunian._liu_op(
        "半合", ["日支", "流年支"], "male", {}, source_ctx
    ) == 1


def test_pillar_sources_cover_all_four_pillars():
    ctx = {
        "liunian": {"nian_zhi": "午"},
        "zw_liunian": {},
        "year": 2026,
        "chart": {"chart": {
            "nian": {"gan": "甲", "zhi": "子"},
            "yue": {"gan": "乙", "zhi": "丑"},
            "ri": {"gan": "丙", "zhi": "寅"},
            "shi": {"gan": "丁", "zhi": "卯"},
        }},
    }
    assert operators_liunian._source_zhi("年支", ctx) == "子"
    assert operators_liunian._source_zhi("月支", ctx) == "丑"
    assert operators_liunian._source_zhi("日支", ctx) == "寅"
    assert operators_liunian._source_zhi("时支", ctx) == "卯"


def test_ten_god_wuxing_follows_day_master_relation_table():
    const = duanyu.load_constants()
    assert operators_liunian._target_wuxing_from_day_master(
        "甲", operators_liunian._target_stars("比劫", "male", const), const
    ) == "木"
    assert operators_liunian._target_wuxing_from_day_master(
        "甲", operators_liunian._target_stars("食伤", "male", const), const
    ) == "火"
    assert operators_liunian._target_wuxing_from_day_master(
        "甲", operators_liunian._target_stars("财星", "male", const), const
    ) == "土"
    assert operators_liunian._target_wuxing_from_day_master(
        "甲", operators_liunian._target_stars("官杀", "male", const), const
    ) == "金"
    assert operators_liunian._target_wuxing_from_day_master(
        "甲", operators_liunian._target_stars("印星", "male", const), const
    ) == "水"


def test_current_limit_does_not_pollute_natal_layers():
    assert duanyu.CURRENT_LIMIT_RULES == frozenset({"大运", "大限"})
    assert "十神" not in duanyu.CURRENT_LIMIT_RULES
    assert "用神" not in duanyu.CURRENT_LIMIT_RULES


def test_yearly_scene_aliases_cover_their_domain_facts():
    assert {"年六亲", "年大运", "年旺衰"} <= set(duanyu.SCENE_ALIASES["yingqi"])
    assert "年子女" in duanyu.SCENE_ALIASES["yearly_family"]


def test_brief_keeps_assertion_id_for_cauijue_traceability():
    source = [{"id": "ymar_101", "事件": "流年婚动", "结论": "婚动"}]
    assert duanyu.brief(source) == source


def test_brief_keeps_controlled_event_taxonomy():
    source = [{
        "id": "ymar_101", "领域": "婚姻", "事件类型": "引动",
        "时间层": "流年", "事件": "流年婚动", "结论": "婚动",
    }]
    assert duanyu.brief(source) == source


def test_common_assertions_merge_both_system_snapshots():
    snapshot = {
        "_snapshot_type": "liunian",
        "八字": {"流年神煞天乙贵人": 1},
        "紫微": {"流年迁移宫禄": 1},
        "context": {},
    }
    result = duanyu._match_rule("年神煞", snapshot)
    assert [row["id"] for row in result["合参"]] == ["ycai_120"]


def test_current_daxian_palace_is_mechanical():
    chart = {"solar": "1990-05-20T12:00:00", "ziwei_daxian": [
        {"gong": "命宫", "start_year": 1990, "end_year": 1999},
        {"gong": "兄弟", "start_year": 2000, "end_year": 2009},
    ]}
    from operators_natal import _op
    assert _op("大限宫位", ["当前", "任意"], "male", chart, 2005) == "兄弟"
    assert _op("大限宫位", ["当前", "夫妻"], "male", chart, 2005) == 0


def test_current_daxian_requires_explicit_or_server_year():
    from operators_natal import _op
    chart = {"ziwei_daxian": _helpers.valid_daxian()}
    assert _op("大限宫位", ["当前", "任意"], "male", chart, 0) == ""
    assert _op("大限宫位", ["当前", "命宫"], "male", chart, 0) == 0


def test_flow_star_palace_is_mechanical():
    ctx = {"liunian": {}, "zw_liunian": {"gong_wei": [
        {"name": "子女", "xing_yao": ["流羊"]},
        {"name": "夫妻", "xing_yao": ["流鸾"]},
    ]}}
    assert operators_liunian._liu_op("流曜入宫", ["流羊", "子女宫"], "male", {}, ctx) == 1
    assert operators_liunian._liu_op("流曜入宫", ["流鸾", "子女宫"], "male", {}, ctx) == 0


def test_explicit_pillar_clash_uses_requested_pillar():
    ctx = {
        "liunian": {"nian_zhi": "午", "natal_interactions": [{
            "zhi_rels": [{"zhi_a": "午", "zhi_b": "子", "type": "六冲"}]
        }]},
        "zw_liunian": {},
        "chart": {"chart": {"nian": {"zhi": "子"}, "ri": {"zhi": "卯"}}},
    }
    assert operators_liunian._liu_op("流年冲", ["年支"], "male", {}, ctx) == 1
    assert operators_liunian._liu_op("流年冲", ["日支"], "male", {}, ctx) == 0
    direct_ctx = {**ctx, "liunian": {"nian_zhi": "午", "natal_interactions": []}}
    assert operators_liunian._liu_op("流年冲", ["年支"], "male", {}, direct_ctx) == 1


def test_query_year_is_rejected_for_pure_natal_rules():
    import pytest
    with pytest.raises(ValueError, match="纯本命域"):
        duanyu.query("十神", {
            "solar": "1990-05-20T12:00:00",
            "lunar": {"year": 1990, "month": 4, "day": 26},
            "gender": "male",
            "chart": {p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
            "full": {p: {"gan": "甲", "zhi": "子"} for p in ("nian", "yue", "ri", "shi")},
            "yongshen": {}, "ziwei": {"gong_wei": []}, "ziwei_daxian": _helpers.valid_daxian(),
        }, year=2005)


def test_required_natal_factors_include_reference_closure():
    tables = [[{
        "id": "test",
        "约束组": [{"比劫夺财": 1}],
    }]]
    assert duanyu._required_natal_factors(tables) >= {
        "比劫夺财", "比劫旺", "财星弱", "财星现"
    }


def test_documented_query_rules_use_runtime_whitelist():
    from pathlib import Path
    import re
    root = Path(__file__).resolve().parents[1] / "skills/liki-bazi"
    documented = set()
    for path in list((root / "app").glob("*.md")) + list((root / "domains").glob("*/*.md")):
        documented.update(re.findall(r"query\(rule=([^,\)\s]+)", path.read_text(encoding="utf-8")))
    assert documented
    assert documented <= duanyu.NATAL_RULES


def test_required_flow_factors_use_assertion_conditions():
    tables = [[{
        "id": "test",
        "约束组": [{"流年配偶星透": 1}],
    }]]
    assert duanyu._required_flow_factors(tables) == {"流年配偶星透"}


def test_dayun_spouse_star_uses_evaluation_gender():
    from operators_natal import _op
    base = {
        "shishen": {},
        "dayun_steps": [
            {"name": "戊辰", "shi_shen": "正财运",
             "start_year": 2011, "end_year": 2020}
        ],
    }
    assert _op("大运十神", ["当前", "配偶星"], "male", base, 2015) == 1
    assert _op("大运十神", ["当前", "配偶星"], "female", base, 2015) == 0


def test_natal_factors_for_flow_cover_explicit_and_implicit_references():
    explicit = duanyu.natal_factors_for_flow({"本命官杀为用"})
    assert "官杀为用" in explicit
    implicit = duanyu.natal_factors_for_flow({"流年支受克"})
    assert {"木旺", "火旺", "土旺", "金旺", "水旺"} <= implicit


def test_domain_configuration_covers_mechanical_operator_contracts():
    const = duanyu.load_constants()
    assert set(const["十神大类日主关系"]) == set(const["十神大类"])
    for relation in const["十神大类日主关系"].values():
        assert relation["关系"] in {"同", "生", "克"}
        if relation["关系"] != "同":
            assert relation["方向"] in {"出", "入"}
    assert const["十神大类"][const["财星十神大类"]] == ["正财", "偏财"]
    assert const["十神大类"][const["印星十神大类"]] == ["正印", "偏印"]
    assert const["官杀取清"]["十神大类"] in const["十神大类"]
    assert set(const["性别闭集"]) == set(const["性别别名"].values())
    assert const["大限段数"] == 12
    assert set(const["关系取合类型"]) == {"六合", "三合", "三会"}
    assert const["关系取冲类型"] == "六冲"
    assert set(const["格局十神"]) <= set(const["月令格局"])
    assert set(const["格局十神"].values()) <= set(const["十神"])
    assert set(const["干支来源"]) == {
        "流年", "大运", "日柱", "流年干", "大运干", "日干",
        "流年支", "大运支", "年支", "月支", "日支", "时支",
    }
    pillar_sources = {
        spec["柱"] for spec in const["干支来源"].values()
        if spec["源"] == "四柱"
    }
    assert pillar_sources == set(const["四柱"])
    assert set(const["算子柱位"]) <= operators_natal._OP_NAMES | {"日主上下文"}
    assert set(const["算子柱位"].values()) <= set(const["四柱序号"])
    assert const["关系字段类型"][const["柱刑关系字段"]] == "三刑组"
    assert load_contract()["柱字段后缀"] == "柱"
    assert set(const["关系取合类型"].values()) == {"两支", "全组"}
    assert const["旬空起点"] in const["天干"]
    assert const["五行旺因子后缀"] == "旺"
    assert set(const["流年年界"]) == {"八字", "紫微", "usage"}
    side_config = const["命理侧"]
    assert list(side_config["标签"]) == side_config["断言代码"]
    assert set(side_config["快照代码"]) <= set(side_config["断言代码"])
    assert side_config["公共代码"] in side_config["断言代码"]
    assert side_config["公共代码"] not in side_config["快照代码"]
    assert set(load_contract()["投影侧代码"].values()) == {
        side_config["标签"][side] for side in side_config["快照代码"]
    }


def test_operator_code_contains_no_domain_member_literals():
    const = duanyu.load_constants()
    operator_names = operators_natal._OP_NAMES | operators_liunian._LIU_OP_NAMES
    domain_members = set(const["十神"])
    domain_members.update(const["五行"])
    domain_members.update(const["天干"])
    domain_members.update(const["地支"])
    domain_members.update(const["关系取合类型"])
    domain_members.add(const["关系取冲类型"])
    domain_members.update(const["紫微主星"])
    domain_members.update(const["格局十神"])
    domain_members.update(const["紫微星曜特殊值"])
    domain_members.update(const["紫微宫位特殊条件"])
    domain_members.add(const["夫妻宫无关系状态"])
    domain_members.update(const["命理侧"]["标签"].values())
    domain_members.update(const["四柱序号"])
    domain_members.update(const["六亲角色"])
    for role in const["六亲角色"].values():
        if isinstance(role, dict):
            domain_members.update(role.values())
        else:
            domain_members.add(role)

    failures = []
    root = Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools"
    for path in sorted(root.glob("*.py")):
        tree = ast.parse(path.read_text(encoding="utf-8"))
        docstrings = set()
        for node in ast.walk(tree):
            if isinstance(node, (ast.Module, ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
                first = node.body[0] if node.body else None
                if isinstance(first, ast.Expr) and isinstance(first.value, ast.Constant):
                    docstrings.add(id(first.value))
        for node in ast.walk(tree):
            if not isinstance(node, ast.Constant) or not isinstance(node.value, str):
                continue
            if id(node) in docstrings or node.value in operator_names:
                continue
            if node.value in domain_members:
                failures.append(f"{path.name}:{node.lineno}:{node.value}")
    assert not failures
