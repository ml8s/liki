"""Current Pydantic tool-contract checks.

The Pydantic models are the only counsel contract source. These tests inspect
the schemas they publish and validate the same values consumed by MCP.
"""


import json
from pathlib import Path

import pytest
from app.contracts import COUNSEL_TOOLS, NATAL_TOOLS
from pydantic import ValidationError

REPO_ROOT = Path(__file__).resolve().parents[2]
CONTRACTS = {contract.name: contract for contract in COUNSEL_TOOLS}
NATAL_FUNCTIONS = {c.name: c.input_schema for c in NATAL_TOOLS}
OPAQUE_CONDITIONALS = ("if", "then", "else", "not")
LEGACY_INPUT_FIELDS = {
    "birth_time": "subject.standard_time",
    "birth_longitude": "subject.longitude",
    "qi_gua": "flat standard_time/longitude",
    "qi_gua_time": "standard_time",
    "qi_gua_longitude": "longitude",
    "qi_ju": "flat standard_time/longitude",
    "qi_ju_time": "standard_time",
    "qi_ju_longitude": "longitude",
    "location": "longitude",
    "place_longitude": "longitude",
    "start_time": "standard_time",
}


# ── published schema fidelity ────────────────────────────────────────────────

def test_qimen_ask_input_uses_method_branches():
    method = CONTRACTS["qimen_ask"].input_schema["properties"]["method"]
    forms = method["anyOf"]
    assert len(forms) == 10
    assert all(form["additionalProperties"] is False for form in forms)
    scopes: dict[str, int] = {}
    for form in forms:
        scope = form["properties"]["scope"]["const"]
        scopes[scope] = scopes.get(scope, 0) + 1
    assert scopes == {"hour": 3, "day": 3, "quarter": 2, "month": 1, "year": 1}


def test_ask_outputs_are_self_contained_readings():
    for name in ("liuyao_ask", "qimen_ask"):
        output = CONTRACTS[name].output_schema["properties"]
        assert {"scope", "evidence", "interpretation"} <= set(output)
        assert "reading_digest" not in output
        assert "case_digest" not in output and "snapshot_digest" not in output
        assert "factors" not in output


@pytest.mark.parametrize("name", sorted(CONTRACTS))
def test_published_input_schemas_are_mcp_objects(name):
    """MCP SDK serializes tool arguments only for object input schemas."""
    assert CONTRACTS[name].input_schema.get("type") == "object"


@pytest.mark.parametrize("name", sorted(CONTRACTS))
def test_published_input_schemas_have_no_legacy_fields(name):
    """Schema bootstrapping itself must reject reintroduced wrappers/aliases."""
    violations: list[str] = []

    def walk(path: str, value: object) -> None:
        if not isinstance(value, dict):
            return
        for key, child in value.items():
            if key in LEGACY_INPUT_FIELDS:
                canonical = LEGACY_INPUT_FIELDS[key]
                violations.append(f"{path}.{key}: use {canonical}")
            walk(f"{path}.{key}", child)

    walk("$", CONTRACTS[name].input_schema)
    assert violations == []


@pytest.mark.parametrize("name", sorted(CONTRACTS))
def test_published_schemas_do_not_ship_conditional_clauses(name):
    """Composition is exposed through anyOf branches, not opaque conditionals."""
    document = json.dumps(CONTRACTS[name].input_schema)
    assert not any(f'"{key}"' in document for key in OPAQUE_CONDITIONALS)


@pytest.mark.parametrize("name", sorted(CONTRACTS))
def test_published_schema_shared_refs_resolve_locally(name):
    for schema in (CONTRACTS[name].input_schema, CONTRACTS[name].output_schema):
        definitions = schema.get("$defs", {})
        refs = []

        def walk(node):
            if isinstance(node, dict):
                if "$ref" in node:
                    refs.append(node["$ref"])
                for value in node.values():
                    walk(value)
            elif isinstance(node, list):
                for value in node:
                    walk(value)

        walk(schema)
        for ref in refs:
            assert ref.startswith("#/$defs/")
            assert ref.rsplit("/", 1)[-1] in definitions


def test_unique_items_is_projected_and_enforced():
    focus_branches = CONTRACTS["qimen_ask"].input_schema["properties"]["focus"]["anyOf"]
    qimen_yongshen = next(
        branch["properties"]["yong_shen"]
        for branch in focus_branches
        if branch.get("properties", {}).get("yong_shen")
    )
    assert qimen_yongshen["uniqueItems"] is True
    with pytest.raises(ValidationError):
        CONTRACTS["qimen_ask"].input.model_validate(
            {
                "question": "该往哪里推进？",
                "longitude": 121.47,
                "focus": {"kind": "yong_shen", "yong_shen": ["生门", "生门"]},
            }
        )


def test_yongshen_contracts_publish_the_typed_core():
    from app.contracts.naming import ConsumableSelectedYongshen
    from app.contracts.yongshen import SelectedYongShen

    selected_schema = CONTRACTS["bazi_yongshen"].output_schema
    definitions = selected_schema.get("$defs", {})
    assert set(selected_schema["required"]) == set(SelectedYongShen.model_fields)
    choice = selected_schema["properties"]["selected_yongshen"]
    yong_schema = choice["properties"]["yong"]
    yong_definition = definitions[yong_schema["$ref"].rsplit("/", 1)[-1]]
    assert yong_definition["type"] == "string"
    assert set(yong_definition["enum"]) == {"木", "火", "土", "金", "水", "无候选"}
    assert set(choice["required"]) == {"yong", "xi", "ji"}


    naming_schema = CONTRACTS["naming_check"].input_schema
    assert set(naming_schema["properties"]) == {
        "given_names",
        "selected_yongshen",
        "forbidden_chars",
        "required_chars",
    }
    naming_selected_schema = naming_schema["properties"]["selected_yongshen"]
    assert set(naming_selected_schema["required"]) == set(ConsumableSelectedYongshen.model_fields)

    assert "factors" not in naming_selected_schema["properties"]
    assert "alternatives" not in naming_selected_schema["properties"]
    evidence_items = selected_schema["properties"]["evidence"]["items"]
    assert {branch["properties"]["source"]["const"] for branch in evidence_items["anyOf"]} == {
        "facts", "fu_yi", "tiao_hou", "ge_ju", "cong_ge",
    }
    assert "$defs" in selected_schema


def test_event_tools_are_not_marked_idempotent():
    expected = {contract.name: contract.annotations for contract in COUNSEL_TOOLS}
    assert expected["liuyao_ask"] == {
        "readOnlyHint": False,
        "idempotentHint": False,
    }
    assert expected["qimen_ask"] == {
        "readOnlyHint": True,
        "idempotentHint": True,
    }
    for name, annotations in expected.items():
        if name != "liuyao_ask":
            assert annotations == {
                "readOnlyHint": True,
                "idempotentHint": True,
            }


def test_yongshen_candidate_state_is_valid_and_empty_is_rejected():
    from app.contracts.yongshen import YongShenFuYiEvidence, YongShenTiaoHouEvidence

    fu_yi = {
        "source": "fu_yi", "qiang_ruo": "身强", "pattern": "",
        "ri_gan_gen": "month_main", "season": "旺", "yin_bi_shu_liang": 1,
        "yong": "无候选", "xi": "无候选", "ji": "无候选",
    }
    tiao_hou = {
        "source": "tiao_hou", "primary_wuxing": "火", "secondary_wuxing": "无辅调候",
        "secondary_condition": "", "condition_ok": False, "secondary_usable": False,
        "primary_present": False, "primary_visible": False,
        "secondary_present": False, "secondary_visible": False,
        "yong": "无候选",
        "yong_source": "none", "yong_present": False, "yong_visible": False,
        "control_wuxing": "无候选",
    }
    assert YongShenFuYiEvidence.model_validate(fu_yi).yong == "无候选"
    assert YongShenTiaoHouEvidence.model_validate(tiao_hou).secondary_wuxing == "无辅调候"

    with pytest.raises(ValidationError):
        YongShenFuYiEvidence.model_validate({**fu_yi, "yong": ""})
    with pytest.raises(ValidationError):
        YongShenTiaoHouEvidence.model_validate({**tiao_hou, "secondary_wuxing": ""})




def test_bazi_yongshen_input_uses_resolved_subject():
    subject = CONTRACTS["bazi_yongshen"].input_schema["properties"]["subject"]
    assert set(subject["required"]) == {"standard_time", "longitude", "gender"}
    assert set(subject["properties"]) == {"standard_time", "longitude", "gender"}
    assert subject["properties"]["gender"]["enum"] == ["male", "female"]
    assert subject["properties"]["longitude"]["minimum"] == -180
    assert subject["properties"]["longitude"]["maximum"] == 180


def _resolved_yongshen():
    from app.natal.tools.yongshen import resolve_yongshen

    input_path = (
        REPO_ROOT / "engine" / "internal" / "mcpcontract" / "testdata"
        / "wire_inputs" / "bazi_chart_1990.json"
    )
    subject = json.loads(input_path.read_text(encoding="utf-8"))["input"]["subject"]
    from app.natal.tools.charts import _bazi_chart

    return resolve_yongshen(_bazi_chart(subject))


def test_real_yongshen_output_conforms_to_the_shared_contract():
    from app.contracts.yongshen import SelectedYongShen
    from app.natal.tools.yongshen import published_yongshen

    result = _resolved_yongshen()
    published = published_yongshen(result)
    typed = SelectedYongShen.model_validate(published)
    assert typed.selected_yongshen.yong == published["selected_yongshen"]["yong"]
    assert typed.status == published["status"]
    assert "bazi_chart_digest" not in published
    assert "yongshen_digest" not in published
    assert "source" not in published
    assert "confidence" not in published
    assert "known_fields" not in published
    assert "factors" not in published
    assert "alternatives" not in published


@pytest.mark.parametrize(
    "mutation",
    [
        lambda result: result.update(status="maybe"),
        lambda result: result["selected_yongshen"].update(yong="风"),
    ],
)
def test_yongshen_contract_rejects_invalid_core_values(mutation):
    from app.contracts.yongshen import SelectedYongShen
    from app.natal.tools.yongshen import published_yongshen

    result = published_yongshen(_resolved_yongshen())
    mutation(result)
    with pytest.raises(ValidationError):
        SelectedYongShen.model_validate(result)


def test_registration_publishes_a_copy_of_the_contract_schema():
    """A registered tool must never hand clients (or itself) the cached schema."""
    from app.mcp_tool_adapter import add_contract_tool
    from mcp.server import MCPServer

    contract = CONTRACTS["qimen_ask"]
    server = MCPServer(name="probe", version="test")
    add_contract_tool(server, contract, lambda arguments: {})
    registered = server._tool_manager.get_tool("qimen_ask")  # noqa: SLF001 - SDK has no getter
    assert registered is not None
    assert registered.parameters == contract.input_schema
    assert registered.parameters is not contract.input_schema
    registered.parameters["type"] = "mutated"
    assert contract.input_schema["type"] == "object"


# ── natal tool closed-args & topic routing ──────────────────────────────────

def test_natal_tools_have_closed_args():
    required = {
        "bazi_yongshen": ["subject"],
        "natal_query": ["subject", "life_topics"],
        "natal_liunian_query": ["subject", "start_year", "year_count", "life_topics"],
    }
    assert set(NATAL_FUNCTIONS) == set(required)
    for name, schema in NATAL_FUNCTIONS.items():
        assert schema["additionalProperties"] is False, name
        assert schema["required"] == required[name], name
        assert schema["type"] == "object", name


def test_natal_topic_enum_is_generated_from_topic_routes():
    routes = json.loads(
        (REPO_ROOT / "counsel/app/natal/tools/topic_routes.json").read_text(encoding="utf-8")
    )
    expected = list(routes["life_topics"])
    for name in ("natal_query", "natal_liunian_query"):
        schema = NATAL_FUNCTIONS[name]["properties"]["life_topics"]
        assert schema["items"]["enum"] == expected, name


def test_natal_topic_routes_are_valid_and_unambiguous():
    from app.natal.tools.factor_constants import load_constants

    routes_doc = json.loads(
        (REPO_ROOT / "counsel/app/natal/tools/topic_routes.json").read_text(encoding="utf-8")
    )
    routes = routes_doc["life_topics"]
    domain_config = load_constants()["命理域"]
    natal = set(domain_config["本命规则"])
    annual = set(domain_config["流年规则"]) | set(domain_config["场景别名"])
    decades = {"大运", "大限"}

    domains_by_topic: dict[str, str] = {}
    for topic, route in routes.items():
        assert route["label"]
        assert route["domains"]
        assert all(rule in natal for rule in route["natal_rules"]), topic
        assert all(rule in annual for rule in route["annual_rules"]), topic
        assert all(rule in decades for rule in route["decade_rules"]), topic
        for domain in route["domains"]:
            if topic == "chart_structure":
                continue  # chart_structure 跨域（含旺衰/用神断语领域），不参与唯一性检查
            assert domain not in domains_by_topic, (domain, topic, domains_by_topic[domain])
            domains_by_topic[domain] = topic
