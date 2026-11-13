"""Naming domain value objects."""

from dataclasses import dataclass, field


@dataclass(frozen=True)
class Character:
    char: str
    frequency: str
    element: str
    stroke: int
    radical: str = ""
    pinyin: str = ""
    tone: int = 0


@dataclass(frozen=True)
class SurnameRecord:
    surname: str
    pinyin: str
    plain_pinyin: str
    baijiaxing_index: int
    aliases: list = field(default_factory=list)

