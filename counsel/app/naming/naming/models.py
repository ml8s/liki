"""Naming domain value objects."""

from __future__ import annotations

from dataclasses import dataclass, field

from app.natal.tools.errors import LikiToolError


@dataclass(frozen=True, slots=True)
class Character:
    char: str
    frequency: str
    element: str
    stroke: int
    radical: str = ""
    pinyin: str = ""
    tone: int = 0


@dataclass(frozen=True, slots=True)
class SurnameRecord:
    surname: str
    pinyin: str
    plain_pinyin: str
    baijiaxing_index: int
    aliases: list = field(default_factory=list)


@dataclass(frozen=True, slots=True)
class SelectedYongshenPlan:
    """Normalized yong/xi/ji roles used by selection and evaluation."""

    yong: str
    xi: tuple[str, ...]
    ji: tuple[str, ...]

    def role_of(self, element: str) -> str:
        if element == self.yong:
            return "yong"
        if element in self.xi:
            return "xi"
        if element in self.ji:
            return "ji"
        return "neutral"

    def elements_for(self, role: str) -> tuple[str, ...]:
        if role == "yong":
            return (self.yong,)
        if role == "xi":
            return self.xi
        if role == "ji":
            return self.ji
        raise LikiToolError("naming_pick 不支持 ji 字池")


@dataclass(frozen=True, slots=True)
class NamingConstraints:
    forbidden: frozenset[str]
    required: frozenset[str]


@dataclass(frozen=True, slots=True)
class NamingError:
    code: str
    char: str

    def to_dict(self) -> dict[str, str]:
        return {"code": self.code, "char": self.char}


@dataclass(frozen=True, slots=True)
class NamingCharacterFact:
    @classmethod
    def from_character(cls, character: Character, wuxing_role: str | None = None) -> NamingCharacterFact:
        return cls(
            char=character.char,
            frequency=character.frequency,
            wuxing=character.element,
            stroke=character.stroke,
            pinyin=character.pinyin,
            tone=character.tone,
            radical=character.radical,
            wuxing_role=wuxing_role,
        )

    char: str
    frequency: str
    wuxing: str
    stroke: int
    pinyin: str
    tone: int
    radical: str = ""
    wuxing_role: str | None = None

    def to_dict(self) -> dict[str, object]:
        item: dict[str, object] = {
            "char": self.char,
            "frequency": self.frequency,
            "wuxing": self.wuxing,
            "stroke": self.stroke,
            "pinyin": self.pinyin,
            "tone": self.tone,
        }
        if self.radical:
            item["radical"] = self.radical
        if self.wuxing_role is not None:
            item["wuxing_role"] = self.wuxing_role
        return item


@dataclass(frozen=True, slots=True)
class NamingEvaluation:
    """Internal evaluation result; projection keeps the published dict shape."""

    given_name: str
    valid: bool
    characters: tuple[NamingCharacterFact, ...] = ()
    errors: tuple[NamingError, ...] = ()
    tones: tuple[int, ...] = ()

    def to_dict(self) -> dict[str, object]:
        if not self.valid:
            return {
                "given_name": self.given_name,
                "valid": False,
                "errors": [error.to_dict() for error in self.errors],
            }
        return {
            "given_name": self.given_name,
            "valid": True,
            "characters": [character.to_dict() for character in self.characters],
            "phonetic": {"tones": "-".join(str(tone) for tone in self.tones)},
        }
