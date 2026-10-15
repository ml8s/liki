"""Naming tool contracts.

Generated once from the legacy hand-written manifests, then maintained as
ordinary source: this module is the single source of truth for these tool
contracts (input schema, output schema, descriptions, enums, constraints).
"""

from __future__ import annotations

from typing import Annotated, Any, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.contracts.schema import ToolContract

# ── naming_surname ──
TOOL_NAMING_SURNAME_DESCRIPTION = "按发音匹配候选姓氏：源姓（拉丁拼音/英文拼写）→ 同音/近音/罗马化别名匹配的百家姓候选（pinyin_exact / romanization_exact / phonetic_close），无匹配时按百家姓顺序回退。"

class NamingSurnameInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    source_surname: str = Field(
        description="源姓氏的拉丁拼音或英文拼写（如 zhang / chan / lee）。",
    )
    max_candidates: int = Field(
        default=None,
        description="返回候选数上限，默认 6，范围 1-12。",
    )


class NamingSurnameOutputCandidatesItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    surname: str = Field(
        description="Chinese surname candidate.",
    )
    pinyin: str = Field(
        description="Candidate pinyin.",
    )
    match_level: str = Field(
        description="Match strategy or fallback level.",
    )
    basis: list[str] = Field(
        default=None,
        description="Match basis tokens.",
    )


class NamingSurnameOutput(BaseModel):
    """Ranked Chinese surname candidates for a Latin source surname."""
    model_config = ConfigDict(extra="allow")
    source_surname: str = Field(
        description="Normalized Latin source surname.",
    )
    strategy: Literal["phonetic", "baijiaxing_fallback"] = Field(
        description="Matching strategy.",
    )
    candidates: list[NamingSurnameOutputCandidatesItem] = Field(
        description="Ordered surname candidates.",
    )

TOOL_NAMING_SURNAME = ToolContract(
    name="naming_surname",
    description=TOOL_NAMING_SURNAME_DESCRIPTION,
    input=NamingSurnameInput,
    output=NamingSurnameOutput,
)


# ── naming_pick ──
TOOL_NAMING_PICK_DESCRIPTION = "按五行取字池：给出名字位（first/second）的五行要求，返回按字排序的可选字池（含字频）。count=1 单名只取 first 池，count=2 双名取 first+second 池。"

class NamingPickInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    wuxing1: Literal["木", "火", "土", "金", "水"] = Field(
        description="名字第一位字的五行（中英文均可，如 木/wood）。",
    )
    wuxing2: Literal["木", "火", "土", "金", "水"] = Field(
        default=None,
        description="名字第二位字的五行（count=2 时；省略则同 wuxing1）。",
    )
    count: Literal[1, 2] = Field(
        default=None,
        description="名字字数：1=单名，2=双名（默认 2）。",
    )


class NamingPickOutputPoolsItemCharsItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    char: str = Field(
        description="Candidate Chinese character.",
    )
    frequency: Literal["common", "standard", "rare"] = Field(
        description="Character usage tier: common, standard, or rare.",
    )


class NamingPickOutputPoolsItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    slot: Literal["first", "second"] = Field(
        description="Name slot.",
    )
    chars: list[NamingPickOutputPoolsItemCharsItem] = Field(
        description="Available characters and frequency.",
    )


class NamingPickOutput(BaseModel):
    """Character pools selected for requested five-element slots."""
    model_config = ConfigDict(extra="allow")
    wuxing1: Literal["木", "火", "土", "金", "水"] = Field(
        description="Normalized first-slot element.",
    )
    wuxing2: Literal["木", "火", "土", "金", "水"] = Field(
        default=None,
        description="Normalized second-slot element, when a double name was requested.",
    )
    pools: list[NamingPickOutputPoolsItem] = Field(
        description="Character pools by name slot.",
    )

TOOL_NAMING_PICK = ToolContract(
    name="naming_pick",
    description=TOOL_NAMING_PICK_DESCRIPTION,
    input=NamingPickInput,
    output=NamingPickOutput,
)


# ── naming_char ──
TOOL_NAMING_CHAR_DESCRIPTION = "查询单个汉字的起名字库信息：字频、五行、笔画、部首、拼音、声调；未收录返回空。"

class NamingCharInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    char: str = Field(
        description="要查询的单个汉字。",
    )


class NamingCharOutput(BaseModel):
    model_config = ConfigDict(extra="allow")
    char: str = Field(
        description="The character.",
    )
    frequency: Literal["common", "standard", "rare"] = Field(
        description="Character usage tier: common, standard, or rare.",
    )
    wuxing: str = Field(
        description="Five element.",
    )
    stroke: int = Field(
        description="Stroke count.",
    )
    pinyin: str = Field(
        description="Pinyin.",
    )
    tone: int = Field(
        description="Tone number.",
    )
    radical: str = Field(
        default=None,
        description="Radical, when known.",
    )

TOOL_NAMING_CHAR = ToolContract(
    name="naming_char",
    description=TOOL_NAMING_CHAR_DESCRIPTION,
    input=NamingCharInput,
    output=NamingCharOutput | None,
    output_description="A single Chinese character, or null when it is not in the naming database.",
)


# ── naming_compose ──
TOOL_NAMING_COMPOSE_DESCRIPTION = "将取字池中的字组合为候选名：单名=first 单字，双名=first×second 两字组合；按 first 顺序生成，最多 max_names 个。"

class NamingComposeInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    first: list[str] = Field(
        description="第一位可选字（1-256 个）。",
    )
    second: list[str] = Field(
        default=None,
        description="第二位可选字（双名用，1-256 个）；省略=单名。",
    )
    max_names: int = Field(
        default=None,
        description="返回候选名上限，默认 100，范围 1-1000。",
    )


class NamingComposeOutput(BaseModel):
    """Candidate given names composed from validated character pools."""
    model_config = ConfigDict(extra="allow")
    total_possible: int = Field(
        description="Total combinations before max_names truncation.",
    )
    names: list[str] = Field(
        description="Generated given-name candidates.",
    )

TOOL_NAMING_COMPOSE = ToolContract(
    name="naming_compose",
    description=TOOL_NAMING_COMPOSE_DESCRIPTION,
    input=NamingComposeInput,
    output=NamingComposeOutput,
)


# ── naming_check ──
TOOL_NAMING_CHECK_DESCRIPTION = "独立评估候选名：长度/字库收录/禁用字校验，音律声调序列，五行命中（用神/喜神/忌神）；返回每名的 valid/errors/characters/phonetic/wuxing。"

class NamingCheckInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    given_names: list[str] = Field(
        description="要评估的名字（单名或双名，1-2 字）。",
    )
    yongshen: Literal["木", "火", "土", "金", "水"] = Field(
        default=None,
        description="八字用神五行（中英文均可）。",
    )
    xishen: list[Literal["木", "火", "土", "金", "水"]] = Field(
        default=None,
        description="喜神五行列表。",
    )
    jishen: list[Literal["木", "火", "土", "金", "水"]] = Field(
        default=None,
        description="忌神五行列表。",
    )


class NamingCheckOutputItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    given_name: str = Field(
        description="Evaluated given name.",
    )
    valid: bool = Field(
        description="Whether the name passed all checks.",
    )
    errors: list[dict[str, Any]] = Field(
        default=None,
        description="Validation errors, if any.",
    )
    characters: list[dict[str, Any]] = Field(
        default=None,
        description="Resolved character facts, when valid.",
    )
    phonetic: dict[str, Any] = Field(
        default=None,
        description="Tone sequence, when valid.",
    )
    wuxing: dict[str, Any] = Field(
        default=None,
        description="Five-element matches, when requested.",
    )

TOOL_NAMING_CHECK = ToolContract(
    name="naming_check",
    description=TOOL_NAMING_CHECK_DESCRIPTION,
    input=NamingCheckInput,
    output=Annotated[list[NamingCheckOutputItem], Field(description="Naming database, phonetic, and five-element evaluations.")],
    output_description="Naming database, phonetic, and five-element evaluations.",
)



NAMING_TOOLS = (
    TOOL_NAMING_SURNAME,
    TOOL_NAMING_PICK,
    TOOL_NAMING_CHAR,
    TOOL_NAMING_COMPOSE,
    TOOL_NAMING_CHECK,
)
