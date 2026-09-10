"""Prompt-layer interaction gates must be ordered, blocking, and self-contained."""

from pathlib import Path
import re


ROOT = Path(__file__).resolve().parents[1]

EXPECTED_NODES = {
    "liki-naming/app/foreign.md": [
        ("0a", "⛔ 阻塞确认"),
        ("0b", "💬 参数收集"),
        ("5a", "⛔ 阻塞确认"),
    ],
    "liki-naming/app/naming.md": [
        ("2", "💬 参数收集"),
        ("6a", "⛔ 阻塞确认"),
    ],
    "liki-bazi/app/mingshu.md": [
        ("2a", "⛔ 阻塞确认"),
    ],
    "liki-bazi/app/mingshu-full.md": [
        ("2a", "⛔ 阻塞确认"),
        ("6a", "⛔ 阻塞确认"),
    ],
    "liki-divination/app/question.md": [
        ("1a", "⛔ 阻塞确认"),
    ],
    "liki-fengshui/app/fengshui.md": [
        ("1a", "⛔ 阻塞确认"),
        ("1b", "💬 参数收集"),
        ("5a", "⛔ 阻塞确认"),
    ],
}


def _flow_text(path: Path) -> tuple[str, list[tuple[str, str]]]:
    text = path.read_text(encoding="utf-8")
    match = re.search(r"^## (?:📖 )?流程(?:$| )|^## LLM 路由流程$", text, re.MULTILINE)
    assert match, path
    start = match.start()
    assert start >= 0, path
    next_match = re.search(r"^## (?!.*流程).*$", text[start + 1:], re.MULTILINE)
    if next_match:
        end = start + 1 + next_match.start()
    else:
        end = len(text)
    flow = text[start:end]
    rows = []
    for line in flow.splitlines():
        if not line.startswith("| ") or line.startswith("| 步骤 "):
            continue
        cells = [cell.strip() for cell in line.strip("|").split("|")]
        if len(cells) >= 2:
            rows.append((cells[0], cells[1]))
    return flow, rows


def test_interaction_nodes_are_ordered_and_limited():
    for relative, expected in EXPECTED_NODES.items():
        path = ROOT / "skills" / relative
        flow, rows = _flow_text(path)
        text = path.read_text(encoding="utf-8")
        actual = [
            (step, marker)
            for step, marker in rows
            if marker.startswith("⛔") or marker.startswith("💬")
        ]
        assert actual == expected, (relative, actual, expected)

        blockers = [step for step, marker in actual if marker.startswith("⛔")]
        assert len(blockers) <= 2, (relative, blockers)

        for step, marker in expected:
            node_lines = [
                line for line in flow.splitlines()
                if line.startswith(f"| {step} | {marker}")
            ]
            assert node_lines, (relative, step)
            if marker.startswith("⛔"):
                assert any(
                    "停在这里" in line and line.lstrip("- ").startswith(f"{step} ⛔")
                    for line in text.splitlines()
                ), (relative, step)
            else:
                assert "都用默认" in text, (relative, step)


def test_interaction_templates_have_options_and_boundaries():
    expected_options = {
        "liki-naming/app/foreign.md": 2,
        "liki-naming/app/naming.md": 2,
        "liki-bazi/app/mingshu.md": 2,
        "liki-bazi/app/mingshu-full.md": 2,
        "liki-divination/app/question.md": 2,
        "liki-fengshui/app/fengshui.md": 2,
    }
    for relative, minimum_options in expected_options.items():
        text = (ROOT / "skills" / relative).read_text(encoding="utf-8")
        option_count = sum(
            text.count(f"{number}.") for number in range(1, minimum_options + 1)
        )
        assert option_count >= minimum_options, (relative, option_count)
        assert "禁止提前输出后续步骤" not in text or "不要输出" in text or "不要展开" in text


def test_root_skills_define_standard_interaction_contract():
    for skill in ("liki-naming", "liki-bazi", "liki-divination", "liki-fengshui"):
        text = (ROOT / "skills" / skill / "SKILL.md").read_text(encoding="utf-8")
        assert "流程表中标记 ⛔ 的步骤为阻塞确认" in text
        assert "等待用户回复后才能继续" in text
        assert "流程表中标记 💬 的步骤为参数收集" in text
        assert "都用默认" in text
        assert "禁止提前输出后续步骤的结果" in text
