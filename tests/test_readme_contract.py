"""README structure contract: one entry document per language, fixed headings, no pseudo headings."""
from __future__ import annotations

import re
from pathlib import Path
from urllib.parse import unquote, urlparse

import pytest

ROOT = Path(__file__).resolve().parents[1]

CONTRACTS = {
    ROOT / "README.md": {
        "h1": "Liki",
        "h2": [
            "安装",
            "快速开始",
            "能力概览",
            "可信度与边界",
            "常见问题",
            "文档",
            "开发者",
            "架构",
            "贡献",
            "许可与声明",
        ],
    },
    ROOT / "README.en.md": {
        "h1": "Liki",
        "h2": [
            "Install",
            "Quick start",
            "What you get",
            "Trust and boundaries",
            "FAQ",
            "Documentation",
            "For developers",
            "Architecture",
            "Contributing",
            "License and disclaimer",
        ],
    },
}

GUIDES = {
    ROOT / "docs/USER_GUIDE.md": {
        "h1": "Liki 用户指南",
        "h2": [
            "开始前准备",
            "第一次提问",
            "对话方式",
            "命理与紫微",
            "起名与改名",
            "问卦",
            "风水",
            "常见问题",
            "输出原则",
        ],
    },
    ROOT / "docs/USER_GUIDE.en.md": {
        "h1": "Liki user guide",
        "h2": [
            "Before you start",
            "First question",
            "Conversation style",
            "Destiny and Ziwei",
            "Naming",
            "Divination",
            "Feng shui",
            "FAQ",
            "Output principles",
        ],
    },
}


def markdown_prose_lines(text: str) -> list[str]:
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


def markdown_headings(text: str) -> list[tuple[int, str]]:
    prose = "\n".join(markdown_prose_lines(text))
    return [
        (len(match.group(1)), match.group(2).strip())
        for match in re.finditer(r"^(#{1,6})\s+(.+?)\s*$", prose, flags=re.MULTILINE)
    ]


def html_h1_values(text: str) -> list[str]:
    return [
        re.sub(r"<[^>]+>", "", match.group(1)).strip()
        for match in re.finditer(r"<h1\b[^>]*>(.*?)</h1>", text, flags=re.IGNORECASE | re.DOTALL)
    ]


@pytest.mark.parametrize("path", CONTRACTS, ids=lambda path: path.name)
def test_readme_has_single_h1(path: Path):
    text = path.read_text(encoding="utf-8")
    markdown_h1 = [title for level, title in markdown_headings(text) if level == 1]
    html_h1 = html_h1_values(text)
    assert len(markdown_h1 + html_h1) == 1
    assert (markdown_h1 + html_h1)[0] == CONTRACTS[path]["h1"]


@pytest.mark.parametrize("path", CONTRACTS, ids=lambda path: path.name)
def test_readme_h2_sequence_is_canonical(path: Path):
    text = path.read_text(encoding="utf-8")
    h2 = [title for level, title in markdown_headings(text) if level == 2]
    assert h2 == CONTRACTS[path]["h2"]


@pytest.mark.parametrize("path", CONTRACTS, ids=lambda path: path.name)
def test_readme_does_not_use_h4_or_deeper(path: Path):
    text = path.read_text(encoding="utf-8")
    assert all(level <= 3 for level, _ in markdown_headings(text))


@pytest.mark.parametrize(
    "path", [*CONTRACTS, *GUIDES], ids=lambda path: path.name
)
def test_has_no_standalone_bold_heading(path: Path):
    for line in markdown_prose_lines(path.read_text(encoding="utf-8")):
        assert not re.fullmatch(r"\*\*[^*]+\*\*[:：]?", line.strip()), line


def html_tag_names(text: str) -> set[str]:
    prose = "\n".join(markdown_prose_lines(text))
    return {
        match.group(2).lower()
        for match in re.finditer(r"<(/?)([a-zA-Z][a-zA-Z0-9-]*)\b[^>]*>", prose)
    }


@pytest.mark.parametrize("path", GUIDES, ids=lambda path: path.name)
def test_user_guides_do_not_use_raw_html(path: Path):
    assert not html_tag_names(path.read_text(encoding="utf-8"))


@pytest.mark.parametrize("path", CONTRACTS, ids=lambda path: path.name)
def test_readme_html_is_limited_to_centered_banner(path: Path):
    lines = markdown_prose_lines(path.read_text(encoding="utf-8"))
    first_h2 = next((index for index, line in enumerate(lines) if line.startswith("## ")), len(lines))
    html_lines = [index for index, line in enumerate(lines) if re.search(r"<[a-zA-Z][^>]*>", line)]
    assert html_lines
    assert all(index < first_h2 for index in html_lines)
    assert html_tag_names(path.read_text(encoding="utf-8")) <= {"h1", "p", "br", "a", "img"}


@pytest.mark.parametrize("path", CONTRACTS, ids=lambda path: path.name)
def test_readme_stays_a_quick_entry_document(path: Path):
    assert len(path.read_text(encoding="utf-8").splitlines()) < 180


@pytest.mark.parametrize("path", GUIDES, ids=lambda path: path.name)
def test_user_guide_h2_sequence_is_canonical(path: Path):
    text = path.read_text(encoding="utf-8")
    h1 = [title for level, title in markdown_headings(text) if level == 1]
    h2 = [title for level, title in markdown_headings(text) if level == 2]
    assert h1 == [GUIDES[path]["h1"]]
    assert h2 == GUIDES[path]["h2"]


def document_relative_links(path: Path) -> list[Path]:
    text = path.read_text(encoding="utf-8")
    links = re.findall(r"\]\(([^)]+)\)", text)
    links += re.findall(r"<a\b[^>]*\bhref=[\"']([^\"']+)[\"']", text, flags=re.IGNORECASE)
    return [
        (path.parent / unquote(urlparse(link).path)).resolve()
        for link in links
        if not urlparse(link).scheme and not link.startswith("#")
    ]


@pytest.mark.parametrize(
    "path",
    [
        ROOT / "README.md",
        ROOT / "README.en.md",
        ROOT / "docs/USER_GUIDE.md",
        ROOT / "docs/USER_GUIDE.en.md",
    ],
    ids=lambda path: path.name,
)
def test_local_links_exist(path: Path):
    missing = [link for link in document_relative_links(path) if not link.is_file()]
    assert not missing


@pytest.mark.parametrize(
    ("readme", "guide"),
    [
        (ROOT / "README.md", "docs/USER_GUIDE.md"),
        (ROOT / "README.en.md", "docs/USER_GUIDE.en.md"),
    ],
    ids=["zh", "en"],
)
def test_readme_links_to_governance_documents(readme: Path, guide: str):
    expected = {
        guide,
        "CONTRIBUTING.md",
        "CHANGELOG.md",
    }
    targets = {link.relative_to(ROOT).as_posix() for link in document_relative_links(readme)}
    assert expected <= targets
