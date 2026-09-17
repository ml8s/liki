"""Unified Liki skill document contract checker.

Checks assertion ids, resolvable skill-relative paths, whitelisted RPC names,
lazy-loading budgets, and reachability of domain knowledge docs.
"""
from __future__ import annotations

import csv
import glob
import os
import re
import sys
from pathlib import Path

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SKILL = sys.argv[1] if len(sys.argv) > 1 else os.path.join(_ROOT, "skills", "liki")
DOMAINS = ("bazi", "divination", "fengshui", "naming")

METHOD_WHITELIST = {
    "rpc.discover",
    "bazi.chart", "bazi.fullchart", "bazi.liunian", "bazi.liuyue", "bazi.liuri",
    "bazi.liushi", "bazi.xiaoyun", "bazi.bond",
    "ziwei.chart", "ziwei.fullchart", "ziwei.liunian", "ziwei.liuyue", "ziwei.liuri",
    "ziwei.liushi", "ziwei.daxian", "ziwei.bond",
    "liuyao.qigua", "liuyao.chart", "qimen.chart", "huangli.days",
    "bazhai.chart", "bazhai.layout", "xuankong.chart", "xuankong.liunian",
    "qiming.surname", "qiming.pick", "qiming.compose", "qiming.check", "qiming.char",
    "time.now", "city.coords", "tianwen.time",
}
_METHOD_PREFIXES = tuple(sorted({m.split(".")[0] for m in METHOD_WHITELIST}))
_SKIP_DOTTED = {"params.properties", "result.methods", "result.info", "result.info.version"}
_PATH_PATTERN = r"(?:(?:bazi|divination|fengshui|naming)/(?:tools|app|domains)|webapp)/"


def load_duanyu_ids() -> set:
    ids: set = set()
    for path in glob.glob(os.path.join(SKILL, "*", "tools", "assertions", "assertions.csv")):
        with open(path, encoding="utf-8") as handle:
            ids.update(row["assertion_id"].strip() for row in csv.DictReader(handle) if row.get("assertion_id", "").strip())
    return ids


def main() -> int:
    ids = load_duanyu_ids()
    prefixes = sorted({item.split("_")[0] + "_" for item in ids if "_" in ids})
    prefix_re = re.compile(r"\b(" + "|".join(map(re.escape, prefixes)) + r")(\d+[a-z]*|x+)\b") if prefixes else None
    path_re = re.compile(r"`(" + _PATH_PATTERN + r"[^`]+)`|(" + _PATH_PATTERN + r"[\w./\-]+\.(?:md|py|json|csv|sh|cmd))")
    method_re = re.compile(r"\b(" + "|".join(map(re.escape, _METHOD_PREFIXES)) + r")\.[a-z_]+\b")

    docs = [os.path.join(SKILL, "SKILL.md")]
    faq = os.path.join(SKILL, "FAQ.md")
    if os.path.exists(faq):
        docs.append(faq)
    docs += [os.path.join(SKILL, domain, "ENTRY.md") for domain in DOMAINS]
    docs += sorted(glob.glob(os.path.join(SKILL, "*", "app", "*.md")))
    docs += sorted(glob.glob(os.path.join(SKILL, "*", "domains", "**", "*.md"), recursive=True))
    docs = [doc for doc in docs if os.path.exists(doc)]
    py_docs = sorted(glob.glob(os.path.join(SKILL, "*", "tools", "*.py")))

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
            if not os.path.exists(os.path.join(SKILL, path)):
                errors.append(f"[{rel}] 引用文件不存在: {path}")
        for match in method_re.finditer(text):
            method = match.group(0)
            if method in _SKIP_DOTTED or method not in METHOD_WHITELIST:
                warnings.append(f"[{rel}] 引用的方法 '{method}' 不在方法白名单")

    rpc_call_re = re.compile(r'\bcall\(\s*["\']([a-z]+\.[a-z_]+)["\']')
    for filename in py_docs:
        rel = os.path.relpath(filename, _ROOT)
        text = open(filename, encoding="utf-8").read()
        for match in rpc_call_re.finditer(text):
            if match.group(1) not in METHOD_WHITELIST:
                errors.append(f"[{rel}] RPC 调用方法 '{match.group(1)}' 不在引擎方法白名单")

    domain_files = {
        os.path.relpath(path, SKILL): Path(path)
        for path in glob.glob(os.path.join(SKILL, "*", "domains", "**", "*.md"), recursive=True)
    }
    all_doc_text = "\n".join(open(doc, encoding="utf-8").read() for doc in docs)
    for relpath, path in domain_files.items():
        if path.name not in all_doc_text and relpath not in all_doc_text:
            errors.append(f"[{os.path.relpath(path, _ROOT)}] domain 文档未被 root/app/domain 引用")

    skill_doc = os.path.join(SKILL, "SKILL.md")
    skill_text = open(skill_doc, encoding="utf-8").read()
    if len(skill_text.splitlines()) > 120:
        errors.append(f"[skills/liki/SKILL.md] 根 SKILL.md {len(skill_text.splitlines())} 行，超过 120 行精简上限")
    if "□" in skill_text:
        errors.append("[skills/liki/SKILL.md] 根 SKILL.md 含过程检查框")

    required_re = re.compile(r"((?:bazi|divination|fengshui|naming)/(?:domains|app|tools)/[\w./\-]+\.md)")
    for pattern in ("*/app/*.md", "*/ENTRY.md"):
        for doc in sorted(glob.glob(os.path.join(SKILL, pattern))):
            rel = os.path.relpath(doc, _ROOT)
            text = open(doc, encoding="utf-8").read()
            app_readme = pattern == "*/app/*.md" and Path(doc).name == "README.md"
            if "□" in text:
                errors.append(f"[{rel}] 文档含过程检查框；用条件/动作/产物表代替")
            if pattern == "*/app/*.md" and not app_readme:
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
            if pattern == "*/app/*.md":
                if "## 红线（强制）" in text or "### ⚠️" in text:
                    errors.append(f"[{rel}] app 卡含重复红线/警示段")
                start = text.find("## 流程")
                if start >= 0:
                    end = text.find("\n## ", start + 1)
                    flow_lines = len(text[start:end if end >= 0 else len(text)].splitlines())
                    if flow_lines > 20:
                        errors.append(f"[{rel}] 流程区 {flow_lines} 行，超过 20 行精简上限")
                max_lines = 75 if Path(doc).name == "mingshu-full.md" else 65
                if len(text.splitlines()) > max_lines:
                    errors.append(f"[{rel}] app 卡 {len(text.splitlines())} 行，超过 {max_lines} 行精简上限")
            required_paths = []
            for line in text.splitlines():
                if not line.lstrip().startswith("[必读]"):
                    continue
                for path in required_re.findall(line):
                    if path in domain_files and path not in required_paths:
                        required_paths.append(path)
            if len(required_paths) > 6:
                errors.append(f"[{rel}] 必读 domain 文件 {len(required_paths)} 个，超过 6 个分支上限")
            loaded_lines = sum(len(domain_files[path].read_text(encoding="utf-8").splitlines()) for path in required_paths)
            if loaded_lines > 650:
                errors.append(f"[{rel}] 必读 domain 共 {loaded_lines} 行，超过 650 行上下文预算")

    for doc in docs:
        text = open(doc, encoding="utf-8").read()
        if "□" in text:
            errors.append(f"[{os.path.relpath(doc, _ROOT)}] 文档含过程检查框")
        if re.search(r"\bStep\s+\d+(?:\.\d+)?\b", text):
            errors.append(f"[{os.path.relpath(doc, _ROOT)}] 文档含旧流程步骤编号")

    readme = os.path.join(_ROOT, "README.md")
    if os.path.exists(readme):
        assertion_path = os.path.join(SKILL, "bazi", "tools", "assertions", "assertions.csv")
        if os.path.exists(assertion_path):
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
