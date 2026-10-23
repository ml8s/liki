"""语义守卫原语：跨工具复用的 fail-closed 检查。

形状守卫（多了/少了/类型/枚举/越界）属于契约模型本身（Pydantic
Field/Literal/union + engine DTO schema），不在本模块重复。这里只收敛
跨工具重复的语义组合守卫。

错误消息是契约的一部分：由 counsel/tests/test_guards.py 钉住，改动
任何消息必须同步更新那里的断言。
"""

_QUESTION_MIN = 2
_MESSAGE_MIN = 2


def require_question(text: str) -> str:
    """strip 后要求非空且至少 2 字；返回规范化文本。"""
    if not isinstance(text, str):
        raise ValueError("question must be non-empty text")
    stripped = text.strip()
    if not stripped:
        raise ValueError("question must be non-empty text")
    if len(stripped) < _QUESTION_MIN:
        raise ValueError("question must be at least 2 characters")
    return stripped


def require_message(text: str) -> str:
    """strip 后要求非空且至少 2 字；返回规范化文本。"""
    if not isinstance(text, str):
        raise ValueError("message must be non-empty text")
    stripped = text.strip()
    if not stripped:
        raise ValueError("message must be non-empty text")
    if len(stripped) < _MESSAGE_MIN:
        raise ValueError("message must be at least 2 characters")
    return stripped


def require_exactly_one(**alternatives: object) -> None:
    """命名字段恰好一个非 None；0 或 ≥2 都拒绝。"""
    provided = [name for name, value in alternatives.items() if value is not None]
    if len(provided) != 1:
        raise ValueError("provide exactly one of " + " or ".join(alternatives))


def require_mutually_exclusive(**fields: object) -> None:
    """命名字段最多一个非 None；并存即拒绝。"""
    if sum(value is not None for value in fields.values()) > 1:
        raise ValueError(" and ".join(fields) + " are mutually exclusive")


def require_object(value: object, name: str) -> None:
    """要求值为 object（dict）。"""
    if not isinstance(value, dict):
        raise ValueError(f"{name} must be an object")
