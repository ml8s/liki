"""Shared wire contracts for the bazi yongshen resolution.

The resolver owns the only yongshen decision. Downstream tools consume this
result verbatim, so the shape is published once here and reused by every MCP
contract that carries it. ``SelectedYongShen`` is the complete, auditable
decision; ``ConsumedYongShen`` is the narrower surface a consumer such as naming
needs, while still accepting the full object because extra fields are allowed.
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.contracts.bazi import Gan, Wuxing, Zhi
from app.contracts.schema import unique_list

YongWuxing = Literal["木", "火", "土", "金", "水", "无候选"]
SecondaryWuxing = Literal["木", "火", "土", "金", "水", "无辅调候"]
TenGod = Literal["比肩", "劫财", "食神", "伤官", "偏财", "正财", "七杀", "正官", "偏印", "正印"]
TenGodNoGod = Literal[
    "无格神", "比肩", "劫财", "食神", "伤官", "偏财", "正财", "七杀", "正官", "偏印", "正印",
]
YongShenStatus = Literal["resolved", "partial", "needs_review", "unresolved"]
YongShenConfidence = Literal["high", "medium", "low"]
KnownYongShenField = Literal["yong", "xi", "ji"]
Sha256Digest = Annotated[str, Field(pattern="^[a-f0-9]{64}$")]


class YongShenChoice(BaseModel):
    """The single selected yong/xi/ji result."""

    model_config = ConfigDict(extra="allow")
    yong: YongWuxing = Field(
        description="Selected yongshen element; 无候选 means the policy did not form a usable choice.",
    )
    xi: unique_list[Wuxing] = Field(
        description="Elements that support the selected yongshen.",
    )
    ji: unique_list[Wuxing] = Field(
        description="Elements that work against the selected yongshen.",
    )
    source: str = Field(
        description="Policy branch that produced this selection, such as fu_yi or tiao_hou.",
    )
    confidence: YongShenConfidence = Field(
        description="Confidence assigned by the yongshen policy.",
    )
    known_fields: unique_list[KnownYongShenField] = Field(
        description="Selected fields that the policy actually resolved.",
    )


class YongShenConflict(BaseModel):
    """A conflict found while resolving the yongshen."""

    model_config = ConfigDict(extra="allow")
    code: str = Field(description="Stable conflict code.")
    message: str = Field(description="Human-readable conflict explanation.")


class YongShenSource(BaseModel):
    """Provenance for a consumed yongshen decision."""

    model_config = ConfigDict(extra="allow")
    policy_id: str = Field(description="Yongshen policy identifier.")
    rule_id: str = Field(description="Matched rule identifier.")
    status: YongShenStatus = Field(
        description="Resolution status; only resolved and partial are consumable.",
    )
    chart_digest: Sha256Digest = Field(
        description="SHA-256 digest of the source bazi full chart.",
    )
    yongshen_digest: Sha256Digest = Field(
        description="SHA-256 digest of the published decision bound to the source chart and evidence.",
    )


class YongShenGeJuContext(BaseModel):
    """The 格局 context attached to the decision."""

    model_config = ConfigDict(extra="allow")
    ge_ju: str = Field(description="格局名称。")
    pattern_god_ten_god: TenGodNoGod = Field(description="格神十神；无格神为 无格神。")
    pattern_god_source: str = Field(description="格神取用来源。")
    yong_fa: str = Field(description="格局顺逆用法。")
    clear: bool = Field(description="格局是否清纯。")
    theme: str = Field(description="格局主题。")


class YongShenFactsEvidence(BaseModel):
    """Engine facts evidence record."""

    model_config = ConfigDict(extra="allow")
    source: Literal["facts"] = Field(description="证据来源。")
    month_zhi: Zhi = Field(description="月支。")
    day_master_stem: Gan = Field(description="日主天干。")
    day_master_element: Wuxing = Field(description="日主五行。")
    season_is_extreme: bool = Field(description="季节是否极端。")


class YongShenFuYiEvidence(BaseModel):
    """扶抑 evidence record."""

    model_config = ConfigDict(extra="allow")
    source: Literal["fu_yi"] = Field(description="证据来源。")
    qiangruo: str = Field(description="身强弱。")
    pattern: str = Field(description="扶抑结构模式。")
    root_type: str = Field(description="通根类型。")
    season: str = Field(description="季节状态。")
    yin_bi_count: int = Field(description="印比数量。")
    yong: YongWuxing = Field(description="扶抑候选用神；缺省为 无候选。")
    xi: YongWuxing = Field(description="扶抑候选喜神；缺省为 无候选。")
    ji: YongWuxing = Field(description="扶抑候选忌神；缺省为 无候选。")


class YongShenTiaoHouEvidence(BaseModel):
    """调候 evidence record."""

    model_config = ConfigDict(extra="allow")
    source: Literal["tiao_hou"] = Field(description="证据来源。")
    primary_wuxing: YongWuxing = Field(description="主调候五行。")
    secondary_wuxing: SecondaryWuxing = Field(description="辅调候五行；缺失时为 无辅调候。")
    secondary_condition: str = Field(description="辅调候条件。")
    condition_ok: bool = Field(description="辅调候条件是否满足。")
    secondary_usable: bool = Field(description="辅调候是否可用。")
    primary_present: bool = Field(description="主调候是否出现。")
    primary_visible: bool = Field(description="主调候是否可见。")
    secondary_present: bool = Field(description="辅调候是否出现。")
    secondary_visible: bool = Field(description="辅调候是否可见。")
    yong: YongWuxing = Field(description="调候用神。")
    yong_source: str = Field(description="调候用神来源。")
    yong_present: bool = Field(description="调候用神是否出现。")
    yong_visible: bool = Field(description="调候用神是否可见。")
    control_wuxing: YongWuxing = Field(description="调候控制五行。")


class YongShenGeJuEvidence(BaseModel):
    """格局 evidence record."""

    model_config = ConfigDict(extra="allow")
    source: Literal["ge_ju"] = Field(description="证据来源。")
    pattern: str = Field(description="格局名称。")
    pattern_god_ten_god: TenGodNoGod = Field(description="格神十神；无格神为 无格神。")
    pattern_god_source: str = Field(description="格神取用来源。")
    clear: bool = Field(description="格局是否清纯。")


class YongShenCongGeEvidence(BaseModel):
    """从格复核 evidence record."""

    model_config = ConfigDict(extra="allow")
    source: Literal["cong_ge"] = Field(description="证据来源。")
    pattern: str = Field(description="从格候选标签。")
    verified: bool = Field(description="从格双证是否通过。")
    fake: bool = Field(description="是否为假从格候选。")
    guan_sha_effective_count: int = Field(description="有效官杀数量：透干，或通根且力量强。")
    cai_effective_count: int = Field(description="有效财星数量：透干，或通根且力量强。")
    shi_shang_effective_count: int = Field(description="有效食伤数量：透干，或通根且力量强。")
    yin_rooted_count: int = Field(description="有根印星数量；任一有根印星即破从弱。")
    bi_jie_rooted_count: int = Field(description="有根比劫数量；任一有根比劫即破从弱。")


YongShenEvidence = (
    YongShenFactsEvidence
    | YongShenFuYiEvidence
    | YongShenTiaoHouEvidence
    | YongShenGeJuEvidence
    | YongShenCongGeEvidence
)


class SelectedYongShenCore(BaseModel):
    """The stable decision fields shared by every consumer."""

    model_config = ConfigDict(extra="allow")
    selected: YongShenChoice = Field(description="The only selected yong/xi/ji result.")
    status: YongShenStatus = Field(
        description="Resolution status; downstream tools consume resolved and partial only.",
    )
    conflicts: list[YongShenConflict] = Field(
        description="Conflicts found while resolving the yongshen; empty when none.",
    )
    rule_id: str = Field(description="Matched rule identifier.")
    policy_id: str = Field(description="Yongshen policy identifier.")
    chart_digest: Sha256Digest = Field(
        description="SHA-256 digest of the source bazi full chart.",
    )
    yongshen_digest: Sha256Digest = Field(
        description="SHA-256 digest of this resolved yongshen result.",
    )


class ConsumedYongShen(SelectedYongShenCore):
    """The subset a downstream consumer needs to accept and consume."""


class SelectedYongShen(SelectedYongShenCore):
    """The complete, auditable yongshen decision published over MCP."""

    ge_ju_context: YongShenGeJuContext = Field(
        description="Pattern context attached to the decision.",
    )
    evidence: list[YongShenEvidence] = Field(
        description="Rule evidence used by the decision, tagged by source.",
    )


__all__ = [
    "BaziYongshenOutput",
    "ConsumedYongShen",
    "SelectedYongShen",
    "SelectedYongShenCore",
    "YongShenChoice",
    "YongShenConflict",
    "YongShenCongGeEvidence",
    "YongShenEvidence",
    "YongShenSource",
]

#: The published bazi_yongshen output is the complete decision object.
BaziYongshenOutput = SelectedYongShen
