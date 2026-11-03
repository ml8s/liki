"""Shared wire contracts for the bazi yongshen resolution.

The resolver owns the only yongshen decision. Downstream tools consume this
result verbatim, so the shape is published once here and reused by every MCP
contract that carries it. ``SelectedYongShen`` is the complete, auditable
decision. Cross-domain consumers project only the field they need; for example,
naming uses ``selected_yongshen.yong`` as its target five element.
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator

from app.contracts.bazi import Gan, Wuxing, Zhi
from app.contracts.schema import unique_list

YongWuxing = Literal["木", "火", "土", "金", "水", "无候选"]
SecondaryWuxing = Literal["木", "火", "土", "金", "水", "无辅调候"]
TenGod = Literal["比肩", "劫财", "食神", "伤官", "偏财", "正财", "七杀", "正官", "偏印", "正印"]
TenGodNoGod = Literal[
    "无格神", "比肩", "劫财", "食神", "伤官", "偏财", "正财", "七杀", "正官", "偏印", "正印",
]
YongShenStatus = Literal["resolved", "partial", "needs_review", "unresolved"]
Sha256Digest = Annotated[str, Field(pattern="^[a-f0-9]{64}$")]


class YongShenChoice(BaseModel):
    """The selected yong/xi/ji five-element plan."""

    model_config = ConfigDict(extra="forbid")
    yong: YongWuxing = Field(
        description="用神五行；无候选表示未形成可用选择。",
    )
    xi: unique_list[Wuxing] = Field(
        description="Elements that support the selected yongshen.",
    )
    ji: unique_list[Wuxing] = Field(
        description="忌神五行。",
    )
    @model_validator(mode="after")
    def _enforce_disjoint_roles(self) -> "YongShenChoice":
        """Reject a plan that assigns the same element to competing roles."""
        assignments: dict[str, str] = {self.yong: "yong"}
        for role in ("xi", "ji"):
            for element in getattr(self, role):
                if element in assignments:
                    raise ValueError(
                        f"五行 {element} 同时属于 {assignments[element]} 和 {role} 角色"
                    )
                assignments[element] = role
        return self


class YongShenConflict(BaseModel):
    """A conflict found while resolving the yongshen."""

    model_config = ConfigDict(extra="forbid")
    code: str = Field(description="Stable conflict code.")
    message: str = Field(description="Human-readable conflict explanation.")


class YongShenGeJuContext(BaseModel):
    """The 格局 context attached to the decision."""

    model_config = ConfigDict(extra="forbid")
    pattern: str = Field(description="格局名称。")
    pattern_god_ten_god: TenGodNoGod = Field(description="格神十神；无格神为 无格神。")
    pattern_god_source: str = Field(description="格神取用来源。")
    yong_fa: str = Field(description="格局顺逆用法。")
    is_clear: bool = Field(description="格局是否清纯。")
    theme: str = Field(description="格局主题。")


class YongShenFactsEvidence(BaseModel):
    """Engine facts evidence record."""

    model_config = ConfigDict(extra="forbid")
    source: Literal["facts"] = Field(description="证据来源。")
    month_zhi: Zhi = Field(description="月支。")
    day_master_stem: Gan = Field(description="日主天干。")
    day_master_element: Wuxing = Field(description="日主五行。")
    season_is_extreme: bool = Field(description="季节是否极端。")


class YongShenFuYiEvidence(BaseModel):
    """扶抑 evidence record."""

    model_config = ConfigDict(extra="forbid")
    source: Literal["fu_yi"] = Field(description="证据来源。")
    qiang_ruo: str = Field(description="身强弱。")
    pattern: str = Field(description="扶抑结构模式。")
    ri_gan_gen: str = Field(description="通根类型。")
    season: str = Field(description="季节状态。")
    yin_bi_shu_liang: int = Field(description="印比数量。")
    yong: YongWuxing = Field(description="扶抑候选用神；缺省为 无候选。")
    xi: YongWuxing = Field(description="扶抑候选喜神；缺省为 无候选。")
    ji: YongWuxing = Field(description="扶抑候选忌神；缺省为 无候选。")


class YongShenTiaoHouEvidence(BaseModel):
    """调候 evidence record."""

    model_config = ConfigDict(extra="forbid")
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

    model_config = ConfigDict(extra="forbid")
    source: Literal["ge_ju"] = Field(description="证据来源。")
    pattern: str = Field(description="格局名称。")
    pattern_god_ten_god: TenGodNoGod = Field(description="格神十神；无格神为 无格神。")
    pattern_god_source: str = Field(description="格神取用来源。")
    is_clear: bool = Field(description="格局是否清纯。")


class YongShenCongGeEvidence(BaseModel):
    """从格复核 evidence record."""

    model_config = ConfigDict(extra="forbid")
    source: Literal["cong_ge"] = Field(description="证据来源。")
    pattern: str = Field(description="从格候选标签。")
    is_verified: bool = Field(description="从格双证是否通过。")
    is_fake_candidate: bool = Field(description="是否为假从格候选。")
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

    model_config = ConfigDict(extra="forbid")
    selected_yongshen: YongShenChoice = Field(description="The only selected yong/xi/ji result.")
    status: YongShenStatus = Field(
        description="裁决状态；下游仅消费 resolved 或 partial。",
    )
    conflicts: list[YongShenConflict] = Field(
        description="用神裁决冲突；无冲突为空。",
    )
    rule_id: str = Field(description="Matched rule identifier.")
    policy_id: str = Field(description="Yongshen policy identifier.")
    bazi_chart_digest: Sha256Digest = Field(
        description="SHA-256 digest of the source bazi full chart.",
    )
    yongshen_digest: Sha256Digest = Field(
        description="用神裁决摘要。",
    )


class SelectedYongShen(SelectedYongShenCore):
    """The complete, auditable yongshen decision published over MCP."""

    ge_ju_context: YongShenGeJuContext = Field(
        description="用神裁决关联的格局上下文。",
    )
    evidence: list[YongShenEvidence] = Field(
        description="按来源标记的裁决证据。",
    )


__all__ = [
    "BaziYongshenOutput",
    "SelectedYongShen",
    "SelectedYongShenCore",
    "YongShenChoice",
    "YongShenConflict",
    "YongShenCongGeEvidence",
    "YongShenEvidence",
]

#: The published bazi_yongshen output is the complete decision object.
BaziYongshenOutput = SelectedYongShen
