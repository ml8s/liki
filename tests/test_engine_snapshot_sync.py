"""engine 输出字段与 Python snapshot 投影契约的同步防漂移。"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


# The engine-schema half of this contract moved to Go: the typed registry owns
# the DTOs now, so engine/internal/mcpcontract/projection_contract_test.go
# checks the fullchart/chart/ziwei field paths against the real schemas.


def test_projection_contract_has_no_stale_or_duplicate_labels():
    contract_path = ROOT / "counsel/app/natal/tools/natal_projection_contract.json"
    contract = json.loads(contract_path.read_text(encoding="utf-8"))
    bazi_labels = contract["八字"]
    ziwei_labels = contract["紫微"]
    assert len(bazi_labels) == len(set(bazi_labels))
    assert len(ziwei_labels) == len(set(ziwei_labels))
    assert "空宫" not in ziwei_labels  # 旧标签已改为“空宫借星”
    assert "fu_yi" in bazi_labels and "tiao_hou" in bazi_labels and "ge_ju" in bazi_labels
    assert "稳定关系组" in bazi_labels
    assert "命理原子事实" in bazi_labels


def test_projection_labels_are_unique_and_generated_from_mappings():
    contract = json.loads(
        (ROOT / "counsel/app/natal/tools/natal_projection_contract.json").read_text(encoding="utf-8")
    )
    bazi_labels = []
    for pillar in contract["四柱"]:
        bazi_labels += [f"{pillar}柱{label}" for label in contract["柱值字段"].values()]
        bazi_labels += [f"{pillar}柱{label}" for label in contract["柱存在字段"].values()]
    bazi_labels += list(contract["八字字段映射"]["full"].values())
    bazi_labels += list(contract["八字字段映射"]["chart"].values())
    ziwei_labels = (
        list(contract["紫微字段映射"]["ziwei"].values())
        + list(contract["紫微字段映射"]["natal_chart"].values())
    )
    assert bazi_labels == contract["八字"]
    assert ziwei_labels == contract["紫微"]
    assert len(bazi_labels) == len(set(bazi_labels))
    assert len(ziwei_labels) == len(set(ziwei_labels))
