"""Typed wire model for the engine ``bazi_chart`` output used as tool input."""

from __future__ import annotations

from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

Gan = Literal["甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"]
Zhi = Literal["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"]
Wuxing = Literal["木", "火", "土", "金", "水"]
PillarName = Literal["nian", "yue", "ri", "shi"]
Gender = Literal["male", "female"]

DateString = Annotated[str, Field(pattern=r"^\d{4}-\d{2}-\d{2}$")]


class BaziPillar(BaseModel):
    """One 干支 pillar from the engine chart."""

    model_config = ConfigDict(extra="allow")
    gan: Gan = Field(description="天干。")
    zhi: Zhi = Field(description="地支。")
    na_yin: str = Field(description="干支纳音。")


class BaziDaYunRootRef(BaseModel):
    """One 通根 reference attached by the engine full chart."""

    model_config = ConfigDict(extra="allow")
    source: str = Field(description="通根来源。")
    pillar: PillarName | None = Field(
        default=None,
        description="相关柱；部分通根来源不绑定柱位。",
    )
    stem: Gan = Field(description="通根天干。")
    branch: Zhi = Field(description="通根地支。")


class BaziDaYunStep(BaseModel):
    """One 大运 step."""

    model_config = ConfigDict(extra="allow")
    gan: Gan = Field(description="大运天干。")
    zhi: Zhi = Field(description="大运地支。")
    start_date: DateString = Field(description="起始公历日 YYYY-MM-DD。")
    end_date: DateString = Field(description="结束公历日 YYYY-MM-DD。")
    start_year: int = Field(description="起始公历年。")
    end_year: int = Field(description="结束公历年。")
    name: str = Field(description="干支名称。")
    wuxing: Wuxing = Field(description="大运五行。")
    shi_shen: str = Field(description="相对日主的十神运。")
    rooted: bool | None = Field(
        default=None,
        description="大运干是否通根；engine fullchart 会补充。",
    )
    root_refs: list[BaziDaYunRootRef] | None = Field(
        default=None,
        description="通根依据；engine fullchart 会补充。",
    )


class BaziDaYun(BaseModel):
    """The 大运 sequence delivered with a bazi chart."""

    model_config = ConfigDict(extra="allow")
    start_date: DateString = Field(description="起运公历日 YYYY-MM-DD。")
    start_year_after: int = Field(description="出生后整年数。")
    start_month_after: int = Field(description="出生后余月数。")
    start_day_after: int = Field(description="出生后余日数。")
    direction: str = Field(description="顺排或逆排。")
    steps: list[BaziDaYunStep] = Field(description="大运步骤。")


class BaziChart(BaseModel):
    """The engine ``bazi_chart`` output accepted by bazi-domain tools."""

    model_config = ConfigDict(extra="allow")
    nian: BaziPillar = Field(description="年柱。")
    yue: BaziPillar = Field(description="月柱。")
    ri: BaziPillar = Field(description="日柱。")
    shi: BaziPillar = Field(description="时柱。")
    da_yun: BaziDaYun | None = Field(
        default=None,
        description="大运；极少数盘无大运时为 null。",
    )
    gender: Gender = Field(description="性别。")
    birth_year: int = Field(description="出生公历年。", ge=1)
    zi_shi_rule: str = Field(
        default="",
        description="子时换日规则说明。",
    )


__all__ = [
    "BaziChart",
    "BaziDaYun",
    "BaziDaYunRootRef",
    "BaziDaYunStep",
    "BaziPillar",
    "DateString",
    "Gan",
    "Gender",
    "PillarName",
    "Wuxing",
    "Zhi",
]
