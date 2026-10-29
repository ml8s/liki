"""Naming tool contracts.

本模块是 naming 域工具契约的单一事实源：input/output schema、描述、枚举与
约束全部由此投影。
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.contracts.schema import ToolContract, unique_list
from app.contracts.yongshen import ConsumedYongShen, YongShenSource

# ── naming_surname ──
TOOL_NAMING_SURNAME_DESCRIPTION = "按发音匹配候选姓氏：源姓（拉丁拼音/英文拼写）→ 同音/近音/罗马化别名匹配的百家姓候选（pinyin_exact / romanization_exact / phonetic_close），无匹配时按百家姓顺序回退。"

class NamingSurnameInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    source_surname: str = Field(
        description="源姓氏的拉丁拼音或英文拼写（如 zhang / chan / lee）。",
    )
    max_candidates: Annotated[int, Field(ge=1, le=12)] = Field(
        default=6,
        description="返回候选数上限。",
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
TOOL_NAMING_PICK_DESCRIPTION = "按五行取字池：给出名字位（first/second）的五行要求，返回按字排序的可选字池（含字频）。second 省略=单名只取 first 池；second 传入=双名取 first+second 池。五行只接受中文闭集（木/火/土/金/水）。"

PickWuxing = Literal["木", "火", "土", "金", "水"]

class NamingPickInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    first: PickWuxing = Field(
        description="名字第一位字的五行。",
    )
    second: PickWuxing | None = Field(
        default=None,
        description="名字第二位字的五行；省略=单名。",
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
    wuxing: PickWuxing = Field(
        description="该字位的五行。",
    )
    slot: Literal["first", "second"] = Field(
        description="Name slot.",
    )
    chars: list[NamingPickOutputPoolsItemCharsItem] = Field(
        description="Available characters and frequency.",
    )


class NamingPickOutput(BaseModel):
    """Character pools selected for requested five-element slots."""
    model_config = ConfigDict(extra="allow")
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


class NamingCharInfo(BaseModel):
    """One naming-database character entry."""

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


class NamingCharOutput(BaseModel):
    """单字查询结果：根始终是对象（≤2025-11-25 的 MCP wire 模型拒绝非 object
    根的 outputSchema），“未收录”用字段级 null 表达。"""

    model_config = ConfigDict(extra="allow")
    character: NamingCharInfo | None = Field(
        description="字库条目；未收录时为 null。",
    )


TOOL_NAMING_CHAR = ToolContract(
    name="naming_char",
    description=TOOL_NAMING_CHAR_DESCRIPTION,
    input=NamingCharInput,
    output=NamingCharOutput,
    output_description="Naming-database entry for one character; character is null when it is not in the naming database.",
)


# ── naming_compose ──
TOOL_NAMING_COMPOSE_DESCRIPTION = "将取字池中的字组合为候选名：单名=first 单字，双名=first×second 两字组合；按 first 顺序生成，最多 max_names 个。"

class NamingComposeInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    first: Annotated[unique_list[Annotated[str, Field(min_length=1, max_length=4)]], Field(min_length=1, max_length=256)] = Field(
        description="第一位可选字（1-256 个）。",
    )
    second: Annotated[unique_list[Annotated[str, Field(min_length=1, max_length=4)]], Field(min_length=1, max_length=256)] = Field(
        default=None,
        description="第二位可选字（双名用，1-256 个）；省略=单名。",
    )
    max_names: Annotated[int, Field(ge=1, le=1000)] = Field(
        default=100,
        description="返回候选名上限。",
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
TOOL_NAMING_CHECK_DESCRIPTION = "独立评估候选名：长度/字库收录/禁用字校验，音律声调序列，五行命中（用神/喜神/忌神）；返回每名的 valid/errors/characters/phonetic/wuxing。用神优先消费八字用神裁决（专用能力或因子快照中的同一对象），不自行重选。"

class _NamingCheckNames(BaseModel):
    model_config = ConfigDict(extra="forbid")
    given_names: Annotated[unique_list[Annotated[str, Field(min_length=1, max_length=2)]], Field(min_length=1, max_length=200)] = Field(
        description="要评估的名字（单名或双名，1-2 字）。",
    )


class NamingCheckManualWuxing(_NamingCheckNames):
    """无完整出生信息时的显式五行策略：直接给出用神/喜神/忌神。"""

    model_config = ConfigDict(extra="forbid")
    yongshen: Literal["木", "火", "土", "金", "水"] = Field(
        default=None,
        description="显式用神五行策略，不是八字裁决结果。",
    )
    xishen: list[Literal["木", "火", "土", "金", "水"]] = Field(
        default=None,
        description="显式喜神五行策略，不是八字裁决结果。",
    )
    jishen: list[Literal["木", "火", "土", "金", "水"]] = Field(
        default=None,
        description="显式忌神五行策略，不是八字裁决结果。",
    )


class NamingCheckSelectedYongshen(_NamingCheckNames):
    """消费八字用神裁决：只消费 selected/status，不重新裁决。"""

    model_config = ConfigDict(extra="forbid")
    selected_yongshen: ConsumedYongShen | None = Field(
        default=None,
        description="八字用神裁决返回对象（专用用神能力或 compute_factors 快照中的同一对象）。",
    )


NamingCheckInput = NamingCheckManualWuxing | NamingCheckSelectedYongshen


class NamingCheckError(BaseModel):
    """One structured naming-validation error."""

    model_config = ConfigDict(extra="forbid")
    code: Literal["invalid_name_length", "character_not_found", "negative_character_forbidden"] = Field(
        description="机器可读错误码。",
    )
    char: Annotated[str, Field(min_length=1, max_length=1)] | None = Field(
        default=None,
        description="出错的单字；长度错误时为 null。",
    )


class NamingCheckCharFacts(BaseModel):
    """Resolved character facts for one evaluated character."""

    model_config = ConfigDict(extra="forbid")
    char: Annotated[str, Field(min_length=1, max_length=1)] = Field(description="The character.")
    frequency: Literal["common", "standard", "rare"] = Field(description="Character usage tier.")
    wuxing: Literal["木", "火", "土", "金", "水"] = Field(description="Character 五行。")
    stroke: int = Field(ge=1, description="Stroke count.")
    pinyin: str = Field(min_length=1, description="Pinyin.")
    tone: int = Field(ge=0, le=5, description="Tone number.")
    radical: str | None = Field(default=None, description="Radical, when known.")


class NamingCheckPhonetic(BaseModel):
    """Tone sequence of a valid name."""

    model_config = ConfigDict(extra="forbid")
    tones: Annotated[str, Field(pattern=r"^\d(-\d)?$")] = Field(
        description="声调序列，按字序以 - 连接（如 2-3）。",
    )


class NamingCheckWuxingAssessment(BaseModel):
    """Five-element hits of a name against the requested 用神/喜神/忌神."""

    model_config = ConfigDict(extra="forbid")
    yong: bool | None = Field(default=None, description="名字是否命中用神五行。")
    xi: bool | None = Field(default=None, description="名字是否命中任一喜神五行。")
    ji: bool | None = Field(default=None, description="名字是否命中任一忌神五行。")


class NamingCheckOutputItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    given_name: str = Field(
        description="Evaluated given name.",
    )
    valid: bool = Field(
        description="Whether the name passed all checks.",
    )
    errors: list[NamingCheckError] | None = Field(
        default=None,
        description="Validation errors, if any.",
    )
    characters: list[NamingCheckCharFacts] | None = Field(
        default=None,
        description="Resolved character facts, when valid.",
    )
    phonetic: NamingCheckPhonetic | None = Field(
        default=None,
        description="Tone sequence, when valid.",
    )
    wuxing: NamingCheckWuxingAssessment | None = Field(
        default=None,
        description="Five-element matches, when requested.",
    )
    yongshen_source: YongShenSource | None = Field(
        default=None,
        description="消费 selected_yongshen 时回传的用神来源：policy_id/rule_id/status/chart_digest/yongshen_digest。",
    )


class NamingCheckOutput(BaseModel):
    """候选名评估结果：根始终是对象（≤2025-11-25 的 MCP wire 模型拒绝非
    object 根的 outputSchema），与 naming_surname.candidates /
    naming_pick.pools 的家族惯例同构。"""

    model_config = ConfigDict(extra="allow")
    evaluations: list[NamingCheckOutputItem] = Field(
        description="每个候选名的评估结果，按输入顺序。",
    )


TOOL_NAMING_CHECK = ToolContract(
    name="naming_check",
    description=TOOL_NAMING_CHECK_DESCRIPTION,
    input=NamingCheckInput,
    output=Annotated[
        NamingCheckOutput,
        Field(description="Naming database, phonetic, and five-element evaluations."),
    ],
    output_description="Naming database, phonetic, and five-element evaluations.",
)



NAMING_TOOLS = (
    TOOL_NAMING_SURNAME,
    TOOL_NAMING_PICK,
    TOOL_NAMING_CHAR,
    TOOL_NAMING_COMPOSE,
    TOOL_NAMING_CHECK,
)
