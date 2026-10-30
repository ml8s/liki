"""Unified Liki skill document checker.

Checks assertion ids, resolvable skill-relative paths (flat
`references/<域>-<app|体系>-<卡>.md`), resolvable section anchors, MCP
method/tool-name prohibition (capabilities are discovered via introspection,
never hard-coded in skill docs), lazy-loading budgets, and reachability of
domain knowledge docs.
"""

import csv
import glob
import json
import os
import re
import sys
from pathlib import Path

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SKILL = sys.argv[1] if len(sys.argv) > 1 else os.path.join(_ROOT, "skills", "liki")
REFS = os.path.join(SKILL, "references")
DOMAINS = ("liki",)
# 域 → 体系段白名单（平铺命名 `<域>-<体系>-<卡>.md` 的第二段）。
DOMAIN_SYSTEMS = {
    "natal": ("bazi", "ziwei"),
    "divination": ("liuyao", "qimen", "huangli"),
    "fengshui": ("bazhai", "xuankong"),
    "naming": ("bazi", "qiming"),
}
ASSERTION_MANIFESTS = os.path.join(_ROOT, "counsel", "app", "*", "tools", "assertions", "assertions.csv")
CATALOG_PATH = os.path.join(_ROOT, "tools", "mcp-tool-catalog.json")

METHOD_WHITELIST = {
    "bazi_chart", "bazi_fullchart", "bazi_liunian", "bazi_liuyue", "bazi_liuri",
    "bazi_liushi", "bazi_xiaoyun", "bazi_bond",
    "ziwei_chart", "ziwei_fullchart", "ziwei_liunian", "ziwei_liuyue", "ziwei_liuri",
    "ziwei_liushi", "ziwei_daxian", "ziwei_bond",
    "liuyao_qigua", "liuyao_chart", "qimen_chart", "huangli_days",
    "bazhai_chart", "bazhai_layout", "xuankong_chart", "xuankong_liunian",
    "time_now", "city_coords", "tianwen_resolve",
}
_METHOD_PREFIXES = tuple(sorted({m.split("_")[0] for m in METHOD_WHITELIST}))

# 场景卡行数上限；命书卡承载全卷输出模板（章节表）、居家风水卡承载八宅×玄空合参，结构固定，单独放宽。
APP_CARD_LINE_BUDGET = {"natal-app-mingshu.md": 75, "fengshui-app-fengshui.md": 75}
_REF_PATH = r"references/[A-Za-z0-9_][A-Za-z0-9_./\-]*\.(?:md|py|json|csv|sh|cmd)"
_ANCHOR_PATH = r"references/[A-Za-z0-9_-]+\.md"
_BARE_MD = r"`([A-Za-z0-9_-]+\.md)`"
_BARE_MD_ALLOWED = {"SKILL.md", "README.md"}


def is_domain_card(name: str) -> bool:
    """平铺命名中的体系真值卡：`<域>-<体系>-<卡>.md`（排除入口卡与 app 卡）。"""
    for domain in DOMAINS:
        prefix = domain + "-"
        if not name.startswith(prefix):
            continue
        rest = name[len(prefix):]
        if rest == "ENTRY.md" or rest.endswith("-ENTRY.md") or rest.startswith("app-"):
            return False
        segment = rest.split("-", 1)[0]
        return segment in DOMAIN_SYSTEMS[domain]
    return False


def load_duanyu_ids() -> set:
    ids: set = set()
    for path in glob.glob(ASSERTION_MANIFESTS):
        with open(path, encoding="utf-8") as handle:
            ids.update(row["assertion_id"].strip() for row in csv.DictReader(handle) if row.get("assertion_id", "").strip())
    return ids


def catalog_tool_names() -> set:
    with open(CATALOG_PATH, encoding="utf-8") as handle:
        catalog = json.load(handle)
    return {
        name
        for server in catalog.get("servers", {}).values()
        for name in server.get("tools", [])
    }


def _resolve_skill_path(path: str) -> str | None:
    """Docs live under references/; resolve against that root first, then skill root."""
    for root in (REFS, SKILL):
        candidate = os.path.join(root, path)
        if os.path.exists(candidate):
            return candidate
    return None


def _flat_cards() -> list[str]:
    return sorted(
        doc
        for doc in glob.glob(os.path.join(REFS, "*.md"))
        if os.path.basename(doc) not in ("FAQ.md",)
    )


def main() -> int:
    ids = load_duanyu_ids()
    prefixes = sorted({item.split("_")[0] + "_" for item in ids if "_" in item})
    prefix_re = re.compile(r"\b(" + "|".join(map(re.escape, prefixes)) + r")(\d+[a-z]*|x+)\b") if prefixes else None
    path_re = re.compile(r"`(" + _REF_PATH + r")`|(?<![`\w/])(" + _REF_PATH + r")")
    anchor_path_re = re.compile(r"`(" + _ANCHOR_PATH + r")`")
    bare_md_re = re.compile(_BARE_MD)
    method_re = re.compile(r"\b(" + "|".join(map(re.escape, _METHOD_PREFIXES)) + r")\.[a-z_]+\b")
    catalog_names = catalog_tool_names()
    catalog_re = re.compile(
        r"(?<![A-Za-z0-9_])(" + "|".join(map(re.escape, sorted(catalog_names))) + r")(?![A-Za-z0-9_])"
    ) if catalog_names else None
    anchor_cache: dict[str, str] = {}

    docs = [os.path.join(SKILL, "SKILL.md")]
    faq = os.path.join(REFS, "FAQ.md")
    if os.path.exists(faq):
        docs.append(faq)
    docs += [os.path.join(REFS, f"{domain}-ENTRY.md") for domain in DOMAINS]
    docs += sorted(glob.glob(os.path.join(REFS, "*-ENTRY.md")))
    docs += sorted(glob.glob(os.path.join(REFS, "*-app-*.md")))
    docs += [path for path in sorted(glob.glob(os.path.join(REFS, "*.md"))) if is_domain_card(os.path.basename(path))]
    docs = sorted({doc for doc in docs if os.path.exists(doc)})
    py_docs = sorted(glob.glob(os.path.join(SKILL, "scripts", "*.py")))

    errors: list[str] = []
    warnings: list[str] = []
    for doc in docs:
        rel = os.path.relpath(doc, _ROOT)
        text = open(doc, encoding="utf-8").read()
        if prefix_re:
            for match in prefix_re.finditer(text):
                token = match.group(0)
                if "x" in match.group(2):
                    continue
                if token not in ids:
                    errors.append(f"[{rel}] 引用断语 id '{token}' 不存在于断语表")
        for match in path_re.finditer(text):
            path = match.group(1) or match.group(2)
            if not path or "*" in path or "xxx" in path or "<" in path:
                continue
            if "/" in path[len("references/"):]:
                errors.append(f"[{rel}] 非平铺引用（须 references/<平铺名>.md）: {path}")
                continue
            if _resolve_skill_path(path) is None:
                errors.append(f"[{rel}] 引用文件不存在: {path}")
        for match in bare_md_re.finditer(text):
            token = match.group(1)
            if token in _BARE_MD_ALLOWED or "xxx" in token:
                continue
            errors.append(f"[{rel}] 裸名引用（须全路径 references/<平铺名>.md）: `{token}`")
        # 小节锚点：`x.md`「小节名」必须紧邻路径（可选空白），且小节名出现在目标文件里。
        for line in text.splitlines():
            for hit in anchor_path_re.finditer(line):
                rel_target = hit.group(1)
                resolved = _resolve_skill_path(rel_target)
                if resolved is None:
                    continue  # 文件缺失已由上面的 path 检查报错
                adjacent = re.match(r"\s*「([^」]+)」", line[hit.end():])
                if not adjacent:
                    continue
                body = anchor_cache.get(resolved)
                if body is None:
                    body = open(resolved, encoding="utf-8").read()
                    anchor_cache[resolved] = body
                anchor = adjacent.group(1)
                if anchor not in body:
                    errors.append(f"[{rel}] 引用小节不存在: {rel_target}「{anchor}」")
        for match in method_re.finditer(text):
            method = match.group(0)
            if method not in METHOD_WHITELIST:
                warnings.append(f"[{rel}] 引用的方法 '{method}' 不在方法白名单")
                continue
            errors.append(f"[{rel}] 硬编码引擎方法名 '{method}'（能力应由 tools/list 自举发现）")
        if catalog_re:
            for match in catalog_re.finditer(text):
                errors.append(f"[{rel}] 硬编码 MCP 工具名 '{match.group(0)}'（能力应由 tools/list 自举发现）")

    rpc_call_re = re.compile(r'\bcall\(\s*["\']([a-z]+\.[a-z_]+)["\']')
    for filename in py_docs:
        rel = os.path.relpath(filename, _ROOT)
        text = open(filename, encoding="utf-8").read()
        for match in rpc_call_re.finditer(text):
            if match.group(1) not in METHOD_WHITELIST:
                errors.append(f"[{rel}] RPC 调用方法 '{match.group(1)}' 不在引擎方法白名单")

    domain_files = {
        os.path.basename(path): Path(path)
        for path in _flat_cards()
        if is_domain_card(os.path.basename(path))
    }
    all_doc_text = "\n".join(open(doc, encoding="utf-8").read() for doc in docs)
    for name, path in domain_files.items():
        if name not in all_doc_text:
            errors.append(f"[{os.path.relpath(path, _ROOT)}] domain 文档未被 root/app/domain 引用")

    skill_doc = os.path.join(SKILL, "SKILL.md")
    skill_text = open(skill_doc, encoding="utf-8").read()
    if len(skill_text.splitlines()) > 120:
        errors.append(f"[skills/liki/SKILL.md] 根 SKILL.md {len(skill_text.splitlines())} 行，超过 120 行精简上限")
    if "□" in skill_text:
        errors.append("[skills/liki/SKILL.md] 根 SKILL.md 含过程检查框")

    app_cards = sorted(glob.glob(os.path.join(REFS, "*-app-*.md")))
    entry_cards = sorted(glob.glob(os.path.join(REFS, "*-ENTRY.md")))
    required_re = re.compile(r"(" + _ANCHOR_PATH + r")")
    for doc in app_cards + entry_cards:
        rel = os.path.relpath(doc, _ROOT)
        text = open(doc, encoding="utf-8").read()
        if "□" in text:
            errors.append(f"[{rel}] 文档含过程检查框；用条件/动作/产物表代替")
        if "-app-" in os.path.basename(doc):
            required_sections = {"## 流程", "## 边界条件", "## 输出模板"}
            actual_sections = {line.strip() for line in text.splitlines()}
            missing_sections = sorted(required_sections - actual_sections)
            if missing_sections:
                errors.append(f"[{rel}] App 卡缺少标准段: {', '.join(missing_sections)}")
            heading_lines = {line.strip() for line in text.splitlines()}
            legacy_sections = {
                heading for heading in (
                    "## 📖 流程", "## 📖 输出模板", "## 边界", "## 输出模板（标准奇门）"
                ) if heading in heading_lines
            }
            if legacy_sections:
                errors.append(f"[{rel}] App 卡含旧标准段: {', '.join(sorted(legacy_sections))}")
            if "## 红线（强制）" in text or "### ⚠️" in text:
                errors.append(f"[{rel}] app 卡含重复红线/警示段")
            start = text.find("## 流程")
            if start >= 0:
                end = text.find("\n## ", start + 1)
                flow_lines = len(text[start:end if end >= 0 else len(text)].splitlines())
                if flow_lines > 20:
                    errors.append(f"[{rel}] 流程区 {flow_lines} 行，超过 20 行精简上限")
            max_lines = APP_CARD_LINE_BUDGET.get(os.path.basename(doc), 65)
            if len(text.splitlines()) > max_lines:
                errors.append(f"[{rel}] app 卡 {len(text.splitlines())} 行，超过 {max_lines} 行精简上限")
        required_paths = []
        for line in text.splitlines():
            if not line.lstrip().startswith("[必读]"):
                continue
            for path in required_re.findall(line):
                name = path[len("references/"):]
                if name in domain_files and name not in required_paths:
                    required_paths.append(name)
        if len(required_paths) > 6:
            errors.append(f"[{rel}] 必读 domain 文件 {len(required_paths)} 个，超过 6 个分支上限")
        loaded_lines = sum(len(domain_files[name].read_text(encoding="utf-8").splitlines()) for name in required_paths)
        if loaded_lines > 650:
            errors.append(f"[{rel}] 必读 domain 共 {loaded_lines} 行，超过 650 行上下文预算")
        if os.path.basename(doc).endswith("-ENTRY.md"):
            # 入口路由表按行加载预算：每行引用的域卡 ≤6 个且 ≤650 行。
            for number, line in enumerate(text.splitlines(), 1):
                if not line.lstrip().startswith("|"):
                    continue
                names = [
                    path[len("references/"):]
                    for path in required_re.findall(line)
                    if path[len("references/"):] in domain_files
                ]
                if not names:
                    continue
                if len(names) > 6:
                    errors.append(f"[{rel}:{number}] 路由行引用域卡 {len(names)} 个，超过 6 个分支上限")
                line_loaded = sum(len(domain_files[name].read_text(encoding="utf-8").splitlines()) for name in names)
                if line_loaded > 650:
                    errors.append(f"[{rel}:{number}] 路由行域卡共 {line_loaded} 行，超过 650 行上下文预算")

    for doc in docs:
        text = open(doc, encoding="utf-8").read()
        if "□" in text:
            errors.append(f"[{os.path.relpath(doc, _ROOT)}] 文档含过程检查框")
        if re.search(r"\bStep\s+\d+(?:\.\d+)?\b", text):
            errors.append(f"[{os.path.relpath(doc, _ROOT)}] 文档含旧流程步骤编号")
        rel = os.path.relpath(doc, _ROOT)
        if rel != os.path.join("skills", "liki", "SKILL.md") and text.startswith("---\n"):
            errors.append(f"[{rel}] 方法论卡带 frontmatter（无消费方）：元数据改放正文或 ENTRY 路由表")
        for match in re.finditer(r"`data/[A-Za-z0-9_./-]+\.(?:json|csv)`", text):
            errors.append(f"[{rel}] 引用实现内部数据文件 {match.group(0)}：改引工具字段名")

    readme = os.path.join(_ROOT, "README.md")
    assertion_path = os.path.join(_ROOT, "counsel", "app", "natal", "tools", "assertions", "assertions.csv")
    if os.path.exists(readme) and os.path.exists(assertion_path):
        actual = sum(1 for row in csv.DictReader(open(assertion_path, encoding="utf-8")) if row.get("assertion_id"))
        match = re.search(r"(\d+)\s*条断语", open(readme, encoding="utf-8").read())
        if match and int(match.group(1)) != actual:
            errors.append(f"[README] 断语统计 {match.group(1)} ≠ 实际 {actual}")

    print(f"扫描文档 {len(docs)} 个（断语 id 全集 {len(ids)}）")
    print(f"错误: {len(errors)} 个")
    for error in errors:
        print("  ✗", error)
    print(f"警告: {len(warnings)} 个")
    for warning in warnings[:30]:
        print("  ⚠", warning)
    if len(warnings) > 30:
        print("  ... 共", len(warnings), "条警告")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
