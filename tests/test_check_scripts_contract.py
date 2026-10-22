"""`make check` must not run an undocumented script.

`make check` is the static half of the push gate. Its script set is listed in
`tests/README.md` ("Static check scripts"), and both directions are pinned here:
a script wired into the recipe but missing from the table, or a table row left
behind after the recipe stopped calling it, would quietly shrink the gate.
"""

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MAKEFILE = ROOT / "Makefile"
README = ROOT / "tests" / "README.md"
SECTION = "## Static check scripts"

_SCRIPT = re.compile(r"scripts/[A-Za-z0-9_.-]+\.(?:py|sh)")


def _target(name: str) -> re.Match[str]:
    match = re.search(
        rf"^{re.escape(name)}:(?P<prereq>[^\n]*)\n(?P<body>(?:\t.*\n)+)", MAKEFILE.read_text(encoding="utf-8"), re.M
    )
    assert match, f"Makefile target {name!r} missing"
    return match


def _referenced(name: str) -> set[str]:
    """Scripts referenced by a target's recipe plus its direct prerequisites."""
    makefile = MAKEFILE.read_text(encoding="utf-8")
    target = _target(name)
    scripts = set(_SCRIPT.findall(target.group("body")))
    for prereq in target.group("prereq").split():
        body = re.search(
            rf"^{re.escape(prereq)}:[^\n]*\n((?:\t.*\n)+)", makefile, re.M
        )
        if body:
            scripts |= set(_SCRIPT.findall(body.group(1)))
    return scripts


def _documented() -> set[str]:
    text = README.read_text(encoding="utf-8")
    assert SECTION in text, f"tests/README.md must document {SECTION!r}"
    section = text.split(SECTION, 1)[1].split("\n## ", 1)[0]
    return set(_SCRIPT.findall(section))


def test_documented_check_scripts_match_the_makefile():
    wired = _referenced("check")
    documented = _documented()
    assert wired == documented, {
        "undocumented": sorted(wired - documented),
        "unwired": sorted(documented - wired),
    }


def test_every_documented_check_script_exists():
    for script in sorted(_documented()):
        assert (ROOT / script).is_file(), script


def test_documented_shell_scripts_are_executable():
    for script in sorted(_documented()):
        if script.endswith(".sh"):
            assert (ROOT / script).stat().st_mode & 0o111, script
