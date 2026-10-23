"""守卫原语单元测试 + 防漂移扫描。

消息字符串是契约：这里的 exact-match 断言防止措辞漂移；源码扫描防止
内联守卫回潮（守卫必须住在 app/guards.py 单点）。
"""

from pathlib import Path

import pytest
from app.guards import (
    require_exactly_one,
    require_message,
    require_mutually_exclusive,
    require_object,
    require_question,
)

TOOLS = Path(__file__).resolve().parents[1] / "app" / "divination" / "tools"
APP = Path(__file__).resolve().parents[1] / "app"


def test_require_question_strips_and_accepts():
    assert require_question("  项目能不能成？  ") == "项目能不能成？"


@pytest.mark.parametrize("bad", ["", "   ", None, 42])
def test_require_question_rejects_empty_or_non_str(bad):
    with pytest.raises(ValueError, match="question must be non-empty text"):
        require_question(bad)


def test_require_question_rejects_single_char():
    with pytest.raises(ValueError, match="question must be at least 2 characters"):
        require_question(" 丢 ")


def test_require_message_canonical_messages():
    assert require_message("  继续推进  ") == "继续推进"
    with pytest.raises(ValueError, match="message must be non-empty text"):
        require_message(None)
    with pytest.raises(ValueError, match="message must be non-empty text"):
        require_message(" ")
    with pytest.raises(ValueError, match="message must be at least 2 characters"):
        require_message("冲")


def test_require_exactly_one():
    require_exactly_one(matter="lost_item", yong_shen=None)
    require_exactly_one(matter=None, yong_shen=["开门"])
    with pytest.raises(ValueError, match="provide exactly one of matter or yong_shen"):
        require_exactly_one(matter=None, yong_shen=None)
    with pytest.raises(ValueError, match="provide exactly one of matter or yong_shen"):
        require_exactly_one(matter="lost_item", yong_shen=["开门"])


def test_require_mutually_exclusive():
    require_mutually_exclusive(matter="wealth", yong_shen=None)
    require_mutually_exclusive(matter=None, yong_shen=["开门"])
    with pytest.raises(ValueError, match="matter and yong_shen are mutually exclusive"):
        require_mutually_exclusive(matter="wealth", yong_shen=["开门"])


def test_require_object():
    require_object({"mode": "auto"}, "casting")
    with pytest.raises(ValueError, match="casting must be an object"):
        require_object("auto", "casting")


def test_no_inline_duplicates_in_tools():
    """内联守卫必须绝迹：这些模式只允许出现在 app/guards.py。"""
    offenders: list[str] = []
    for path in APP.rglob("*.py"):
        if path.name == "guards.py" or "test" in path.name:
            continue
        text = path.read_text(encoding="utf-8")
        for pattern in (
            "matter 与 yong_shen 互斥",
            "provide exactly one of matter or yong_shen",
            "question must be non-empty text",
            "message must be non-empty text",
            "question must be at least 2 characters",
        ):
            if pattern in text:
                offenders.append(f"{path.relative_to(APP)}: {pattern}")
    assert offenders == []


def test_tools_delegate_to_guards():
    expected = {
        "liuyao_snapshot.py": "require_question",
        "qimen_snapshot.py": "require_question",
        "liuyao_board.py": "require_mutually_exclusive",
        "qimen_charts.py": "require_mutually_exclusive",
        "liuyao_ask.py": "require_message",
        "qimen_ask.py": "require_message",
    }
    for name, symbol in expected.items():
        text = (TOOLS / name).read_text(encoding="utf-8")
        assert "app.guards" in text and symbol in text, name
