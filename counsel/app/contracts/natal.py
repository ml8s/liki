"""Natal tool contracts.

本模块是 natal 域工具契约的单一事实源：input/output schema、描述、枚举与
约束全部由此投影。
"""


from typing import Annotated, Any, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.contracts.bazi import BaziChart
from app.contracts.schema import ToolContract, unique_list
from app.contracts.yongshen import BaziYongshenOutput
from app.contracts.ziwei import EngineChart

# ── bazi_yongshen ──
TOOL_BAZI_YONGSHEN_DESCRIPTION = "engine 八字排盘结果（chart）→ 唯一用神裁决。只接受八字盘；返回最终用神、喜神、忌神、状态、冲突、命中规则与来源摘要。起名与断语消费同一份裁决，不重新裁决。"

class BaziYongshenInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    chart: BaziChart = Field(
        description="engine bazi_chart 输出（含四柱与性别）；只接受八字盘。",
    )

TOOL_BAZI_YONGSHEN = ToolContract(
    name="bazi_yongshen",
    description=TOOL_BAZI_YONGSHEN_DESCRIPTION,
    input=BaziYongshenInput,
    output=BaziYongshenOutput,
    output_description="唯一用神裁决结果。",
)


# ── compute_factors ──
TOOL_COMPUTE_FACTORS_DESCRIPTION = "engine 排盘结果（chart）→ {factors, factors_digest}。八字侧内部调 engine fullchart 后按表裁决唯一用神，并将同一份裁决写入 factors.selected_yongshen；来源 context 随 factors._provenance.context 一起进入 digest，digest 为 SHA-256 canonical 防篡改校验。"

class ComputeFactorsInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    chart: EngineChart = Field(
        description="engine 排盘结果：bazi_chart（八字，含四柱）或 ziwei_chart（紫微，含十二宫）",
    )


class ComputeFactorsOutput(BaseModel):
    """Immutable factor snapshot derived from an engine chart."""
    model_config = ConfigDict(extra="allow")
    factors: dict[str, Any] = Field(
        description="Canonical factor snapshot; includes _provenance with side, chart digest, and context. Bazi snapshots include selected_yongshen as the reserved domain result element.",
    )
    factors_digest: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(
        description="SHA-256 hex digest of the canonical factor snapshot.",
    )
TOOL_COMPUTE_FACTORS = ToolContract(
    name="compute_factors",
    description=TOOL_COMPUTE_FACTORS_DESCRIPTION,
    input=ComputeFactorsInput,
    output=ComputeFactorsOutput,
)


# ── natal_query ──
TOOL_NATAL_QUERY_DESCRIPTION = "已绑定来源的因子快照 + 人生问题 → 本命断语。来源 context 由服务端从 factors._provenance 读取并与 digest 一并校验，调用方无需重复传递。"

class NatalQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    factors: dict[str, Any] = Field(
        description="compute_factors 返回的因子快照",
    )
    factors_digest: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(
        description="compute_factors 返回的 SHA-256 因子摘要；必须与 factors 原样匹配",
    )
    topics: Annotated[unique_list[Annotated[Literal["marriage", "family", "children", "career", "wealth", "health", "mental", "study", "personality", "appearance", "origin", "relocation", "social", "property", "adversity", "chart_structure"], Field(description="受控人生问题领域闭集")]], Field(min_length=1, max_length=16)] = Field(
        description="人生问题域：marriage/family/children/career/wealth/health/mental/study/personality/appearance/origin/relocation/social/property/adversity/chart_structure 等",
    )


class NatalQueryOutputAssertionsItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    assertion_id: str = Field(
        description="Stable assertion id.",
    )
    side: Literal["bazi", "ziwei"] = Field(
        description="Factor side: bazi or ziwei.",
    )
    topic: str = Field(
        description="Requested life-topic key.",
    )
    method: str | None = Field(
        default=None,
        description="Judgment method id.",
    )
    conclusion: str = Field(
        description="Deterministic assertion conclusion.",
    )
    source: str | dict[str, Any] = Field(
        default=None,
        description="Classical/rule source. May be a text citation or structured source object.",
    )
    evidence: list[dict[str, Any]] = Field(
        default=None,
        description="Factor evidence used by the rule.",
    )


class NatalQueryOutput(BaseModel):
    """Deterministic natal assertions for requested topics."""
    model_config = ConfigDict(extra="allow")
    assertions: list[NatalQueryOutputAssertionsItem] = Field(
        description="Traceable natal assertions matching the requested topics.",
    )

TOOL_NATAL_QUERY = ToolContract(
    name="natal_query",
    description=TOOL_NATAL_QUERY_DESCRIPTION,
    input=NatalQueryInput,
    output=NatalQueryOutput,
)


# ── period_query ──
TOOL_PERIOD_QUERY_DESCRIPTION = "因子快照 + 时间层 + 人生问题 → 应期断语（大运/大限/流年）。内部据 chart 调 engine 流年/大限取应期字段后匹配断语表；来源 context 与 chart 绑定校验由服务端从 factors._provenance 完成。"

IanaTimezone = Annotated[str, Field(min_length=1, description="IANA 时区，如 Asia/Shanghai。")]


class TimeScopeYear(BaseModel):
    """时间层变体：指定公历年的流年。"""
    model_config = ConfigDict(extra="forbid")
    type: Literal["year"] = Field(description="指定公历年的流年断语。")
    year: int = Field(ge=1, description="目标公历年。")
    timezone: IanaTimezone


class TimeScopeYearRange(BaseModel):
    """时间层变体：指定公历年区间的逐年断语。"""
    model_config = ConfigDict(extra="forbid")
    type: Literal["year_range"] = Field(description="指定公历年区间的逐年断语。")
    start_year: int = Field(ge=1, description="起始公历年。")
    end_year: int = Field(ge=1, description="结束公历年（含）。")
    timezone: IanaTimezone


class TimeScopeDecade(BaseModel):
    """时间层变体：指定公历年所在的大限（十年层）。"""
    model_config = ConfigDict(extra="forbid")
    type: Literal["decade"] = Field(description="指定公历年所在大限（十年层）。")
    year: int = Field(ge=1, description="目标公历年。")
    timezone: IanaTimezone


class TimeScopeCurrentYear(BaseModel):
    """时间层变体：以当前时间解析的流年；时区为变体必填字段。"""
    model_config = ConfigDict(extra="forbid")
    type: Literal["current_year"] = Field(description="以当前时间解析的流年；时区必填。")
    timezone: IanaTimezone


class TimeScopeCurrentDecade(BaseModel):
    """时间层变体：以当前时间解析的大限；时区为变体必填字段。"""
    model_config = ConfigDict(extra="forbid")
    type: Literal["current_decade"] = Field(description="以当前时间解析的大限；时区必填。")
    timezone: IanaTimezone


TimeScope = (
    TimeScopeYear
    | TimeScopeYearRange
    | TimeScopeDecade
    | TimeScopeCurrentYear
    | TimeScopeCurrentDecade
)


class PeriodQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    factors: dict[str, Any] = Field(
        description="compute_factors 返回的因子快照",
    )
    factors_digest: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(
        description="compute_factors 返回的 SHA-256 因子摘要；必须与 factors 原样匹配",
    )
    time_scope: TimeScope = Field(
        description="时间层：year=指定公历年的流年，year_range=逐年区间，decade=指定年所在大限，current_year/current_decade=按当前时间解析。年界与“当前”都按 timezone 确定。",
    )
    topics: Annotated[unique_list[Annotated[Literal["marriage", "family", "children", "career", "wealth", "health", "mental", "study", "personality", "appearance", "origin", "relocation", "social", "property", "adversity", "chart_structure"], Field(description="受控人生问题领域闭集")]], Field(min_length=1, max_length=16)] = Field(
        description="人生问题域（同 natal_query）",
    )
    chart: EngineChart = Field(
        description="engine 排盘结果；必须与 factors 来自同一次 compute_factors 调用（服务端按 chart digest 校验绑定），内部据此调 engine 流年/大限",
    )


class PeriodQueryOutputChart(BaseModel):
    """Chart identity and digest used for period computation."""
    model_config = ConfigDict(extra="allow")
    id: str = Field(
        description="Chart id.",
    )
    digest: str | None = Field(
        description="Chart canonical digest.",
    )
    completeness: str = Field(
        description="Chart completeness.",
    )


class PeriodQueryOutputPeriodsItem(BaseModel):
    model_config = ConfigDict(extra="allow")
    time_scope: dict[str, Any] = Field(
        description="Resolved output scope.",
    )
    assertions: list[dict[str, Any]] = Field(
        default=None,
        description="Assertions for this scope.",
    )
    counts: dict[str, Any] = Field(
        default=None,
        description="Table match and returned counts.",
    )
    error: dict[str, Any] = Field(
        default=None,
        description="Engine error for this scope, if any.",
    )


class PeriodQueryOutputCalibrationFactsHourPillar(BaseModel):
    """Hour stem and branch."""
    model_config = ConfigDict(extra="allow")
    gan: str
    zhi: str


class PeriodQueryOutputCalibrationFacts(BaseModel):
    """Hour-pillar-sensitive facts for birth-time calibration. Present for bazi period queries."""
    model_config = ConfigDict(extra="allow")
    hour_pillar: PeriodQueryOutputCalibrationFactsHourPillar = Field(
        description="Hour stem and branch.",
    )
    hour_stem_ten_god: str = Field(
        description="Ten-god name of the hour stem.",
    )
    hour_shen_sha: list[str] = Field(
        description="Shen sha attached to the hour pillar.",
    )


class PeriodQueryOutput(BaseModel):
    """Deterministic period assertions grouped by requested time scopes."""
    model_config = ConfigDict(extra="allow")
    chart: PeriodQueryOutputChart = Field(
        description="Chart identity and digest used for period computation.",
    )
    query: dict[str, Any] = Field(
        description="Echo of period query scope, topics, and rules.",
    )
    year_basis: dict[str, Any] | None = Field(
        default=None,
        description="Engine year basis, when available.",
    )
    current_year: int | None = Field(
        default=None,
        description="Engine current year, when available.",
    )
    current_year_source: str | None = Field(
        default=None,
        description="Source of current year.",
    )
    periods: list[PeriodQueryOutputPeriodsItem] = Field(
        description="Per-year or per-decade assertion scopes.",
    )
    calibration_facts: PeriodQueryOutputCalibrationFacts = Field(
        default=None,
        description="Hour-pillar-sensitive facts for birth-time calibration. Present for bazi period queries.",
    )

TOOL_PERIOD_QUERY = ToolContract(
    name="period_query",
    description=TOOL_PERIOD_QUERY_DESCRIPTION,
    input=PeriodQueryInput,
    output=PeriodQueryOutput,
)



NATAL_TOOLS = (
    TOOL_BAZI_YONGSHEN,
    TOOL_COMPUTE_FACTORS,
    TOOL_NATAL_QUERY,
    TOOL_PERIOD_QUERY,
)
