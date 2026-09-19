"""Skill markdown 结构契约：heading 层级、代码块闭合、裸粗体标题、HTML 白名单等。

覆盖 skills/liki 下全部 .md 文件，自动化 TRACE Convention 子项中可规则化的部分。
"""
from __future__ import annotations

import re
from pathlib import Path

import pytest

SKILL_ROOT = Path(__file__).resolve().parents[1] / "skills/liki"

# 所有 .md 文件
ALL_MD = sorted(p for p in SKILL_ROOT.rglob("*.md") if "README" not in p.name)

# App 卡有 frontmatter，不需要 H1；其余文件应有且仅有 1 个 H1
APP_CARDS = {
    p for p in ALL_MD
    if p.parent.name == "app" and p.name != "README.md"
}
INDEX_FILES = {
    p for p in ALL_MD
    if p.name in ("README.md",) and p.parent.name == "app"
}

# 最大 heading 层级
MAX_HEADING_LEVEL = 4

# 允许的 HTML 标签（README banner 区域用到的）
ALLOWED_HTML = {"h1", "p", "br", "a", "img"}


def _prose_lines(text: str) -> list[str]:
    """去掉 fenced code block 内容后返回 prose 行。"""
    lines: list[str] = []
    fence: str | None = None
    for line in text.splitlines():
        marker = re.match(r"^\s*(`{3,}|~{3,})", line)
        if marker:
            token = marker.group(1)
            if fence is None:
                fence = token
            elif token[0] == fence[0] and len(token) >= len(fence):
                fence = None
            continue
        if fence is None:
            lines.append(line)
    return lines


def _prose_text(text: str) -> str:
    return "\n".join(_prose_lines(text))


def _headings(text: str) -> list[tuple[int, str]]:
    prose = _prose_text(text)
    return [
        (len(m.group(1)), m.group(2).strip())
        for m in re.finditer(r"^(#{1,6})\s+(.+?)\s*$", prose, flags=re.MULTILINE)
    ]


def _rel(path: Path) -> str:
    return str(path.relative_to(SKILL_ROOT))


def _has_frontmatter(path: Path) -> bool:
    return path.read_text(encoding="utf-8").startswith("---\n")


# ── 单一 H1 ────────────────────────────────────────────────────────────

@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_h1_count(path: Path):
    text = path.read_text(encoding="utf-8")
    h1s = [t for level, t in _headings(text) if level == 1]
    if path in APP_CARDS and _has_frontmatter(path):
        # App 卡 frontmatter 的 description 顶替 H1，不要求 body 有 H1
        return
    if path in INDEX_FILES:
        return  # index 文件允许 0 个 H1
    assert len(h1s) == 1, f"{_rel(path)}: H1 count = {len(h1s)}, expected 1"


# ── heading 层级不超限 ─────────────────────────────────────────────────

@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_max_heading_level(path: Path):
    text = path.read_text(encoding="utf-8")
    for level, title in _headings(text):
        assert level <= MAX_HEADING_LEVEL, f"{_rel(path)}: H{level} '{title}' 超过最大层级 {MAX_HEADING_LEVEL}"


# ── 无空 heading ──────────────────────────────────────────────────────

@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_no_empty_heading(path: Path):
    prose = _prose_text(path.read_text(encoding="utf-8"))
    for m in re.finditer(r"^#{1,6}\s*$", prose, flags=re.MULTILINE):
        pytest.fail(f"{_rel(path)}: 空 heading 在行 {prose[:m.start()].count(chr(10)) + 1}")


# ── 无裸粗体标题 ──────────────────────────────────────────────────────

@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_no_standalone_bold_heading(path: Path):
    for line in _prose_lines(path.read_text(encoding="utf-8")):
        stripped = line.strip()
        if re.fullmatch(r"\*\*[^*]+\*\*[:：]?", stripped):
            pytest.fail(f"{_rel(path)}: 裸粗体标题 '{stripped}'，请改用 heading")


# ── 代码块闭合 ────────────────────────────────────────────────────────

@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_code_fences_closed(path: Path):
    text = path.read_text(encoding="utf-8")
    fence: str | None = None
    for i, line in enumerate(text.splitlines(), 1):
        marker = re.match(r"^\s*(`{3,}|~{3,})", line)
        if marker:
            token = marker.group(1)
            if fence is None:
                fence = token
            elif token[0] == fence[0] and len(token) >= len(fence):
                fence = None
    assert fence is None, f"{_rel(path)}: 未闭合的 code fence '{fence}'"


# ── HTML 标签白名单 ───────────────────────────────────────────────────

def _html_tags(text: str) -> set[str]:
    prose = _prose_text(text)
    return {
        m.group(2).lower()
        for m in re.finditer(r"<(/?)([a-zA-Z][a-zA-Z0-9-]*)\b[^>]*>", prose)
    }


@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_no_raw_html_outside_whitelist(path: Path):
    tags = _html_tags(path.read_text(encoding="utf-8"))
    bad = tags - ALLOWED_HTML
    assert not bad, f"{_rel(path)}: 不允许的 HTML 标签 {bad}"


# ── 无乱码 ────────────────────────────────────────────────────────────

@pytest.mark.parametrize("path", ALL_MD, ids=_rel)
def test_no_mojibake(path: Path):
    text = path.read_text(encoding="utf-8")
    assert "\ufffd" not in text, f"{_rel(path)}: 含 Unicode 替换字符（乱码）"
