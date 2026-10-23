"""Current Pydantic tool-contract checks.

The Pydantic models are the only counsel contract source. These tests inspect
the schemas they publish and validate the same values consumed by MCP.
"""


import hashlib
import json
import subprocess
import sys
from pathlib import Path

import pytest
from app.contracts import COUNSEL_TOOLS, TOOL_GROUPS, tool_vocabulary
from pydantic import ValidationError

REPO_ROOT = Path(__file__).resolve().parents[2]
CONTRACTS = {contract.name: contract for contract in COUNSEL_TOOLS}
OPAQUE_CONDITIONALS = ("if", "then", "else", "not")


# ── published schema fidelity ────────────────────────────────────────────────

def test_qimen_snapshot_input_uses_method_branches():
    method = CONTRACTS["qimen_snapshot"].input_schema["properties"]["method"]
    forms = method["anyOf"]
    assert len(forms) == 10
    assert all(form["additionalProperties"] is False for form in forms)
    scopes: dict[str, int] = {}
    for form in forms:
        scope = form["properties"]["scope"]["const"]
        scopes[scope] = scopes.get(scope, 0) + 1
    assert scopes == {"hour": 3, "day": 3, "quarter": 2, "month": 1, "year": 1}


def test_query_inputs_consume_complete_snapshot_outputs():
    pairs = {"liuyao_query": "liuyao_snapshot", "qimen_query": "qimen_snapshot"}
    for query, snapshot in pairs.items():
        consumed = CONTRACTS[query].input_schema["properties"]["snapshot"]
        published = CONTRACTS[snapshot].output_schema
        assert consumed["required"] == published["required"]
        assert consumed["properties"] == published["properties"]


@pytest.mark.parametrize("name", sorted(CONTRACTS))
def test_published_input_schemas_are_mcp_objects(name):
    """MCP SDK serializes tool arguments only for object input schemas."""
    assert CONTRACTS[name].input_schema.get("type") == "object"


@pytest.mark.parametrize("name", sorted(CONTRACTS))
def test_published_schemas_do_not_ship_conditional_clauses(name):
    """Composition is exposed through anyOf branches, not opaque conditionals."""
    document = json.dumps(CONTRACTS[name].input_schema)
    assert not any(f'"{key}"' in document for key in OPAQUE_CONDITIONALS)


@pytest.mark.parametrize("name", sorted(CONTRACTS))
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
    yong_schema = choice["properties"]["yong"]
    assert yong_schema["type"] == ["string", "null"]
    assert yong_schema["enum"] == ["木", "火", "土", "金", "水"]
    assert set(choice["required"]) == {"yong", "xi", "ji", "source", "confidence", "known_fields"}


    compute_schema = CONTRACTS["compute_factors"].output_schema
    assert "selected_yongshen" not in compute_schema["properties"]
    assert "selected_yongshen" in compute_schema["properties"]["factors"]["description"]

    naming_schema = CONTRACTS["naming_check"].input_schema
    naming_selected = _yongshen_object_schema(
        # input 是 anyOf 两分支（显式五行 / selected_yongshen），
        # selected_yongshen 属性住在第二分支里。
        next(
            branch for branch in naming_schema["anyOf"]
            if "selected_yongshen" in branch.get("properties", {})
        )["properties"]["selected_yongshen"]
    )
    assert set(naming_selected["required"]) == set(ConsumedYongShen.model_fields)

    source_schema = _yongshen_object_schema(
        CONTRACTS["naming_check"].output_schema["items"]["properties"]["yongshen_source"]
    )
    assert set(source_schema["required"]) == {
        "policy_id", "rule_id", "status",
        "chart_digest", "yongshen_digest",
    }

    assert "factors" not in selected_schema["properties"]
    assert "alternatives" not in selected_schema["properties"]
    evidence_items = selected_schema["properties"]["evidence"]["items"]
    assert {branch["properties"]["source"]["const"] for branch in evidence_items["anyOf"]} == {
        "facts", "fu_yi", "tiao_hou", "ge_ju", "cong_ge",
    }
    assert "#/$defs" not in json.dumps(selected_schema)


def test_yongshen_candidate_none_is_valid_and_empty_is_rejected():
    from app.contracts.yongshen import YongShenFuYiEvidence, YongShenTiaoHouEvidence

    fu_yi = {
        "source": "fu_yi", "qiangruo": "身强", "pattern": "",
        "root_type": "month_main", "season": "旺", "yin_bi_count": 1,
        "yong": None, "xi": None, "ji": None,
    }
    tiao_hou = {
        "source": "tiao_hou", "primary_wuxing": "火", "secondary_wuxing": None,
        "secondary_condition": "", "condition_ok": False, "secondary_usable": False,
        "primary_present": False, "primary_visible": False,
        "secondary_present": False, "secondary_visible": False,
        "season": "春", "detail": "test", "model": "test", "yong": None,
        "yong_source": "none", "yong_present": False, "yong_visible": False,
        "control_wuxing": None,
    }
    assert YongShenFuYiEvidence.model_validate(fu_yi).yong is None
    assert YongShenTiaoHouEvidence.model_validate(tiao_hou).secondary_wuxing is None

    with pytest.raises(ValidationError):
        YongShenFuYiEvidence.model_validate({**fu_yi, "yong": ""})
    with pytest.raises(ValidationError):
        YongShenTiaoHouEvidence.model_validate({**tiao_hou, "secondary_wuxing": ""})


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
