"""Natal tool contracts.

Generated once from the legacy hand-written manifests, then maintained as
ordinary source: this module is the single source of truth for these tool
contracts (input schema, output schema, descriptions, enums, constraints).
"""

from __future__ import annotations

from typing import Annotated, Any, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.bazi import BaziChart
from app.contracts.schema import ToolContract, unique_list
from app.contracts.yongshen import BaziYongshenOutput

# ── bazi_yongshen ──
TOOL_BAZI_YONGSHEN_DESCRIPTION = "engine 八字排盘结果（chart）→ 唯一用神裁决。只接受八字盘；返回最终用神、喜神、忌神、状态、冲突、命中规则与来源摘要。起名与断语消费同一份裁决，不重新裁决。"

class BaziYongshenInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    chart: BaziChart = Field(
        description="engine bazi_chart 输出（含四柱与性别）；只接受八字盘。",
    )

    @model_validator(mode="before")
    @classmethod
    def _require_bazi_chart(cls, data):
        if isinstance(data, dict):
            chart = data.get("chart")
            if isinstance(chart, dict) and "ri" not in chart:
                raise ValueError("counsel/bazi 只接收八字盘（chart 含四柱），收到紫微盘。")
        return data


TOOL_BAZI_YONGSHEN = ToolContract(
    name="bazi_yongshen",
    description=TOOL_BAZI_YONGSHEN_DESCRIPTION,
    input=BaziYongshenInput,
    output=BaziYongshenOutput,
    output_description="唯一用神裁决结果。",
)


# ── compute_factors ──
TOOL_COMPUTE_FACTORS_DESCRIPTION = "engine 排盘结果（chart）→ {factors, factors_digest, context}。八字侧内部调 engine fullchart 后按表裁决唯一用神，并将同一份裁决写入 factors.selected_yongshen；digest 为 SHA-256 canonical 防篡改校验。"

class ComputeFactorsInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    chart: dict[str, Any] = Field(
        description="engine 排盘结果：本域 bazi_chart（八字）或 ziwei_chart（紫微）输出",
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
    context: dict[str, Any] = Field(
        description="Minimal source context bound to the factor snapshot.",
    )
TOOL_COMPUTE_FACTORS = ToolContract(
    name="compute_factors",
    description=TOOL_COMPUTE_FACTORS_DESCRIPTION,
    input=ComputeFactorsInput,
    output=ComputeFactorsOutput,
)


# ── natal_query ──
TOOL_NATAL_QUERY_DESCRIPTION = "已绑定来源的因子快照 + 人生问题 → 本命断语。context 必须原样来自 compute_factors，并与因子 provenance 校验。"

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
    context: dict[str, Any] = Field(
        description="compute_factors 返回的原始 context；会与因子 provenance 绑定校验",
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
    method: str = Field(
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
TOOL_PERIOD_QUERY_DESCRIPTION = "因子快照 + 时间层 + 人生问题 → 应期断语（大运/大限/流年）。内部据 chart 调 engine 流年/大限取应期字段后匹配断语表。"

class PeriodQueryInputTimeScope(BaseModel):
    """时间层：year/year_range/current_year/decade/current_decade"""
    model_config = ConfigDict(extra="forbid")
    type: Literal["year", "year_range", "current_year", "decade", "current_decade"]
    year: int = Field(
        default=None,
        description="type=year/decade 时的锚定年份",
    )
    start_year: int = Field(
        default=None,
        description="type=year_range 的起始年份",
    )
    end_year: int = Field(
        default=None,
        description="type=year_range 的结束年份",
    )

    @model_validator(mode="after")
    def _enforce_scope_anchor(self) -> "PeriodQueryInputTimeScope":
        """Require the anchor fields the legacy schema conditioned on ``type``."""
        required = {
            "year": ("year",),
            "decade": ("year",),
            "year_range": ("start_year", "end_year"),
        }.get(self.type, ())
        missing = [name for name in required if getattr(self, name) is None]
        if missing:
            raise ValueError(f"time_scope type={self.type} requires {', '.join(missing)}")
        return self


class PeriodQueryInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    factors: dict[str, Any] = Field(
        description="compute_factors 返回的因子快照",
    )
    factors_digest: Annotated[str, Field(pattern="^[a-f0-9]{64}$")] = Field(
        description="compute_factors 返回的 SHA-256 因子摘要；必须与 factors 原样匹配",
    )
    time_scope: PeriodQueryInputTimeScope = Field(
        description="时间层：year/year_range/current_year/decade/current_decade",
    )
    topics: Annotated[unique_list[Annotated[Literal["marriage", "family", "children", "career", "wealth", "health", "mental", "study", "personality", "appearance", "origin", "relocation", "social", "property", "adversity", "chart_structure"], Field(description="受控人生问题领域闭集")]], Field(min_length=1, max_length=16)] = Field(
        description="人生问题域（同 natal_query）",
    )
    chart: dict[str, Any] = Field(
        description="engine 排盘结果（同 compute_factors 输入），内部据此调 engine 流年/大限",
    )
    context: dict[str, Any] = Field(
        default=None,
        description="compute_factors 返回的原始 context；提供时必须与因子 provenance 校验一致",
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
