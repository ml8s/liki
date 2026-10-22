"""Fidelity between the Pydantic contracts and the frozen legacy manifests.

Before the migration every counsel tool contract was a hand-written JSON
manifest. Those manifests are frozen under ``tests/testdata`` as the migration
golden; these tests prove that the Pydantic contracts in ``app.contracts``
reproduce them, and that the clauses JSON Schema can express but Pydantic cannot
project (``if``/``then``) are still enforced at runtime.
"""

from __future__ import annotations

import hashlib
import json
import subprocess
import sys
from pathlib import Path

import pytest
from app.contracts import COUNSEL_TOOLS, TOOL_GROUPS, tool_vocabulary
from jsonschema import Draft202012Validator
from pydantic import ValidationError

REPO_ROOT = Path(__file__).resolve().parents[2]
LEGACY_DIR = Path(__file__).resolve().parent / "testdata" / "legacy-tool-manifests"
LEGACY_MANIFESTS = (
    LEGACY_DIR / "natal-counsel-tools.json",
    LEGACY_DIR / "naming-skill-tools.json",
    LEGACY_DIR / "divination-skill-tools.json",
)

CONTRACTS = {contract.name: contract for contract in COUNSEL_TOOLS}
LEGACY = {
    item["function"]["name"]: item["function"]
    for path in LEGACY_MANIFESTS
    for item in json.loads(path.read_text(encoding="utf-8"))["tools"]
}
CONDITIONALS = ("allOf", "oneOf", "not", "if", "then")

#: These fields were intentionally upgraded from free-form objects during the
#: yongshen migration. The frozen manifests remain the golden for every other
#: part of the contract; the typed shape is asserted separately below.
MIGRATED_TYPED_FIELDS = {
    # 时间基成对契约：gregorian_time 为年/月柱判定所需的 Gregorian time；
    # 时间字段契约演进（solar_time → gregorian_time + longitude）随工具输入同步。
    ("liuyao_snapshot", "input"): (
        ("properties", "instant"),
        ("properties", "timezone"),
        ("properties", "solar_time"),
        ("properties", "longitude"),
    ),
    # 快照不再生成 solar_time（历史字段，保留可空以兼容旧快照）。
    ("liuyao_snapshot", "output"): (
        ("properties", "question", "properties", "solar_time"),
        ("properties", "question", "required"),
    ),
    ("qimen_snapshot", "output"): (
        ("properties", "input", "properties", "solar_time"),
        ("properties", "input", "properties", "local_time"),
        ("properties", "input", "properties", "gregorian_time"),
        ("properties", "input", "properties", "instant"),
        ("properties", "input", "properties", "timezone"),
        ("properties", "input", "required"),
    ),
    # 起局地点二选一：city/longitude 改为可空类型，运行时强校验互斥。
    ("qimen_snapshot", "input"): (
        ("properties", "city"),
        ("properties", "longitude"),
        ("properties", "instant"),
        ("properties", "timezone"),
        ("properties", "time"),
    ),
    ("period_query", "input"): (("properties", "timezone"),),
    ("compute_factors", "output"): (("properties", "selected_yongshen"),),
    ("naming_check", "input"): (("properties", "selected_yongshen"),),
    ("naming_check", "output"): (("items", "properties", "yongshen_source"),),
}


# ── JSON Schema normalisation ────────────────────────────────────────────────
# The legacy manifests and the Pydantic generator spell the same JSON Schema
# differently: nullable types use ``type: [T, "null"]`` vs ``anyOf``, absent
# ``additionalProperties`` means the same as ``true``, an empty ``items`` schema
# means "anything", annotation-only ``title``/``default`` keys carry no
# constraint, and an object with ``properties`` is always of type object.

def _strip_titles(node):
    if isinstance(node, dict):
        node.pop("title", None)
        for value in node.values():
            _strip_titles(value)
    elif isinstance(node, list):
        for value in node:
            _strip_titles(value)
    return node


def _strip_conditionals(node):
    """Drop the clauses the legacy manifests expressed as ``allOf``."""
    if isinstance(node, dict):
        for key in CONDITIONALS:
            node.pop(key, None)
        for value in node.values():
            _strip_conditionals(value)
    elif isinstance(node, list):
        for value in node:
            _strip_conditionals(value)
    return node


def _normalise(node):
    if isinstance(node, dict):
        types = node.get("type")
        if isinstance(types, list) and "null" in types and len(types) == 2:
            other = [item for item in types if item != "null"][0]
            outer = {key: node[key] for key in ("description", "title") if key in node}
            siblings = {
                key: value
                for key, value in node.items()
                if key not in ("type", "anyOf", "description", "title", "default")
            }
            node.clear()
            node.update(outer)
            node["anyOf"] = [{**siblings, "type": other}, {"type": "null"}]
        if node.get("default") is None:
            node.pop("default", None)
        if node.get("properties") and node.get("type") in (None, "object"):
            node["type"] = "object"
        if "const" in node:
            node.pop("type", None)
        if "enum" in node and node.get("type") == "string":
            node.pop("type", None)
        if node.get("type") == "object" and "additionalProperties" not in node:
            node["additionalProperties"] = True
        if node.get("items") == {}:
            node.pop("items")
        if isinstance(node.get("required"), list):
            node["required"] = sorted(node["required"])
        properties = node.get("properties")
        if isinstance(properties, dict):
            for key in [key for key, value in properties.items() if value == {}]:
                properties.pop(key)
            if not properties:
                node.pop("properties")
        anyof = node.get("anyOf")
        if (
            isinstance(anyof, list)
            and len(anyof) == 2
            and anyof[1] == {"type": "null"}
            and isinstance(anyof[0], dict)
        ):
            arm = anyof[0]
            moved = arm.pop("description", None)
            if moved is not None and "description" not in node:
                node["description"] = moved
        for value in node.values():
            _normalise(value)
        return node
    if isinstance(node, list):
        return [_normalise(value) for value in node]
    return node


def _diff(got, want, path=""):
    out = []
    if isinstance(got, dict) and isinstance(want, dict):
        for key in sorted(set(got) | set(want)):
            if key not in got:
                out.append(f"{path}/{key}: missing from the contract")
            elif key not in want:
                out.append(f"{path}/{key}: extra in the contract")
            else:
                out.extend(_diff(got[key], want[key], f"{path}/{key}"))
    elif isinstance(got, list) and isinstance(want, list):
        if len(got) != len(want):
            out.append(f"{path}: length {len(got)} != {len(want)}")
        else:
            for index, (left, right) in enumerate(zip(got, want)):
                out.extend(_diff(left, right, f"{path}[{index}]"))
    elif got != want:
        out.append(f"{path}: {got!r} != {want!r}")
    return out


def _project(document):
    return _normalise(json.loads(json.dumps(document)))


def _drop_path(document, path):
    node = document
    owner = None
    for part in path[:-1]:
        if part == "properties":
            owner = node
        node = node.get(part, {})
    if isinstance(node, dict):
        key = path[-1]
        node.pop(key, None)
        if isinstance(owner, dict):
            required = owner.get("required")
            if isinstance(required, list) and key in required:
                required.remove(key)


def _drop_migrated_typed_fields(name, direction, document):
    for path in MIGRATED_TYPED_FIELDS.get((name, direction), ()):
        _drop_path(document, path)
    return document


def _schema_deltas(contract, legacy_function):
    legacy_input = _strip_conditionals(
        json.loads(json.dumps(legacy_function["parameters"]))
    )
    legacy_output = _strip_conditionals(
        json.loads(json.dumps(legacy_function.get("output_schema") or {}))
    )
    got_input = _drop_migrated_typed_fields(
        contract.name, "input", _project(contract.input_schema)
    )
    got_output = _drop_migrated_typed_fields(
        contract.name, "output", _project(contract.output_schema or {})
    )
    want_input = _drop_migrated_typed_fields(
        contract.name, "input", _project(legacy_input)
    )
    want_output = _drop_migrated_typed_fields(
        contract.name, "output", _project(legacy_output)
    )
    deltas = _diff(got_input, want_input, "input")
    deltas += _diff(got_output, want_output, "output")
    return deltas


# ── published schema fidelity ────────────────────────────────────────────────

def test_contracts_cover_every_legacy_tool():
    assert set(LEGACY) <= set(CONTRACTS)
    assert set(CONTRACTS) - set(LEGACY) == {"bazi_yongshen"}
    assert len(CONTRACTS) == 13


@pytest.mark.parametrize("name", sorted(LEGACY))
def test_published_schemas_match_frozen_legacy_manifests(name):
    assert _schema_deltas(CONTRACTS[name], LEGACY[name]) == []


@pytest.mark.parametrize("name", sorted(LEGACY))
def test_published_schemas_do_not_ship_conditional_clauses(name):
    """Pydantic cannot project ``if``/``then``; the models enforce them instead."""
    document = json.dumps(CONTRACTS[name].input_schema)
    assert not any(f'"{key}"' in document for key in CONDITIONALS[1:])


@pytest.mark.parametrize("name", sorted(LEGACY))
def test_published_schemas_are_flat_self_contained_documents(name):
    for schema in (CONTRACTS[name].input_schema, CONTRACTS[name].output_schema):
        document = json.dumps(schema)
        assert "$defs" not in document
        assert "$ref" not in document


def test_unique_items_is_projected_and_enforced():
    qimen = CONTRACTS["qimen_snapshot"].input_schema["properties"]["yong_shen"]
    assert qimen["uniqueItems"] is True
    with pytest.raises(ValidationError):
        CONTRACTS["qimen_snapshot"].input.model_validate(
            {"question": "该往哪里推进？", "longitude": 121.47, "yong_shen": ["生门", "生门"]}
        )


def _yongshen_object_schema(property_schema):
    return next(
        arm for arm in property_schema["anyOf"]
        if arm.get("type") == "object"
    )


def test_yongshen_contracts_publish_the_typed_core():
    from app.contracts.yongshen import ConsumedYongShen, SelectedYongShen

    selected_schema = CONTRACTS["bazi_yongshen"].output_schema
    assert set(selected_schema["required"]) == set(SelectedYongShen.model_fields)
    choice = selected_schema["properties"]["selected"]
    assert choice["properties"]["yong"]["enum"] == ["", "木", "火", "土", "金", "水"]
    assert set(choice["required"]) == {"yong", "xi", "ji", "source", "confidence", "known_fields"}

    compute_schema = CONTRACTS["compute_factors"].output_schema
    assert "selected_yongshen" not in compute_schema["properties"]
    assert "selected_yongshen" in compute_schema["properties"]["factors"]["description"]

    naming_schema = CONTRACTS["naming_check"].input_schema
    naming_selected = _yongshen_object_schema(
        naming_schema["properties"]["selected_yongshen"]
    )
    assert set(naming_selected["required"]) == set(ConsumedYongShen.model_fields)

    source_schema = _yongshen_object_schema(
        CONTRACTS["naming_check"].output_schema["items"]["properties"]["yongshen_source"]
    )
    assert set(source_schema["required"]) == {
        "policy_id", "rule_version", "rule_id", "status",
        "chart_digest", "yongshen_digest",
    }

    assert "factors" not in selected_schema["properties"]
    assert "alternatives" not in selected_schema["properties"]
    evidence_items = selected_schema["properties"]["evidence"]["items"]
    assert {branch["properties"]["source"]["const"] for branch in evidence_items["anyOf"]} == {
        "facts", "fu_yi", "tiao_hou", "ge_ju", "cong_ge",
    }
    assert "#/$defs" not in json.dumps(selected_schema)


def test_bazi_yongshen_chart_input_is_fully_typed():
    chart = CONTRACTS["bazi_yongshen"].input_schema["properties"]["chart"]
    assert set(chart["required"]) == {
        "nian", "yue", "ri", "shi", "gender", "birth_year",
    }
    assert set(chart["properties"]) == {
        "nian", "yue", "ri", "shi", "da_yun", "gender", "birth_year", "zi_shi_rule",
    }
    assert chart["properties"]["gender"]["enum"] == ["male", "female"]
    assert chart["properties"]["birth_year"]["minimum"] == 1
    for pillar in ("nian", "yue", "ri", "shi"):
        assert set(chart["properties"][pillar]["required"]) == {"gan", "zhi", "na_yin"}
        assert len(chart["properties"][pillar]["properties"]["gan"]["enum"]) == 10
        assert len(chart["properties"][pillar]["properties"]["zhi"]["enum"]) == 12


def _resolved_yongshen():
    from app.natal.tools.yongshen import resolve_yongshen

    path = (
        REPO_ROOT / "engine" / "internal" / "mcpcontract" / "testdata"
        / "actual" / "bazi_chart_1990_fullchart.json"
    )
    return resolve_yongshen(json.loads(path.read_text(encoding="utf-8"))["output"])


def test_real_yongshen_output_conforms_to_the_shared_contract():
    from app.contracts.yongshen import SelectedYongShen
    from app.natal.tools.yongshen import published_yongshen

    result = _resolved_yongshen()
    published = published_yongshen(result)
    typed = SelectedYongShen.model_validate(published)
    assert typed.selected.yong == published["selected"]["yong"]
    assert typed.status == published["status"]
    assert typed.yongshen_digest == published["yongshen_digest"]
    assert "source" not in published
    assert "confidence" not in published
    assert "known_fields" not in published
    assert "factors" not in published
    assert "alternatives" not in published
    digest_input = {
        key: value for key, value in published.items() if key != "yongshen_digest"
    }
    digest = hashlib.sha256(
        json.dumps(
            digest_input, ensure_ascii=False, sort_keys=True, separators=(",", ":")
        ).encode("utf-8")
    ).hexdigest()
    assert digest == published["yongshen_digest"]


@pytest.mark.parametrize(
    "mutation",
    [
        lambda result: result.update(status="maybe"),
        lambda result: result.pop("yongshen_digest"),
        lambda result: result["selected"].update(yong="风"),
    ],
)
def test_yongshen_contract_rejects_invalid_core_values(mutation):
    from app.contracts.yongshen import SelectedYongShen

    result = _resolved_yongshen()
    mutation(result)
    with pytest.raises(ValidationError):
        SelectedYongShen.model_validate(result)


def test_fidelity_comparator_detects_drift():
    """Guard the guard: the comparison must fail when a contract changes."""
    contract = CONTRACTS["compute_factors"]
    legacy = LEGACY["compute_factors"]
    mutated = _project(contract.input_schema)
    mutated["properties"].pop("chart")
    deltas = _diff(mutated, _project(legacy["parameters"]), "input")
    assert deltas


# ── runtime validation equivalence ───────────────────────────────────────────

QIMEN = {"question": "该往哪里推进？"}
LIUYAO = {"question": "这次面试能不能通过？"}

#: The legacy qimen manifest wrapped two ``then`` clauses without the required
#: ``properties`` keyword, so JSON Schema treated the whole clause as an
#: annotation and never enforced it. The models enforce the documented intent
#: (and the engine catalog rejects the same combinations), so they are strictly
#: stricter than the frozen manifest here.
LEGACY_MANIFEST_NOOP_CONDITIONALS = [
    (  # allOf[7]: quarter_rule=ten_minute_sanyuan forbids the advanced fields
        "qimen_snapshot",
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "quarter",
            "school": "zhuanpan",
            "quarter_rule": "ten_minute_sanyuan",
            "base_dingju_method": "chaibu",
        },
    ),
    (  # allOf[9]: scope=month/year forbids the hour-scope method fields
        "qimen_snapshot",
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "month",
            "matter": "wealth",
            "dingju_method": "chaibu",
        },
    ),
]

def _period(time_scope):
    return {
        "factors": {},
        "factors_digest": "a" * 64,
        "time_scope": time_scope,
        "topics": ["career"],
        "chart": {},
    }


CORPUS = {
    "liuyao_snapshot": [
        {**LIUYAO, "matter": "career"},
        {**LIUYAO, "yong_shen": "官鬼"},
        {**LIUYAO, "matter": "career", "yong_shen": "官鬼"},
        {**LIUYAO},
        {**LIUYAO, "matter": "career", "mode": "coins"},
        {**LIUYAO, "matter": "career", "mode": "yaos"},
        {
            **LIUYAO,
            "matter": "career",
            "mode": "coins",
            "rounds": [["正", "反", "反"]] * 6,
        },
        {**LIUYAO, "matter": "career", "mode": "yaos", "yaos": [7] * 6},
        {**LIUYAO, "matter": "career", "rounds": [["正", "反", "反"]] * 6},
        {**LIUYAO, "matter": "career", "mode": "auto", "yaos": [7] * 6},
        {**LIUYAO, "matter": "career", "perspective": "male"},
        {**LIUYAO, "matter": "relationship", "perspective": "male"},
        {**LIUYAO, "matter": "relationship"},
        {**LIUYAO, "yong_shen": "官鬼", "perspective": "male"},
        {**LIUYAO, "matter": "career", "unexpected": True},
        {**LIUYAO, "matter": "career", "mode": "coins", "rounds": [["正", "反"]] * 6},
    ],
    "qimen_snapshot": [
        {**QIMEN, "longitude": 121.47, "matter": "wealth"},
        {**QIMEN, "city": "上海"},
        {**QIMEN, "longitude": 121.47},
        {**QIMEN},
        {**QIMEN, "city": "上海", "longitude": 121.47},
        {**QIMEN, "longitude": 121.47, "yong_shen": ["生门"]},
        {**QIMEN, "longitude": 121.47, "matter": "wealth", "yong_shen": ["生门"]},
        {**QIMEN, "longitude": 121.47, "quarter_rule": "ten_minute_sanyuan"},
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "quarter",
            "school": "zhuanpan",
            "quarter_rule": "ten_minute_sanyuan",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "quarter",
            "school": "zhuanpan",
            "quarter_rule": "ten_minute_sanyuan",
            "dingju_method": "chaibu",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "quarter",
            "school": "zhuanpan",
            "quarter_rule": "twelve_minute_ten_division",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "quarter",
            "school": "zhuanpan",
            "quarter_rule": "twelve_minute_ten_division",
            "base_dingju_method": "zhirun",
            "dun_source": "solar_term",
            "hour_boundary": "zi_zheng",
        },
        {**QIMEN, "longitude": 121.47, "scope": "quarter"},
        {**QIMEN, "longitude": 121.47, "scope": "year", "matter": "wealth"},
        {**QIMEN, "longitude": 121.47, "scope": "month", "matter": "wealth"},
        {**QIMEN, "longitude": 121.47, "dingju_method": "maoshan"},
        {**QIMEN, "longitude": 121.47, "scope": "hour", "dingju_method": "maoshan", "matter": "wealth"},
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "hour",
            "school": "zhuanpan",
            "dingju_method": "maoshan",
            "matter": "wealth",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "school": "mingfa_feipan",
            "scope": "hour",
            "dingju_method": "chaibu",
            "matter": "wealth",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "school": "mingfa_feipan",
            "scope": "day",
            "dingju_method": "chaibu",
            "matter": "wealth",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "school": "mingfa_feipan",
            "scope": "hour",
            "matter": "wealth",
        },
        {**QIMEN, "longitude": 121.47, "school": "jinhan_yujing", "scope": "day"},
        {
            **QIMEN,
            "longitude": 121.47,
            "school": "jinhan_yujing",
            "scope": "day",
            "matter": "wealth",
        },
        {**QIMEN, "longitude": 121.47, "school": "jinhan_yujing", "scope": "hour"},
        {**QIMEN, "longitude": 121.47, "scope": "day"},
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "hour",
            "school": "zhuanpan",
            "rule": "lost_property",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "hour",
            "school": "zhuanpan",
            "rule": "lost_property",
            "matter": "wealth",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "hour",
            "school": "zhuanpan",
            "rule": "missing_person",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "hour",
            "school": "zhuanpan",
            "rule": "missing_person",
            "matter": "missing_person",
        },
        {
            **QIMEN,
            "longitude": 121.47,
            "scope": "hour",
            "school": "zhuanpan",
            "rule": "missing_person",
            "matter": "career",
        },
    ],
    "period_query": [
        _period(scope)
        for scope in (
            {"type": "current_year"},
            {"type": "year"},
            {"type": "year", "year": 2024},
            {"type": "year_range", "start_year": 2020, "end_year": 2024},
            {"type": "year_range", "start_year": 2020},
            {"type": "decade"},
            {"type": "decade", "year": 2024},
            {"type": "current_decade"},
        )
    ],
}


def _schema_accepts(name, instance):
    validator = Draft202012Validator(LEGACY[name]["parameters"])
    return not list(validator.iter_errors(instance))


def _model_accepts(name, instance):
    try:
        CONTRACTS[name].input.model_validate(instance)
    except ValidationError:
        return False
    return True


@pytest.mark.parametrize(
    ("name", "instance"),
    [(name, instance) for name, cases in CORPUS.items() for instance in cases],
    ids=lambda value: value if isinstance(value, str) else "",
)
def test_models_never_accept_more_than_the_frozen_manifest(name, instance):
    """Every call the model accepts was accepted by the legacy manifest too."""
    if _model_accepts(name, instance):
        assert _schema_accepts(name, instance)


@pytest.mark.parametrize(
    ("name", "instance"),
    LEGACY_MANIFEST_NOOP_CONDITIONALS,
)
def test_models_enforce_conditionals_the_manifest_failed_to_express(name, instance):
    assert _schema_accepts(name, instance)
    assert not _model_accepts(name, instance)


def test_registration_publishes_a_copy_of_the_contract_schema():
    """A registered tool must never hand clients (or itself) the cached schema."""
    from app.mcp_tool_adapter import add_contract_tool
    from mcp.server import MCPServer

    contract = CONTRACTS["qimen_snapshot"]
    server = MCPServer(name="probe", version="test")
    add_contract_tool(server, contract, lambda arguments: {})
    registered = server._tool_manager.get_tool("qimen_snapshot")  # noqa: SLF001 - SDK has no getter
    assert registered is not None
    assert registered.parameters == contract.input_schema
    assert registered.parameters is not contract.input_schema
    registered.parameters["type"] = "mutated"
    assert contract.input_schema["type"] == "object"


def test_tool_vocabulary_publishes_every_group_and_tool():
    vocabulary = tool_vocabulary()
    assert set(vocabulary) == set(TOOL_GROUPS)
    for group, tools in TOOL_GROUPS.items():
        names = vocabulary[group]
        assert names == sorted(set(names)), group
        for tool in tools:
            assert tool.name in names, f"{group}: {tool.name}"


def test_tool_vocabulary_reaches_names_nested_below_the_root_properties():
    """The walk descends into property children, where nested enums own names."""
    vocabulary = tool_vocabulary()
    # ``self``/``other`` are values of the ``matter`` enum, one level below the
    # root ``properties`` object of liuyao_snapshot.
    assert {"self", "other"} <= set(vocabulary["divination"])
    assert {"qimen_snapshot", "liuyao_snapshot"} <= set(vocabulary["divination"])
    assert "compute_factors" in vocabulary["natal"]
    assert "naming_surname" in vocabulary["naming"]


def test_tool_vocabulary_bridge_publishes_exactly_what_the_models_expose():
    """The Go suite loads the vocabulary through this script, not a copy."""
    bridge = REPO_ROOT / "scripts" / "counsel_tool_vocabulary.py"
    result = subprocess.run(
        [sys.executable, str(bridge)],
        check=True,
        capture_output=True,
        text=True,
    )
    assert json.loads(result.stdout) == tool_vocabulary()
