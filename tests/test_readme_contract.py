"""Public documentation structure and wiring contract.

This intentionally avoids pinning marketing titles, section wording, or SEO
copy.  It only checks portable properties: one entry heading, bounded
structure, allowed HTML, sane document size, and working local wiring.
"""

import re
from pathlib import Path
from urllib.parse import unquote, urlparse

import pytest

ROOT = Path(__file__).resolve().parents[1]

README_PATHS = (
    ROOT / "README.md",
    ROOT / "README.en.md",
)
GUIDE_PATHS = (
    ROOT / "docs/USER_GUIDE.md",
    ROOT / "docs/USER_GUIDE.en.md",
)
PUBLIC_DOC_PATHS = (*README_PATHS, *GUIDE_PATHS)


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
        for match in re.finditer(
            r"<h1\b[^>]*>(.*?)</h1>", text, flags=re.IGNORECASE | re.DOTALL
        )
    ]


def html_tag_names(text: str) -> set[str]:
    prose = "\n".join(markdown_prose_lines(text))
    return {
        match.group(2).lower()
        for match in re.finditer(r"<(/?)([a-zA-Z][a-zA-Z0-9-]*)\b[^>]*>", prose)
    }


@pytest.mark.parametrize("path", README_PATHS, ids=lambda path: path.name)
def test_readme_has_one_entry_heading(path: Path):
    values = [
        title
        for level, title in markdown_headings(path.read_text(encoding="utf-8"))
        if level == 1
    ] + html_h1_values(path.read_text(encoding="utf-8"))
    assert len(values) == 1


@pytest.mark.parametrize("path", PUBLIC_DOC_PATHS, ids=lambda path: path.name)
def test_public_doc_heading_levels_are_bounded(path: Path):
    assert all(level <= 3 for level, _ in markdown_headings(path.read_text(encoding="utf-8")))


@pytest.mark.parametrize("path", PUBLIC_DOC_PATHS, ids=lambda path: path.name)
def test_public_doc_has_no_standalone_bold_heading(path: Path):
    for line in markdown_prose_lines(path.read_text(encoding="utf-8")):
        assert not re.fullmatch(r"\*\*[^*]+\*\*[:：]?", line.strip()), line


@pytest.mark.parametrize("path", GUIDE_PATHS, ids=lambda path: path.name)
def test_user_guides_do_not_use_raw_html(path: Path):
    assert not html_tag_names(path.read_text(encoding="utf-8"))


@pytest.mark.parametrize("path", README_PATHS, ids=lambda path: path.name)
def test_readme_html_tags_are_limited(path: Path):
    assert html_tag_names(path.read_text(encoding="utf-8")) <= {"h1", "p", "br", "a", "img"}


@pytest.mark.parametrize("path", README_PATHS, ids=lambda path: path.name)
def test_readme_stays_a_quick_entry_document(path: Path):
    assert len(path.read_text(encoding="utf-8").splitlines()) < 180


@pytest.mark.parametrize("path", GUIDE_PATHS, ids=lambda path: path.name)
def test_user_guide_has_one_entry_heading(path: Path):
    h1 = [level for level, _ in markdown_headings(path.read_text(encoding="utf-8")) if level == 1]
    assert h1 == [1]


def document_relative_links(path: Path) -> list[Path]:
    text = path.read_text(encoding="utf-8")
    links = re.findall(r"\]\(([^)]+)\)", text)
    links += re.findall(r"<a\b[^>]*\bhref=[\"']([^\"']+)[\"']", text, flags=re.IGNORECASE)
    return [
        (path.parent / unquote(urlparse(link).path)).resolve()
        for link in links
        if not urlparse(link).scheme and not link.startswith("#")
    ]


@pytest.mark.parametrize("path", PUBLIC_DOC_PATHS, ids=lambda path: path.name)
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
    expected = {guide, "CONTRIBUTING.md", "CHANGELOG.md"}
    targets = {
        link.relative_to(ROOT).as_posix()
        for link in document_relative_links(readme)
        if link.is_relative_to(ROOT)
    }
    assert expected <= targets
