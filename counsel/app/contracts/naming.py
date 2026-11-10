"""Naming tool contracts.

本模块是 naming 域工具契约的单一事实源：input/output schema、描述、枚举与
约束全部由此投影。
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.schema import ToolContract, unique_list
from app.contracts.yongshen import YongShenChoice

# ── naming_surname ──
TOOL_NAMING_SURNAME_DESCRIPTION = "按发音匹配候选姓氏：源姓必须是 Latin 拼音/英文拼写，匹配前会 trim、lowercase 并折叠常见拉丁变体；按 pinyin_exact / romanization_exact / phonetic_close 排序，无匹配时按百家姓顺序回退。max_candidates 是硬上限。仅处理姓氏；取字池用 naming_pick，整名评估用 naming_check。"

class NamingSurnameInput(BaseModel):
    """按发音匹配姓氏的输入。"""

    model_config = ConfigDict(extra="forbid")
    source_surname: str = Field(
        description="源姓氏的拉丁拼音或英文拼写（如 zhang / chan / lee）。",
    )
    max_candidates: Annotated[int, Field(ge=1, le=12)] = Field(
        default=6,
        description="返回候选数上限。",
    )


class NamingSurnameOutputCandidatesItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    surname: str = Field(
        description="Chinese surname candidate.",
    )
    pinyin: str = Field(
        description="Candidate pinyin.",
    )
    tone: int = Field(
        ge=0,
        le=5,
        description="Candidate surname tone number.",
    )
    baijiaxing_index: int = Field(
        ge=1,
        description="Classic Baijiaxing order index.",
    )
    match_kind: Literal[
        "pinyin_exact",
        "romanization_exact",
        "phonetic_close",
        "fallback_baijiaxing",
    ] = Field(
        description="确定性匹配类型。",
    )
    basis: list[str] = Field(
        default=None,
        description="Match basis tokens.",
    )


class NamingSurnameOutput(BaseModel):
    """Ordered Chinese surname candidates for a Latin source surname."""
    model_config = ConfigDict(extra="forbid")
    source_surname: str = Field(
        description="Normalized Latin source surname.",
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
TOOL_NAMING_PICK_DESCRIPTION = "按用/喜五行角色取字池：selected_yongshen 直接使用八字用神裁决返回；first_role 与 second_role 选择字位角色，second_role 省略表示单名；生成阶段不提供 ji 字池。字池产出后组合候选名用 naming_compose，组合前查单字事实用 naming_character。"

PickWuxing = Literal["木", "火", "土", "金", "水"]
PickRole = Literal["yong", "xi"]

class ConsumableSelectedYongshen(BaseModel):
    """起名只消费用神裁决中的五行计划与状态；其他裁决上下文不参与校验。"""

    model_config = ConfigDict(extra="allow")
    selected_yongshen: YongShenChoice
    status: Literal["resolved", "partial"] = Field(description="用神裁决状态；起名只接受 resolved 或 partial。")


class NamingPickInput(BaseModel):
    """按用/喜角色取字池的输入。"""

    model_config = ConfigDict(extra="forbid")
    selected_yongshen: ConsumableSelectedYongshen = Field(
        description="八字用神裁决返回对象。",
    )
    first_role: PickRole = Field(
        default="yong",
        description="第一字位角色；只能是 yong 或 xi。",
    )
    second_role: PickRole | None = Field(
        default=None,
        description="第二字位角色；省略表示单名，只能是 yong 或 xi。",
    )

    @model_validator(mode="after")
    def _enforce_pick_roles(self) -> "NamingPickInput":
        """双名至少保留一个 yong 字位；单名只允许 yong。"""
        if self.selected_yongshen.status not in ("resolved", "partial"):
            raise ValueError("selected_yongshen.status 必须是 resolved 或 partial")
        if self.selected_yongshen.selected_yongshen.yong == "无候选":
            raise ValueError("selected_yongshen.selected_yongshen.yong 必须是具体五行")
        if self.second_role is None and self.first_role != "yong":
            raise ValueError("单名 first_role 只能是 yong")
        if self.first_role != "yong" and self.second_role != "yong":
            raise ValueError("双名至少一个字位必须是 yong")
        return self


class NamingPickOutputPoolsItemCharsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    char: str = Field(
        description="Candidate Chinese character.",
    )
    frequency: Literal["common", "standard", "rare"] = Field(
        description="汉字使用频层。",
    )


class NamingPickOutputPoolsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    wuxing: PickWuxing = Field(
        description="该字位的五行。",
    )
    char_position: Literal["first", "second"] = Field(
        description="Position of the character in the given name.",
    )
    role: PickRole = Field(
        description="五行角色。",
    )
    chars: list[NamingPickOutputPoolsItemCharsItem] = Field(
        description="Available characters and frequency.",
    )


class NamingPickOutput(BaseModel):
    """Character pools selected for requested five-element positions."""
    model_config = ConfigDict(extra="forbid")
    pools: list[NamingPickOutputPoolsItem] = Field(
        description="按字位返回的字池。",
    )

TOOL_NAMING_PICK = ToolContract(
    name="naming_pick",
    description=TOOL_NAMING_PICK_DESCRIPTION,
    input=NamingPickInput,
    output=NamingPickOutput,
)


# ── naming_character ──
TOOL_NAMING_CHAR_DESCRIPTION = "按精确单字查询起名字库事实：字频、五行、笔画、部首、拼音、声调。不做繁简/异体折叠，不接受多字或空白变体；未收录时 character 为 null。适用于核对单字属性；取字池用 naming_pick，组合候选名用 naming_compose，整名评估用 naming_check。"

class NamingCharInput(BaseModel):
    """单字字库查询输入。"""

    model_config = ConfigDict(extra="forbid")
    char: str = Field(
        description="要查询的单个汉字。",
    )


class NamingCharInfo(BaseModel):
    """One naming-database character entry."""

    model_config = ConfigDict(extra="forbid")
    char: str = Field(
        description="The character.",
    )
    frequency: Literal["common", "standard", "rare"] = Field(
        description="汉字使用频层。",
    )
    wuxing: PickWuxing = Field(
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

    model_config = ConfigDict(extra="forbid")
    character: NamingCharInfo | None = Field(
        description="字库条目；未收录时为 null。",
    )


TOOL_NAMING_CHAR = ToolContract(
    name="naming_character",
    description=TOOL_NAMING_CHAR_DESCRIPTION,
    input=NamingCharInput,
    output=NamingCharOutput,
    output_description="Naming-database entry for one character; character is null when it is not in the naming database.",
)


# ── naming_compose ──
TOOL_NAMING_COMPOSE_DESCRIPTION = "将已收录且非禁用的字组合为候选名：单名=first_chars 单字，双名=first_chars×second_chars 两字组合；按 first_chars 顺序生成并在 max_names 处截断。输入字池须来自 naming_pick，重复字会被拒绝。组合后的候选名评估用 naming_check，单字事实查询用 naming_character。"

class NamingComposeInput(BaseModel):
    """候选名组合输入。"""

    model_config = ConfigDict(extra="forbid")
    first_chars: Annotated[unique_list[Annotated[str, Field(min_length=1, max_length=4)]], Field(min_length=1, max_length=256)] = Field(
        description="第一位可选字（1-256 个）。",
    )
    second_chars: Annotated[unique_list[Annotated[str, Field(min_length=1, max_length=4)]], Field(min_length=1, max_length=256)] = Field(
        default=None,
        description="第二位可选字（双名用，1-256 个）；省略=单名。",
    )
    max_names: Annotated[int, Field(ge=1, le=1000)] = Field(
        default=100,
        description="返回候选名上限。",
    )


class NamingComposeOutput(BaseModel):
    """Candidate given names composed from validated character pools."""
    model_config = ConfigDict(extra="forbid")
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
TOOL_NAMING_CHECK_DESCRIPTION = "独立评估整名，不生成候选名；生成用 naming_compose，取字池用 naming_pick。接受 1-200 个去重候选名，逐个返回 evaluations：invalid 不中断批量，以 valid=false 与结构化 errors 说明长度、字库、禁用/必含字或音律问题。selected_yongshen 直接使用八字裁决对象且仅接受 resolved/partial；省略时不做五行角色评估。"

class NamingCheckInput(BaseModel):
    """候选名评估输入。"""

    model_config = ConfigDict(extra="forbid")
    given_names: Annotated[unique_list[Annotated[str, Field(min_length=1, max_length=2)]], Field(min_length=1, max_length=200)] = Field(
        description="要评估的名字（单名或双名，1-2 字）。",
    )
    selected_yongshen: ConsumableSelectedYongshen | None = Field(
        default=None,
        description="八字用神裁决返回对象；省略时不做五行角色评估。",
    )
    forbidden_chars: Annotated[
        unique_list[Annotated[str, Field(min_length=1, max_length=1)]],
        Field(max_length=32),
    ] = Field(
        default_factory=list,
        description="用户明确避讳的单字；候选名命中任一字即为 invalid。",
    )
    required_chars: Annotated[
        unique_list[Annotated[str, Field(min_length=1, max_length=1)]],
        Field(max_length=2),
    ] = Field(
        default_factory=list,
        description="用户要求必含的单字；候选名必须包含全部字符。",
    )
    @model_validator(mode="after")
    def _enforce_consumable_yongshen(self) -> "NamingCheckInput":
        if self.selected_yongshen is None:
            return self
        if self.selected_yongshen.status not in ("resolved", "partial"):
            raise ValueError("selected_yongshen.status 必须是 resolved 或 partial")
        if self.selected_yongshen.selected_yongshen.yong == "无候选":
            raise ValueError("selected_yongshen.selected_yongshen.yong 必须是具体五行")
        if set(self.forbidden_chars) & set(self.required_chars):
            raise ValueError("同一字不能同时属于 forbidden_chars 和 required_chars")
        return self


class NamingCheckError(BaseModel):
    """One structured naming-validation error."""

    model_config = ConfigDict(extra="forbid")
    code: Literal[
        "invalid_name_length",
        "character_not_found",
        "negative_character_forbidden",
        "forbidden_character",
        "required_character_missing",
    ] = Field(
        description="机器可读错误码。",
    )
    char: Annotated[str, Field(min_length=1, max_length=1)] | None = Field(
        default=None,
        description="出错的单字；长度错误、禁用字或缺失必含字时为对应字，否则为 null。",
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
    wuxing_role: Literal["yong", "xi", "ji", "neutral"] | None = Field(
        default=None,
        description="该字对照五行计划的实际角色；未评估时为 null。",
    )


class NamingCheckPhonetic(BaseModel):
    """Tone sequence of a valid name."""

    model_config = ConfigDict(extra="forbid")
    tones: Annotated[str, Field(pattern=r"^\d(-\d)?$")] = Field(
        description="声调序列，按字序以 - 连接（如 2-3）。",
    )


class NamingCheckOutputItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
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
class NamingCheckOutput(BaseModel):
    """候选名评估结果：根始终是对象（≤2025-11-25 的 MCP wire 模型拒绝非
    object 根的 outputSchema），与 naming_surname.candidates /
    naming_pick.pools 的家族惯例同构。"""

    model_config = ConfigDict(extra="forbid")
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
