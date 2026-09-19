"""engine 输出字段与 Python snapshot 投影契约的同步防漂移。"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def _engine_method_schema(go_file: str, method: str) -> dict:
    text = (ROOT / "engine/internal/agent" / go_file).read_text(encoding="utf-8")
    marker = f'Name: "{method}"'
    start = text.index(marker)
    result_at = text.index("Result:", start)
    backtick = text.index("`", result_at)
    end = text.index("`),", backtick)
    return json.loads(text[backtick + 1:end])


def _fullchart_schema() -> dict:
    text = (ROOT / "engine/internal/agent/tools_bazi.go").read_text(encoding="utf-8")
    marker = 'Name: "bazi.fullchart"'
    start = text.index(marker)
    result_at = text.index("Result:", start)
    backtick = text.index("`", result_at)
    end = text.index("`),", backtick)
    return json.loads(text[backtick + 1:end])


def test_projection_contract_covers_engine_fullchart_fields():
    contract_path = ROOT / "skills/liki/bazi/tools/natal_projection_contract.json"
    contract = json.loads(contract_path.read_text(encoding="utf-8"))
    schema = _fullchart_schema()
    top = set(schema["properties"])

    pillar_fields = set(contract["柱值字段"]) | set(contract["柱存在字段"])
    mapped_full = set(contract["八字字段映射"]["full"])
    mapped_chart = set(contract["八字字段映射"]["chart"])
    # gender is projected into context; birth_year is available through chart mapping.
    # Other new top-level engine facts must be added to projection mapping deliberately.
    allowed_unprojected = {"gender"}
    pillars = {"nian", "yue", "ri", "shi"}

    missing = top - pillars - mapped_full - mapped_chart - allowed_unprojected
    assert not missing, f"engine fullchart 字段未进入 projection contract: {sorted(missing)}"

    for pillar in pillars:
        engine_fields = set(schema["properties"][pillar].get("properties", {}))
        missing = engine_fields - pillar_fields
        assert not missing, f"{pillar} 字段未进入 projection contract: {sorted(missing)}"


def test_projection_contract_has_no_stale_or_duplicate_labels():
    contract_path = ROOT / "skills/liki/bazi/tools/natal_projection_contract.json"
    contract = json.loads(contract_path.read_text(encoding="utf-8"))
    bazi_labels = contract["八字"]
    ziwei_labels = contract["紫微"]
    assert len(bazi_labels) == len(set(bazi_labels))
    assert len(ziwei_labels) == len(set(ziwei_labels))
    assert "空宫" not in ziwei_labels  # 旧标签已改为“空宫借星”
    assert "fu_yi" in bazi_labels and "tiao_hou" in bazi_labels and "ge_ju" in bazi_labels
    assert "稳定关系组" in bazi_labels
    assert "命理原子事实" in bazi_labels

def _engine_method_schema(go_file: str, method: str) -> dict:
    text = (ROOT / "engine/internal/agent" / go_file).read_text(encoding="utf-8")
    marker = f'Name: "{method}"'
    start = text.index(marker)
    result_at = text.index("Result:", start)
    backtick = text.index("`", result_at)
    end = text.index("`),", backtick)
    return json.loads(text[backtick + 1:end])


def test_projection_contract_paths_exist_in_engine_schema():
    contract = json.loads(
        (ROOT / "skills/liki/bazi/tools/natal_projection_contract.json").read_text(encoding="utf-8")
    )
    full_schema = _fullchart_schema()["properties"]
    ziwei_schema = _engine_method_schema("tools_ziwei.go", "ziwei.chart")["properties"]
    chart_schema = _engine_method_schema("tools_bazi.go", "bazi.chart")["properties"]

    missing = []
    for pillar in contract["四柱"]:
        pillar_props = full_schema[pillar]["properties"]
        for source, label in contract["柱值字段"].items():
            if source not in pillar_props:
                missing.append(f"full.{pillar}.{source} ({label})")
        for source, label in contract["柱存在字段"].items():
            if source not in pillar_props:
                missing.append(f"full.{pillar}.{source} ({label})")

    for source, label in contract["八字字段映射"]["full"].items():
        if source not in full_schema:
            missing.append(f"full.{source} ({label})")
    for source, label in contract["八字字段映射"]["chart"].items():
        if source not in chart_schema:
            missing.append(f"chart.{source} ({label})")
    for source, label in contract["紫微字段映射"]["ziwei"].items():
        if source not in ziwei_schema:
            missing.append(f"ziwei.{source} ({label})")
    for source, label in contract["紫微字段映射"]["pan"].items():
        # full_paipan surface is assembled by Python; contract keys are checked in tool contracts.
        continue

    assert not missing, f"projection contract 引用的 engine 字段缺失: {missing}"


def test_projection_labels_are_unique_and_generated_from_mappings():
    contract = json.loads(
        (ROOT / "skills/liki/bazi/tools/natal_projection_contract.json").read_text(encoding="utf-8")
    )
    bazi_labels = []
    for pillar in contract["四柱"]:
        bazi_labels += [f"{pillar}柱{label}" for label in contract["柱值字段"].values()]
        bazi_labels += [f"{pillar}柱{label}" for label in contract["柱存在字段"].values()]
    bazi_labels += list(contract["八字字段映射"]["full"].values())
    bazi_labels += list(contract["八字字段映射"]["chart"].values())
    ziwei_labels = (
        list(contract["紫微字段映射"]["ziwei"].values())
        + list(contract["紫微字段映射"]["pan"].values())
    )
    assert bazi_labels == contract["八字"]
    assert ziwei_labels == contract["紫微"]
    assert len(bazi_labels) == len(set(bazi_labels))
    assert len(ziwei_labels) == len(set(ziwei_labels))
