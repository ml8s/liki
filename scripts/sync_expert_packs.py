#!/usr/bin/env python3
"""Expert-pack 生成器：从权威源生成包内全部派生物。

两类生成（唯一生成入口）：

1. 方法论卡镜像：权威源 `skills/liki/references/`（domains 真值卡 + app 编排卡）
   → `expert-packs/<pack>/skills/<domain>/references/`。顶层领域即专家包
   （1:1），包内 references/ 是根 `references/<domain>/` 子树的逐字节镜像
   （保留 `app/`、`domains/<体系>/` 布局），只把卡内互引的
   `references/<domain>/…` 前缀改写成包内相对路径——缺失、多余（孤儿）、内容
   差异均判漂移；引用别的领域子树（跨域 token）判死链，包必须自包含。
2. 专家身份派生物：权威源身份文件 `expert-packs/<pack>/SKILL.md`（frontmatter 元数据
   + 身份正文）→ 域 `skills/<domain>/SKILL.md`（ADK skill 入口）与 `.codebuddy-plugin/plugin.json`
   的 agent 指针/description/profession/version（插件声明的 agent 定义文件就是 `./SKILL.md`
   本身）。身份正文同时就是内部 AgentDeployment 的 instruction 骨架，三者同源，
   除身份文件外一律生成，禁止手改，包内不得再有第二份身份副本。

    python3 scripts/sync_expert_packs.py            # 镜像生成（覆盖 + 清孤儿）
    python3 scripts/sync_expert_packs.py --check    # 双向漂移检测（make check / CI）
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import workbuddy_version  # noqa: E402  制品版本的唯一投影

SKILL = ROOT / "skills" / "liki" / "references"
PACKS = ROOT / "expert-packs"

# <根 skill 子目录>:<包名>:<包内 skill 目录>。顶层领域 = 专家包（1:1）：
# `skills/liki/references/<domain>/` 整棵子树（ENTRY/TOOLS/app/domains）镜像进
# 包内 `skills/<domain>/references/`，保留子目录布局。
MAPPINGS: tuple[tuple[str, str, str], ...] = (
    ("natal", "liki-natal", "natal"),
    ("divination", "liki-divination", "divination"),
    ("fengshui", "liki-fengshui", "fengshui"),
    ("naming", "liki-naming", "naming"),
)

# 域 → 包名。身份文件固定是包根 expert-packs/<pack>/SKILL.md；
# 包内域 skills/<domain>/SKILL.md 由它派生（包内不再有第二份身份副本）。
DOMAIN_PACKS: dict[str, str] = {domain: pack for domain, pack, _skill in MAPPINGS}

# 包内卡片互引统一形如 references/<path>.md（ADK load_skill_resource 只认
# references/assets/scripts 前缀）：包内路径 = 根路径去掉 `references/<domain>/`。
_REF_PATTERN = re.compile(r"references/([A-Za-z0-9_./-]+\.md)")

# 制品版本 = 发行 CalVer 的平台投影（见 scripts/version_projection.py）。
RUNTIME_VERSION_PATH = ROOT / "skills" / "liki" / "VERSION.txt"
# 已上架连接器：平台分配 ID 的唯一源（source → connector id）。专家不声明自带 MCP，
# 只依赖这两个聚合连接器，见 workbuddy/README.md。
CONNECTOR_IDS_PATH = ROOT / "workbuddy" / "connector-ids.json"
CONNECTOR_SOURCES_DIR = ROOT / "workbuddy"
PACK_AUTHOR = "Liki"

# 身份文件 frontmatter 契约：每个键都有明确消费方，无一「写而无用」。
# 消费方一＝生成器（本文件）：description → 域 skill 描述；description_en →
# plugin.json.description 与部署 agent 描述；display_name(_en) → 域 skill metadata；
# profession → plugin.json.profession；keywords → plugin.json.keywords（市场检索词）；
# skills → plugin.json.skills（预加载 skill 路径，由 skill 名 rebase 成 ./skills/<name>）。
# 消费方二＝WorkBuddy agent md（`agents/<pack>.md`，平台按 `agents/` 目录约定取文件）：
# 该文件由身份文件投影，frontmatter 用平台字段名（camelCase：`displayName` / `maxTurns`），
# `name` 与文件名一致（不含 `.md`），正文即身份正文；身份仍是包根 SKILL.md 一处手写。
# 正文＝身份骨架：既是包内两处入口的正文，也是内部 AgentDeployment 的 instruction。
IDENTITY_KEYS = (
    "name",
    "description",
    "description_en",
    "display_name",
    "display_name_en",
    "profession",
    "keywords",
    "skills",
    "maxTurns",
)
IDENTITY_LOCALES = ("en", "zh")


def identity_path(domain: str) -> Path:
    return PACKS / DOMAIN_PACKS[domain] / "SKILL.md"


def split_frontmatter(path: Path, label: str) -> tuple[list[str], str] | list[str]:
    """解析 frontmatter：返回 (frontmatter 行, 正文)；损坏时返回错误行列表。"""
    text = path.read_text(encoding="utf-8")
    rel = path.relative_to(ROOT)
    if not text.startswith("---\n"):
        return [f"{label} 缺少 frontmatter 起始标记: {rel}"]
    end = text.find("\n---\n")
    if end < 0:
        return [f"{label} 缺少 frontmatter 结束标记: {rel}"]
    front = text[4:end].splitlines()
    body = text[end + len("\n---\n"):]
    if body.startswith("\n"):
        body = body[1:]
    return front, body


def load_identity(domain: str) -> tuple[dict, str] | list[str]:
    """身份文件 → (frontmatter, 身份正文)；损坏或元数据不全时返回错误行列表。"""
    path = identity_path(domain)
    if not path.is_file():
        return [f"身份文件缺失: {path.relative_to(ROOT)}"]
    parsed = split_frontmatter(path, "身份文件")
    if isinstance(parsed, list):
        return parsed
    front, body = parsed
    rel = path.relative_to(ROOT)
    try:
        metadata = yaml.safe_load("\n".join(front))
    except Exception as exc:
        return [f"身份文件 frontmatter YAML 无效: {rel}: {exc}"]
    if not isinstance(metadata, dict):
        return [f"身份文件 frontmatter 必须是对象: {rel}"]
    errors = [f"{rel}: 身份文件 frontmatter 缺少 {key}" for key in IDENTITY_KEYS if key not in metadata]
    pack = DOMAIN_PACKS[domain]
    if metadata.get("name") != pack:
        errors.append(f"{rel}: 身份文件 name 必须是 {pack}")
    value = metadata.get("profession")
    if not isinstance(value, dict):
        errors.append(f"{rel}: 身份文件 profession 必须是 en/zh 映射")
    else:
        errors.extend(
            f"{rel}: 身份文件 profession 缺少 {locale}" for locale in IDENTITY_LOCALES if not value.get(locale)
        )
    for key in ("keywords", "skills"):
        value = metadata.get(key)
        if not isinstance(value, list) or not value:
            errors.append(f"{rel}: 身份文件 {key} 必须是非空列表")
    if not isinstance(metadata.get("maxTurns"), int):
        errors.append(f"{rel}: 身份文件 maxTurns 必须是整数")
    return errors if errors else (metadata, body)


class _PackDumper(yaml.SafeDumper):
    """块序列缩进两格（PyYAML 默认顶格），与包内其余 md 保持一致。"""

    def increase_indent(self, flow: bool = False, indentless: bool = False):  # noqa: D102
        return super().increase_indent(flow, False)


def _frontmatter_block(metadata: dict) -> str:
    dumped = yaml.dump(
        metadata,
        Dumper=_PackDumper,
        allow_unicode=True,
        sort_keys=False,
        width=4096,
        default_flow_style=False,
    )
    return "---\n" + dumped + "---\n\n"


def _card_title(card: Path) -> str:
    for line in card.read_text(encoding="utf-8").splitlines():
        if line.startswith("# "):
            return line[2:].strip()
    return card.stem


def methodology_index(skill_dir: Path) -> list[str]:
    """包内方法论卡索引（`references/**` 逐卡一行），随镜像同步自动跟随。"""
    references = skill_dir / "references"
    return [
        f"- `{card.relative_to(skill_dir).as_posix()}` {_card_title(card)}"
        for card in sorted(references.rglob("*.md"))
        if card.name != "SKILL.md"
    ]


def connector_ids() -> tuple[list[str], list[str]]:
    """返回 (按 source 排序的已上架连接器 ID 列表, 错误列表)。

    与 `workbuddy/*-connector/connector-meta.json` 的 `source` 双向对齐：少一个 ID 或
    多一个 ID 都是漂移，fail-closed。
    """
    errors: list[str] = []
    try:
        declared = json.loads(CONNECTOR_IDS_PATH.read_text(encoding="utf-8"))
    except Exception as exc:  # noqa: BLE001 - 单一源读不出来就不能生成制品
        return [], [f"{CONNECTOR_IDS_PATH.relative_to(ROOT)}: 读取失败: {exc}"]
    if not isinstance(declared, dict):
        return [], [f"{CONNECTOR_IDS_PATH.relative_to(ROOT)}: 须为 source → connector id 映射"]

    sources: list[str] = []
    for meta_path in sorted(CONNECTOR_SOURCES_DIR.glob("*-connector/connector-meta.json")):
        try:
            source = json.loads(meta_path.read_text(encoding="utf-8")).get("source")
        except Exception as exc:  # noqa: BLE001 - 同上
            errors.append(f"{meta_path.relative_to(ROOT)}: 读取失败: {exc}")
            continue
        if not isinstance(source, str) or not source:
            errors.append(f"{meta_path.relative_to(ROOT)}: 缺 source")
            continue
        sources.append(source)
    for source in sources:
        if source not in declared:
            errors.append(f"{CONNECTOR_IDS_PATH.relative_to(ROOT)}: 缺已上架连接器 `{source}` 的平台 ID")
    for source in declared:
        if source not in sources:
            errors.append(f"{CONNECTOR_IDS_PATH.relative_to(ROOT)}: `{source}` 不在 workbuddy/*-connector 提交源里")

    # 顺序取声明文件里的书写顺序（engine → counsel），不重排，保持可读。
    ids: list[str] = []
    for source in declared:
        value = declared.get(source)
        if isinstance(value, str) and value.strip():
            ids.append(value.strip())
        elif value is not None:
            errors.append(f"{CONNECTOR_IDS_PATH.relative_to(ROOT)}: `{source}` 的 ID 须为非空字符串")
    return ids, errors


def generated_documents() -> tuple[dict[Path, bytes], list[str]]:
    """身份文件 → 域 SKILL.md 的目标字节（包内不再有其他身份副本）。"""
    version = workbuddy_version(RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip())
    documents: dict[Path, bytes] = {}
    errors: list[str] = []
    for domain, pack in DOMAIN_PACKS.items():
        loaded = load_identity(domain)
        if isinstance(loaded, list):
            errors.extend(loaded)
            continue
        metadata, body = loaded
        identity_body = body.rstrip() + "\n"
        pack_dir = PACKS / pack
        index = methodology_index(pack_dir / "skills" / domain)
        agent_doc = _frontmatter_block(
            {
                "name": pack,
                "description": metadata["description_en"],
                "displayName": {
                    "en": metadata["display_name_en"],
                    "zh": metadata["display_name"],
                },
                "profession": metadata["profession"],
                "maxTurns": metadata["maxTurns"],
            }
        ) + identity_body
        documents[pack_dir / "agents" / f"{pack}.md"] = agent_doc.encode("utf-8")
        skill_doc = _frontmatter_block(
            {
                "name": domain,
                "description": metadata["description"],
                "metadata": {
                    "display_name": metadata["display_name"],
                    "display_name_en": metadata["display_name_en"],
                    "version": version,
                    "author": PACK_AUTHOR,
                },
            }
        ) + identity_body + (
            f"\n## 方法论（{len(index)} 卡）\n\n" + "\n".join(index) + "\n"
        )
        documents[pack_dir / "skills" / domain / "SKILL.md"] = skill_doc.encode("utf-8")
    return documents, errors


def plugin_metadata_updates() -> tuple[list[tuple[Path, dict, dict]], list[str]]:
    """Return plugin manifests and the identity-derived fields that must match."""
    version = workbuddy_version(RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip())
    connectors, errors = connector_ids()
    updates: list[tuple[Path, dict, dict]] = []
    for domain, pack in DOMAIN_PACKS.items():
        loaded = load_identity(domain)
        if isinstance(loaded, list):
            errors.extend(loaded)
            continue
        metadata, _body = loaded
        plugin_path = PACKS / pack / ".codebuddy-plugin" / "plugin.json"
        try:
            doc = json.loads(plugin_path.read_text(encoding="utf-8"))
        except Exception as exc:
            errors.append(f"plugin manifest读取失败: {plugin_path.relative_to(ROOT)}: {exc}")
            continue
        expected = {
            "agents": [f"./agents/{pack}.md"],
            "agentName": pack,
            "description": metadata["description_en"],
            "profession": metadata["profession"],
            "version": version,
            # 市场检索词与预加载 skill 路径：单一源是身份文件，plugin.json 只做投影。
            "keywords": list(metadata["keywords"]),
            "skills": [f"./skills/{name}" for name in metadata["skills"]],
            # 运行时依赖：连接器 ID 的单一源是 workbuddy/connector-ids.json，专家不自带 MCP。
            "dependencies": {"connectors": list(connectors)},
        }
        updates.append((plugin_path, doc, expected))
    return updates, errors


def _plugin_metadata_drift(
    updates: list[tuple[Path, dict, dict]]
) -> list[tuple[Path, dict, dict]]:
    drift: list[tuple[Path, dict, dict]] = []
    for path, doc, expected in updates:
        mismatched = {
            key: expected[key]
            for key, value in expected.items()
            if doc.get(key) != value
        }
        if mismatched:
            drift.append((path, doc, mismatched))
    return drift


def _dest_groups() -> dict[Path, list[Path]]:
    groups: dict[Path, list[Path]] = {}
    for src_rel, pack, skill in MAPPINGS:
        dest = PACKS / pack / "skills" / skill / "references"
        groups.setdefault(dest, []).append(SKILL / src_rel)
    return groups


def _domain_of(dest: Path) -> str:
    """包内 references 目录 → 领域名（= 包内 skill 目录名）。"""
    return dest.relative_to(PACKS).parts[2]


def _rebase(payload: bytes, domain: str, valid: set[str], errors: list[str], where: str) -> bytes:
    """根 skill 互引 → 包内路径：去掉 `references/<domain>/` 前缀。

    引用别的领域子树的 token 在包内必然死链（域专家包只镜像本域），fail-closed。
    """
    text = payload.decode("utf-8")

    def replace(match: re.Match[str]) -> str:
        token = match.group(0)
        parts = match.group(1).split("/")
        if len(parts) >= 2 and parts[0] == domain:
            target = "references/" + "/".join(parts[1:])
            if target in valid:
                return target
            errors.append(f"包内死链: {where}: {token} → {target}")
            return token
        errors.append(f"跨域引用（包必须自包含）: {where}: {token}")
        return token

    return _REF_PATTERN.sub(replace, text).encode("utf-8")


def _expected(dest: Path, srcs: list[Path]) -> dict[str, bytes] | list[str]:
    """根 `references/<domain>/` 子树 → 包内 references/ 相对路径映射（已 rebase）。

    键是相对包内 `references/` 的路径（`ENTRY.md`、`domains/bazi/dayun.md`），
    正文里的互引 token 是相对 skill 根的 `references/<键>`（ADK load_skill_resource）。
    """
    merged: dict[str, bytes] = {}
    errors: list[str] = []
    domain = _domain_of(dest)
    for src in srcs:
        if not src.is_dir():
            errors.append(f"源不存在: {src.relative_to(ROOT)}")
            continue
        for card in sorted(src.rglob("*.md")):
            if card.name == "SKILL.md":
                continue
            merged[card.relative_to(src).as_posix()] = card.read_bytes()
    valid = {f"references/{name}" for name in merged}
    rebased: dict[str, bytes] = {}
    for name, payload in merged.items():
        rebased[name] = _rebase(payload, domain, valid, errors, f"{dest.relative_to(ROOT)}/{name}")
    return errors if errors else rebased


def _actual_cards(dest: Path) -> dict[str, bytes]:
    return {card.relative_to(dest).as_posix(): card.read_bytes() for card in dest.rglob("*.md")}


def _mirror_cards() -> int:
    failed = False
    groups = _dest_groups()
    for dest, srcs in groups.items():
        result = _expected(dest, srcs)
        if isinstance(result, list):
            for line in result:
                print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
            failed = True
            continue
        dest.mkdir(parents=True, exist_ok=True)
        for name, payload in result.items():
            target = dest / name
            if not target.exists() or target.read_bytes() != payload:
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(payload)
        for name, _payload in _actual_cards(dest).items():
            if name not in result:
                (dest / name).unlink()
        for directory in sorted((d for d in dest.rglob("*") if d.is_dir()), reverse=True):
            if not any(directory.iterdir()):
                directory.rmdir()
        src_rel = ", ".join(str(s.relative_to(SKILL)) for s in srcs)
        print(f"[sync-expert-packs] ✓ {src_rel} → {dest.relative_to(ROOT)}")
    return 1 if failed else 0


def _agent_orphans(documents: dict[Path, bytes]) -> list[Path]:
    """`agents/` 只允许身份文件投影出的那一个 md；其余都是身份副本回流。"""
    allowed = {path for path in documents if path.parent.name == "agents"}
    packs_dirs = {path.parent for path in allowed}
    found = {
        extra
        for packs_dir in packs_dirs
        if packs_dir.is_dir()
        for extra in packs_dir.glob("*.md")
    }
    return sorted(found - allowed)


def _write_documents(documents: dict[Path, bytes]) -> None:
    for path, payload in documents.items():
        if not path.exists() or path.read_bytes() != payload:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(payload)
            print(f"[sync-expert-packs] ✓ 身份文件 → {path.relative_to(ROOT)}")


def sync() -> int:
    failed = _mirror_cards() != 0

    documents, doc_errors = generated_documents()
    for line in doc_errors:
        print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
    failed = failed or bool(doc_errors)
    _write_documents(documents)
    for orphan in _agent_orphans(documents):
        orphan.unlink()
        print(f"[sync-expert-packs] ✓ 清理身份副本: {orphan.relative_to(ROOT)}")

    updates, plugin_errors = plugin_metadata_updates()
    for line in plugin_errors:
        print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
    failed = failed or bool(plugin_errors)
    for path, doc, expected in updates:
        if any(doc.get(key) != value for key, value in expected.items()):
            doc.update(expected)
            path.write_text(
                json.dumps(doc, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"[sync-expert-packs] ✓ plugin metadata → {path.relative_to(ROOT)}")

    if failed:
        return 1
    print("[sync-expert-packs] 完成。")
    return 0


def check() -> int:
    drift: list[str] = []
    groups = _dest_groups()
    for dest, srcs in groups.items():
        result = _expected(dest, srcs)
        if isinstance(result, list):
            drift.extend(result)
            continue
        if not dest.is_dir():
            drift.append(f"目标不存在: {dest.relative_to(ROOT)}")
            continue
        actual = _actual_cards(dest)
        for name in sorted(set(result) - set(actual)):
            drift.append(f"缺失: {dest.relative_to(ROOT)}/{name}")
        for name in sorted(set(actual) - set(result)):
            drift.append(f"多余（孤儿卡，双向漂移）: {dest.relative_to(ROOT)}/{name}")
        for name in sorted(set(result) & set(actual)):
            if result[name] != actual[name]:
                drift.append(f"内容漂移: {dest.relative_to(ROOT)}/{name}")

    documents, doc_errors = generated_documents()
    drift.extend(doc_errors)
    for path, payload in documents.items():
        if not path.is_file():
            drift.append(f"缺失: {path.relative_to(ROOT)}")
        elif path.read_bytes() != payload:
            drift.append(f"内容漂移: {path.relative_to(ROOT)}")
    for orphan in _agent_orphans(documents):
        drift.append(f"多余（身份副本，唯一源是包根 SKILL.md）: {orphan.relative_to(ROOT)}")

    updates, plugin_errors = plugin_metadata_updates()
    drift.extend(plugin_errors)
    for path, _doc, expected in _plugin_metadata_drift(updates):
        for key, value in expected.items():
            drift.append(f"plugin metadata漂移: {path.relative_to(ROOT)}/{key}")

    if drift:
        print("[sync-expert-packs] ✗ expert-packs 与权威源存在漂移：", file=sys.stderr)
        for line in drift:
            print(f"  {line}", file=sys.stderr)
        print("运行 `make sync-expert-packs` 重新生成。", file=sys.stderr)
        return 1
    total = len(_dest_groups())
    print(
        f"[sync-expert-packs] ✓ 无漂移（{total} 目标 × 集合+内容双向一致；"
        f"{len(documents)} 份入口（域 skill + agent md）由身份文件生成）。"
    )
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description="Sync expert-pack artifacts from authoritative sources")
    ap.add_argument("--check", action="store_true", help="只检测漂移，不写文件")
    args = ap.parse_args()
    return check() if args.check else sync()


if __name__ == "__main__":
    sys.exit(main())
