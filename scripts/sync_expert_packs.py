#!/usr/bin/env python3
"""Expert-pack 生成器：从权威源生成包内全部派生物。

两类生成（唯一生成入口）：

1. 方法论卡镜像：权威源 `skills/liki/references/`（domains 真值卡 + app 编排卡）
   → `expert-packs/<pack>/skills/<skill>/references/`。多对一聚合（如
   naming ← qiming+bazi+app），目标目录的 `*.md` 必须与源并集逐名逐字节一致
   ——缺失、多余（孤儿）、内容差异均判漂移。
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
SKILL = ROOT / "skills" / "liki" / "references"
PACKS = ROOT / "expert-packs"

# <根相对路径>:<包名>:<skill 目录>[:<子目录>]（domains 全集 + app 编排卡副本）
MAPPINGS: tuple[tuple[str, str, str, str | None], ...] = (
    ("natal/domains/bazi", "liki-bazi", "bazi", None),
    ("natal/domains/ziwei", "liki-ziwei", "ziwei", None),
    ("divination/domains/liuyao", "liki-liuyao", "liuyao", None),
    ("divination/domains/qimen", "liki-qimen", "qimen", None),
    ("divination/domains/huangli", "liki-qimen", "qimen", "huangli"),
    ("fengshui/domains/bazhai", "liki-fengshui", "fengshui", None),
    ("fengshui/domains/xuankong", "liki-fengshui", "fengshui", None),
    ("fengshui/domains/fengshui", "liki-fengshui", "fengshui", None),
    ("fengshui/app", "liki-fengshui", "fengshui", None),
    ("naming/domains/qiming", "liki-naming", "naming", None),
    ("naming/domains/bazi", "liki-naming", "naming", None),
    ("naming/app", "liki-naming", "naming", None),
)

# 域 → 包名。身份文件固定是包根 expert-packs/<pack>/SKILL.md；
# 包内域 skills/<domain>/SKILL.md 由它派生（包内不再有第二份身份副本）。
DOMAIN_PACKS: dict[str, str] = {
    "bazi": "liki-bazi",
    "ziwei": "liki-ziwei",
    "liuyao": "liki-liuyao",
    "qimen": "liki-qimen",
    "fengshui": "liki-fengshui",
    "naming": "liki-naming",
}

# 包内卡片互引统一形如 references/<name>.md（ADK load_skill_resource 只认
# references/assets/scripts 前缀）；根 skill 的子树布局 token 由 sync rebase。
_REF_PATTERN = re.compile(r"references/[A-Za-z0-9_./-]+\.md")

RUNTIME_VERSION_PATH = ROOT / "skills" / "liki" / "VERSION.txt"
PACK_AUTHOR = "Liki"

# 身份文件 frontmatter 契约：每个键都有明确消费方（生成器/校验器/客户端）。
# name → 插件/skill 名；description → 域 skill 描述；description_en → plugin.json.description
# 与部署 agent 描述；display_name(_en) → 市场展示名；profession → plugin.json.profession；
# keywords/skills → 插件入口检索词与 skill 标签；max_turns → agent 回合上限；
# plugin.json 的 agents/agentName 固定指向本文件（包内不再有单独 agent 定义）。
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
    "max_turns",
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
    if not isinstance(metadata.get("max_turns"), int):
        errors.append(f"{rel}: 身份文件 max_turns 必须是整数")
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


def generated_documents() -> tuple[dict[Path, bytes], list[str]]:
    """身份文件 → 域 SKILL.md 的目标字节（包内不再有其他身份副本）。"""
    version = RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip()
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
        if (pack_dir / "agents").exists():
            errors.append(
                f"包内不得再有身份副本（唯一源是 {identity_path(domain).relative_to(ROOT)}）: "
                f"{(pack_dir / 'agents').relative_to(ROOT)}"
            )
        documents[pack_dir / "skills" / domain / "SKILL.md"] = skill_doc.encode("utf-8")
    return documents, errors


def plugin_metadata_updates() -> tuple[list[tuple[Path, dict, dict]], list[str]]:
    """Return plugin manifests and the identity-derived fields that must match."""
    version = RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip()
    updates: list[tuple[Path, dict, dict]] = []
    errors: list[str] = []
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
            "agents": [f"./{identity_path(domain).name}"],
            "agentName": pack,
            "description": metadata["description_en"],
            "profession": metadata["profession"],
            "version": version,
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


def _source_names(srcs: list[Path]) -> set[str]:
    names: set[str] = set()
    for src in srcs:
        if not src.is_dir():
            continue
        for card in src.glob("*.md"):
            if card.name != "SKILL.md":
                names.add(card.name)
    return names


def _pack_name_maps(groups: dict[Path, list[Path]]) -> tuple[dict[str, dict[str, str]], list[str]]:
    """pack → {basename: 包内 references 路径}。同包同名异位 = 歧义，fail-closed。"""
    maps: dict[str, dict[str, str]] = {}
    errors: list[str] = []
    for dest, srcs in groups.items():
        pack = dest.relative_to(PACKS).parts[0]
        rel = dest.relative_to(PACKS).parts
        prefix = "references" if len(rel) <= 4 else f"references/{rel[4]}"
        name_map = maps.setdefault(pack, {})
        for name in _source_names(srcs):
            target = f"{prefix}/{name}"
            if name in name_map and name_map[name] != target:
                errors.append(f"包内同名卡歧义: {pack} {name} → {name_map[name]} vs {target}")
            else:
                name_map[name] = target
    return maps, errors


def _dest_groups() -> dict[Path, list[Path]]:
    groups: dict[Path, list[Path]] = {}
    for src_rel, pack, skill, sub in MAPPINGS:
        dest = PACKS / pack / "skills" / skill / "references"
        if sub:
            dest = dest / sub
        groups.setdefault(dest, []).append(SKILL / src_rel)
    return groups


def _rebase(payload: bytes, name_map: dict[str, str], errors: list[str], where: str) -> bytes:
    text = payload.decode("utf-8")

    def replace(match: re.Match[str]) -> str:
        token = match.group(0)
        target = name_map.get(Path(token).name)
        if target is None:
            errors.append(f"包内死链（该包不存在此卡）: {where}: {token}")
            return token
        return target

    return _REF_PATTERN.sub(replace, text).encode("utf-8")


def _expected(dest: Path, srcs: list[Path], name_map: dict[str, str]) -> dict[str, bytes] | list[str]:
    merged: dict[str, bytes] = {}
    errors: list[str] = []
    for src in srcs:
        if not src.is_dir():
            errors.append(f"源不存在: {src.relative_to(ROOT)}")
            continue
        for card in src.glob("*.md"):
            if card.name == "SKILL.md":
                continue
            payload = card.read_bytes()
            if card.name in merged and merged[card.name] != payload:
                errors.append(f"多源同名卡内容冲突: {card.name} ← {src.relative_to(ROOT)}")
            merged[card.name] = payload
    rebased: dict[str, bytes] = {}
    for name, payload in merged.items():
        rebased[name] = _rebase(payload, name_map, errors, f"{dest.relative_to(ROOT)}/{name}")
    return errors if errors else rebased


def _mirror_cards() -> int:
    failed = False
    groups = _dest_groups()
    pack_maps, map_errors = _pack_name_maps(groups)
    for line in map_errors:
        print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
    failed = bool(map_errors)
    for dest, srcs in groups.items():
        pack = dest.relative_to(PACKS).parts[0]
        result = _expected(dest, srcs, pack_maps.get(pack, {}))
        if isinstance(result, list):
            for line in result:
                print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
            failed = True
            continue
        dest.mkdir(parents=True, exist_ok=True)
        for name, payload in result.items():
            target = dest / name
            if not target.exists() or target.read_bytes() != payload:
                target.write_bytes(payload)
        for orphan in dest.glob("*.md"):
            if orphan.name not in result:
                orphan.unlink()
        src_rel = ", ".join(str(s.relative_to(SKILL)) for s in srcs)
        print(f"[sync-expert-packs] ✓ {src_rel} → {dest.relative_to(ROOT)}")
    return 1 if failed else 0


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
    pack_maps, map_errors = _pack_name_maps(groups)
    drift.extend(map_errors)
    for dest, srcs in groups.items():
        pack = dest.relative_to(PACKS).parts[0]
        result = _expected(dest, srcs, pack_maps.get(pack, {}))
        if isinstance(result, list):
            drift.extend(result)
            continue
        if not dest.is_dir():
            drift.append(f"目标不存在: {dest.relative_to(ROOT)}")
            continue
        actual = {c.name: c.read_bytes() for c in dest.glob("*.md")}
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
        f"{len(documents)} 份包入口由身份文件生成）。"
    )
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description="Sync expert-pack artifacts from authoritative sources")
    ap.add_argument("--check", action="store_true", help="只检测漂移，不写文件")
    args = ap.parse_args()
    return check() if args.check else sync()


if __name__ == "__main__":
    sys.exit(main())
