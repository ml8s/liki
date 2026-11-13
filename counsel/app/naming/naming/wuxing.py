"""Canonical Chinese five-element helper."""


def wuxing_from_chinese(name: str) -> str:
    return {
        "木": "木", "火": "火", "土": "土", "金": "金", "水": "水",
    }.get(name, "")
