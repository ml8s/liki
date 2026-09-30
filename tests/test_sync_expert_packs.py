"""生成器契约：sync_expert_packs 的映射完整性与无漂移状态。

生成器语义（唯一生成入口）：
- 权威源 = skills/liki/references/（domains 真值卡 + app 编排卡副本）
- expert-packs/*/skills/<skill>/references/ 全部由生成器镜像产出（手改=漂移）
- make check 跑 --check 双向检测（缺失/孤儿/内容差异）
"""
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts" / "sync_expert_packs.py"
SKILL = ROOT / "skills" / "liki" / "references"

sys.path.insert(0, str(ROOT / "scripts"))
import sync_expert_packs as gen  # noqa: E402


def test_check_reports_no_drift():
    proc = subprocess.run(
        [sys.executable, str(SCRIPT), "--check"],
        capture_output=True, text=True, cwd=ROOT,
    )
    assert proc.returncode == 0, proc.stderr


def test_mappings_cover_every_root_domain_dir():
    """根 references/*/domains/* 全集必须被 MAPPINGS 覆盖（新增域防漏映射）。"""
    root_dirs = {
        str(p.relative_to(SKILL))
        for p in SKILL.glob("*/domains/*")
        if p.is_dir()
    }
    mapped = {m[0] for m in gen.MAPPINGS}
    missing = root_dirs - mapped
    assert not missing, f"根域目录未进 MAPPINGS（生成器会漏卡）: {sorted(missing)}"


def test_app_orchestration_cards_are_mapped():
    """app 编排卡副本（WorkBuddy 运行时需要）必须由生成器管，禁止手维护。"""
    mapped = {m[0] for m in gen.MAPPINGS}
    assert "naming/app" in mapped
    assert "fengshui/app" in mapped


def test_every_mapped_source_exists_and_dest_inside_expert_packs():
    for src_rel, pack, skill, sub in gen.MAPPINGS:
        assert (SKILL / src_rel).is_dir(), f"源缺失: {src_rel}"
        dest = gen.PACKS / pack / "skills" / skill / "references"
        if sub:
            dest = dest / sub
        assert str(dest).startswith(str(gen.PACKS)), dest
        assert (gen.PACKS / pack / "SKILL.md").is_file(), pack


def test_dest_groups_aggregate_multi_source():
    """多对一聚合：naming ← qiming+bazi+app；fengshui ← bazhai+xuankong+domains+fengshui/app。"""
    groups = gen._dest_groups()
    naming_dest = gen.PACKS / "liki-naming" / "skills" / "naming" / "references"
    fengshui_dest = gen.PACKS / "liki-fengshui" / "skills" / "fengshui" / "references"
    assert len(groups[naming_dest]) == 3, groups[naming_dest]
    assert len(groups[fengshui_dest]) == 4, groups[fengshui_dest]


def test_expected_rebases_root_layout_refs_into_pack_paths():
    """包内互引必须 rebase：根子树 token（references/<x>/domains/<y>/<name>.md）
    → 包平铺路径（references/<name>.md），否则 ADK load_skill_resource 死链。"""
    groups = gen._dest_groups()
    maps, errors = gen._pack_name_maps(groups)
    assert not errors, errors
    dest = gen.PACKS / "liki-fengshui" / "skills" / "fengshui" / "references"
    result = gen._expected(dest, groups[dest], maps["liki-fengshui"])
    assert isinstance(result, dict), result
    text = result["fengshui.md"].decode("utf-8")
    assert "references/fengshui/domains/" not in text, "根子树 token 未 rebase"
    assert "references/youxing.md" in text
    assert "references/caijue.md" in text


def test_rebase_reports_cross_pack_dead_link():
    """引用本包不存在的卡 = 包必须自包含，fail-closed 报错且不改写。"""
    errors: list[str] = []
    payload = gen._rebase(
        b"read `references/elsewhere.md`",
        {"youxing.md": "references/youxing.md"},
        errors,
        "pack/references/x.md",
    )
    assert len(errors) == 1 and "死链" in errors[0], errors
    assert payload == b"read `references/elsewhere.md`"


def test_pack_name_maps_rejects_ambiguous_basename(tmp_path, monkeypatch):
    """同包同名异位（references/a.md 与 references/sub/a.md）必须报歧义。"""
    monkeypatch.setattr(gen, "PACKS", tmp_path / "packs")
    src1, src2 = tmp_path / "src1", tmp_path / "src2"
    src1.mkdir()
    src2.mkdir()
    (src1 / "a.md").write_text("main", encoding="utf-8")
    (src2 / "a.md").write_text("sub", encoding="utf-8")
    groups = {
        tmp_path / "packs" / "liki-x" / "skills" / "x" / "references": [src1],
        tmp_path / "packs" / "liki-x" / "skills" / "x" / "references" / "sub": [src2],
    }
    maps, errors = gen._pack_name_maps(groups)
    assert any("歧义" in e for e in errors), errors
    assert maps["liki-x"]["a.md"] == "references/a.md"
