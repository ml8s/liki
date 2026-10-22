"""Typed wire model for the engine ``ziwei_chart`` output used as tool input.

Ported 1:1 from ``engine/internal/mcpcontract/ziwei_wire.go`` so the counsel
chain (compute_factors / period_query) consumes the engine's chart with the
same closed shapes the engine publishes — the bazi side is ``contracts/bazi.py``.
"""


from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.contracts.bazi import BaziChart, Gan, Gender, Zhi

Palace = Literal[
    "命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
    "迁移", "仆役", "官禄", "田宅", "福德", "父母",
]
SiHua = Literal["禄", "权", "科", "忌"]
JuShu = Literal["水二局", "木三局", "金四局", "土五局", "火六局"]
PalaceFactsKind = Literal["star", "brightness", "si_hua", "special"]


class ZiweiChartSchool(BaseModel):
    model_config = ConfigDict(extra="forbid")
    leap_month: Literal["first_fifteen_days_current_month_next_fifteen_days_following_month"] = Field(
        description="闰月拆分规则。",
    )
    source: Literal["iztro_compatible"] = Field(
        description="排盘来源。",
    )


class ZiweiXingYao(BaseModel):
    model_config = ConfigDict(extra="forbid")
    xing: Annotated[str, Field(min_length=1)] = Field(description="星名。")
    name: Annotated[str, Field(min_length=1)] = Field(description="全名。")
    is_zhu_xing: bool = Field(description="是否主星。")
    si_hua: SiHua | None = Field(default=None, description="四化(禄/权/科/忌)。")
    liang_du: str | None = Field(default=None, description="亮度(庙/旺/利/平/陷)。")


class ZiweiGongWeiItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    index: Palace = Field(description="宫名。")
    name: Palace = Field(description="宫名。")
    gan: Gan = Field(description="宫干。")
    zhi: Zhi = Field(description="宫支。")
    is_shen_gong: bool | None = Field(default=None, description="是否身宫。")
    is_yuan_gong: bool | None = Field(default=None, description="是否来因宫。")
    xing_yao: Annotated[list[ZiweiXingYao], Field(min_length=0)] = Field(
        description="星曜列表。",
    )
    zi_wei: str | None = Field(default=None, description="紫微星（本宫）。")


class ZiweiPalaceFactsItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    palace: Palace = Field(description="宫位。")
    kind: PalaceFactsKind = Field(description="类别。")
    target: Annotated[str, Field(min_length=1)] = Field(description="target。")
    star: str | None = Field(default=None, description="star。")
    value: str | None = Field(default=None, description="取值。")


class ZiweiKongGongItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    gong_name: Palace = Field(description="空宫宫名。")
    jie_xing: Annotated[list[str], Field(min_length=1)] = Field(
        description="借星列表。",
    )
    dui_gong: Palace = Field(description="对宫宫名。")


class ZiweiPatternItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: Annotated[str, Field(min_length=1)] = Field(description="格局名。")
    description: str = Field(description="格局结构注解（非吉凶结论）。")


class ZiweiSanFangItem(BaseModel):
    model_config = ConfigDict(extra="forbid")
    name: Palace = Field(description="三方四正宫名。")
    zhu_xing: Annotated[list[str], Field(min_length=0)] = Field(description="主星。")
    fu_xing: Annotated[list[str], Field(min_length=0)] = Field(description="辅星。")
    si_hua: Annotated[list[SiHua], Field(min_length=0)] = Field(
        description="本宫全部四化，按星曜顺序输出。",
    )


class ZiweiChart(BaseModel):
    """The engine ``ziwei_chart`` output accepted by ziwei-domain tools."""

    model_config = ConfigDict(extra="forbid")
    school: ZiweiChartSchool = Field(description="流派。")
    gong_wei: Annotated[list[ZiweiGongWeiItem], Field(min_length=12, max_length=12)] = Field(
        description="12 宫信息。",
    )
    palace_facts: Annotated[list[ZiweiPalaceFactsItem], Field(min_length=0)] = Field(
        description="宫位原子事实；Python 因子层只读。",
    )
    ming_gong: Palace | None = Field(default=None, description="命宫。")
    si_hua: dict[str, SiHua] | None = Field(default=None, description="四化: {星名: 化名}。")
    shen_gong: Palace | None = Field(default=None, description="身宫。")
    ju_shu: JuShu | None = Field(default=None, description="局数。")
    ju_shu_name: str | None = Field(default=None, description="ju_shu_name。")
    ming_zhu: str | None = Field(default=None, description="命主。")
    shen_zhu: str | None = Field(default=None, description="身主。")
    nian_gan: Gan | None = Field(default=None, description="年干。")
    nian_zhi: Zhi | None = Field(default=None, description="年支。")
    shi_zhi: Zhi | None = Field(default=None, description="时支。")
    birth_year: Annotated[int, Field(ge=1)] | None = Field(default=None, description="出生公历年。")
    ziwei_pos: Palace | None = Field(default=None, description="紫微落宫。")
    gender: Gender | None = Field(default=None, description="性别。")
    birth_lunar_month: int | None = Field(default=None, description="出生农历月。")
    birth_is_leap: bool | None = Field(default=None, description="出生农历月是否闰月。")
    lunar_month: int | None = Field(default=None, description="排盘用农历月（闰月后半月按下月）。")
    lunar_day: int | None = Field(default=None, description="排盘用农历日。")
    kong_gong: list[ZiweiKongGongItem] | None = Field(
        default=None, description="空宫借星（无主星宫位借对宫主星）。",
    )
    patterns: list[ZiweiPatternItem] | None = Field(
        default=None, description="命盘格局（如辅弼拱主/禄马交驰）。",
    )
    san_fang: list[ZiweiSanFangItem] | None = Field(
        default=None, description="三方四正（命/财帛/官禄/迁移）。",
    )


# The union every chart-consuming counsel tool accepts. Pydantic resolves the
# branch structurally: bazi charts carry 四柱 (nian/ri), ziwei charts carry
# school/gong_wei — no discriminator field exists on the wire.
EngineChart = BaziChart | ZiweiChart

__all__ = [
    "EngineChart",
    "Palace",
    "PalaceFactsKind",
    "SiHua",
    "JuShu",
    "ZiweiChart",
    "ZiweiChartSchool",
    "ZiweiGongWeiItem",
    "ZiweiKongGongItem",
    "ZiweiPalaceFactsItem",
    "ZiweiPatternItem",
    "ZiweiSanFangItem",
    "ZiweiXingYao",
]
