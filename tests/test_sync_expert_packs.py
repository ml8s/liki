"""生成器契约：sync_expert_packs 的映射完整性与无漂移状态。

生成器语义（唯一生成入口）：
- 权威源 = skills/liki/references/（平铺命名 `<域>-<app|体系>-<卡>.md`）
- 顶层领域 = 专家包（1:1）：expert-packs/liki-<domain>/skills/<domain>/references/
  是根"按域前缀过滤的逐字节视图"（保留平铺名、引用不改写；手改=漂移）
- 根级共享卡（`FAQ.md` / `ops-*` / `payment-*`）只存在于根，不进任何包
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


def test_every_root_card_is_classified_and_flat():
    """根 references/ 必须平铺；每个卡要么归域（有包）、要么在根级共享白名单。"""
    assert not [p for p in SKILL.iterdir() if p.is_dir()], "references 必须平铺（无子目录）"
    assert not [p for p in SKILL.iterdir() if p.is_file() and p.suffix != ".md"], "只允许 .md 卡"
    unclassified = [
        name
        for name in sorted(p.name for p in SKILL.glob("*.md"))
        if name not in gen.ROOT_SHARED_FILES
        and not any(name.startswith(d + "-") for d, _pack in gen.MAPPINGS)
    ]
    assert not unclassified, f"未分类卡（须归域或进根级共享白名单）: {unclassified}"


def test_every_root_domain_is_one_pack():
    """顶层领域 = 专家包（1:1）：每个域有平铺 ENTRY 与对应包/身份文件。"""
    for domain, pack in gen.MAPPINGS:
        assert (SKILL / f"{domain}-ENTRY.md").is_file(), domain
        assert list(SKILL.glob(f"{domain}-*.md")), domain
        assert (gen.PACKS / pack / "SKILL.md").is_file(), pack
        assert (gen.PACKS / pack / "skills" / domain).is_dir(), pack
    assert gen.DOMAIN_PACKS == dict(gen.MAPPINGS)


def test_dest_groups_filter_by_domain_prefix():
    """dest 分组 = 按域前缀过滤；全量覆盖（共享卡除外）、且各组只含本域前缀卡。"""
    groups, errors = gen._dest_groups()
    assert not errors, errors
    assert len(groups) == len(gen.MAPPINGS)
    total = 0
    for dest, srcs in groups.items():
        domain = gen._domain_of(dest)
        assert srcs and all(src.name.startswith(domain + "-") for src in srcs), (dest, srcs)
        total += len(srcs)
    all_cards = [p.name for p in SKILL.glob("*.md") if p.name not in gen.ROOT_SHARED_FILES]
    assert total == len(all_cards)


def test_expected_is_byte_identical_domain_view():
    """包内 = 源卡的逐字节视图：字节相等、引用不改写、无旧嵌套 token、无路径改写。"""
    groups, errors = gen._dest_groups()
    assert not errors, errors
    dest = gen.PACKS / "liki-natal" / "skills" / "natal" / "references"
    result = gen._expected(dest, groups[dest])
    assert isinstance(result, dict), result
    sample = "natal-app-marriage.md"
    assert sample in result
    assert result[sample] == (SKILL / sample).read_bytes()
    text = result[sample].decode("utf-8")
    assert f"references/{sample}" not in text  # 卡不自引
    # 跨层索引禁止：app 卡不得引用域卡（场景到卡的映射在侧入口）
    assert "references/natal-bazi-" not in text
    assert "references/natal-ziwei-" not in text
    assert "references/natal/" not in text, "包内不得出现旧嵌套前缀 token"


def test_expected_fail_closed_on_invalid_refs(tmp_path):
    """非平铺 / 裸名 / 跨域 / 根级共享卡 / scripts 依赖 / 死链一律 fail-closed。"""
    dest = gen.PACKS / "liki-natal" / "skills" / "natal" / "references"
    cases = [
        ("references/divination-qimen-yongshen.md", "跨域引用"),
        ("references/natal-nope.md", "死链"),
        ("references/natal/app/x.md", "非平铺引用"),
        ("references/FAQ.md", "根级共享卡"),
        ("geju.md", "裸名引用"),
        ("python3 scripts/aipay.py status", "包内无 scripts/ 运行时"),
    ]
    for token, want in cases:
        src = tmp_path / "natal-app-probe.md"
        src.write_text(f"see `{token}`\n", encoding="utf-8")
        result = gen._expected(dest, [src])
        assert isinstance(result, list) and any(want in error for error in result), (token, result)


def test_identity_file_is_the_only_authored_expert_identity():
    """专家身份唯一手写源＝包根 SKILL.md；仓库根不再保留 persona 目录。"""
    assert not (ROOT / "agents" / "personas").exists()
    for domain, pack in gen.DOMAIN_PACKS.items():
        identity = gen.identity_path(domain)
        assert identity == gen.PACKS / pack / "SKILL.md", identity
        assert identity.is_file(), identity


def test_pack_agents_dir_holds_only_the_projected_agent_doc():
    """`agents/` 只放身份文件投影出的 agent md（平台入口），不出现手写副本。"""
    documents, errors = gen.generated_documents()
    assert not errors, errors
    assert not gen._agent_orphans(documents), gen._agent_orphans(documents)
    for domain, pack in gen.DOMAIN_PACKS.items():
        pack_dir = gen.PACKS / pack
        assert (pack_dir / "SKILL.md").is_file(), pack_dir
        assert sorted((pack_dir / "agents").glob("*.md")) == [pack_dir / "agents" / f"{pack}.md"]


def test_skill_entry_is_rendered_from_identity_file():
    """两处入口都由身份文件渲染：域 SKILL.md（ADK，skill 元数据 + 卡索引）与
    `agents/<pack>.md`（WorkBuddy agent md，平台 camelCase 字段 + 同源正文）。"""
    documents, errors = gen.generated_documents()
    assert not errors, errors
    expected: set[Path] = set()
    for domain, pack in gen.DOMAIN_PACKS.items():
        loaded = gen.load_identity(domain)
        assert isinstance(loaded, tuple), loaded
        metadata, body = loaded
        agent_doc = gen.PACKS / pack / "agents" / f"{pack}.md"
        expected.add(agent_doc)
        agent_payload = documents[agent_doc].decode("utf-8")
        agent_head, agent_body = agent_payload.split("\n---\n\n", 1)
        assert f"name: {pack}" in agent_head, pack
        assert f"description: {metadata['description_en']}" in agent_head, pack
        assert f"en: {metadata['display_name_en']}" in agent_head, pack
        assert f"zh: {metadata['display_name']}" in agent_head, pack
        assert f"maxTurns: {metadata['maxTurns']}" in agent_head, pack
        assert agent_body == body.rstrip() + "\n", pack
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


def test_skill_index_tracks_mirrored_cards():
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
    from version_projection import workbuddy_version

    version = workbuddy_version(
        (ROOT / "skills" / "liki" / "VERSION.txt").read_text(encoding="utf-8").strip()
    )
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
        bump_version.replace_frontmatter_version(copy, gen.workbuddy_version(fake_version))
        assert copy.read_text(encoding="utf-8") == payload.decode("utf-8"), target


def test_plugin_agent_pointer_resolves_to_projected_agent_doc():
    """plugin.json 的 agents/agentName 指向包内 `agents/<pack>.md`（由身份文件投影）。"""
    for domain, pack in gen.DOMAIN_PACKS.items():
        plugin = json.loads(
            (gen.PACKS / pack / ".codebuddy-plugin" / "plugin.json").read_text(encoding="utf-8")
        )
        assert plugin["agents"] == [f"./agents/{pack}.md"], pack
        assert plugin["agentName"] == pack, pack
        identity = gen.PACKS / pack / "SKILL.md"
        assert identity.is_file(), identity
        for declared in plugin["agents"]:
            assert (gen.PACKS / pack / declared.removeprefix("./")).is_file(), pack
        metadata, _body = gen.load_identity(domain)
        assert metadata["name"] == plugin["agentName"], pack
