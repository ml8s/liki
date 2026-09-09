"""问卦场景路由；只判断场景与方法，不做排盘和命理解释。"""
from __future__ import annotations

import re

from divination_safety import assess as assess_safety


SCHEMA_VERSION = "divination-route-v1"

METHODS = {
    "liuyao": "app/liuyao-chart.md",
    "qimen": "app/qimen-chart.md",
    "huangli": "app/auspicious.md",
}

CATEGORIES = {
    "event_outcome": {
        "method": "liuyao",
        "name": "事件成败 / 结果",
        "reason": "问题核心是具体事件的结果或趋势，默认用六爻。",
    },
    "strategy_decision": {
        "method": "qimen",
        "name": "策略 / 决策 / 进退",
        "reason": "问题核心是行动选择、路径或当前态势，默认用奇门。",
    },
    "direction_space": {
        "method": "qimen",
        "name": "方向 / 方位 / 空间",
        "reason": "问题核心是方向、方位或空间路径，默认用奇门。",
    },
    "timing_event": {
        "method": "liuyao",
        "name": "事件应期",
        "reason": "问题核心是具体事件何时有结果，默认用六爻应期。",
    },
    "timing_action": {
        "method": "qimen",
        "name": "行动时机",
        "reason": "问题核心是现在是否适合行动或选择时机，默认用奇门。",
    },
    "date_selection": {
        "method": "huangli",
        "name": "择日",
        "reason": "问题核心是选择日期，默认用黄历。",
    },
    "mixed_outcome_strategy": {
        "method": "clarify",
        "name": "结果与策略混合",
        "reason": "结果和策略是两个目标；默认不双排，请用户选择本次主问题。",
    },
    "life_pattern": {
        "method": "clarify",
        "name": "长期命局",
        "reason": "长期命局更适合八字 / 命理盘，不应临时用六爻或奇门替代。",
    },
    "ambiguous": {
        "method": "clarify",
        "name": "场景不明确",
        "reason": "问题目标不足以选择方法，先澄清。",
    },
}

_KEYWORDS = [
    ("date_selection", r"哪一天|哪天|择日|择吉|吉日|日子适合|日期"),
    ("direction_space", r"方向|方位|哪边|往.{0,6}(南|北|东|西)|东南|西北|东北|西南|路径"),
    ("strategy_decision", r"该不该|要不要|怎么办|怎么选|策略|该.{0,8}还是|主动.{0,6}还是|等待.{0,6}还是|联系.{0,6}还是|进退|推进|方案"),
    ("timing_action", r"现在适合|今天适合|适合.{0,4}(行动|推进|谈判|签约|出门)|时机"),
    ("timing_event", r"什么时候有结果|多久.{0,4}(结果|到账|回复)|何时.{0,4}(成|回|到|结束)"),
    ("event_outcome", r"能不能|能否|会不会|可不可以|是否|成不成|结果|成功率|找得到|通过|答应|复合|到账|回款"),
    ("life_pattern", r"终身|命格|一生运势|适合什么行业|婚姻整体|未来几年运势"),
]

_METHOD_WORDS = [
    ("liuyao", r"六爻|摇卦|起卦|铜钱|硬币"),
    ("qimen", r"奇门|起局|遁甲|奇门盘"),
    ("huangli", r"黄历|老黄历|择日|择吉"),
]


def _safety(question: str) -> dict:
    return assess_safety(question)


def _explicit_method(question: str) -> str | None:
    for method, pattern in _METHOD_WORDS:
        if re.search(pattern, question, re.IGNORECASE):
            return method
    return None


def _heuristic_category(question: str) -> tuple[str, list[str]]:
    matches: list[str] = []
    for category, pattern in _KEYWORDS:
        if re.search(pattern, question, re.IGNORECASE):
            matches.append(category)
    if not matches:
        return "ambiguous", []
    if "date_selection" in matches:
        return "date_selection", matches
    if {"event_outcome", "strategy_decision"} & set(matches) and {"timing_action", "direction_space"} & set(matches):
        return "mixed_outcome_strategy", matches
    if "strategy_decision" in matches:
        return "strategy_decision", matches
    if "direction_space" in matches:
        return "direction_space", matches
    if "timing_action" in matches:
        return "timing_action", matches
    if "timing_event" in matches:
        return "timing_event", matches
    return matches[0], matches


def route_question(
    question: str,
    *,
    category: str | None = None,
    specified_method: str | None = None,
) -> dict:
    """返回单一默认路由；双法合参不是本函数的输出。"""
    if not isinstance(question, str) or not question.strip():
        raise ValueError("question must be non-empty text")
    question = question.strip()
    safety = _safety(question)
    if safety["status"] == "redirect":
        return {
            "schema_version": SCHEMA_VERSION,
            "route": "blocked",
            "category": "safety_redirect",
            "safety_category": safety["category"],
            "confidence": "safety",
            "reason": safety["message"],
            "entry": None,
            "next_tool": None,
            "dual_divination": False,
            "user_question": question,
        }

    specified = specified_method
    if specified is not None and specified not in METHODS:
        raise ValueError(f"specified_method must be one of: {', '.join(METHODS)}")
    if specified is None:
        specified = _explicit_method(question)

    keyword_matches: list[str] = []
    if category is None:
        category, keyword_matches = _heuristic_category(question)
    elif category not in CATEGORIES:
        raise ValueError(f"unknown category: {category}; valid: {', '.join(sorted(CATEGORIES))}")

    fact = CATEGORIES[category]
    method = specified or fact["method"]
    confidence = "specified" if specified else "semantic"
    reason = fact["reason"]
    if specified:
        reason = f"用户或调用方已指定方法，覆盖场景默认；场景为{fact['name']}。"

    clarify = None
    if method == "clarify":
        if category == "mixed_outcome_strategy":
            clarify = {
                "question": "本次优先看哪一个？",
                "options": [
                    {"value": "event_outcome", "label": "先看事情能不能成", "method": "liuyao"},
                    {"value": "strategy_decision", "label": "先看该怎么做 / 从哪里推进", "method": "qimen"},
                ],
            }
        elif category == "life_pattern":
            clarify = {
                "question": "长期命局建议使用八字命理盘；若仍要问具体事件，请补充本次要判断的事件。",
                "options": [],
            }
        else:
            clarify = {
                "question": "你这次最想判断的是结果，还是行动策略 / 时机 / 方向？",
                "options": [
                    {"value": "event_outcome", "label": "看结果", "method": "liuyao"},
                    {"value": "strategy_decision", "label": "看策略 / 进退", "method": "qimen"},
                    {"value": "date_selection", "label": "选日期", "method": "huangli"},
                ],
            }

    return {
        "schema_version": SCHEMA_VERSION,
        "route": method,
        "category": category,
        "category_name": fact["name"],
        "confidence": confidence,
        "reason": reason,
        "entry": METHODS.get(method),
        "next_tool": {
            "liuyao": "liuyao_qigua",
            "qimen": "qimen_read",
            "huangli": None,
            "clarify": None,
            "blocked": None,
        }.get(method),
        "clarify": clarify,
        "keyword_matches": keyword_matches,
        "dual_divination": False,
        "safety": safety,
        "user_question": question,
    }
