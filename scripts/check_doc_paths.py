#!/usr/bin/env python3
"""Reject Markdown references to repository paths that do not exist.

Two reference forms are checked across every Markdown file outside
``CHANGELOG.md`` (a historical log, where past paths are legitimately gone):

* link targets -- ``[text](path)`` / ``![alt](path)``, the standard
  markdown-link-check surface;
* inline code spans anchored at a repository top-level directory
  (``docs/RELEASE_MODEL.md``), which is the form this repository uses to point
  at live files.

A bare relative span such as ``TOOLS.md`` or ``app/mingshu.md`` is prose, not a
pointer, so it is left alone: those short names resolve against several
different roots (skill root, references root) and cannot be checked without a
per-file rule. Targets are accepted if they resolve against the Markdown file's
own directory or against the repository root.

Division of labour with ``check_docs.py``: that checker owns skill-document
semantics (assertion ids, section budgets, method whitelist) inside ``skills/liki``;
this one owns the path surface of every other Markdown file.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SKIP_DIRS = {
    ".git", ".venv", ".agents", "__pycache__", "node_modules",
    "dist", "bin", "build", "htmlcov", ".mypy_cache", ".pytest_cache", ".ruff_cache",
}
SKIP_FILES = {"CHANGELOG.md"}
#: MCP JSON-RPC method names look like ``namespace/verb``; they are not paths.
MCP_METHODS = frozenset({
    "tools/list", "tools/call", "tools/list_changed",
    "server/discover", "notifications/cancelled", "notifications/initialized",
})
_LINK = re.compile(r"!?\[[^\]]*\]\(\s*([^)\s]+?)\s*\)")
_CODE = re.compile(r"`([^`\n]+)`")
_IGNORED_SCHEMES = ("http://", "https://", "mailto:", "tel:", "#")
_TOP_LEVEL = {
    path.name
    for path in ROOT.iterdir()
    if path.is_dir() and not path.name.startswith(".") and path.name not in SKIP_DIRS
}


def _markdown_files() -> list[Path]:
    files = []
    for path in sorted(ROOT.rglob("*.md")):
        rel = path.relative_to(ROOT)
        if rel.name in SKIP_FILES or rel.parts[0] in SKIP_DIRS:
            continue
        if path.is_file() and not path.is_symlink():
            files.append(path)
    return files


def _resolves(source: Path, target: str) -> bool:
    if target.startswith("./"):
        target = target[2:]
    if not target or target.startswith("/"):
        return True
    return (source.parent / target).exists() or (ROOT / target).exists()


def _code_path(span: str) -> str | None:
    span = span.strip()
    if "/" not in span or span in MCP_METHODS or span.endswith("/"):
        return None
    if any(char in span for char in "*<>{} $`\\"):
        return None
    if span.split("/", 1)[0] not in _TOP_LEVEL:
        return None
    return span


def main() -> int:
    errors: list[str] = []
    files = _markdown_files()
    for path in files:
        rel = path.relative_to(ROOT)
        text = path.read_text(encoding="utf-8")
        for match in _LINK.finditer(text):
            target = match.group(1)
            if target.startswith(_IGNORED_SCHEMES):
                continue
            if not _resolves(path, target.split("#", 1)[0]):
                errors.append(f"[{rel}] 链接目标不存在: {target}")
        for match in _CODE.finditer(text):
            span = _code_path(match.group(1))
            if span and not _resolves(path, span):
                errors.append(f"[{rel}] 引用的仓库路径不存在: {span}")

    print(f"扫描 Markdown 文件 {len(files)} 个（跳过 {', '.join(sorted(SKIP_FILES))}）")
    for error in errors:
        print("  ✗", error)
    print(f"错误: {len(errors)} 个")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
