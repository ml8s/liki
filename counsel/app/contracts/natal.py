"""Natal tool contracts.

本模块是 natal 域工具契约的单一事实源：input/output schema、描述、枚举与
约束全部由此投影。
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, StrictStr

from app.contracts.schema import ToolContract, unique_list
from app.contracts.yongshen import SelectedYongShen


class NatalSubjectInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    standard_time: Annotated[
        str,
        Field(
            min_length=1,
            json_schema_extra={"format": "date-time"},
            description="出生地标准时间，RFC3339 fixed offset；来自用户明确事实或 subject_candidates 候选，不得编造。",
        ),
    ]
    longitude: Annotated[
        float,
        Field(ge=-180, le=180, description="出生地经度；来自用户确认地点、地点检索结果或既有事实。"),
    ]
    gender: Literal["male", "female"] = Field(description="性别；来自用户明确事实。")



# ── bazi_yongshen ──
TOOL_BAZI_YONGSHEN_DESCRIPTION = "schema 合法的出生主体 → 八字排盘与唯一用神裁决。subject 可由用户明确事实组装，也可取自用户确认后的 subject_candidates 候选；时间模糊或多候选时先归一化并确认。返回最终用神、喜神、忌神、状态、冲突、命中规则与来源摘要；起名只消费 selected_yongshen.yong 投影出的目标五行。"

class BaziYongshenInput(BaseModel):
    """八字用神裁决输入。"""

    model_config = ConfigDict(extra="forbid")
    subject: NatalSubjectInput = Field(description="已考时归一化的出生主体。")

TOOL_BAZI_YONGSHEN = ToolContract(
    name="bazi_yongshen",
    description=TOOL_BAZI_YONGSHEN_DESCRIPTION,
    input=BaziYongshenInput,
    output=SelectedYongShen,
    output_description="唯一用神裁决结果。",
)


# ── natal_query ──
TOOL_NATAL_QUERY_DESCRIPTION = "schema 合法的出生主体 + 人生问题 → 八字与紫微本命合参断语。subject 可由用户明确事实组装，也可取自用户确认后的 subject_candidates 候选；时间模糊或多候选时先归一化并确认。因子计算由服务端内部完成，不对外返回。"

LifeTopic = Literal[
    "marriage", "family", "children", "career", "wealth", "health", "mental",
    "study", "personality", "appearance", "origin", "relocation", "social",
    "property", "adversity", "chart_structure",
]
LifeTopicList = Annotated[
    unique_list[LifeTopic],
    Field(min_length=1, max_length=16, description="受控人生问题域闭集。"),
]

class NatalQueryInput(BaseModel):
    """八字与紫微本命合参断语输入。"""

    model_config = ConfigDict(extra="forbid")
    subject: NatalSubjectInput = Field(description="已考时归一化的出生主体。")
    life_topics: LifeTopicList = Field(
        description="人生问题域。",
    )


class NatalAssertionEvidence(BaseModel):
    model_config = ConfigDict(extra="forbid")
    condition_group: int | None = Field(default=None, description="条件组编号。")
    factor: str = Field(description="命中的因子名。")
    expected: StrictStr | None = Field(default=None, description="因子期望值。")
    actual: StrictStr | None = Field(default=None, description="因子实际值。")
    gender_scope: Literal["男", "女"] | None = Field(
        default=None, description="性别专属因子的作用范围。",
    )


class NatalAssertionSource(BaseModel):
    model_config = ConfigDict(extra="forbid")
    kind: Literal["assertion_table"] = Field(description="断语来源类型。")
    rule_domain: str = Field(description="断语规则在来源表中的命理领域。")
    basis: str = Field(description="规则依据。")
    classic_basis: str = Field(description="经典依据。")


class NatalAssertion(BaseModel):
    model_config = ConfigDict(extra="forbid")
    assertion_id: str = Field(
        description="Stable assertion id.",
    )
    life_topics: Annotated[list[LifeTopic], Field(min_length=1)] = Field(
        description="命中的受控人生问题域。",
    )
    rule_id: str = Field(description="命理断语规则 ID。")
    event_type: str = Field(description="事件类型。")
    event: str = Field(description="事件摘要。")
    conclusion: str = Field(
        description="Deterministic assertion conclusion.",
    )
    source: NatalAssertionSource = Field(
        description="断语规则来源与依据。",
    )
    evidence: list[NatalAssertionEvidence] = Field(
        default_factory=list,
        description="Factor evidence used by the rule.",
    )


class NatalSideResult(BaseModel):
    """One technique side of a combined natal result."""
    model_config = ConfigDict(extra="forbid")
    assertions: list[NatalAssertion] = Field(description="该侧命理断语结果。")


class NatalQueryOutput(BaseModel):
    """Deterministic natal assertions for requested topics."""
    model_config = ConfigDict(extra="forbid")
    natal_chart_digest: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(description="完整本命盘 SHA-256 摘要。")
    bazi: NatalSideResult = Field(description="八字本命断语。")
    ziwei: NatalSideResult = Field(description="紫微本命断语。")

TOOL_NATAL_QUERY = ToolContract(
    name="natal_query",
    description=TOOL_NATAL_QUERY_DESCRIPTION,
    input=NatalQueryInput,
    output=NatalQueryOutput,
)


# ── natal_liunian_query ──
TOOL_LIUNIAN_QUERY_DESCRIPTION = "schema 合法的出生主体 + 起始年号/年数 + 人生问题 → 八字与紫微逐年流年合参断语。subject 可由用户明确事实组装，也可取自用户确认后的 subject_candidates 候选；时间模糊或多候选时先归一化并确认。因子计算由服务端内部完成；year_count=1 表示单年，单次最多 50 年。"


class LiunianQueryInput(BaseModel):
    """八字与紫微流年合参断语输入。"""

    model_config = ConfigDict(extra="forbid")
    subject: NatalSubjectInput = Field(description="已考时归一化的出生主体。")
    start_year: Annotated[int, Field(ge=1, description="起始流年年号；八字按公历立春年，紫微按农历年。")] = Field(
        description="起始流年年号。",
    )
    year_count: Annotated[int, Field(ge=1, le=50, description="连续查询年数；1 表示单年，最多 50 年。")] = Field(
        description="连续查询年数。",
    )
    life_topics: LifeTopicList = Field(
        description="人生问题域（同 natal_query）。",
    )


class LiunianCounts(BaseModel):
    model_config = ConfigDict(extra="forbid")
    matched: int = Field(ge=0, description="来源断语命中次数。")
    returned: int = Field(ge=0, description="过滤后返回的断语数。")


class LiunianYearError(BaseModel):
    model_config = ConfigDict(extra="forbid")
    code: Literal["ENGINE_UNAVAILABLE"] = Field(description="错误码。")
    message: str = Field(description="错误信息。")


class LiunianSideResult(BaseModel):
    """One technique side of a combined yearly result."""
    model_config = ConfigDict(extra="forbid")
    assertions: list[NatalAssertion] = Field(description="该侧流年断语。")
    counts: LiunianCounts = Field(description="命中与返回计数。")


class LiunianYearSuccess(BaseModel):
    model_config = ConfigDict(extra="forbid")
    year: int = Field(description="流年年号。")
    bazi: LiunianSideResult = Field(description="八字流年断语。")
    ziwei: LiunianSideResult = Field(description="紫微流年断语。")


class LiunianYearFailure(BaseModel):
    model_config = ConfigDict(extra="forbid")
    year: int = Field(description="流年年号。")
    error: LiunianYearError = Field(description="流年计算失败信息。")


LiunianYearResult = LiunianYearSuccess | LiunianYearFailure
class LiunianQueryOutput(BaseModel):
    """Deterministic Liunian assertions."""
    model_config = ConfigDict(extra="forbid")
    natal_chart_digest: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(
        description="流年计算使用的完整本命盘摘要。",
    )
    years: list[LiunianYearResult] = Field(
        description="逐年流年结果。",
    )

TOOL_LIUNIAN_QUERY = ToolContract(
    name="natal_liunian_query",
    description=TOOL_LIUNIAN_QUERY_DESCRIPTION,
    input=LiunianQueryInput,
    output=LiunianQueryOutput,
)



NATAL_TOOLS = (
    TOOL_BAZI_YONGSHEN,
    TOOL_NATAL_QUERY,
    TOOL_LIUNIAN_QUERY,
)
