"""生成器契约：sync_expert_packs 的映射完整性与无漂移状态。

生成器语义（唯一生成入口）：
- 权威源 = skills/liki/references/（domains 真值卡 + app 编排卡副本）
- 顶层领域 = 专家包（1:1）：expert-packs/liki-<domain>/skills/<domain>/references/
  是根 references/<domain>/ 子树的逐字节镜像（手改=漂移）
- make check 跑 --check 双向检测（缺失/孤儿/内容差异）
"""
from __future__ import annotations

import json
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
    """根 references/*/domains/* 全集必须被 MAPPINGS 覆盖（新增体系防漏镜像）。"""
    root_dirs = [p.relative_to(SKILL) for p in SKILL.glob("*/domains/*") if p.is_dir()]
    assert root_dirs
    mapped = {m[0] for m in gen.MAPPINGS}
    missing = [str(d) for d in root_dirs if str(d).split("/")[0] not in mapped]
    assert not missing, f"根域目录未进 MAPPINGS（生成器会漏卡）: {missing}"


def test_every_root_domain_is_one_pack():
    """顶层领域 = 专家包（1:1）：每个 references/<domain>/ 都有对应包与身份文件。"""
    mapped = {m[0]: (m[1], m[2]) for m in gen.MAPPINGS}
    # 领域目录 = 带 ENTRY.md 的顶层目录（references/ 下还有共享的 ops/、payment/）。
    domains = {p.name for p in SKILL.iterdir() if p.is_dir() and (p / "ENTRY.md").is_file()}
    assert domains == set(mapped), (domains, set(mapped))
    for domain, (pack, skill) in mapped.items():
        assert pack == gen.DOMAIN_PACKS[domain], domain
        assert skill == domain, domain
        assert (SKILL / domain / "ENTRY.md").is_file(), domain
        assert (gen.PACKS / pack / "SKILL.md").is_file(), pack
        assert (gen.PACKS / pack / "skills" / domain).is_dir(), pack


def test_every_mapped_source_exists_and_dest_inside_expert_packs():
    for src_rel, pack, skill in gen.MAPPINGS:
        assert (SKILL / src_rel).is_dir(), f"源缺失: {src_rel}"
        dest = gen.PACKS / pack / "skills" / skill / "references"
        assert str(dest).startswith(str(gen.PACKS)), dest
        assert (gen.PACKS / pack / "SKILL.md").is_file(), pack


def test_dest_groups_mirror_one_subtree_per_pack():
    """1:1 镜像：每个目标 references/ 只由一个根子树喂入（不再多对一聚合）。"""
    groups = gen._dest_groups()
    assert len(groups) == len(gen.MAPPINGS)
    for dest, srcs in groups.items():
        assert len(srcs) == 1, (dest, srcs)


def test_expected_rebases_root_layout_refs_into_pack_paths():
    """包内互引必须 rebase：`references/<domain>/…` → `references/…`（相对 skill 根），
    否则 ADK load_skill_resource 死链；除此之外镜像逐字节保真。"""
    groups = gen._dest_groups()
    dest = gen.PACKS / "liki-natal" / "skills" / "natal" / "references"
    result = gen._expected(dest, groups[dest])
    assert isinstance(result, dict), result
    text = result["app/marriage.md"].decode("utf-8")
    assert "references/natal/" not in text, "根子树 token 未 rebase"
    assert "references/domains/bazi/shishen.md" in text
    assert "references/domains/ziwei/yingqi.md" in text
    root_card = (SKILL / "natal" / "app" / "marriage.md").read_text(encoding="utf-8")
    assert text == root_card.replace("references/natal/", "references/")
    # 子树布局保留（app/ 与 domains/<体系>/ 分目录，不再平铺）
    assert "domains/bazi/geju.md" in result and "domains/ziwei/geju.md" in result


def test_rebase_reports_cross_domain_dead_link():
    """引用别的领域子树 = 包必须自包含，fail-closed 报错且不改写。"""
    errors: list[str] = []
    payload = gen._rebase(
        b"read `references/naming/app/naming.md`",
        "natal",
        {"references/app/marriage.md"},
        errors,
        "pack/references/x.md",
    )
    assert len(errors) == 1 and "跨域引用" in errors[0], errors
    assert payload == b"read `references/naming/app/naming.md`"


def test_rebase_reports_intra_domain_dead_link():
    """本域 token 指向包内不存在的卡（如根子树漏镜像）同样 fail-closed。"""
    errors: list[str] = []
    gen._rebase(b"read `references/natal/app/nope.md`", "natal", set(), errors, "pack/x.md")
    assert len(errors) == 1 and "死链" in errors[0], errors
    assert "references/app/nope.md" in errors[0], errors

def test_identity_file_is_the_only_authored_expert_identity():
    """专家身份唯一手写源＝包根 SKILL.md；仓库根不再保留 persona 目录。"""
    assert not (ROOT / "agents" / "personas").exists()
    assert (ROOT / "agents" / "router" / "router.md").is_file()
    for domain, pack in gen.DOMAIN_PACKS.items():
        identity = gen.identity_path(domain)
        assert identity == gen.PACKS / pack / "SKILL.md", identity
        assert identity.is_file(), identity


def test_pack_has_no_second_identity_copy():
    """包内只有一处身份文件：不再有 agents/<domain>-expert.md 之类的副本。"""
    documents, errors = gen.generated_documents()
    assert not errors, errors
    for domain, pack in gen.DOMAIN_PACKS.items():
        pack_dir = gen.PACKS / pack
        assert not (pack_dir / "agents").exists(), pack_dir / "agents"
        assert (pack_dir / "SKILL.md").is_file(), pack_dir


def test_skill_entry_is_rendered_from_identity_file():
    """域 SKILL.md 由身份文件渲染：skill 元数据映射 + 同源正文 + 自动卡索引。"""
    documents, errors = gen.generated_documents()
    assert not errors, errors
    expected: set[Path] = set()
    for domain, pack in gen.DOMAIN_PACKS.items():
        loaded = gen.load_identity(domain)
        assert isinstance(loaded, tuple), loaded
        metadata, body = loaded
        target = gen.PACKS / pack / "skills" / domain / "SKILL.md"
        expected.add(target)
        assert target.read_bytes() == documents[target], target
        payload = documents[target].decode("utf-8")
        head, rendered = payload.split("\n---\n\n", 1)
        assert f"name: {domain}" in head
        assert f"description: {metadata['description']}" in head
        assert f"display_name: {metadata['display_name']}" in head
        assert f"display_name_en: {metadata['display_name_en']}" in head
        assert body.strip() in rendered, target
    assert set(documents) == expected, set(documents) ^ expected


def test_nested_skill_index_tracks_mirrored_cards():
    """域 SKILL.md 的方法论索引随镜像自动生成，不得手维护卡名清单。"""
    documents, errors = gen.generated_documents()
    assert not errors, errors
    for domain, pack in gen.DOMAIN_PACKS.items():
        target = gen.PACKS / pack / "skills" / domain / "SKILL.md"
        assert target.read_bytes() == documents[target], target
        payload = documents[target].decode("utf-8")
        _, rendered = payload.split("\n---\n\n", 1)
        indexed = {
            line.split("`")[1]
            for line in rendered.splitlines()
            if line.startswith("- `references/")
        }
        cards = {
            card.relative_to(target.parent).as_posix()
            for card in (target.parent / "references").rglob("*.md")
            if card.name != "SKILL.md"
        }
        assert indexed == cards, (domain, sorted(cards - indexed), sorted(indexed - cards))


def test_plugin_metadata_is_identity_derived():
    updates, errors = gen.plugin_metadata_updates()
    assert not errors, errors
    assert len(updates) == len(gen.DOMAIN_PACKS)
    version = (ROOT / "skills" / "liki" / "VERSION.txt").read_text(encoding="utf-8").strip()
    for path, doc, expected in updates:
        assert expected["version"] == version, path
        for key, value in expected.items():
            assert doc.get(key) == value, (path, key)


def test_version_bump_keeps_generated_pack_entries_in_sync(tmp_path, monkeypatch):
    """bump_version 直接改写包内生成物：改写后的字节必须仍等于生成器输出。

    否则每次发版 bump 都会立刻让 `sync_expert_packs.py --check` 报漂移。
    """
    fake_version = "2999.01.01.0"
    version_file = tmp_path / "VERSION.txt"
    version_file.write_text(fake_version + "\n", encoding="utf-8")
    monkeypatch.setattr(gen, "RUNTIME_VERSION_PATH", version_file)
    documents, errors = gen.generated_documents()
    assert not errors, errors

    import bump_version  # noqa: E402  （scripts/ 已在文件头插入 sys.path）

    versioned = {path: payload for path, payload in documents.items() if path.name == "SKILL.md"}
    assert len(versioned) == len(gen.DOMAIN_PACKS)
    for index, (target, payload) in enumerate(sorted(versioned.items())):
        copy = tmp_path / str(index) / target.name
        copy.parent.mkdir(parents=True, exist_ok=True)
        copy.write_text(target.read_text(encoding="utf-8"), encoding="utf-8")
        bump_version.replace_frontmatter_version(copy, fake_version)
        assert copy.read_text(encoding="utf-8") == payload.decode("utf-8"), target

def test_plugin_agent_pointer_resolves_to_identity_file():
    """plugin.json 的 agents/agentName 指向身份文件本身，包内不再有单独 agent 定义。"""
    for domain, pack in gen.DOMAIN_PACKS.items():
        plugin = json.loads(
            (gen.PACKS / pack / ".codebuddy-plugin" / "plugin.json").read_text(encoding="utf-8")
        )
        assert plugin["agents"] == ["./SKILL.md"], pack
        assert plugin["agentName"] == pack, pack
        identity = gen.PACKS / pack / "SKILL.md"
        assert identity.is_file(), identity
        for declared in plugin["agents"]:
            assert (gen.PACKS / pack / declared.removeprefix("./")).is_file(), pack
        metadata, _body = gen.load_identity(domain)
        assert metadata["name"] == plugin["agentName"], pack
